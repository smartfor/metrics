package utils

import (
	"fmt"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/polling"
)

func CreateReportURL(metric polling.Metric) string {
	var url = "/update"
	if metric.Type == metrics.Gauge {
		url += "/gauge"
	} else {
		url += "/counter"
	}

	return fmt.Sprintf("%s/%s/%s", url, metric.Key, metric.Value)
}
