package psutil

import (
	"net"
	"strings"
	"sync"

	"github.com/rehiy/libgo/request"
)

// 公网 IP 缓存
var (
	publicIPv4   string
	publicIPv6   string
	publicAddrMu sync.RWMutex
)

// PublicAddress 获取公网 IP 地址
func PublicAddress(force bool) (string, string) {
	publicAddrMu.RLock()
	v4, v6 := publicIPv4, publicIPv6
	publicAddrMu.RUnlock()

	if !force && v4 != "" && v6 != "" {
		return v4, v6
	}

	newV4 := strings.TrimSpace(request.TimingGet("http://ipv4.rehi.org/ip", request.H{}, 10))
	newV6 := strings.TrimSpace(request.TimingGet("http://ipv6.rehi.org/ip", request.H{}, 10))

	publicAddrMu.Lock()
	defer publicAddrMu.Unlock()

	// 只更新非空值
	if newV4 != "" {
		publicIPv4 = newV4
	}
	if newV6 != "" {
		publicIPv6 = newV6
	}

	return publicIPv4, publicIPv6
}

// InterfaceAddrs 获取网卡 IP 地址列表
func InterfaceAddrs(name string) ([]string, []string) {
	var addrs []net.Addr

	if name == "" {
		addrs, _ = net.InterfaceAddrs()
	} else if ift, err := net.InterfaceByName(name); err == nil {
		addrs, _ = ift.Addrs()
	}

	if len(addrs) == 0 {
		return nil, nil
	}

	// 预分配容量
	ipv4 := make([]string, 0, len(addrs))
	ipv6 := make([]string, 0, len(addrs))

	for _, ip := range addrs {
		if ipnet, ok := ip.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil {
				ipv4 = append(ipv4, ipnet.IP.String())
			} else {
				ipv6 = append(ipv6, ipnet.IP.String())
			}
		}
	}

	return ipv4, ipv6
}
