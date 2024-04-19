package core

import (
	"errors"
	"github.com/smartfor/metrics/internal/metrics"
)

type Storage interface {
	Set(metric metrics.MetricType, key string, value string) error
	Get(metric metrics.MetricType, key string) (string, error)
	GetAll() (map[string]string, error)
}

var (
	ErrUnknownMetricType = errors.New("unknown metric type")
	ErrBadMetricValue    = errors.New("bad metric value")
	ErrNotFound          = errors.New("not found")
)
