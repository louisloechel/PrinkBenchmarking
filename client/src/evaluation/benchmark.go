package evaluation

import (
	"fmt"
	"log"
	"net"
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
			return fmt.Errorf("could not write to Flink: %v", err)
		}

		count++
	}

	passedTime := time.Since(start)
	log.Printf("Wrote %d records to Flink in %s.", count, passedTime.String())

	return nil
}


