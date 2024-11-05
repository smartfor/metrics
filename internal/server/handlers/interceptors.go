package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/smartfor/metrics/api/metricapi"
	"github.com/smartfor/metrics/internal/crypto"
	"github.com/smartfor/metrics/internal/server/config"
	"github.com/smartfor/metrics/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxMetricsKey string

const ctxMetrics ctxMetricsKey = "metricsAsBytes"

func MakeGrpcAuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cfg.Secret == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unknown, "Empty Metadata")
		}

		fmt.Println("md : ", md)

		authHeader := md.Get(utils.AuthHeaderName)
		fmt.Println(authHeader)
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

		metrics := ctx.Value(ctxMetrics)
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
			ctx = context.WithValue(ctx, ctxMetrics, bodyBytes)
		} else {
			bodyBytes = metrics.([]byte)
		}

		// проверяем хеш
		if !utils.Verify(cfg.Secret, string(hashBytes), bodyBytes) {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid hash")
		}

		return handler(ctx, req)
	}
}

func MakeGrpcCryptoInterceptor(privateKey []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if privateKey == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unknown, "Empty Metadata")
		}

		cryptoKey := md.Get(utils.CryptoKey)
		if len(cryptoKey) < 1 {
			return nil, status.Errorf(codes.Unauthenticated, "Empty crypto key")
		}

		// ключ содержит слайс строк, получаем первую строку
		hexKey := cryptoKey[0]
		// декодируем ключ
		key, err := hex.DecodeString(hexKey)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid crypto key")
		}

		var bodyBytes []byte

		metrics := ctx.Value(ctxMetrics)
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
			ctx = context.WithValue(ctx, ctxMetrics, bodyBytes)
		} else {
			bodyBytes = metrics.([]byte)
		}

		// Decrypt the message using the private key
		decodedMessage, err := crypto.DecryptWithPrivateKey(bodyBytes, key, privateKey)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Decryption error")
		}

		return handler(ctx, decodedMessage)
	}
}
