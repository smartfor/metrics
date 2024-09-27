package core

// MetricType - тип метрики
type MetricType string

const (
	// Gauge - обычное значение метрики, которое перезаписывается при каждом обновлении
	Gauge MetricType = "gauge"
	// Counter - значение, которое при установке накапливает свое значение
	Counter MetricType = "counter"
	// Unknown - Неизвестный тип метрики.
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
