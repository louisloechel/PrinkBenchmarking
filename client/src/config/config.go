package config

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"prinkbenchmarking/src/types"

	"gopkg.in/yaml.v2"
)


func LoadConfig() *types.Config {
	env_config := os.Getenv("CLIENT_CONFIG")
	var (
		config *types.Config
		err    error
	)
	if env_config != "" {
		reader := strings.NewReader(env_config)
		config, err = ReadConfig(reader)
		log.Print("Using environment variable for configuration")
	} else {
		// Load the .yaml file
		configPath := dir("config.yaml")
		config, err = ReadConfigFromFile(configPath)
		log.Print("Using config.yaml for configuration")
	}

	if err != nil {
		log.Fatal(err)
	}

	return config
}

// https://github.com/joho/godotenv/issues/126#issuecomment-1474645022
// dir returns the absolute path of the given environment file (envFile) in the Go module's
// root directory. It searches for the 'go.mod' file from the current working directory upwards
// and appends the envFile to the directory containing 'go.mod'.
// It panics if it fails to find the 'go.mod' file.
func dir(envFile string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found"))
		}
		currentDir = parent
	}

	return filepath.Join(currentDir, envFile)
}

// NewConfig returns a new decoded Config struct
func ReadConfigFromFile(configPath string) (*types.Config, error) {
	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ReadConfig(file)
}

func ReadConfig(reader io.Reader) (*types.Config, error) {
	d := yaml.NewDecoder(reader)
	var config types.Config

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
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

