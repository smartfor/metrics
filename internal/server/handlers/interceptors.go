package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/smartfor/metrics/api/metricapi"
	"github.com/smartfor/metrics/internal/server/config"
	"github.com/smartfor/metrics/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ctxMetricsKey = "metricsAsBytes"

func MakeGrpcAuthInterceptor(cfg *config.Config) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cfg.Secret != "" {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unknown, "Empty Metadata")
			}

			authHeader := md.Get(utils.AuthHeaderName)
			if len(authHeader) < 1 {
				// разорвать соединение при отсутствии хеша
				return nil, status.Errorf(codes.Unauthenticated, "Empty hash")
			}

			fmt.Println("authHeader", authHeader)

			// ключ содержит слайс строк, получаем первую строку
			hexHash := authHeader[0]
			// декодируем хеш
			hashBytes, err := hex.DecodeString(hexHash)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid hash")
			}

			var bodyBytes []byte

			metrics := ctx.Value(ctxMetricsKey)
			if metrics == nil {
				metrics, ok = (req).(*metricapi.UpdateRequest)
				if !ok {
					return nil, fmt.Errorf("invalid request type")
				}

				bodyBytes, err = json.Marshal(metrics)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "Marshalling error")
				}

				// кэшируем в контексте чтобы другим перехватчикам не нужно было заново маршалить весь запрос
				ctx = context.WithValue(ctx, ctxMetricsKey, bodyBytes)
			}

			// проверяем хеш
			if !utils.Verify(cfg.Secret, string(hashBytes), bodyBytes) {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid hash")
			}
		}

		return handler(ctx, req)
	}
}
