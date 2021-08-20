package netutil

import (
	"net"
	"time"

	"github.com/pubgo/xerror"
)

// LocalIP gets the first NIC's IP address.
func LocalIP() (string, error) {
	addrList, err := net.InterfaceAddrs()

	if nil != err {
		return "", err
	}

	for _, address := range addrList {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if nil != ipnet.IP.To4() {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", xerror.Fmt("can't get local IP")
}

func ScanPort(protocol string, addr string) bool {
	conn, err := net.DialTimeout(protocol, addr, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
