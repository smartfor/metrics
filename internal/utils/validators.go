package utils

import (
	"net"
	"time"
)

func ValidateAddress(addr string) error {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	return nil
}

func ValidateDuration(durAsStr string) error {
	_, err := time.ParseDuration(durAsStr)
	if err != nil {
		return err
	}
	return nil
}
