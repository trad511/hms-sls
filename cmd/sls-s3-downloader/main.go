// Copyright 2020 Hewlett Packard Enterprise Development LP

package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/namsral/flag"
	"go.uber.org/zap"
	hms_s3 "stash.us.cray.com/HMS/hms-s3"
)

const SLS_FILE = "sls_input_file.json"

var (
	outputDir = flag.String("output_dir", "/", "Destination directory to put files.")

	logger   *zap.Logger
	s3Client *hms_s3.S3Client
)

func writeOutputFile(file string) {
	var err error
	var objectOutput *s3.GetObjectOutput

	// We really need these files, try forever to get them.
	for true {
		objectOutput, err = s3Client.GetObject(file)
		if err != nil {
			logger.Error("Failed to get file from S3!",
				zap.String("file", file),
				zap.Error(err),
			)

			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}

	fullPath := path.Join(*outputDir, file)
	outFile, err := os.Create(fullPath)
	if err != nil {
		logger.Fatal("Failed to create output file!", zap.Error(err))
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, objectOutput.Body)
	if err != nil {
		logger.Fatal("Failed to write to output file!", zap.Error(err))
	}

	logger.Info("Downloaded file.", zap.String("fullPath", fullPath))
}

func main() {
	// Parse the arguments.
	flag.Parse()

	// Setup logging.
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Can't initialize Zap logger: %v", err))
	}
	defer logger.Sync()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	logger.Info("Beginning downloading SLS input files S3.")

	s3Connection, err := hms_s3.LoadConnectionInfoFromEnvVars()
	if err != nil {
		logger.Fatal("Failed to load connection info for S3!", zap.Error(err))
	}

	s3Client, err = hms_s3.NewS3Client(s3Connection, httpClient)
	if err != nil {
		// An error here is uncorrectable.
		logger.Fatal("Failed to setup new S3 client!", zap.Error(err))
	}

	for true {
		err = s3Client.PingBucket()
		if err != nil {
			logger.Warn("Failed to ping bucket.", zap.Error(err))
			time.Sleep(time.Second)
		} else {
			logger.Info("Connected to S3 bucket.", zap.String("bucket", s3Client.ConnInfo.Bucket))
			break
		}
	}

	// Now pull down the one file we need.
	writeOutputFile(SLS_FILE)
}
