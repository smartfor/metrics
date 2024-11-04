package metric_sender

import (
	"context"

	"github.com/smartfor/metrics/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

// Прокси-Codec, который передаёт заранее сериализованные данные
type proxyCodec struct {
	encoding.Codec
	serializedData []byte
}

func (c *proxyCodec) Marshal(v interface{}) ([]byte, error) {
	// Возвращаем заранее сериализованные данные вместо повторной сериализации
	return c.serializedData, nil
}

func (c *proxyCodec) Unmarshal(data []byte, v interface{}) error {
	// Используем стандартный Unmarshal
	return c.Codec.Unmarshal(data, v)
}

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
		//body, ok := req.(*metricapi.UpdateRequest)
		//if ok {
		//	return status.Errorf(codes.InvalidArgument, "bad request")
		//}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
