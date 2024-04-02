package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Addr string
}

func GetConfig() Config {
	addr := flag.String("a", ":8080", "address and port to run server")
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Println("Error: unknown flags:", flag.Args())
		flag.PrintDefaults()
		os.Exit(1)
	}

	if a := os.Getenv("ADDRESS"); a != "" {
		*addr = a
	}

	return Config{Addr: *addr}
}
