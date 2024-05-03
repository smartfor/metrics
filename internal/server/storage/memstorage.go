package storage

import (
	"context"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/utils"
	"sync"
)

type MemStorage struct {
	core.BaseMetricStorage
	backup      core.Storage
	synchronize bool
	mu          *sync.Mutex
}

func NewMemStorage(backup core.Storage, restore bool, synchronize bool) (*MemStorage, error) {
	s := &MemStorage{
		BaseMetricStorage: core.NewBaseMetricStorage(),
		backup:            backup,
		synchronize:       synchronize,
		mu:                &sync.Mutex{},
	}

	if restore {
		if err := core.Sync(backup, s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *MemStorage) Set(metric core.MetricType, key string, value string) error {
	s.Lock()
	defer s.Unlock()

	switch metric {
	case core.Gauge:
		{
			val, err := utils.GaugeFromString(value)
			if err != nil {
				return core.ErrBadMetricValue
			}

			s.SetGauge(key, val)
		}

	case core.Counter:
		{
			val, err := utils.CounterFromString(value)
			if err != nil {
				return core.ErrBadMetricValue
			}

			s.SetCounter(key, val)
		}

	case core.Unknown:
		{
			return core.ErrUnknownMetricType
		}
	}

	if s.synchronize {
		// Получаем значение счетчика, потому что оно увеличилось и нужно синхронизировать
		if metric == core.Counter {
			v, _ := s.GetCounter(key)
			value = utils.CounterAsString(v)
		}
		if err := s.backup.Set(metric, key, value); err != nil {
			return err
		}
	}

	return nil
}

func (s *MemStorage) Get(metric core.MetricType, key string) (string, error) {
	s.Lock()
	defer s.Unlock()

	switch metric {
	case core.Gauge:
		{
			v, ok := s.GetGauge(key)
			if !ok {
				return "", core.ErrNotFound
			}

			return utils.GaugeAsString(v), nil
		}
	case core.Counter:
		{
			v, ok := s.GetCounter(key)
			if !ok {
				return "", core.ErrNotFound
			}
			return utils.CounterAsString(v), nil
		}
	default:
		{
			return "", core.ErrUnknownMetricType
		}
	}
}

func (s *MemStorage) GetAll() (core.BaseMetricStorage, error) {
	s.Lock()
	defer s.Unlock()

	return core.CloneBaseMetricStorage(&s.BaseMetricStorage), nil
}

func (s *MemStorage) Close() error {
	s.Lock()
	defer s.Unlock()

	if err := s.backup.Close(); err != nil {
		return err
	}

	return nil
}

func (s *MemStorage) Lock() {
	s.mu.Lock()
}

func (s *MemStorage) Unlock() {
	s.mu.Unlock()
}

func (s *MemStorage) Ping(_ context.Context) error {
	return nil
}
