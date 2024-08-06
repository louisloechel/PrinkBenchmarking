package main

import (
	"bufio"
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
	writer.Comma = ';' // use semicolon as delimiter
	defer writer.Flush()

	// Write header if the file is empty
	if info.Size() == 0 {
		err = writer.Write([]string{"t_e", "m_id", "t_s", "building_id", "timestamp", "meter_reading", "primary_use", "square_feet", "year_built", "floor_count", "air_temperature", "cloud_coverage", "dew_temperature", "precip_depth_1_hr", "sea_level_pressure", "wind_direction", "wind_speed"})
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

func benchmark(dataset [][]string, conn net.Conn) {
	// benchmark the SUT
	// Iterate over the records for duration seconds and write them to console (for now)
	count := 0
	start := time.Now()
	for i, record := range dataset {
		if i == 0 {
			continue
		}
		t_s := time.Now()

		// Bench fields:
		// t_s, message_id

		// Data fields:
		// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed

		message := fmt.Sprintf("%d;%v;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s;%s\n",
			count, t_s, record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10], record[11], record[12], record[13])

		// Write the message to Flink socket
		_, err := conn.Write([]byte(message))
		if err != nil {
			log.Fatalf("Could not write to Flink: %v", err)
		}

		count++
	}

	passedTime := time.Since(start)
	log.Printf("Wrote %d records to Flink in %s.", count, passedTime.String())
}

func experimentDone() {
	// Create experiment_done.txt (to be used later in tf script for stopping the cluster)
	file, err := os.Create("../results/experiment_done.txt")
	if err != nil {
		log.Fatalf("Could not create experiment_done.txt: %v", err)
	}
	defer file.Close()
}

func socketConnection(u url.URL, dataset [][]string) {
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
		go handleConnection(conn, dataset)
	}
}

func handleConnection(conn net.Conn, dataset [][]string) {
	// close connection when done
	defer conn.Close()

	log.Printf("Starting the benchmark.")
	benchmark(dataset, conn)
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

	writer := bufio.NewWriter(file)
	defer writer.Flush()
	log.Printf("Reading from connection")

	reader := bufio.NewScanner(conn)

	// Read the data
	for reader.Scan() {
		response := reader.Text()

		output := fmt.Sprintf("%v; %s\n", time.Now(), response)
		log.Printf("Writing: %s", output)
		_, err = writer.Write([]byte(output))
		if err != nil {
			log.Fatalf("Could not write to buffer: %v", err)
		}

	}

	// flush the buffer
	writer.Flush()
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
		socketConnection(*u, dataset)
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
