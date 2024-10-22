// Package internal содержит всю логику работы с метриками.
package internal

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"math/rand"
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

var ErrAgentClosed = errors.New("agent closed")

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
	client             *resty.Client
	mu                 *sync.Mutex
	config             config.Config
	pollCounter        atomic.Int64
	privateKey         []byte
	inShutdown         atomic.Bool
	activeWorkersCount atomic.Int64
}

func NewService(cfg *config.Config, privateKey []byte) Service {
	client := resty.
		New().
		SetBaseURL(cfg.HostEndpoint).
		SetHeader("Content-Type", "application/json").
		SetTimeout(cfg.ResponseTimeoutDuration)

	return Service{
		config:             *cfg,
		client:             client,
		privateKey:         privateKey,
		mu:                 &sync.Mutex{},
		inShutdown:         atomic.Bool{},
		activeWorkersCount: atomic.Int64{},
	}
}

func (s *Service) Run(ctx context.Context) error {
	var (
		mainPollCh     = polling.CreateMainPollChannel(ctx, s.config.PollIntervalDuration)
		advancedPollCh = polling.CreateAdvancedPollChannel(ctx, s.config.PollIntervalDuration)
		fanIn          = polling.FanInPolling(ctx, mainPollCh, advancedPollCh)
		jobs           = make(chan Job, s.config.RateLimit)
		results        = make(chan JobResult, s.config.RateLimit)
		messages       = make([]polling.PollMessage, 0, 1024)
		ticker         = time.NewTicker(s.config.ReportIntervalDuration)
	)

	for w := 0; w <= s.config.RateLimit; w++ {
		go s.worker(jobs, results)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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
			if s.inShutdown.Load() {
				close(jobs)
				return ErrAgentClosed
			}

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
	s.activeWorkersCount.Add(1)
	defer s.activeWorkersCount.Add(-1)

	for j := range jobs {
		if s.inShutdown.Load() {
			return
		}

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
		key        []byte
		compressed []byte
		sign       hash.Hash
		hexHash    string
		metric     *metrics.Metrics
	)

	store["PoolCount"] = polling.MetricsModel{
		Type:  core.Gauge,
		Key:   "PoolCount",
		Value: strconv.FormatInt(pollCounter, 10),
	}

	for _, v := range store {
		metric, err = metrics.FromMetricModel(v)
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

	if s.privateKey != nil {
		body, key, err = utils.EncryptWithPublicKey(body, s.privateKey)
		if err != nil {
			fmt.Println("Encryption error: ", err)
			return err
		}
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

		if s.privateKey != nil {
			r = r.SetHeader(utils.CryptoKey, hex.EncodeToString(key))
		}

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

func (s *Service) Shutdown(ctx context.Context) error {
	s.inShutdown.Store(true)
	defer s.inShutdown.Store(false)

	shutdownPollIntervalMax := 10 * time.Second
	pollIntervalBase := time.Millisecond
	nextPollInterval := func() time.Duration {
		// Add 10% jitter.
		interval := pollIntervalBase + time.Duration(rand.Intn(int(pollIntervalBase/10)))
		// Double and clamp for next time.
		pollIntervalBase *= 2
		if pollIntervalBase > shutdownPollIntervalMax {
			pollIntervalBase = shutdownPollIntervalMax
		}
		return interval
	}

	timer := time.NewTimer(nextPollInterval())
	defer timer.Stop()
	for {
		if s.activeWorkersCount.Load() == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			timer.Reset(nextPollInterval())
		}
	}
}
