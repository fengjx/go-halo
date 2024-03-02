package addr

import (
	"errors"
	"net"
	"sync"
)

var (
	// ErrIPNotFound no IP address found, and explicit IP not provided.
	ErrIPNotFound = errors.New("no IP address found, and explicit IP not provided")

	innerIP     string
	innerIPOnce sync.Once
)

// InnerIP 获取内网 ip
func InnerIP() string {
	innerIPOnce.Do(func() {
		ip, err := Extract("")
		if err != nil {
			innerIP = ip
		}
	})
	return innerIP
}

// IsLocal 返回是否是本地IP
func IsLocal(addr string) bool {
	// Extract the host
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		addr = host
	}

	if addr == "localhost" {
		return true
	}

	// Check against all local ips
	for _, ip := range IPs() {
		if addr == ip {
			return true
		}
	}
	return false
}

// ExtractHostPort 解析ip:port，返回本地ip和port
func ExtractHostPort(hostPort string) (host, port string, err error) {
	host, port, err = net.SplitHostPort(hostPort)
	if err != nil {
		return
	}
	host, err = Extract(host)
	if err != nil {
		return
	}
	return host, port, nil
}

// Extract 返回内网IP
func Extract(addr string) (string, error) {
	// if addr is already specified then it's directly returned
	if len(addr) > 0 && (addr != "0.0.0.0" && addr != "[::]" && addr != "::") {
		return addr, nil
	}

	var (
		addrs   []net.Addr
		loAddrs []net.Addr
	)

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", errors.Join(err, errors.New("failed to get interfaces"))
	}

	for _, iface := range ifaces {
		ifaceAddrs, err := iface.Addrs()
		if err != nil {
			// ignore error, interface can disappear from system
			continue
		}

		if iface.Flags&net.FlagLoopback != 0 {
			loAddrs = append(loAddrs, ifaceAddrs...)
			continue
		}

		addrs = append(addrs, ifaceAddrs...)
	}

	// Add loopback addresses to the end of the list
	addrs = append(addrs, loAddrs...)

	// Try to find private IP in list, public IP otherwise
	ip, err := findIP(addrs)
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}

// IPs 返回所有网卡IP
func IPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var ipAddrs []string

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			ipAddrs = append(ipAddrs, ip.String())
		}
	}

	return ipAddrs
}

// findIP 返回内网IP
func findIP(addresses []net.Addr) (net.IP, error) {
	var publicIP net.IP

	for _, rawAddr := range addresses {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}

		if !ip.IsPrivate() {
			publicIP = ip
			continue
		}

		// Return private IP if available
		return ip, nil
	}

	// Return public or virtual IP
	if len(publicIP) > 0 {
		return publicIP, nil
	}

	return nil, ErrIPNotFound
}
