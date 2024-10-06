// Package storage Модуль опредялет основные типы хранения метрик
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/utils"
	utils2 "github.com/smartfor/metrics/internal/utils"
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

// FileStorage - тип для хранения состояния метрик в файле
type FileStorage struct {
	file    *os.File
	mu      *sync.Mutex
	encoder *json.Encoder
}

// NewFileStorage - конструктор для создания файлового хранилища
// где filepath - это путь к файлу в котором будут храниться метрики.
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

func (f *FileStorage) SetBatch(_ context.Context, batch core.BaseMetricStorage) error {
	f.lock()
	defer f.unlock()

	metrics := &metrics{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
	}

	if err := f.read(metrics); err != nil {
		return err
	}

	for k, v := range batch.Gauges() {
		metrics.Gauges[k] = v
	}

	for k, v := range batch.Counters() {
		current, ok := metrics.Counters[k]
		if !ok {
			current = 0
		}
		metrics.Counters[k] = current + v
	}

	if err := utils2.RetryVoid(func() error {
		return f.write(metrics)
	}, nil); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) Set(ctx context.Context, key string, value string, metric core.MetricType) error {
	f.lock()
	defer f.unlock()

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

	if err := utils2.RetryVoid(func() error {
		return f.write(metrics)
	}, nil); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) Get(ctx context.Context, key string, metric core.MetricType) (string, error) {
	f.lock()
	defer f.unlock()

	var lMetrics metrics
	if err := f.read(&lMetrics); err != nil {
		return "", err
	}

	switch metric {
	case core.Gauge:
		if v, ok := lMetrics.Gauges[key]; ok {
			return utils.GaugeAsString(v), nil
		} else {
			return "", core.ErrNotFound
		}
	case core.Counter:
		if v, ok := lMetrics.Counters[key]; ok {
			return utils.CounterAsString(v), nil
		} else {
			return "", core.ErrNotFound
		}
	default:
		return "", core.ErrUnknownMetricType
	}
}

func (f *FileStorage) GetAll(context.Context) (core.BaseMetricStorage, error) {
	f.lock()
	defer f.unlock()

	var lMetrics metrics
	if err := f.read(&lMetrics); err != nil {
		return core.NewBaseMetricStorage(), err
	}

	return *lMetrics.ToBaseStorage(), nil
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

func (f *FileStorage) lock() {
	f.mu.Lock()
}

func (f *FileStorage) unlock() {
	f.mu.Unlock()
}

func (f *FileStorage) Ping(_ context.Context) error {
	return nil
}
