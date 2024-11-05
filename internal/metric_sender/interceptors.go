package metric_sender

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/smartfor/metrics/api/metricapi"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func MakeClientInterceptor(cfg *config.Config) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		meta := map[string]string{}

		if cfg.Secret != "" {
			body, ok := req.(*metricapi.UpdateRequest)
			if !ok {
				return fmt.Errorf("invalid request type")
			}

			asJson, err := json.Marshal(body)
			if err != nil {
				return err
			}

			sign := utils.Sign(asJson, cfg.Secret)
			hexHash := hex.EncodeToString(sign.Sum(nil))
			meta[utils.AuthHeaderName] = hexHash
		}

		ctx = metadata.NewOutgoingContext(ctx, metadata.New(meta))
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
