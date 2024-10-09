package evaluation

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)



func benchmark(dataset [][]string, conn net.Conn) error {
	// benchmark the SUT
	// Iterate over the records for duration seconds and write them to console (for now)
	count := 0
	start := time.Now()
	for i, record := range dataset {
		if i == 0 {
			continue
		}

		// Data fields:
		// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed, building_id2, unix_timestamp,
		// 
		// Benchmark fields (append to the end): 
		// m_id, ts

		ts := time.Now()
	
		message := strings.Join(record, ";") + fmt.Sprintf(";%d;%v\n", count, ts)

		// Write the message to Flink socket
		_, err := conn.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("could not write to Flink: %v", err)
		}

		count++
	}

	passedTime := time.Since(start)
	log.Printf("Wrote %d records to Flink in %s.", count, passedTime.String())

	return nil
}


