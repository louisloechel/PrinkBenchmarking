package evaluation

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"prinkbenchmarking/src/types"
	"strings"
	"time"
)


func handleReadConnection(conn net.Conn, config types.Config, experiment *types.Experiment) error {
	// close connection when done

	writer, file := initialiseResults(config.OutputFolder, experiment)
	defer file.Close()
	defer writer.Flush()
	log.Printf("Reading from connection")

	reader := bufio.NewScanner(conn)

	// Read the data
	for reader.Scan() {
		response := reader.Text()

		output := fmt.Sprintf("%v; %s\n", time.Now(), response)
		_, err := writer.Write([]byte(output))
		if err != nil {
			return fmt.Errorf("could not write to buffer: %v", err)
		}

	}

	// flush the buffer
	writer.Flush()

	return nil
}



func readSocketConnection(e *types.Experiment, config types.Config) error {
	// Open socket connection
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "0.0.0.0", e.SutPortRead))
	if err != nil {
		return fmt.Errorf("could not open read socket connection: %v", err)
	}
	defer ln.Close()

	// Accept connection
	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("could not accept read connection: %v", err)
	}
	defer conn.Close()

	// Handle connection
	return handleReadConnection(conn, config, e)
}



func initialiseResults(path string, experiment *types.Experiment) (*bufio.Writer, *os.File) {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatalf("Could not create output directory: %v", err)
	}

	// open file for appending, create if it doesn't exist
	file, err := os.OpenFile(path+"/results." + time.Now().Format("2006-01-02_15:04:05") + "." + experiment.ToFileName() + ".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open results file: %v", err)
	}

	// Check if the file is empty
	info, err := file.Stat()
	if err != nil {
		log.Fatalf("Could not get file info: %v", err)
	}

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header if the file is empty
	if info.Size() == 0 {
				
		_, err = writer.WriteString(strings.Join([]string{"t_e", "m_id", "t_s", "building_id", "timestamp", "meter_reading", "primary_use", "square_feet", "year_built", "floor_count", "air_temperature", "cloud_coverage", "dew_temperature", "precip_depth_1_hr", "sea_level_pressure", "wind_direction", "wind_speed"}, ";"))
		writer.Write([]byte("\n"))
		if err != nil {
			log.Fatalf("Could not write to results.csv: %v", err)
		}
	}

	if err != nil {
		log.Fatalf("Could not write to results.csv: %v", err)
	}

	return writer, file
}