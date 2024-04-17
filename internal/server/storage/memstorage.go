package storage

import (
	"fmt"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/server/utils"
	"strconv"
	"sync"
)

type MemStorage struct {
	store map[metrics.MetricType]map[string]interface{}
	mu    *sync.Mutex
}

func NewMemStorage() *MemStorage {
	s := &MemStorage{
		store: make(map[metrics.MetricType]map[string]interface{}),
		mu:    &sync.Mutex{},
	}

	s.store[metrics.Gauge] = make(map[string]interface{})
	s.store[metrics.Counter] = make(map[string]interface{})

	return s
}

func (s *MemStorage) Set(metric metrics.MetricType, key string, value string) *core.StorageError {
	s.mu.Lock()
	defer s.mu.Unlock()

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

			s.store[metric][key] = val
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

			if _, ok := s.store[metric][key]; !ok {
				s.store[metric][key] = int64(0)
			}

			s.store[metric][key] = s.store[metric][key].(int64) + val
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

func (s *MemStorage) Get(metric metrics.MetricType, key string) (string, *core.StorageError) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if metric == metrics.Unknown {
		return "", &core.StorageError{
			Msg:  "unknown Metric Type",
			Key:  key,
			Type: core.UnknownMetricType,
		}
	}

	value, ok := s.store[metric][key]
	if !ok {
		return "", &core.StorageError{
			Msg:  fmt.Sprintf("not found by type: %s key: %s", metric, key),
			Key:  key,
			Type: core.NotFound,
		}
	}

	if metric == metrics.Gauge {
		return utils.GaugeAsString(value), nil
	} else {
		return utils.CounterAsString(value), nil
	}
}

func (s *MemStorage) GetAll() (map[string]string, *core.StorageError) {
	var out = make(map[string]string)

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.store[metrics.Gauge] {
		out[k] = utils.GaugeAsString(v)
	}

	for k, v := range s.store[metrics.Counter] {
		out[k] = utils.CounterAsString(v)
	}

	return out, nil
}
