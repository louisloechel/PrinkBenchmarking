package evaluation

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RawGaugeMeterReading = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raw_gauge_meter_reading",
			Help: "Values of the raw data",
		},
		[]string{"building_id", "primary_use"},
	)
)
var (
	RawGaugeSquareFeet = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raw_gauge_square_feet",
			Help: "Values of the raw data",
		},
		[]string{"building_id", "primary_use"},
	)
)

var (
	PrinkGaugeMeterReading = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prink_gauge_meter_reading",
			Help: "Values of the prink data",
		},
		[]string{"building_id", "primary_use"},
	)
)
var (
	PrinkGaugeSquareFeet = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "prink_gauge_square_feet",
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
	prometheus.MustRegister(RawGaugeMeterReading)
	prometheus.MustRegister(PrinkGaugeMeterReading)
	prometheus.MustRegister(RawGaugeSquareFeet)
	prometheus.MustRegister(PrinkGaugeSquareFeet)
}

func ExportRecordAsPrometheusGaugeRaw(record []string) {
	// Data fields:
	// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed, building_id2, unix_timestamp,

	// Export record as prometheus Gauge
	buildingID := record[0]
	meter_reading, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		log.Printf("Error converting meter reading to float: %v", err)
		return
	}
	primaryUse := record[3]
	square_feet, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		log.Printf("Error converting square feet to float: %v", err)
		return
	}
	RawGaugeMeterReading.WithLabelValues(buildingID, primaryUse).Set(meter_reading)
	RawGaugeSquareFeet.WithLabelValues(buildingID, primaryUse).Set(square_feet)
}

func ExportRecordAsPrometheusGaugePrink(record []string) {
	// Data fields:
	// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed, building_id2, unix_timestamp,

	// Export record as prometheus Gauge
	buildingID := record[0]
	meter_reading, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		log.Printf("Error converting meter reading to float: %v", err)
		return
	}
	primaryUse := record[3]
	square_feet_tuple := record[4]

	// Convert Tuples into single values
	// rm first and last character and split by ','
	square_feet_tuple = square_feet_tuple[1 : len(square_feet_tuple)-1]
	squre_feet_list := strings.Split(square_feet_tuple, ",")
	square_feet_sum := 0.0

	for _, meter_reading := range squre_feet_list {
		meter_reading, err := strconv.ParseFloat(meter_reading, 64)
		if err != nil {
			log.Printf("Error converting meter reading to float: %v", err)
			return
		}
		square_feet_sum += meter_reading
	}
	square_feet := square_feet_sum / float64(len(squre_feet_list))

	// Expose the data as prometheus gauges
	PrinkGaugeSquareFeet.WithLabelValues(buildingID, primaryUse).Set(square_feet)
	PrinkGaugeMeterReading.WithLabelValues(buildingID, primaryUse).Set(meter_reading)
}
