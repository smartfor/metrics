package polling

import "github.com/smartfor/metrics/internal/metrics"

type Metric struct {
	Type  metrics.MetricType
	Key   string
	Value string
}
