package core

import "github.com/smartfor/metrics/internal/metrics"

type Storage interface {
	Set(metric metrics.MetricType, key string, value string) *StorageError
	Get(metric metrics.MetricType, key string) (string, *StorageError)
	GetAll() (map[string]string, *StorageError)
}

type StorageErrorType int

const (
	UnknownMetricType StorageErrorType = iota
	BadMetricValue
	NotFound
)

type StorageError struct {
	Metric metrics.MetricType
	Key    string
	Value  string
	Msg    string
	Type   StorageErrorType
}

func (e *StorageError) Error() string {
	return e.Msg
}
