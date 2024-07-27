package internal

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/polling"
	"github.com/smartfor/metrics/internal/utils"
)

var UpdateBatchURL string = "/updates/"

type Metric = polling.MetricsModel

type Job struct {
	Store       polling.MetricStore
	PoolCounter int64
}

type JobResult struct {
	Err         error
	PoolCounter int64
}

type Service struct {
	config      config.Config
	client      *resty.Client
	mu          *sync.Mutex
	pollCounter atomic.Int64
}

func NewService(cfg *config.Config) Service {
	client := resty.
		New().
		SetBaseURL(cfg.HostEndpoint).
		SetHeader("Content-Type", "application/json").
		SetTimeout(cfg.ResponseTimeout)

	return Service{
		config: *cfg,
		client: client,
		mu:     &sync.Mutex{},
	}
}

func (s *Service) Run(ctx context.Context) {
	var (
		mainPollCh     = polling.CreateMainPollChannel(ctx, s.config.PollInterval)
		advancedPollCh = polling.CreateAdvancedPollChannel(ctx, s.config.PollInterval)
		fanIn          = polling.FanInPolling(ctx, mainPollCh, advancedPollCh)
		jobs           = make(chan Job, s.config.RateLimit)
		results        = make(chan JobResult, s.config.RateLimit)
		messages       = make([]polling.PollMessage, 0, 1024)
		ticker         = time.NewTicker(s.config.ReportInterval)
	)

	for w := 0; w <= s.config.RateLimit; w++ {
		go s.worker(jobs, results)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-fanIn:
			if msg.Err != nil {
				continue
			}

			messages = append(messages, msg)
		case result := <-results:
			if result.Err != nil {
				s.pollCounter.Store(result.PoolCounter)
				continue
			}
		case <-ticker.C:
			if len(messages) == 0 {
				continue
			}

			slices.Reverse(messages)
			var (
				store       = make(polling.MetricStore)
				hasMain     bool
				hasAdvanced bool
			)

			for _, m := range messages {
				if m.Type == polling.PollMainMetricsType {
					for k, v := range m.Msg {
						store[k] = v
					}
					hasMain = true
				}

				if m.Type == polling.PollAdvancedMetricsType {
					for k, v := range m.Msg {
						store[k] = v
					}
					hasAdvanced = true
				}
			}

			if !hasMain || !hasAdvanced {
				continue
			}

			messages = messages[:0]
			counter := s.pollCounter.Load()
			s.pollCounter.Store(0)

			jobs <- Job{
				Store:       store,
				PoolCounter: counter,
			}
		}
	}
}

func (s *Service) worker(jobs <-chan Job, results chan<- JobResult) {
	for j := range jobs {
		if err := s.send(j.Store, j.PoolCounter); err != nil {
			results <- JobResult{Err: err, PoolCounter: j.PoolCounter}
		}

		results <- JobResult{Err: nil, PoolCounter: j.PoolCounter}
	}
}

func (s *Service) send(store polling.MetricStore, pollCounter int64) error {
	var (
		batch      []metrics.Metrics
		err        error
		body       []byte
		compressed []byte
		sign       hash.Hash
		hexHash    string
	)

	store["PoolCount"] = polling.MetricsModel{
		Type:  core.Gauge,
		Key:   "PoolCount",
		Value: strconv.FormatInt(pollCounter, 10),
	}

	for _, v := range store {
		metric, err := metrics.FromMetricModel(v)
		if err != nil {
			fmt.Println("Extract metric from model error: ", err)
			return err
		}
		batch = append(batch, *metric)
	}

	if body, err = json.Marshal(batch); err != nil {
		fmt.Println("Marshalling batch error: ", err)
		return err
	}

	if s.config.Secret != "" {
		sign = utils.Sign(body, s.config.Secret)
		hexHash = hex.EncodeToString(sign.Sum(nil))
	}

	if compressed, err = utils.GzipCompress(body); err != nil {
		fmt.Println("Compressed body error: ", err)
		return err
	}

	_, err = utils.Retry(func() (*resty.Response, error) {
		r := s.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept-Encoding", "gzip").
			SetHeader("Content-Encoding", "gzip").
			SetBody(compressed)

		if s.config.Secret != "" {
			r = r.SetHeader(utils.AuthHeaderName, hexHash)
		}

		return r.Post(UpdateBatchURL)
	}, nil)
	if err != nil {
		return err
	}

	return nil
}
