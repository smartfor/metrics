package internal

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/polling"
	"github.com/smartfor/metrics/internal/utils"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var UpdateURL string = "/update/"
var UpdateBatchURL string = "/updates/"

type Metric = polling.MetricsModel

type Service struct {
	config config.Config
	store  map[string]Metric
	client *resty.Client
	mu     *sync.Mutex
}

func NewService(cfg *config.Config) Service {
	client := resty.
		New().
		SetBaseURL(cfg.HostEndpoint).
		SetHeader("Content-Type", "application/json").
		SetTimeout(cfg.ResponseTimeout)

	return Service{
		config: *cfg,
		store:  make(map[string]Metric),
		client: client,
		mu:     &sync.Mutex{},
	}
}

func (s *Service) Run() {
	fmt.Println("Metrics Agent is started...")

	go func() {
		for {
			if s.isEmptyStore() {
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
	s.mu.Lock()
	defer s.mu.Unlock()

	var (
		batch      []metrics.Metrics
		err        error
		body       []byte
		compressed []byte
	)

	for _, v := range s.store {
		metric, err := metrics.FromMetricModel(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Extract metric from model error: ", err)
			return
		}
		batch = append(batch, *metric)
	}

	if body, err = json.Marshal(batch); err != nil {
		fmt.Fprintln(os.Stderr, "Marshalling batch error: ", err)
		return
	}

	if compressed, err = utils.GzipCompress(body); err != nil {
		fmt.Fprintln(os.Stderr, "Compressed body error: ", err)
		return
	}

	fmt.Println("start send..")
	_, err = s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressed).
		Post(UpdateBatchURL)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Send report error: ", err)
		return
	}

	s.resetPollCounter()
	fmt.Println("...End send")
}

func (s *Service) poll() {
	var ms = runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	s.updateGaugeMetrics(&ms)
	s.updatePollCounter()
}

func (s *Service) isEmptyStore() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.store) == 0
}

