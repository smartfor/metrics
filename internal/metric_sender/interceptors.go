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

type ctxMetricsType string

const ctxMetricsKey ctxMetricsType = "metricsAsBytes"

func MakeClientAuthInterceptor(cfg *config.Config) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if cfg.Secret == "" {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		var (
			meta      = make(map[string]string)
			bodyBytes []byte
			cache     any
			err       error
		)

		cache = ctx.Value(ctxMetricsKey)
		if cache == nil {
			body, ok := req.(*metricapi.UpdateRequest)
			if !ok {
				return fmt.Errorf("invalid request type")
			}

			cache, err = json.Marshal(body)
			if err != nil {
				return err
			}

			// кэшируем в контексте запрос для последующего использования в других интерцепторах
			ctx = context.WithValue(ctx, ctxMetricsKey, body)
		}
		bodyBytes = cache.([]byte)

		sign := utils.Sign(bodyBytes, cfg.Secret)
		hexHash := hex.EncodeToString(sign.Sum(nil))

		meta[utils.AuthHeaderName] = hexHash

		ctx = metadata.AppendToOutgoingContext(ctx, utils.AuthHeaderName, hexHash)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func MakeClientCryptoInterceptor(cfg *config.Config, publicKey []byte) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if cfg.CryptoKey != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, utils.CryptoKey, hex.EncodeToString(publicKey))
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
