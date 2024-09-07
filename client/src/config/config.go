package config

import (
	"encoding/csv"
	"log"
	"os"

	"prinkbenchmarking/src/types"

	"gopkg.in/yaml.v2"
)

func LoadConfig(path string) types.Config {

	configData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	log.Printf("Loaded config file: %v", path)
	log.Printf("Config data: %v", string(configData))

	// Parse the YAML data into the Config struct
	var config types.Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	// print config
	log.Printf("Config: %v", config)

	return config
}

func LoadDataset(path string) [][]string {
	// load the dataset
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open dataset file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	dataset, err := reader.ReadAll()

	if err != nil {
		log.Fatalf("Could not read dataset file: %v", err)
	}

	return dataset
}

