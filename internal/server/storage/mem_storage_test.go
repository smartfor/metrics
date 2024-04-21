package storage

import (
	"github.com/smartfor/metrics/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemStorage_Get(t *testing.T) {
	tests := []struct {
		name       string
		metricType core.MetricType
		key        string
		expected   interface{}
		wantErr    bool
	}{
		{
			name:       "Gauge: positive 1",
			metricType: core.Gauge,
			key:        "k1",
			expected:   "1",
			wantErr:    false,
		},
		{
			name:       "Counter: positive 42",
			metricType: core.Counter,
			key:        "c1",
			expected:   "42",
			wantErr:    false,
		},
		{
			name:       "Negative - bad metric type",
			metricType: core.Gauge,
			key:        "key1",
			expected:   nil,
			wantErr:    true,
		},
		{
			name:       "Gauge: Negative - not found by key",
			metricType: core.Gauge,
			key:        "badKey",
			expected:   nil,
			wantErr:    true,
		},
		{
			name:       "Counter: Negative - not found by key",
			metricType: core.Counter,
			key:        "badKey",
			expected:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStorage("/tmp/metric.json")
			if err != nil {
				t.Fatal(err)
			}
			s, err := NewMemStorage(fs, false, false)
			if err != nil {
				t.Fatal(err)
			}

			s.Set(core.Gauge, "k1", "1")
			s.Set(core.Counter, "c1", "42")

			value, err := s.Get(tt.metricType, tt.key)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestMemStorage_Set(t *testing.T) {
	tests := []struct {
		name       string
		metricType core.MetricType
		key        string
		value      string
		expected   interface{}
		wantErr    bool
	}{
		{
			name:       "Gauge: positive 123.123",
			metricType: core.Gauge,
			key:        "key1",
			value:      "123.123",
			expected:   "123.123",
			wantErr:    false,
		},
		{
			name:       "Gauge: positive 123",
			metricType: core.Gauge,
			key:        "key1",
			value:      "123",
			expected:   "123",
			wantErr:    false,
		},
		{
			name:       "Counter: positive 123",
			metricType: core.Counter,
			key:        "key1",
			value:      "123",
			expected:   "123",
			wantErr:    false,
		},
		{
			name:       "Negative - unknown metric type",
			metricType: "someUnknown",
			key:        "key1",
			value:      "123",
			expected:   "",
			wantErr:    true,
		},
		{
			name:       "Gauge: Negative - bad value type",
			metricType: core.Gauge,
			key:        "key1",
			value:      "sdfsdf",
			expected:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStorage("/tmp/metric.json")
			if err != nil {
				t.Fatal(err)
			}
			s, err := NewMemStorage(fs, false, false)
			if err != nil {
				t.Fatal(err)
			}

			mType := core.NewMetricType(string(tt.metricType))
			err = s.Set(mType, tt.key, tt.value)

			if tt.wantErr {
				require.Error(t, err)
			}

			m, _ := s.Get(tt.metricType, tt.key)
			assert.Equal(t, tt.expected, m)
		})
	}

	t.Run("Counter - incremented few times", func(t *testing.T) {
		fs, err := NewFileStorage("/tmp/metric.json")
		if err != nil {
			t.Fatal(err)
		}
		s, err := NewMemStorage(fs, false, false)
		if err != nil {
			t.Fatal(err)
		}

		k := "someCounter"
		s.Set(core.Counter, k, "12")
		s.Set(core.Counter, k, "2")
		s.Set(core.Counter, k, "8")
		s.Set(core.Counter, k, "-7")

		actual, _ := s.Get(core.Counter, k)
		assert.Equal(t, "15", actual)
	})
}
