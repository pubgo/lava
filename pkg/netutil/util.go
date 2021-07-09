package netutil

import (
	"github.com/pubgo/xerror"

	"net"
)

// LocalIP gets the first NIC's IP address.
func LocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if nil != err {
		return "", err
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if nil != ipnet.IP.To4() {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", xerror.Fmt("can't get local IP")
}
