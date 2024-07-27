package utils

import (
	"fmt"

	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/polling"
)

func CreateReportURL(metric polling.MetricsModel) string {
	var url = "/update"
	if metric.Type == core.Gauge {
		url += "/gauge"
	} else {
		url += "/counter"
	}

	return fmt.Sprintf("%s/%s/%s", url, metric.Key, metric.Value)
}
