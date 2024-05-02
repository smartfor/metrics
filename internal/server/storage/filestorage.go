package storage

import (
	"encoding/json"
	"errors"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/utils"
	"io"
	"os"
	"sync"
)

type metrics struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func (metrics *metrics) ToBaseStorage() *core.BaseMetricStorage {
	s := core.NewBaseMetricStorage()

	for k, v := range metrics.Gauges {
		s.SetGauge(k, v)
	}

	for k, v := range metrics.Counters {
		s.SetCounter(k, v)
	}

	return &s
}

type FileStorage struct {
	file    *os.File
	mu      *sync.Mutex
	encoder *json.Encoder
}

func NewFileStorage(filepath string) (*FileStorage, error) {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")

	return &FileStorage{
		file:    file,
		mu:      &sync.Mutex{},
		encoder: encoder,
	}, nil
}

func (f *FileStorage) Set(metric core.MetricType, key string, value string) error {
	f.Lock()
	defer f.Unlock()

	metrics := &metrics{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
	}

	if err := f.read(metrics); err != nil {
		return err
	}

	switch metric {
	case core.Gauge:
		val, err := utils.GaugeFromString(value)
		if err != nil {
			return core.ErrBadMetricValue
		}
		metrics.Gauges[key] = val
	case core.Counter:
		val, err := utils.CounterFromString(value)
		if err != nil {
			return core.ErrBadMetricValue
		}
		metrics.Counters[key] = val
	default:
		return core.ErrUnknownMetricType
	}

	if err := f.write(metrics); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) Get(metric core.MetricType, key string) (string, error) {
	f.Lock()
	defer f.Unlock()

	var metrics metrics
	if err := f.read(&metrics); err != nil {
		return "", err
	}

	switch metric {
	case core.Gauge:
		if v, ok := metrics.Gauges[key]; ok {
			return utils.GaugeAsString(v), nil
		} else {
			return "", core.ErrNotFound
		}
	case core.Counter:
		if v, ok := metrics.Counters[key]; ok {
			return utils.CounterAsString(v), nil
		} else {
			return "", core.ErrNotFound
		}
	default:
		return "", core.ErrUnknownMetricType
	}
}

func (f *FileStorage) GetAll() (core.BaseMetricStorage, error) {
	f.Lock()
	defer f.Unlock()

	var metrics metrics
	if err := f.read(&metrics); err != nil {
		return core.NewBaseMetricStorage(), err
	}

	return *metrics.ToBaseStorage(), nil
}

func (f *FileStorage) write(metrics *metrics) error {
	if err := f.clear(); err != nil {
		return err
	}

	if err := f.encoder.Encode(metrics); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) read(content *metrics) error {
	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	if err := json.NewDecoder(f.file).Decode(content); err != nil {
		if errors.Is(err, io.EOF) {
			*content = metrics{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			}
		} else {
			return err
		}
	}

	return nil
}

func (f *FileStorage) clear() error {
	if err := f.file.Truncate(0); err != nil {
		return err
	}

	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}

func (f *FileStorage) Lock() {
	f.mu.Lock()
}

func (f *FileStorage) Unlock() {
	f.mu.Unlock()
}
