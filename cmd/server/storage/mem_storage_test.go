package storage

import (
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemStorage_Get(t *testing.T) {

	tests := []struct {
		name       string
		metricType metrics.MetricType
		key        string
		expected   interface{}
		wantErr    bool
	}{
		{
			name:       "Gauge: positive 1",
			metricType: metrics.Gauge,
			key:        "k1",
			expected:   float64(1),
			wantErr:    false,
		},
		{
			name:       "Counter: positive 42",
			metricType: metrics.Counter,
			key:        "c1",
			expected:   int64(42),
			wantErr:    false,
		},
		{
			name:       "Negative - bad metric type",
			metricType: metrics.Gauge,
			key:        "key1",
			expected:   nil,
			wantErr:    true,
		},
		{
			name:       "Gauge: Negative - not found by key",
			metricType: metrics.Gauge,
			key:        "badKey",
			expected:   nil,
			wantErr:    true,
		},
		{
			name:       "Counter: Negative - not found by key",
			metricType: metrics.Counter,
			key:        "badKey",
			expected:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			s.Set(metrics.Gauge, "k1", "1")
			s.Set(metrics.Counter, "c1", "42")

			value, err := s.Get(tt.metricType, tt.key)
			if tt.wantErr {
				require.Error(t, err)
			}

			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestMemStorage_Set(t *testing.T) {
	tests := []struct {
		name       string
		metricType metrics.MetricType
		key        string
		value      string
		expected   interface{}
		wantErr    bool
	}{
		{
			name:       "Gauge: positive 123.123",
			metricType: metrics.Gauge,
			key:        "key1",
			value:      "123.123",
			expected:   float64(123.123),
			wantErr:    false,
		},
		{
			name:       "Gauge: positive 123",
			metricType: metrics.Gauge,
			key:        "key1",
			value:      "123",
			expected:   float64(123),
			wantErr:    false,
		},
		{
			name:       "Counter: positive 123",
			metricType: metrics.Counter,
			key:        "key1",
			value:      "123",
			expected:   int64(123),
			wantErr:    false,
		},
		{
			name:       "Negative - unknown metric type",
			metricType: "someUnknown",
			key:        "key1",
			value:      "123",
			expected:   nil,
			wantErr:    true,
		},
		{
			name:       "Gauge: Negative - bad value type",
			metricType: metrics.Gauge,
			key:        "key1",
			value:      "sdfsdf",
			expected:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStorage()
			err := s.Set(tt.metricType, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
			}

			m, _ := s.Get(tt.metricType, tt.key)
			assert.Equal(t, tt.expected, m)
		})
	}

	t.Run("Counter - incremented few times", func(t *testing.T) {
		s := NewMemStorage()
		k := "someCounter"
		s.Set(metrics.Counter, k, "12")
		s.Set(metrics.Counter, k, "2")
		s.Set(metrics.Counter, k, "8")
		s.Set(metrics.Counter, k, "-7")

		actual, _ := s.Get(metrics.Counter, k)
		assert.Equal(t, int64(15), actual)
	})
}
