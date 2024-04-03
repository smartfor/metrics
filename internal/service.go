package internal

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/polling"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Metric = polling.Metric

type Service struct {
	config config.Config
	store  map[string]Metric
	client *resty.Client
	mu     sync.Mutex
}

func NewService(cfg *config.Config) Service {
	client := resty.
		New().
		SetBaseURL(cfg.HostEndpoint).
		SetTimeout(cfg.ResponseTimeout)

	return Service{
		config: *cfg,
		store:  make(map[string]Metric),
		client: client,
	}
}

func (s *Service) Run() {
	fmt.Println("Metrics Agent is started...")

	go func() {
		for {
			s.mu.Lock()
			if len(s.store) == 0 {
				s.mu.Unlock()
				time.Sleep(1 * time.Second)
				continue
			} else {
				s.mu.Unlock()
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
	fmt.Println("start send..")
	var wg sync.WaitGroup
	for _, v := range s.store {
		wg.Add(1)

		go func(m Metric) {
			defer wg.Done()

			str := s.createURL(m)

			_, err := s.client.R().
				SetHeader("Content-Type", "application/json").
				Post(str)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Send report error: ", err)
			}
		}(v)
	}

	wg.Wait()
	fmt.Println("end send..")

	s.resetPollCounter()
}

func (s *Service) poll() {
	var ms = runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	s.updateGaugeMetrics(&ms)
	s.updatePollCounter()
}

func (s *Service) updateGaugeMetrics(ms *runtime.MemStats) {

	s.mu.Lock()
	fmt.Println("start update gauges...")
	s.store["Alloc"] = Metric{Type: metrics.Gauge, Key: "Alloc", Value: strconv.FormatUint(ms.Alloc, 10)}
	s.store["BuckHashSys"] = Metric{Type: metrics.Gauge, Key: "BuckHashSys", Value: strconv.FormatUint(ms.BuckHashSys, 10)}
	s.store["Frees"] = Metric{Type: metrics.Gauge, Key: "Frees", Value: strconv.FormatUint(ms.Frees, 10)}
	s.store["GCCPUFraction"] = Metric{Type: metrics.Gauge, Key: "GCCPUFraction", Value: strconv.FormatFloat(ms.GCCPUFraction, 'f', -1, 64)}
	s.store["GCSys"] = Metric{Type: metrics.Gauge, Key: "GCSys", Value: strconv.FormatUint(ms.GCSys, 10)}
	s.store["HeapAlloc"] = Metric{Type: metrics.Gauge, Key: "HeapAlloc", Value: strconv.FormatUint(ms.HeapAlloc, 10)}
	s.store["HeapIdle"] = Metric{Type: metrics.Gauge, Key: "HeapIdle", Value: strconv.FormatUint(ms.HeapIdle, 10)}
	s.store["HeapInuse"] = Metric{Type: metrics.Gauge, Key: "HeapInuse", Value: strconv.FormatUint(ms.HeapInuse, 10)}
	s.store["HeapReleased"] = Metric{Type: metrics.Gauge, Key: "HeapReleased", Value: strconv.FormatUint(ms.HeapReleased, 10)}
	s.store["HeapSys"] = Metric{Type: metrics.Gauge, Key: "HeapSys", Value: strconv.FormatUint(ms.HeapSys, 10)}
	s.store["LastGC"] = Metric{Type: metrics.Gauge, Key: "LastGC", Value: strconv.FormatUint(ms.LastGC, 10)}
	s.store["Lookups"] = Metric{Type: metrics.Gauge, Key: "Lookups", Value: strconv.FormatUint(ms.Lookups, 10)}
	s.store["MCacheInuse"] = Metric{Type: metrics.Gauge, Key: "MCacheInuse", Value: strconv.FormatUint(ms.MCacheInuse, 10)}
	s.store["MCacheSys"] = Metric{Type: metrics.Gauge, Key: "MCacheSys", Value: strconv.FormatUint(ms.MCacheSys, 10)}
	s.store["MSpanInuse"] = Metric{Type: metrics.Gauge, Key: "MSpanInuse", Value: strconv.FormatUint(ms.MSpanInuse, 10)}
	s.store["MSpanSys"] = Metric{Type: metrics.Gauge, Key: "MSpanSys", Value: strconv.FormatUint(ms.MSpanSys, 10)}
	s.store["Mallocs"] = Metric{Type: metrics.Gauge, Key: "Mallocs", Value: strconv.FormatUint(ms.Mallocs, 10)}
	s.store["NextGC"] = Metric{Type: metrics.Gauge, Key: "NextGC", Value: strconv.FormatUint(ms.NextGC, 10)}
	s.store["NumForcedGC"] = Metric{Type: metrics.Gauge, Key: "NumForcedGC", Value: strconv.FormatUint(uint64(ms.NumForcedGC), 10)}
	s.store["NumGC"] = Metric{Type: metrics.Gauge, Key: "NumGC", Value: strconv.FormatUint(uint64(ms.NumGC), 10)}
	s.store["OtherSys"] = Metric{Type: metrics.Gauge, Key: "OtherSys", Value: strconv.FormatUint(ms.OtherSys, 10)}
	s.store["PauseTotalNs"] = Metric{Type: metrics.Gauge, Key: "PauseTotalNs", Value: strconv.FormatUint(ms.PauseTotalNs, 10)}
	s.store["StackInuse"] = Metric{Type: metrics.Gauge, Key: "StackInuse", Value: strconv.FormatUint(ms.StackInuse, 10)}
	s.store["StackSys"] = Metric{Type: metrics.Gauge, Key: "StackSys", Value: strconv.FormatUint(ms.StackSys, 10)}
	s.store["Sys"] = Metric{Type: metrics.Gauge, Key: "Sys", Value: strconv.FormatUint(ms.Sys, 10)}
	s.store["TotalAlloc"] = Metric{Type: metrics.Gauge, Key: "TotalAlloc", Value: strconv.FormatUint(ms.TotalAlloc, 10)}
	s.store["RandomValue"] = Metric{Type: metrics.Gauge, Key: "RandomValue", Value: strconv.FormatFloat(rand.Float64(), 'f', -1, 64)}
	fmt.Println("end update gauges")

	s.mu.Unlock()
}

func (s *Service) updatePollCounter() {
	s.mu.Lock()

	defer s.mu.Unlock()

	key := "PollCount"

	counterStr, ok := s.store[key]
	fmt.Println("start update poll counter, before", counterStr)
	if !ok {
		s.store[key] = Metric{Type: metrics.Counter, Key: key, Value: strconv.FormatInt(0, 10)}
		counterStr = s.store[key]
	}
	counter, _ := strconv.ParseInt(counterStr.Value, 10, 64)
	s.store[key] = Metric{Type: metrics.Counter, Key: key, Value: strconv.FormatInt(counter+1, 10)}
	fmt.Println("start update poll counter, after", s.store[key].Value)
}

func (s *Service) resetPollCounter() {
	s.mu.Lock()
	fmt.Println("reset poll counter, before :: ", s.store["PollCount"])
	s.store["PollCount"] = Metric{Type: metrics.Counter, Key: "PollCount", Value: strconv.FormatInt(0, 10)}
	fmt.Println("reset poll counter, after :: ", s.store["PollCount"])
	s.mu.Unlock()
	fmt.Println("reset poll counter")

}

func (s *Service) createURL(metric Metric) string {
	var url = "/update"
	if metric.Type == metrics.Gauge {
		url += "/gauge"
	} else {
		url += "/counter"
	}

	return fmt.Sprintf("%s/%s/%s", url, metric.Key, metric.Value)
}
