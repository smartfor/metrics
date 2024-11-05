package metric_sender

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/go-resty/resty/v2"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/crypto"
	"github.com/smartfor/metrics/internal/ip"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/utils"
)

var UpdateBatchURL string = "/updates/"

type HttpMetricSender struct {
	client    *resty.Client
	realIP    string
	publicKey []byte
	secret    string
}

func NewHttpMetricSender(cfg *config.Config, publicKey []byte) (MetricSender, error) {
	client := resty.
		New().
		SetBaseURL(cfg.HostEndpoint).
		SetHeader("Content-Type", "application/json").
		SetTimeout(cfg.ResponseTimeoutDuration)

	realIP, err := ip.GetExternalIP()
	if err != nil {
		return nil, err
	}

	return &HttpMetricSender{
		client:    client,
		realIP:    realIP,
		secret:    cfg.Secret,
		publicKey: publicKey,
	}, nil
}

var _ MetricSender = &HttpMetricSender{}

func (s *HttpMetricSender) Send(batch []metrics.Metrics) error {
	var (
		err        error
		body       []byte
		key        []byte
		compressed []byte
		sign       hash.Hash
		hexHash    string
	)

	if body, err = json.Marshal(batch); err != nil {
		fmt.Println("Marshalling batch error: ", err)
		return err
	}

	if s.secret != "" {
		sign = utils.Sign(body, s.secret)
		hexHash = hex.EncodeToString(sign.Sum(nil))
	}

	if s.publicKey != nil {
		body, key, err = crypto.EncryptWithPublicKey(body, s.publicKey)
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
			SetHeader("X-Real-IP", s.realIP).
			SetBody(compressed)

		if s.publicKey != nil {
			r = r.SetHeader(utils.CryptoKey, hex.EncodeToString(key))
		}

		if s.secret != "" {
			r = r.SetHeader(utils.AuthHeaderName, hexHash)
		}

		return r.Post(UpdateBatchURL)
	}, nil)

	return nil
}
