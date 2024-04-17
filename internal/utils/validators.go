package utils

import "net"

func ValidateAddress(addr string) error {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	return nil
}
