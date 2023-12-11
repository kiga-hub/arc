package utils

import (
	"math"
	"net"
)

// GetAllIPv4Address 获取所有IP v4地址
func GetAllIPv4Address() ([]net.IP, error) {
	result := []net.IP{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, value := range addrs {
		if ipnet, ok := value.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				result = append(result, ipnet.IP)
			}
		}
	}
	return result, nil
}

// IsKigaOverlayIP check if ip is in the kiga overlay network
func IsKigaOverlayIP(ip net.IP) bool {
	l := len(ip)
	return ip[l-4] == 172 && ip[l-3] == 128
}

// FindNearestIP 找到最近的ip
func FindNearestIP(source []net.IP, target net.IP) net.IP {
	if len(source) == 0 {
		return net.ParseIP("127.0.0.1")
	}
	min := math.MaxInt64
	var result net.IP
	for _, ip := range source {
		diff := 0
		for i := 12; i < 16; i++ {
			diff *= 256
			diff += int(ip[i] ^ target[i])
		}
		if diff < min {
			min = diff
			result = ip
		}
	}
	return result
}
