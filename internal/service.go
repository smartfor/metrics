package internal

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/polling"
	"github.com/smartfor/metrics/internal/utils"
	"hash"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var UpdateBatchURL string = "/updates/"

type PollMessageType int

const (
	PollMainMetricsType     PollMessageType = 0
	PollAdvancedMetricsType PollMessageType = 1
)

type PollMessage struct {
	Msg  polling.MetricStore
	Err  error
	Type PollMessageType
}

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
	fmt.Println("Metrics Agent is started...  :::::::::::::::::: ")

	mainPollCh := createMainPollChannel(ctx, s.config.PollInterval)
	advancedPollCh := createAdvancedPollChannel(ctx, s.config.PollInterval)
	fanIn := fanInPolling(ctx, mainPollCh, advancedPollCh)

	jobs := make(chan Job, s.config.RateLimit)
	results := make(chan JobResult, s.config.RateLimit)

	for w := 0; w <= s.config.RateLimit; w++ {
		go s.worker(jobs, results)
	}

	messages := make([]PollMessage, 0, 1024)
	ticker := time.NewTicker(s.config.ReportInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-fanIn:
			fmt.Println("FanIn received:  :::::::::::::::::: ")

			if msg.Err != nil {
				fmt.Println(" ::::::::::::::::::  Polling error: ", msg.Err)
				continue
			}

			messages = append(messages, msg)
		case result := <-results:
			fmt.Println(" ::::::::::::::::::  Job result: ")
			if result.Err != nil {
				fmt.Println(" ::::::::::::::::::  Job error: ", result.Err)
				s.pollCounter.Store(result.PoolCounter)
				continue
			}
		case <-ticker.C:
			fmt.Println(" ::::::::::::::::::  Ticker tick: ")
			if len(messages) == 0 {
				fmt.Println("No messages to send  :::::::::::::::::: ")
				continue
			}

			fmt.Println("Start sending messages  :::::::::::::::::: ")
			slices.Reverse(messages)
			var (
				store       = make(polling.MetricStore)
				hasMain     bool
				hasAdvanced bool
			)

			for _, m := range messages {
				if m.Type == PollMainMetricsType {
					for k, v := range m.Msg {
						store[k] = v
					}
					hasMain = true
				}

				if m.Type == PollAdvancedMetricsType {
					for k, v := range m.Msg {
						store[k] = v
					}
					hasAdvanced = true
				}
			}

			if !hasMain || !hasAdvanced {
				fmt.Println("No main or advanced metrics  :::::::::::::::::: ")
				continue
			}

			fmt.Println("Reset messages and counter  :::::::::::::::::: ")
			messages = messages[:0]
			counter := s.pollCounter.Load()
			s.pollCounter.Store(0)

			fmt.Println("Send Job  :::::::::::::::::: ")
			jobs <- Job{
				Store:       store,
				PoolCounter: counter,
			}
		}

	}
}

func (s *Service) worker(jobs <-chan Job, results chan<- JobResult) {
	for j := range jobs {
		fmt.Println("started job :::::::::::::::::: ")
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
		fmt.Println("...End send")
		return err
	}

	if s.config.Secret != "" {
		sign = utils.Sign(body, s.config.Secret)
		hexHash = hex.EncodeToString(sign.Sum(nil))
	}

	if compressed, err = utils.GzipCompress(body); err != nil {
		fmt.Println("Compressed body error: ", err)
		fmt.Println("...End send")
		return err
	}

	fmt.Println("start send..")
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
		fmt.Println("End report error: ", err)
		fmt.Println("...End send")
		return err
	}

	return nil
}

func fanInPolling(ctx context.Context, chs ...<-chan PollMessage) <-chan PollMessage {
	var wg sync.WaitGroup
	outCh := make(chan PollMessage, 1024)

	output := func(ch <-chan PollMessage) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-ch:
				if !ok {
					return
				}

				fmt.Println("fanInPolling: :::::::::::::::::::: ")
				outCh <- msg
			}
		}
	}

	wg.Add(len(chs))
	for _, ch := range chs {
		go output(ch)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}

func createMainPollChannel(ctx context.Context, interval time.Duration) <-chan PollMessage {
	ch := make(chan PollMessage)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return

			case <-time.After(interval):
				fmt.Println("Polling main metrics...")
				ch <- PollMessage{
					Msg:  polling.PollMainMetrics(),
					Type: PollMainMetricsType,
				}
			}
		}
	}()

	return ch
}

func createAdvancedPollChannel(ctx context.Context, interval time.Duration) <-chan PollMessage {
	ch := make(chan PollMessage)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return

			case <-time.After(interval):
				fmt.Println("Polling advanced metrics...")
				m, err := polling.PollAdvancedMetrics()
				ch <- PollMessage{
					Msg:  m,
					Type: PollAdvancedMetricsType,
					Err:  err,
				}
			}
		}
	}()

	return ch
}
