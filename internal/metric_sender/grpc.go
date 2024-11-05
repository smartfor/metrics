package metric_sender

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/smartfor/metrics/api/metricapi"
	"github.com/smartfor/metrics/internal/config"
	crypto_codec "github.com/smartfor/metrics/internal/crypto-codec"
	"github.com/smartfor/metrics/internal/ip"
	"github.com/smartfor/metrics/internal/metrics"
	"github.com/smartfor/metrics/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcMetricSender struct {
	client    metricapi.MetricsClient
	realIP    string
	publicKey []byte
	secret    string
}

func NewGrpcMetricSender(cfg *config.Config, publicKey []byte) (MetricSender, error) {
	backoffConfig := backoff.Config{
		BaseDelay:  1 * time.Second,   // Начальная задержка
		Multiplier: 1.6,               // Множитель экспоненциального увеличения
		MaxDelay:   120 * time.Second, // Максимальная задержка между попытками
		Jitter:     0.2,               // Процент случайного разброса
	}

	conn, err := grpc.NewClient(
		cfg.HostEndpoint,
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoffConfig,
			MinConnectTimeout: 5 * time.Second, // Минимальное время ожидания соединения
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(crypto_codec.MakeCryptoCodec()),
			grpc.UseCompressor(gzip.Name),
		),
		grpc.WithChainUnaryInterceptor(
			MakeClientAuthInterceptor(cfg),
			MakeClientCryptoInterceptor(cfg, publicKey),
		),
	)
	if err != nil {
		return nil, err
	}

	client := metricapi.NewMetricsClient(conn)

	realIP, err := ip.GetExternalIP()
	if err != nil {
		return nil, err
	}

	return &GrpcMetricSender{
		client:    client,
		realIP:    realIP,
		publicKey: publicKey,
		secret:    cfg.Secret,
	}, nil
}

var _ MetricSender = &GrpcMetricSender{}

func (s *GrpcMetricSender) Send(batch []metrics.Metrics) error {
	md := make(map[string]string)

	outBatch := make([]*metricapi.Metric, 0)
	for _, m := range batch {
		outMetric := &metricapi.Metric{
			Id:    m.ID,
			Mtype: m.MType,
		}

		if m.Value != nil {
			outMetric.Value = *m.Value
		}

		if m.Delta != nil {
			outMetric.Delta = *m.Delta
		}

		outBatch = append(outBatch, outMetric)
	}

	if s.publicKey != nil {
		md[utils.CryptoKey] = hex.EncodeToString(s.publicKey)
	}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))

	_, err := utils.Retry(func() (*emptypb.Empty, error) {
		_, err := s.client.Update(ctx, &metricapi.UpdateRequest{
			Metrics: outBatch,
		})
		if err != nil {
			return nil, err
		}

		return &emptypb.Empty{}, nil
	}, nil)

	if err != nil {
		fmt.Println("Sending batch error:	", err)
		return err
	}

	fmt.Println("Batch sent successfully")
	return nil
}
