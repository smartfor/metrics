package ip

import (
	"io"
	"net"
	"net/http"
)

const (
	// ExternalIPURL - URL для получения внешнего IP-адреса
	ExternalIPURL = "https://ifconfig.me"
)

// GetExternalIP - функция для получения внешнего IP-адреса через внешний API
func GetExternalIP() (string, error) {
	resp, err := http.Get(ExternalIPURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func InSubnet(ip string, subnet string) bool {
	realIP := net.ParseIP(ip)
	if realIP == nil {
		return false
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}

	return ipNet.Contains(realIP)
}

// ExtractXRealIP - функция для извлечения IP-адреса из заголовка X-Real-IP
func ExtractXRealIP(header http.Header) string {
	return header.Get("X-Real-IP")
}
