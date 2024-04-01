package internal

import (
	"fmt"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/polling"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type Metric = polling.Metric

type Service struct {
	config      config.Config
	pollCounter int64
	store       map[string]Metric
	httpClient  http.Client
}

func NewService(cfg *config.Config) Service {
	if cfg == nil {
		cfg = &config.DefaultConfig
	}

	return Service{
		config:      *cfg,
		store:       make(map[string]Metric),
		pollCounter: 0,
		httpClient: http.Client{
			Timeout: time.Second,
		},
	}
}

func (s *Service) Run() {
	fmt.Println("Metrics Agent is started...")

	go func() {
		for {
			if len(s.store) == 0 {
				time.Sleep(1 * time.Second)
				continue
			}

			s.send()
			time.Sleep(s.config.ReportInterval)
		}
	}()

	for {
		s.poll()
		time.Sleep(s.config.PollInterval)
	}
}

func (s *Service) send() {
	for _, v := range s.store {
		go func(m Metric) {
			str := s.createURL(m)
			//fmt.Printf("Строка для формирования урла: %s\n", str)
			res, err := s.httpClient.Post(str, "text/plain", nil)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Send report error: ", err)
				return
			}

			fmt.Println(res.Status)

			defer res.Body.Close()
		}(v)
	}
}

func (s *Service) poll() {
	var memStats = runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	s.store["Alloc"] = Metric{Type: metrics.Gauge, Key: "Alloc", Value: float64(memStats.Alloc)}
	s.store["BuckHashSys"] = Metric{Type: metrics.Gauge, Key: "BuckHashSys", Value: float64(memStats.BuckHashSys)}
	s.store["Frees"] = Metric{Type: metrics.Gauge, Key: "Frees", Value: float64(memStats.Frees)}
	s.store["GCCPUFraction"] = Metric{Type: metrics.Gauge, Key: "GCCPUFraction", Value: memStats.GCCPUFraction}
	s.store["GCSys"] = Metric{Type: metrics.Gauge, Key: "GCSys", Value: float64(memStats.GCSys)}
	s.store["HeapAlloc"] = Metric{Type: metrics.Gauge, Key: "HeapAlloc", Value: float64(memStats.HeapAlloc)}
	s.store["HeapIdle"] = Metric{Type: metrics.Gauge, Key: "HeapIdle", Value: float64(memStats.HeapIdle)}
	s.store["HeapInuse"] = Metric{Type: metrics.Gauge, Key: "HeapInuse", Value: float64(memStats.HeapInuse)}
	s.store["HeapReleased"] = Metric{Type: metrics.Gauge, Key: "HeapReleased", Value: float64(memStats.HeapReleased)}
	s.store["HeapSys"] = Metric{Type: metrics.Gauge, Key: "HeapSys", Value: float64(memStats.HeapSys)}
	s.store["LastGC"] = Metric{Type: metrics.Gauge, Key: "LastGC", Value: float64(memStats.LastGC)}
	s.store["Lookups"] = Metric{Type: metrics.Gauge, Key: "Lookups", Value: float64(memStats.Lookups)}
	s.store["MCacheInuse"] = Metric{Type: metrics.Gauge, Key: "MCacheInuse", Value: float64(memStats.MCacheInuse)}
	s.store["MCacheSys"] = Metric{Type: metrics.Gauge, Key: "MCacheSys", Value: float64(memStats.MCacheSys)}
	s.store["MSpanInuse"] = Metric{Type: metrics.Gauge, Key: "MSpanInuse", Value: float64(memStats.MSpanInuse)}
	s.store["MSpanSys"] = Metric{Type: metrics.Gauge, Key: "MSpanSys", Value: float64(memStats.MSpanSys)}
	s.store["Mallocs"] = Metric{Type: metrics.Gauge, Key: "Mallocs", Value: float64(memStats.Mallocs)}
	s.store["NextGC"] = Metric{Type: metrics.Gauge, Key: "NextGC", Value: float64(memStats.NextGC)}
	s.store["NumForcedGC"] = Metric{Type: metrics.Gauge, Key: "NumForcedGC", Value: float64(memStats.NumForcedGC)}
	s.store["NumGC"] = Metric{Type: metrics.Gauge, Key: "NumGC", Value: float64(memStats.NumGC)}
	s.store["OtherSys"] = Metric{Type: metrics.Gauge, Key: "OtherSys", Value: float64(memStats.OtherSys)}
	s.store["PauseTotalNs"] = Metric{Type: metrics.Gauge, Key: "PauseTotalNs", Value: float64(memStats.PauseTotalNs)}
	s.store["StackInuse"] = Metric{Type: metrics.Gauge, Key: "StackInuse", Value: float64(memStats.StackInuse)}
	s.store["StackSys"] = Metric{Type: metrics.Gauge, Key: "StackSys", Value: float64(memStats.StackSys)}
	s.store["Sys"] = Metric{Type: metrics.Gauge, Key: "Sys", Value: float64(memStats.Sys)}
	s.store["TotalAlloc"] = Metric{Type: metrics.Gauge, Key: "TotalAlloc", Value: float64(memStats.TotalAlloc)}

	s.store["RandomValue"] = Metric{Type: metrics.Gauge, Key: "RandomValue", Value: rand.Float64()}

	s.pollCounter += 1
	s.store["PollCount"] = Metric{Type: metrics.Counter, Key: "PollCount", Value: s.pollCounter}
	//fmt.Printf("Сбор Метрик окончен: %d\n", pollCounter)
}

func (s *Service) createURL(metric Metric) string {
	var url = strings.Clone(s.config.UpdateURL)
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
