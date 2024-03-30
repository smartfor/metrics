package main

import (
	"fmt"
	"github.com/smartfor/metrics/internal/metrics"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const BaseURL = "http://localhost:8080/update"

const PollInterval = 2
const ReportInterval = 10

type Metric struct {
	Type  metrics.MetricType
	Key   string
	Value interface{}
}

var metricStore = make(map[string]Metric)
var pollCounter = int64(0)

func main() {
	go runReportLoop()

	for {
		var memStats = runtime.MemStats{}
		runtime.ReadMemStats(&memStats)

		metricStore["Alloc"] = Metric{Type: metrics.Gauge, Key: "Alloc", Value: float64(memStats.Alloc)}
		metricStore["BuckHashSys"] = Metric{Type: metrics.Gauge, Key: "BuckHashSys", Value: float64(memStats.BuckHashSys)}
		metricStore["Frees"] = Metric{Type: metrics.Gauge, Key: "Frees", Value: float64(memStats.Frees)}
		metricStore["GCCPUFraction"] = Metric{Type: metrics.Gauge, Key: "GCCPUFraction", Value: memStats.GCCPUFraction}
		metricStore["GCSys"] = Metric{Type: metrics.Gauge, Key: "GCSys", Value: float64(memStats.GCSys)}
		metricStore["HeapAlloc"] = Metric{Type: metrics.Gauge, Key: "HeapAlloc", Value: float64(memStats.HeapAlloc)}
		metricStore["HeapIdle"] = Metric{Type: metrics.Gauge, Key: "HeapIdle", Value: float64(memStats.HeapIdle)}
		metricStore["HeapInuse"] = Metric{Type: metrics.Gauge, Key: "HeapInuse", Value: float64(memStats.HeapInuse)}
		metricStore["HeapReleased"] = Metric{Type: metrics.Gauge, Key: "HeapReleased", Value: float64(memStats.HeapReleased)}
		metricStore["HeapSys"] = Metric{Type: metrics.Gauge, Key: "HeapSys", Value: float64(memStats.HeapSys)}
		metricStore["LastGC"] = Metric{Type: metrics.Gauge, Key: "LastGC", Value: float64(memStats.LastGC)}
		metricStore["Lookups"] = Metric{Type: metrics.Gauge, Key: "Lookups", Value: float64(memStats.Lookups)}
		metricStore["MCacheInuse"] = Metric{Type: metrics.Gauge, Key: "MCacheInuse", Value: float64(memStats.MCacheInuse)}
		metricStore["MCacheSys"] = Metric{Type: metrics.Gauge, Key: "MCacheSys", Value: float64(memStats.MCacheSys)}
		metricStore["MSpanInuse"] = Metric{Type: metrics.Gauge, Key: "MSpanInuse", Value: float64(memStats.MSpanInuse)}
		metricStore["MSpanSys"] = Metric{Type: metrics.Gauge, Key: "MSpanSys", Value: float64(memStats.MSpanSys)}
		metricStore["Mallocs"] = Metric{Type: metrics.Gauge, Key: "Mallocs", Value: float64(memStats.Mallocs)}
		metricStore["NextGC"] = Metric{Type: metrics.Gauge, Key: "NextGC", Value: float64(memStats.NextGC)}
		metricStore["NumForcedGC"] = Metric{Type: metrics.Gauge, Key: "NumForcedGC", Value: float64(memStats.NumForcedGC)}
		metricStore["NumGC"] = Metric{Type: metrics.Gauge, Key: "NumGC", Value: float64(memStats.NumGC)}
		metricStore["OtherSys"] = Metric{Type: metrics.Gauge, Key: "OtherSys", Value: float64(memStats.OtherSys)}
		metricStore["PauseTotalNs"] = Metric{Type: metrics.Gauge, Key: "PauseTotalNs", Value: float64(memStats.PauseTotalNs)}
		metricStore["StackInuse"] = Metric{Type: metrics.Gauge, Key: "StackInuse", Value: float64(memStats.StackInuse)}
		metricStore["StackSys"] = Metric{Type: metrics.Gauge, Key: "StackSys", Value: float64(memStats.StackSys)}
		metricStore["Sys"] = Metric{Type: metrics.Gauge, Key: "Sys", Value: float64(memStats.Sys)}
		metricStore["TotalAlloc"] = Metric{Type: metrics.Gauge, Key: "TotalAlloc", Value: float64(memStats.TotalAlloc)}

		pollCounter += 1
		metricStore["PollCount"] = Metric{Type: metrics.Counter, Key: "PollCount", Value: pollCounter}
		metricStore["RandomValue"] = Metric{Type: metrics.Gauge, Key: "RandomValue", Value: rand.Float64()}

		//fmt.Printf("Сбор Метрик окончен: %d\n", pollCounter)
		time.Sleep(PollInterval * time.Second)
	}
}

func runReportLoop() {
	for {
		if len(metricStore) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		for _, v := range metricStore {
			go func(m Metric) {
				str := createURL(m)
				//fmt.Printf("Строка для формирования урла: %s\n", str)
				post, err := http.Post(str, "text/plain", nil)
				if err != nil {
					return
				}

				defer post.Body.Close()
			}(v)
		}

		//fmt.Printf("Отчет закончен: %d\n", pollCounter)
		time.Sleep(ReportInterval * time.Second)
	}
}

func createURL(metric Metric) string {
	var url = strings.Clone(BaseURL)
	if metric.Type == metrics.Gauge {
		url += "/gauge"
	} else {
		url += "/counter"
	}

	if metric.Type == metrics.Counter {
		return fmt.Sprintf("%s/%s/%d", url, metric.Key, metric.Value)
	} else {
		return fmt.Sprintf("%s/%s/%f", url, metric.Key, metric.Value)
	}
}
