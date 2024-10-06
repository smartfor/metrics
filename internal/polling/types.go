// Package polling содержит логику сбора метрик.
package polling

import (
	"github.com/smartfor/metrics/internal/core"
)

type MetricsModel struct {
	Type  core.MetricType
	Key   string
	Value string
}

type PollMessage struct {
	Msg  MetricStore
	Err  error
	Type PollMessageType
}
