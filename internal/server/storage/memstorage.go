package storage

import (
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/utils"
	"sync"
)

type MemStorage struct {
	store map[core.MetricType]map[string]interface{}
	mu    *sync.Mutex
}

func NewMemStorage() *MemStorage {
	s := &MemStorage{
		store: make(map[core.MetricType]map[string]interface{}),
		mu:    &sync.Mutex{},
	}

	s.store[core.Gauge] = make(map[string]interface{})
	s.store[core.Counter] = make(map[string]interface{})

	return s
}

func (s *MemStorage) Set(metric core.MetricType, key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch metric {
	case core.Gauge:
		{
			val, err := utils.GaugeFromString(value)
			if err != nil {
				return core.ErrBadMetricValue
			}

			s.store[metric][key] = val
		}

	case core.Counter:
		{
			val, err := utils.CounterFromString(value)
			if err != nil {
				return core.ErrBadMetricValue
			}

			if _, ok := s.store[metric][key]; !ok {
				s.store[metric][key] = int64(0)
			}

			s.store[metric][key] = s.store[metric][key].(int64) + val
		}

	case core.Unknown:
		{
			return core.ErrUnknownMetricType
		}
	}

	return nil
}

func (s *MemStorage) Get(metric core.MetricType, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if metric == core.Unknown {
		return "", core.ErrUnknownMetricType
	}

	value, ok := s.store[metric][key]
	if !ok {
		return "", core.ErrNotFound
	}

	if metric == core.Gauge {
		return utils.GaugeAsString(value), nil
	} else {
		return utils.CounterAsString(value), nil
	}
}

func (s *MemStorage) GetAll() (map[string]string, error) {
	var out = make(map[string]string)

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.store[core.Gauge] {
		out[k] = utils.GaugeAsString(v)
	}

	for k, v := range s.store[core.Counter] {
		out[k] = utils.CounterAsString(v)
	}

	return out, nil
}
