// Copyright 2020 Hewlett Packard Enterprise Development LP

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/namsral/flag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	shcd_parser "stash.us.cray.com/HMS/hms-shcd-parser/pkg/shcd-parser"
	sls_common "stash.us.cray.com/HMS/hms-sls/pkg/sls-common"
)

var (
	hmnConnectionsFile = flag.String("hmn_connections_file", "",
		"Location for the JSON file containing the HMN connections.")
	slsInputStateFile = flag.String("sls_generator_input_state_file", "",
		"Location for the SLS Generator Input State JSON file.")
	outputFile = flag.String("sls_file_path", "sls_input_file.json",
		"Location to dump generated configuration.")

	atomicLevel zap.AtomicLevel
	logger      *zap.Logger
)

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

func main() {
	setupLogging()

	// Parse the arguments.
	flag.Parse()

	logger.Info("Beginning SLS configuration generation.")

	if *hmnConnectionsFile == "" {
		logger.Fatal("HMN connections file not specified!")
	}
	if *slsInputStateFile == "" {
		logger.Fatal("SLS Generator Input State file not specified!")
	}

	// Parse the input files
	slsInputState := parseSLSInputState()
	hmnRows := parseHMNConnectionsFile()

	// Generate SLS State
	g := NewSLSStateGenerator(logger, slsInputState, hmnRows)

	payload := g.GenerateSLSState()

	payloadJSON, _ := json.Marshal(payload)
	logger.Debug("Generated JSON.", zap.String("payloadJSON", string(payloadJSON)))

	// Write JSON to file if applicable.
	if *outputFile != "" {
		writeErr := ioutil.WriteFile(*outputFile, payloadJSON, os.ModePerm)

		if writeErr != nil {
			logger.Fatal("Failed to write JSON!", zap.Error(writeErr))
		} else {
			logger.Info("Wrote configuration to file.", zap.String("outputFile", *outputFile))
		}
	}

	logger.Info("Configuration generated.")
}

func parseHMNConnectionsFile() []shcd_parser.HMNRow {
	// Open and parse the file.
	jsonFile, err := os.Open(*hmnConnectionsFile)
	if err != nil {
		logger.Fatal("Unable to open HMN Connections file",
			zap.String("filename", *slsInputStateFile),
			zap.Error(err),
		)

	}

	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	jsonString := string(jsonBytes)

	_ = jsonFile.Close()

	if jsonString == "" {
		logger.Fatal("HMN connections file empty!")
	}

	rows := []shcd_parser.HMNRow{}
	err = json.Unmarshal(jsonBytes, &rows)
	if err != nil {
		logger.Panic("Failed to unmarshal HMN connections file!", zap.Error(err))
	}

	logger.Debug("Parsed HMN connections file.")

	return rows
}

func parseSLSInputState() sls_common.SLSGeneratorInputState {
	// Open and parse the file.
	jsonFile, err := os.Open(*slsInputStateFile)
	if err != nil {
		logger.Fatal("Unable to open SLS State Input file",
			zap.String("filename", *slsInputStateFile),
			zap.Error(err),
		)
	}

	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	jsonString := string(jsonBytes)

	_ = jsonFile.Close()

	if jsonString == "" {
		logger.Fatal("SLS Input State file empty!")
	}

	inputState := sls_common.SLSGeneratorInputState{}
	err = json.Unmarshal(jsonBytes, &inputState)
	if err != nil {
		logger.Panic("Failed to unmarshal SLS Input State file!", zap.Error(err))
	}

	logger.Debug("Parsed SLS Input State file.")

	return inputState
}
