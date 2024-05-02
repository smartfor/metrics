package polling

import (
	"github.com/smartfor/metrics/internal/core"
)

type MetricsModel struct {
	Type  core.MetricType
	Key   string
	Value string
}
