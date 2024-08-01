package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-yaml/yaml"
)

type Metric struct {
	Duration time.Duration
}

type Config struct {
	Address      string `yaml:"sut_address"`
	Port         int    `yaml:"sut_port"`
	Concurrency  int    `yaml:"concurrency"`
	Mode         string `yaml:"mode"`
	Interval     int    `yaml:"interval"`
	Duration     int    `yaml:"duration"`
	Warmup       int    `yaml:"warmup"`
	OutputFolder string `yaml:"output_folder"`
	InputData    string `yaml:"input_data"`
}

func loadConfig(path string) Config {

	configData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	log.Printf("Loaded config file: %v", path)
	log.Printf("Config data: %v", string(configData))

	// Parse the YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	// print config
	log.Printf("Config: %v", config)

	return config
}

func initialiseResults(path string) {
	// open file for appending, create if it doesn't exist
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open results file: %v", err)
	}
	defer file.Close()

	// Check if the file is empty
	info, err := file.Stat()
	if err != nil {
		log.Fatalf("Could not get file info: %v", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if the file is empty
	if info.Size() == 0 {
		err = writer.Write([]string{"Timestamp", "Stay Time", "Stay Count", "Throughput"})
		if err != nil {
			log.Fatalf("Could not write to results.csv: %v", err)
		}
	}

	// Write data
	err = writer.Write([]string{
		fmt.Sprintf("%v", time.Now()),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%v", 0),
	})
	if err != nil {
		log.Fatalf("Could not write to results.csv: %v", err)
	}

}

func loadDataset(path string) [][]string {
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

func warmUp() {

}

func benchmark() {

}

func experimentDone() {

}

func main() {
	config := loadConfig("/client/config.yml")

	// endpoint := fmt.Sprintf("%s:%d", config.Address, config.Port)

	resultFile := config.OutputFolder + "/results.csv"
	initialiseResults(resultFile)

	loadDataset()

	log.Printf("Warming up the SUT for %d Seconds", config.Warmup)
	warmUp()
	log.Printf("Warmup done.")

	log.Printf("Starting the benchmark for %d Seconds", config.Duration)
	benchmark()

	experimentDone()
	log.Printf("Experiment done. Created output file: %s", resultFile)
}