func (s *Service) updateGaugeMetrics(ms *runtime.MemStats) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("start update gauges...")
	s.store["Alloc"] = Metric{Type: core.Gauge, Key: "Alloc", Value: strconv.FormatUint(ms.Alloc, 10)}
	s.store["BuckHashSys"] = Metric{Type: core.Gauge, Key: "BuckHashSys", Value: strconv.FormatUint(ms.BuckHashSys, 10)}
	s.store["Frees"] = Metric{Type: core.Gauge, Key: "Frees", Value: strconv.FormatUint(ms.Frees, 10)}
	s.store["GCCPUFraction"] = Metric{Type: core.Gauge, Key: "GCCPUFraction", Value: strconv.FormatFloat(ms.GCCPUFraction, 'f', -1, 64)}
	s.store["GCSys"] = Metric{Type: core.Gauge, Key: "GCSys", Value: strconv.FormatUint(ms.GCSys, 10)}
	s.store["HeapAlloc"] = Metric{Type: core.Gauge, Key: "HeapAlloc", Value: strconv.FormatUint(ms.HeapAlloc, 10)}
	s.store["HeapIdle"] = Metric{Type: core.Gauge, Key: "HeapIdle", Value: strconv.FormatUint(ms.HeapIdle, 10)}
	s.store["HeapInuse"] = Metric{Type: core.Gauge, Key: "HeapInuse", Value: strconv.FormatUint(ms.HeapInuse, 10)}
	s.store["HeapReleased"] = Metric{Type: core.Gauge, Key: "HeapReleased", Value: strconv.FormatUint(ms.HeapReleased, 10)}
	s.store["HeapObjects"] = Metric{Type: core.Gauge, Key: "HeapObjects", Value: strconv.FormatUint(ms.HeapObjects, 10)}
	s.store["HeapSys"] = Metric{Type: core.Gauge, Key: "HeapSys", Value: strconv.FormatUint(ms.HeapSys, 10)}
	s.store["LastGC"] = Metric{Type: core.Gauge, Key: "LastGC", Value: strconv.FormatUint(ms.LastGC, 10)}
	s.store["Lookups"] = Metric{Type: core.Gauge, Key: "Lookups", Value: strconv.FormatUint(ms.Lookups, 10)}
	s.store["MCacheInuse"] = Metric{Type: core.Gauge, Key: "MCacheInuse", Value: strconv.FormatUint(ms.MCacheInuse, 10)}
	s.store["MCacheSys"] = Metric{Type: core.Gauge, Key: "MCacheSys", Value: strconv.FormatUint(ms.MCacheSys, 10)}
	s.store["MSpanInuse"] = Metric{Type: core.Gauge, Key: "MSpanInuse", Value: strconv.FormatUint(ms.MSpanInuse, 10)}
	s.store["MSpanSys"] = Metric{Type: core.Gauge, Key: "MSpanSys", Value: strconv.FormatUint(ms.MSpanSys, 10)}
	s.store["Mallocs"] = Metric{Type: core.Gauge, Key: "Mallocs", Value: strconv.FormatUint(ms.Mallocs, 10)}
	s.store["NextGC"] = Metric{Type: core.Gauge, Key: "NextGC", Value: strconv.FormatUint(ms.NextGC, 10)}
	s.store["NumForcedGC"] = Metric{Type: core.Gauge, Key: "NumForcedGC", Value: strconv.FormatUint(uint64(ms.NumForcedGC), 10)}
	s.store["NumGC"] = Metric{Type: core.Gauge, Key: "NumGC", Value: strconv.FormatUint(uint64(ms.NumGC), 10)}
	s.store["OtherSys"] = Metric{Type: core.Gauge, Key: "OtherSys", Value: strconv.FormatUint(ms.OtherSys, 10)}
	s.store["PauseTotalNs"] = Metric{Type: core.Gauge, Key: "PauseTotalNs", Value: strconv.FormatUint(ms.PauseTotalNs, 10)}
	s.store["StackInuse"] = Metric{Type: core.Gauge, Key: "StackInuse", Value: strconv.FormatUint(ms.StackInuse, 10)}
	s.store["StackSys"] = Metric{Type: core.Gauge, Key: "StackSys", Value: strconv.FormatUint(ms.StackSys, 10)}
	s.store["Sys"] = Metric{Type: core.Gauge, Key: "Sys", Value: strconv.FormatUint(ms.Sys, 10)}
	s.store["TotalAlloc"] = Metric{Type: core.Gauge, Key: "TotalAlloc", Value: strconv.FormatUint(ms.TotalAlloc, 10)}
	s.store["RandomValue"] = Metric{Type: core.Gauge, Key: "RandomValue", Value: strconv.FormatFloat(rand.Float64(), 'f', -1, 64)}
	fmt.Println("end update gauges")

}

func (s *Service) updatePollCounter() {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := "PollCount"

	counterStr, ok := s.store[key]
	//fmt.Println("start update poll counter, before", counterStr)
	if !ok {
		s.store[key] = Metric{Type: core.Counter, Key: key, Value: strconv.FormatInt(0, 10)}
		counterStr = s.store[key]
	}
	counter, _ := strconv.ParseInt(counterStr.Value, 10, 64)
	s.store[key] = Metric{Type: core.Counter, Key: key, Value: strconv.FormatInt(counter+1, 10)}
	//fmt.Println("start update poll counter, after", s.store[key].Value)
}

func (s *Service) resetPollCounter() {
	s.mu.Lock()
	//fmt.Println("reset poll counter, before :: ", s.store["PollCount"])
	s.store["PollCount"] = Metric{Type: core.Counter, Key: "PollCount", Value: strconv.FormatInt(0, 10)}
	//fmt.Println("reset poll counter, after :: ", s.store["PollCount"])
	s.mu.Unlock()
	//fmt.Println("reset poll counter")
}
