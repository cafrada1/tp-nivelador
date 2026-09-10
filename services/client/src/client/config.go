package client

import (
	"fmt"
	"os"
	"strconv"
)

const (
	agencyIdKey   = "AGENCY_ID"
	inputFileKey  = "INPUT_FILE"
	outputFileKey = "OUTPUT_FILE"
	serverHostKey = "SERVER_HOST"
	serverPortKey = "SERVER_PORT"
	batchSizeKey  = "BATCH_SIZE"
)

type Config struct {
	ServerHost string
	ServerPort string
	AgencyId   int
	InputFile  string
	OutputFile string
	BatchSize  int
}

func getEnv(key string) (string, error) {
	value := os.Getenv(key)
	if len(value) == 0 {
		return "", fmt.Errorf("%s environment variable is required", key)
	}
	return value, nil
}

func toInt(key, value string) (int, error) {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s environment variable must be a number and got %s", key, value)
	}
	return intValue, nil
}

func LoadConfig() (Config, error) {
	configsKeys := []string{
		agencyIdKey,
		inputFileKey,
		outputFileKey,
		serverHostKey,
		serverPortKey,
		batchSizeKey,
	}

	configs := make(map[string]string)

	for _, key := range configsKeys {
		value, err := getEnv(key)
		if err != nil {
			return Config{}, err
		}
		configs[key] = value
	}

	agencyId, err := toInt(agencyIdKey, configs[agencyIdKey])
	if err != nil {
		return Config{}, err
	}

	batchSize, err := toInt(batchSizeKey, configs[batchSizeKey])
	if err != nil {
		return Config{}, err
	}

	if batchSize <= 0 {
		return Config{}, fmt.Errorf("batch size must be greater than zero and got %d", batchSize)
	}

	return Config{
		ServerHost: configs[serverHostKey],
		ServerPort: configs[serverPortKey],
		AgencyId:   agencyId,
		InputFile:  configs[inputFileKey],
		OutputFile: configs[outputFileKey],
		BatchSize:  batchSize,
	}, nil
}
