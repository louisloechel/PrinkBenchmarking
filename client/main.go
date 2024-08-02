package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
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
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatalf("Could not create output directory: %v", err)
	}

	// open file for appending, create if it doesn't exist
	file, err := os.OpenFile(path+"/results.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

func warmUp(dataset [][]string, duration int, u url.URL) {
	// warm up the SUT
	// Iterate over the records for duration seconds and write them to Kafka
	count := 0
	start := time.Now()
	for _, record := range dataset {
		t_s := time.Now()

		// Message fields:
		// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed

		message := fmt.Sprintf("t_s: %v, message_id: %d, building_id: %s, timestamp: %s, meter_reading: %s, primary_use: %s, square_feet: %s, year_built: %s, floor_count: %s, air_temperature: %s, cloud_coverage: %s, dew_temperature: %s, precip_depth_1_hr: %s, sea_level_pressure: %s, wind_direction: %s, wind_speed: %s",
			t_s, count, record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12], record[13])

		// TODO: Write the message to Flink instead of just printing it (rm if condition)
		if count%1000000 == 0 {
			log.Printf("Wrote record: %s to %s", message, u.String())
		}
		// Check if the duration has passed
		if time.Since(start) > time.Duration(duration)*time.Second {
			break
		}

		count++
	}
	passedTime := time.Since(start)
	log.Printf("Wrote all %d records to Flink in %s before warmup time of %ds ended.", len(dataset), passedTime.String(), duration)
}

func benchmark(dataset [][]string, duration int, interval int, u url.URL) {
	// benchmark the SUT
	// Iterate over the records for duration seconds and write them to Kafka
	count := 0
	start := time.Now()
	for _, record := range dataset {
		t_s := time.Now()

		// Bench fields:
		// t_s, message_id

		// Data fields:
		// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed

		message := fmt.Sprintf("t_s: %v, message_id: %d, building_id: %s, timestamp: %s, meter_reading: %s, primary_use: %s, square_feet: %s, year_built: %s, floor_count: %s, air_temperature: %s, cloud_coverage: %s, dew_temperature: %s, precip_depth_1_hr: %s, sea_level_pressure: %s, wind_direction: %s, wind_speed: %s",
			t_s, count, record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12], record[13])

		// TODO: Write the message to Flink instead of just printing it (rm if condition)
		if count%10 == 0 {
			log.Printf("Wrote record: %s to %s", message, u.String())
		}

		// Check if the duration has passed
		if time.Since(start) > time.Duration(duration)*time.Second {
			return
		}

		// Sleep for the interval if count%82==0
		if count%82 == 0 {
			time.Sleep(time.Duration(interval) * time.Second)
			log.Printf("Slept for %d seconds", interval)
		}

		count++
	}

	passedTime := time.Since(start)
	log.Printf("Wrote %d records to Flink in %s before Benchmark time of %ds ended.", count, passedTime.String(), duration)
}

func experimentDone() {
	// Create experiment_done.txt
	file, err := os.Create("../results/experiment_done.txt")
	if err != nil {
		log.Fatalf("Could not create experiment_done.txt: %v", err)
	}
	defer file.Close()
}

func main() {
	config := loadConfig("config.yml")

	// parse url
	u, err := url.Parse("https://" + config.Address + ":" + fmt.Sprintf("%d", config.Port))
	if err != nil {
		log.Fatalf("Could not parse URL: %v", err)
	}

	// TODO: set up connection to Flink

	initialiseResults(config.OutputFolder)

	dataset := loadDataset(config.InputData)

	log.Printf("Warming up the SUT for %d Seconds", config.Warmup)
	warmUp(dataset, config.Warmup, *u)
	log.Printf("Warmup done.")

	log.Printf("Starting the benchmark for %d Seconds", config.Duration)
	benchmark(dataset, config.Duration, config.Interval, *u)
	log.Printf("Benchmark done.")

	// TODO: Close connection

	experimentDone()
	log.Printf("Created output file in: %s", config.OutputFolder)
}
