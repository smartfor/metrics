package storage

import (
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"strconv"
)

type MemStorage struct {
	store map[metrics.MetricType]map[string]interface{}
}

func NewMemStorage() core.Storage {
	var s = MemStorage{
		store: make(map[metrics.MetricType]map[string]interface{}),
	}

	s.store[metrics.Gauge] = make(map[string]interface{})
	s.store[metrics.Counter] = make(map[string]interface{})

	return s
}

func (storage MemStorage) Set(metric metrics.MetricType, key string, value string) *core.StorageError {
	switch metric {
	case metrics.Gauge:
		{
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return &core.StorageError{
					Msg:   err.Error(),
					Key:   key,
					Value: value,
					Type:  core.BadMetricValue,
				}
			}

			storage.store[metric][key] = val
		}

	case metrics.Counter:
		{
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return &core.StorageError{
					Msg:   err.Error(),
					Key:   key,
					Value: value,
					Type:  core.BadMetricValue,
				}
			}

			if _, ok := storage.store[metric][key]; !ok {
				storage.store[metric][key] = int64(0)
			}

			storage.store[metric][key] = storage.store[metric][key].(int64) + val
		}

	case metrics.Unknown:
		{
			return &core.StorageError{
				Msg:   "unknown metric type",
				Key:   key,
				Value: value,
				Type:  core.UnknownMetricType,
			}
		}
	}

	return nil
}

func (storage MemStorage) Get(metric metrics.MetricType, key string) (interface{}, *core.StorageError) {
	if metric == metrics.Unknown {
		return nil, &core.StorageError{
			Msg:  "Unknown Metric Type",
			Key:  key,
			Type: core.UnknownMetricType,
		}
	}

	return storage.Get(metric, key)
}
