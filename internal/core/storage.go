package core

import (
	"errors"
)

type Storage interface {
	Set(metric MetricType, key string, value string) error
	Get(metric MetricType, key string) (string, error)
	GetAll() (map[string]string, error)
}

var (
	ErrUnknownMetricType = errors.New("unknown metric type")
	ErrBadMetricValue    = errors.New("bad metric value")
	ErrNotFound          = errors.New("not found")
)
