package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-yaml/yaml"
)

type Metric struct {
	Duration time.Duration
}

type Config struct {
	Address      string `yaml:"sut_address"`
	PortWrite    int    `yaml:"sut_port_write"`
	PortRead     int    `yaml:"sut_port_read"`
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
		err = writer.Write([]string{"t_e", "t_s", "m_id", "building_id", "timestamp", "meter_reading", "primary_use", "square_feet", "year_built", "floor_count", "air_temperature", "cloud_coverage", "dew_temperature", "precip_depth_1_hr", "sea_level_pressure", "wind_direction", "wind_speed"})
		if err != nil {
			log.Fatalf("Could not write to results.csv: %v", err)
		}
	}

	// Write data
	err = writer.Write([]string{
		fmt.Sprintf("%v", time.Now()),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
		fmt.Sprintf("%d", 0),
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

func warmUp(dataset [][]string, duration int, conn net.Conn) {
	// warm up the SUT
	// Iterate over the records for duration seconds and write them to console (for now)
	count := 0
	start := time.Now()
	for _, record := range dataset {
		t_s := time.Now()

		// Message fields:
		// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed

		message := fmt.Sprintf("t_s: %v, message_id: %d, building_id: %s, timestamp: %s, meter_reading: %s, primary_use: %s, square_feet: %s, year_built: %s, floor_count: %s, air_temperature: %s, cloud_coverage: %s, dew_temperature: %s, precip_depth_1_hr: %s, sea_level_pressure: %s, wind_direction: %s, wind_speed: %s",
			t_s, count, record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12], record[13])

		// Write the message to Flink socket
		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Fatalf("Could not write to Flink: %v", err)
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

func benchmark(dataset [][]string, duration int, conn net.Conn) {
	// benchmark the SUT
	// Iterate over the records for duration seconds and write them to console (for now)
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

		// Write the message to Flink socket
		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Fatalf("Could not write to Flink: %v", err)
		}

		// Check if the duration has passed
		if time.Since(start) > time.Duration(duration)*time.Second {
			return
		}

		count++
	}

	passedTime := time.Since(start)
	log.Printf("Wrote %d records to Flink in %s before Benchmark time of %ds ended.", count, passedTime.String(), duration)
}

func experimentDone() {
	// Create experiment_done.txt (to be used later in tf script for stopping the cluster)
	file, err := os.Create("../results/experiment_done.txt")
	if err != nil {
		log.Fatalf("Could not create experiment_done.txt: %v", err)
	}
	defer file.Close()
}

func socketConnection(u url.URL, config Config, dataset [][]string) {
	// Open socket connection
	ln, err := net.Listen("tcp", u.Host)
	if err != nil {
		log.Fatalf("Could not open socket connection: %v", err)
	}
	defer ln.Close()

	// Accept connection
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("Could not accept connection: %v", err)
		}

		// Handle connection
		go handleConnection(conn, config, dataset)
	}
}

func handleConnection(conn net.Conn, config Config, dataset [][]string) {
	// close connection when done
	defer conn.Close()

	log.Printf("Warming up the SUT for %d Seconds", config.Warmup)
	warmUp(dataset, config.Warmup, conn)
	log.Printf("Warmup done.")

	log.Printf("Starting the benchmark for %d Seconds", config.Duration)
	benchmark(dataset, config.Duration, conn)
	log.Printf("Benchmark done.")

	// close the connection
	conn.Close()
}

func readSocketConnection(v url.URL, config Config) {
	// Open socket connection
	ln, err := net.Listen("tcp", v.Host)
	if err != nil {
		log.Fatalf("Could not open read socket connection: %v", err)
	}
	defer ln.Close()

	// Accept connection
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("Could not accept read connection: %v", err)
		}

		// Handle connection
		go handleReadConnection(conn, config)
	}
}

func handleReadConnection(conn net.Conn, config Config) {
	// close connection when done
	defer conn.Close()

	// Write the data to results.csv
	file, err := os.OpenFile(config.OutputFolder+"/results.csv", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open results.csv: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	log.Printf("Reading from connection")

	// Read the data
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Fatalf("Could not read from connection: %v", err)
		}

		// Print the data
		message := string(buffer[:n])
		log.Printf("Read from connection: %s", message)

		err = writer.Write([]string{
			fmt.Sprintf("%v", time.Now()),
			message,
		})
		if err != nil {
			log.Fatalf("Could not write to results.csv: %v", err)
		}

		writer.Flush()
	}
}

func main() {
	config := loadConfig("config.yml")

	// parse url (write)
	u, err := url.Parse("https://" + config.Address + ":" + fmt.Sprintf("%d", config.PortWrite))
	if err != nil {
		log.Fatalf("Could not parse write URL: %v", err)
	}

	// parse url (read)
	v, err := url.Parse("https://" + config.Address + ":" + fmt.Sprintf("%d", config.PortRead))
	if err != nil {
		log.Fatalf("Could not parse read URL: %v", err)
	}

	initialiseResults(config.OutputFolder)

	dataset := loadDataset(config.InputData)

	var wg sync.WaitGroup

	// Increment the WaitGroup counter
	wg.Add(4)

	// write socket connection
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		socketConnection(*u, config, dataset)
	}()

	// read socket connection
	go func() {
		defer wg.Done() // Decrement the counter when the goroutine completes
		readSocketConnection(*v, config)
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	experimentDone()
	log.Printf("Created output file in: %s", config.OutputFolder)
}
