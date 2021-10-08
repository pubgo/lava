package netutil

import (
	"errors"
	"net"
	"time"

	"github.com/pubgo/xerror"
)

// LocalIP gets the first NIC's IP address.
func LocalIP() (string, error) {
	addrList, err := net.InterfaceAddrs()

	if nil != err {
		return "", xerror.Wrap(err)
	}

	for _, address := range addrList {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", xerror.Fmt("can't get local IP")
}

// CheckPort 检查端口是否被占用
func CheckPort(protocol string, addr string) bool {
	conn, err := net.DialTimeout(protocol, addr, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// DiscoverDNS ...
func DiscoverDNS(service, proto string, address string) ([]*net.SRV, error) {
	_, addresses, err := net.LookupSRV(service, proto, address)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, errors.New("discovery: srv lookup nothing")
	}

	return addresses, nil
}
