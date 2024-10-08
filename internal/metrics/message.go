// Package metrics содержит структуры и функции для работы с метриками.
package metrics

import (
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/polling"
	"github.com/smartfor/metrics/internal/server/utils"
)

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
}

func FromMetricModel(m polling.MetricsModel) (*Metrics, error) {
	var (
		metric Metrics
		err    error
	)

	metric.ID = m.Key
	metric.MType = string(m.Type)

	switch m.Type {
	case core.Counter:
		var delta int64
		if delta, err = utils.CounterFromString(m.Value); err != nil {
			return nil, core.ErrBadMetricValue
		}
		metric.Delta = &delta
	case core.Gauge:
		var value float64
		if value, err = utils.GaugeFromString(m.Value); err != nil {
			return nil, core.ErrBadMetricValue
		}
		metric.Value = &value
	default:
		return nil, core.ErrUnknownMetricType
	}

	return &metric, nil
}
