package handlers

import (
	"context"
	"errors"

	"github.com/smartfor/metrics/api/metricapi"
	"github.com/smartfor/metrics/internal/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MetricsServer struct {
	metricapi.UnimplementedMetricsServer

	storage       core.Storage
	logger        *zap.Logger
	secret        string
	cryptoKey     []byte
	trustedSubnet string
}

func NewGRPCServer(s core.Storage, logger *zap.Logger, secret string, cryptoKey []byte, trustedSubnet string) *MetricsServer {
	return &MetricsServer{
		storage:       s,
		logger:        logger,
		secret:        secret,
		cryptoKey:     cryptoKey,
		trustedSubnet: trustedSubnet,
	}
}

func (s *MetricsServer) Update(ctx context.Context, req *metricapi.UpdateRequest) (*emptypb.Empty, error) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	for _, m := range req.GetMetrics() {
		switch core.NewMetricType(m.Mtype) {
		case core.Gauge:
			gauges[m.Id] = m.Value
		case core.Counter:
			v, ok := counters[m.Id]
			if !ok {
				v = 0
			}
			counters[m.Id] = v + m.Delta
		default:
			return nil, core.ErrUnknownMetricType
		}
	}

	batch := core.NewBaseMetricStorageWithValues(gauges, counters)
	if err := s.storage.SetBatch(ctx, batch); err != nil {
		if errors.Is(err, core.ErrBadMetricValue) {
			return nil, status.Errorf(codes.InvalidArgument, "bad metric value")
		}

		return nil, err
	}

	return nil, nil
}
