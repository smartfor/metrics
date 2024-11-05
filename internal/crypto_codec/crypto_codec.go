package crypto_codec

import (
	"fmt"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
)

func MakeCryptoCodec() CryptoCodec {
	return CryptoCodec{}
}

// CryptoCodec - кастомный codec, расширяющий стандартный proto codec
type CryptoCodec struct {
	encoding.Codec
}

func (c CryptoCodec) Marshal(v interface{}) ([]byte, error) {
	// Здесь можно добавить свою логику перед сериализацией
	fmt.Println("Custom Marshal")
	return proto.Marshal(v.(proto.Message))
}

func (c CryptoCodec) Unmarshal(data []byte, v interface{}) error {
	// Здесь можно добавить свою логику перед десериализацией
	fmt.Println("Custom Unmarshal")
	return proto.Unmarshal(data, v.(proto.Message))
}

func (c CryptoCodec) Name() string {
	return "my-custom-codec"
}
