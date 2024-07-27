package core

import (
	"context"
	"errors"
	"io"

	"github.com/smartfor/metrics/internal/server/utils"
)

var (
	ErrUnknownMetricType = errors.New("unknown metric type")
	ErrBadMetricValue    = errors.New("bad metric value")
	ErrNotFound          = errors.New("not found")
)

type Storage interface {
	io.Closer
	Set(ctx context.Context, key string, value string, metric MetricType) error
	SetBatch(ctx context.Context, batch BaseMetricStorage) error
	Get(ctx context.Context, key string, metric MetricType) (string, error)
	GetAll(ctx context.Context) (BaseMetricStorage, error)
	Ping(ctx context.Context) error
}

type BaseMetricStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewBaseMetricStorage() BaseMetricStorage {
	return BaseMetricStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func NewBaseMetricStorageWithValues(gauges map[string]float64, counters map[string]int64) BaseMetricStorage {
	return BaseMetricStorage{
		gauges:   gauges,
		counters: counters,
	}
}

func CloneBaseMetricStorage(storage *BaseMetricStorage) BaseMetricStorage {
	gaugesCopy := make(map[string]float64)
	for k, v := range storage.Gauges() {
		gaugesCopy[k] = v
	}

	countersCopy := make(map[string]int64)
	for k, v := range storage.Counters() {
		countersCopy[k] = v
	}

	return BaseMetricStorage{
		gauges:   gaugesCopy,
		counters: countersCopy,
	}
}

func (bs *BaseMetricStorage) Gauges() map[string]float64 {
	return bs.gauges
}

func (bs *BaseMetricStorage) Counters() map[string]int64 {
	return bs.counters
}

func (bs *BaseMetricStorage) GetCounter(key string) (int64, bool) {
	c, ok := bs.Counters()[key]
	return c, ok
}

func (bs *BaseMetricStorage) GetGauge(key string) (float64, bool) {
	g, ok := bs.Gauges()[key]
	return g, ok
}

func (bs *BaseMetricStorage) SetGauge(key string, value float64) {
	bs.gauges[key] = value
}

func (bs *BaseMetricStorage) SetCounter(key string, delta int64) {
	if _, ok := bs.counters[key]; !ok {
		bs.counters[key] = 0
	}

	bs.counters[key] += delta
}

func Sync(ctx context.Context, source Storage, target Storage) error {
	main, err := source.GetAll(ctx)
	if err != nil {
		return err
	}

	for k, v := range main.Gauges() {
		if err := target.Set(ctx, k, utils.GaugeAsString(v), Gauge); err != nil {
			return err
		}
	}

	for k, v := range main.Counters() {
		if err := target.Set(ctx, k, utils.CounterAsString(v), Counter); err != nil {
			return err
		}
	}

	return nil
}
