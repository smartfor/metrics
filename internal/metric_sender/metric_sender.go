package metric_sender

import (
	"github.com/smartfor/metrics/internal/metrics"
)

type SendOptions struct {
	PrivateKey []byte
	Secret     string
}

// MetricSender интерфейс для отправки метрик
type MetricSender interface {
	Send(metrics []metrics.Metrics, options SendOptions) error
}
