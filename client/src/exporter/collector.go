package exporter

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var collector *prinkCollector = newPrinkCollector()

type prinkMetricValue struct {
	value float64
	timestamp time.Time
	labelValues []string
}

type prinkMetric struct {
	desc *prometheus.Desc
	valueType prometheus.ValueType
	values []prinkMetricValue
	mtx sync.RWMutex
}

func newPrinkMetric(name string, help string, variableLabels []string) *prinkMetric {
	return &prinkMetric{
		desc: prometheus.NewDesc(name, help, variableLabels, nil),
		valueType: prometheus.GaugeValue,
		values: make([]prinkMetricValue, 0),
	}
}

func (metric *prinkMetric) Add(value float64, timestamp time.Time, labelValues []string) {
	metric.mtx.Lock()
	defer metric.mtx.Unlock()
	metric.values = append(metric.values, prinkMetricValue{value: value, timestamp: timestamp, labelValues: labelValues})
}


type prinkCollector struct {
	rawGaugeMeterReading *prinkMetric
	rawGaugeSquareFeet *prinkMetric
	prinkGaugeMeterReading *prinkMetric
	prinkGaugeSquareFeet *prinkMetric
}


//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func newPrinkCollector() *prinkCollector {
	return &prinkCollector{
		rawGaugeMeterReading: newPrinkMetric("raw_gauge_meter_reading","P", []string{"building_id", "primary_use"}),
		rawGaugeSquareFeet: newPrinkMetric("raw_gauge_square_feet","P", []string{"building_id", "primary_use"}),
		prinkGaugeMeterReading: newPrinkMetric("prink_gauge_meter_reading","P", []string{"building_id", "primary_use"}),
		prinkGaugeSquareFeet: newPrinkMetric("prink_gauge_square_feet","P", []string{"building_id", "primary_use"}),
	}
}



//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *prinkCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.rawGaugeMeterReading.desc
	ch <- collector.rawGaugeSquareFeet.desc
	ch <- collector.prinkGaugeMeterReading.desc
	ch <- collector.prinkGaugeSquareFeet.desc
}

//Collect implements required collect function for all promehteus collectors
func (collector *prinkCollector) Collect(ch chan<- prometheus.Metric) {

	for _, metric := range []*prinkMetric{collector.rawGaugeMeterReading, collector.rawGaugeSquareFeet, collector.prinkGaugeMeterReading, collector.prinkGaugeSquareFeet} {
		metric.mtx.Lock()
		for _, value := range metric.values {
			ts := value.timestamp
			normalizedTs := time.Date(time.Now().Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, ts.Location())
			m := prometheus.MustNewConstMetric(metric.desc, metric.valueType, value.value, value.labelValues...)
			m = prometheus.NewMetricWithTimestamp(normalizedTs, m)
			ch <- m
		}
		// clear values
		metric.values = make([]prinkMetricValue, 0)
		metric.mtx.Unlock()
	}
	
}


func StartPrometheusExporter(addr string) {
	prometheus.MustRegister(collector)
	// Expose /metrics HTTP endpoint using the created custom registry.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(addr, nil))
}


func ExportRecordAsPrometheusGaugeRaw(record []string) {
	// Data fields:
	// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed, building_id2, unix_timestamp,

	// Export record as prometheus Gauge
	buildingID := record[0]
	timestamp := record[1]
	primaryUse := record[3]
	meter_reading, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		log.Printf("Error converting meter reading to float: %v", err)
		return
	}

	ts, err := time.Parse(time.DateTime, timestamp)
	if err != nil {
		log.Printf("Error converting timestamp to time: %v in ExportRecordAsPrometheusGaugeRaw", err)
		return
	}

	square_feet, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		log.Printf("Error converting square feet to float: %v", err)
		return
	}

	collector.rawGaugeMeterReading.Add(meter_reading, ts, []string{buildingID, primaryUse})
	collector.rawGaugeSquareFeet.Add(square_feet, ts, []string{buildingID, primaryUse})
}

func ExportRecordAsPrometheusGaugePrink(record []string) {
	// Data fields:
	// building_id, timestamp, meter_reading, primary_use, square_feet, year_built, floor_count, air_temperature, cloud_coverage, dew_temperature, precip_depth_1_hr, sea_level_pressure, wind_direction, wind_speed, building_id2, unix_timestamp,

	// Export record as prometheus Gauge
	buildingID := record[0]
	timestamp := record[1]
	primaryUse := record[3]
	square_feet_tuple := record[4]

	meter_reading, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		log.Printf("Error converting meter reading to float: %v", err)
		return
	}
	
	ts, err := time.Parse(time.DateTime, timestamp)
	if err != nil {
		log.Printf("Error converting timestamp to time: %v in ExportRecordAsPrometheusGaugeRaw", err)
		return
	}

	// Convert Tuples into single values
	// rm first and last character and split by ','
	square_feet_tuple = square_feet_tuple[1 : len(square_feet_tuple)-1]
	squre_feet_list := strings.Split(square_feet_tuple, ",")
	square_feet_sum := 0.0

	for _, square_feet_entry := range squre_feet_list {
		square_feet_entry, err := strconv.ParseFloat(square_feet_entry, 64)
		if err != nil {
			log.Printf("Error converting square_feet to float: %v", err)
			return
		}
		square_feet_sum += square_feet_entry
	}
	square_feet := square_feet_sum / float64(len(squre_feet_list))

	// Expose the data as prometheus gauges
	
	collector.prinkGaugeMeterReading.Add(meter_reading, ts, []string{buildingID, primaryUse})
	collector.prinkGaugeSquareFeet.Add(square_feet, ts, []string{buildingID, primaryUse})
}
