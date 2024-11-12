package evaluation

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RawGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raw_gauge",
			Help: "Values of the raw data",
		},
		[]string{"building_id", "primary_use"},
	)
)

var (
	PrinkGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prink_gauge",
			Help: "Values of the prink data",
		},
		[]string{"building_id", "primary_use"},
	)
)

func StartPrometheusExporter(addr string) {

	// Expose /metrics HTTP endpoint using the created custom registry.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}

func RegisterMetrics() {
	prometheus.MustRegister(RawGauge)
	prometheus.MustRegister(PrinkGauge)
}

func ExportRecordAsPrometheusGaugeRaw(record []string) {
	// Export record as prometheus Gauge
	buildingID := record[0]
	timestamp := record[1]
	primaryUse := record[3]

	// Data fields:
	// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed, building_id2, unix_timestamp,
	meter_reading, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		log.Printf("Error converting meter reading to float: %v", err)
		return
	}

	// meter_reading := meter_reading_sum / float64(len(meter_readings))
	ts, err := time.Parse(time.DateTime, timestamp)
	if err != nil {
		log.Printf("Error converting timestamp to time: %v in ExportRecordAsPrometheusGaugeRaw", err)
		return
	}

	prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(RawGauge.WithLabelValues(buildingID, primaryUse).Desc(), prometheus.GaugeValue, meter_reading, buildingID, primaryUse))
	// RawGauge.WithLabelValues(buildingID, primaryUse, timestamp).Set(meter_reading)
}

func ExportRecordAsPrometheusGaugePrink(record []string) {
	// Export record as prometheus Gauge
	buildingID := record[0]
	timestamp := record[1]
	primaryUse := record[3]

	// Data fields:
	// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed, building_id2, unix_timestamp,
	meter_reading_tuple := record[2]
	// rm first and last character and split by ','
	meter_reading_tuple = meter_reading_tuple[1 : len(meter_reading_tuple)-1]
	meter_readings := strings.Split(meter_reading_tuple, ",")
	meter_reading_sum := 0.0

	for _, meter_reading := range meter_readings {
		meter_reading, err := strconv.ParseFloat(meter_reading, 64)
		if err != nil {
			log.Printf("Error converting meter reading to float: %v", err)
			return
		}
		meter_reading_sum += meter_reading
	}
	meter_reading, err := strconv.ParseFloat(meter_readings[1], 64)
	if err != nil {
		log.Printf("Error converting meter reading to float: %v", err)
		return
	}
	// meter_reading := meter_reading_sum / float64(len(meter_readings))
	ts, err := time.Parse(time.DateTime, timestamp)
	if err != nil {
		log.Printf("Error converting timestamp to time: %v in ExportRecordAsPrometheusGaugePrink", err)
		return
	}

	prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(PrinkGauge.WithLabelValues(buildingID, primaryUse).Desc(), prometheus.GaugeValue, meter_reading, buildingID, primaryUse))
	// PrinkGauge.WithLabelValues(buildingID, primaryUse).Set(meter_reading)
}
