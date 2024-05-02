package core

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"

	Unknown MetricType = "unknown"
)

func NewMetricType(str string) MetricType {
	if str == "gauge" {
		return Gauge
	} else if str == "counter" {
		return Counter
	} else {
		return Unknown
	}
}
