/*
 * MIT License
 *
 * (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/namsral/flag"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	hms_s3 "stash.us.cray.com/HMS/hms-s3"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

var (
	// CLI/ENV Flags
	slsFilePath = flag.String("sls_file_path", "sls_input_file.json", "Path to the SLS file")
	slsURL      = flag.String("sls_url", "http://cray-sls", "URL to SLS instance")

	uploadCheckS3Marker = flag.Bool("sls_loader_check_s3_marker", true,
		"Check S3 bucket to see if the SLS file has been uploaded before")
	uploadCheckSLSContents = flag.Bool("sls_loader_check_sls_contents", false,
		"Check if SLS is if it is empty before uploading")
	forceUpload = flag.Bool("sls_loader_force_upload", false, "Force upload of SLS file")

	// Globals
	logger      *zap.Logger
	atomicLevel zap.AtomicLevel
)

func main() {
	// Setup Context
	ctx := setupContext()

	// Setup logging.
	setupLogging()

	// Configuration from environment flags
	flag.Parse()
	logger.Info("Loader Configuration",
		zap.Stringp("sls_file_path", slsFilePath),
		zap.Stringp("sls_url", slsURL),
		zap.Boolp("sls_loader_check_s3_marker", uploadCheckS3Marker),
		zap.Boolp("sls_loader_check_sls_contents", uploadCheckSLSContents),
		zap.Boolp("sls_loader_force_upload", forceUpload),
	)

	// Connection to S3
	s3Connection, err := hms_s3.LoadConnectionInfoFromEnvVars()
	if err != nil {
		logger.Fatal("Failed to load connection info for S3!", zap.Error(err))
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	s3Client, err := hms_s3.NewS3Client(s3Connection, httpClient)
	if err != nil {
		// An error here is uncorrectable.
		logger.Fatal("Failed to setup new S3 client!", zap.Error(err))
	}

	if *uploadCheckS3Marker {
		// Check S3 to see if the SLS file has been uploaded before
		uploaded, err := hasS3FileBeenUploadedBefore(s3Client)
		if err != nil {
			logger.Fatal("Failed to check S3", zap.Error(err))
		}
		logger.Info("Has the SLS file been uploaded before", zap.Bool("uploaded", uploaded))

		if !*forceUpload && uploaded {
			logger.Warn("The SLS file has been uploaded before")
			return
		}
	}

	if *uploadCheckSLSContents {
		// Only Upload the SLS file if SLS is empty
		empty, err := isSLSEmpty(ctx, *slsURL)
		if err != nil {
			logger.Error("Failed to check wether SLS is empty", zap.Error(err))
		}

		logger.Info("Is SLS Empty", zap.Bool("empty", empty))

		if !*forceUpload && !empty {
			logger.Warn("SLS is not empty. Not uploading SLS file")
			return
		}
	}

	// Proceed to upload the SLS file, as SLS is empty
	err = uploadFileToSLS(ctx, *slsFilePath, *slsURL)
	if err != nil {
		logger.Fatal("Failed to upload file to SLS", zap.Error(err))
	}

	// Create 'uploaded' file in S3 to mark that the SLS file has been uploaded
	s3Client.PutObject("uploaded", []byte("SLS File has been uploaded"))
}

func setupContext() context.Context {
	var cancel context.CancelFunc
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c

		// Cancel the context to cancel any in progress HTTP requests.
		cancel()
	}()

	return ctx
}

func setupLogging() {
	logLevel := os.Getenv("LOG_LEVEL")
	logLevel = strings.ToUpper(logLevel)

	atomicLevel = zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()
	logger = zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atomicLevel,
	))

	switch logLevel {
	case "DEBUG":
		atomicLevel.SetLevel(zap.DebugLevel)
	case "INFO":
		atomicLevel.SetLevel(zap.InfoLevel)
	case "WARN":
		atomicLevel.SetLevel(zap.WarnLevel)
	case "ERROR":
		atomicLevel.SetLevel(zap.ErrorLevel)
	case "FATAL":
		atomicLevel.SetLevel(zap.FatalLevel)
	case "PANIC":
		atomicLevel.SetLevel(zap.PanicLevel)
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
	}
}

func hasS3FileBeenUploadedBefore(s3client *hms_s3.S3Client) (bool, error) {
	_, err := s3client.GetObject("uploaded")

	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == s3.ErrCodeNoSuchKey {
			// The marker file is not present, so the SLS file has not been uploaded
			// This is the only case where we should return false
			return false, nil
		}
	}

	// If we get here it could mean one of the following
	// - If err is nil, then the file is truly in S3
	// - If err is not nil, then we are assuming the worst that the file is present in S3,
	//   and returning the error that we encountered.
	return true, err
}

func isSLSEmpty(ctx context.Context, slsURL string) (bool, error) {
	logger.Info("Checking wether SLS is empty")

	httpClient := retryablehttp.NewClient()
	httpClient.RetryMax = 100

	// Dump the contents of SLS
	req, err := retryablehttp.NewRequest("GET", slsURL+"/v1/dumpstate", nil)
	if err != nil {
		return false, err
	}

	req = req.WithContext(ctx)

	resp, doErr := httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if doErr != nil {
		return false, doErr
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Dumpstate returned unexpected status code: %d", resp.StatusCode)
		return false, err
	}

	slsStateRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	slsState := sls_common.SLSState{}
	err = json.Unmarshal(slsStateRaw, &slsState)
	if err != nil {
		return false, err
	}

	// Check SLS contents
	hardwareEmpty := len(slsState.Hardware) == 0
	networksEmpty := len(slsState.Networks) == 0

	logger.Debug("Is SLS empty",
		zap.Bool("hardware_empty", hardwareEmpty), zap.Bool("networks_empty", networksEmpty))

	return hardwareEmpty && networksEmpty, nil
}

func uploadFileToSLS(ctx context.Context, slsFilePath, slsURL string) error {
	fmt.Printf("Uploading SLS file (%s) to SLS (%s)...\n", slsFilePath, slsURL)

	// Open and parse the file.
	jsonFile, err := os.Open(slsFilePath)
	if err != nil {
		return err
	}

	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	jsonString := string(jsonBytes)

	_ = jsonFile.Close()

	if jsonString == "" {
		return fmt.Errorf("SLS file is empty")
	}

	fmt.Printf("SLS file contents:\n%s\n", string(jsonString))

	// Create a buffer with the file contents.

	slsDumpReader := strings.NewReader(jsonString)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fw, err := writer.CreateFormFile("sls_dump", "sls-input-file.json")
	if err != nil {
		logger.Error("Failed to create form file for dump", zap.Error(err))
		return err
	}
	_, err = io.Copy(fw, slsDumpReader)
	if err != nil {
		logger.Error("Failed to copy form file for dump", zap.Error(err))
		return err
	}

	writer.Close()

	// Setup a client and request.
	httpClient := retryablehttp.NewClient()
	httpClient.RetryMax = 100

	req, err := retryablehttp.NewRequest("POST", slsURL+"/v1/loadstate", &buf)
	if err != nil {
		logger.Error("Failed to build HTTP request for loadstate", zap.Error(err))
		return err
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("HMS-Service", "sls-loader")

	// Send it.
	resp, doErr := httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if doErr != nil {
		logger.Error("Failed to pop")
		panic(doErr)
	}

	if resp.StatusCode != http.StatusNoContent {
		logger.Fatal("Upload returned unexpected status code", zap.Int("status_code", resp.StatusCode))
	}

	logger.Info("SLS file successfully uploaded!")

	return nil
}
