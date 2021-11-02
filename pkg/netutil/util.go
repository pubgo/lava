package netutil

import (
	"errors"
	"io"
	"net"
	"regexp"
	"strings"
	"syscall"
	"time"
)

func GetLocalIP() string {
	localIP := "localhost"

	// skip the error since we don't want to break RPC calls because of it
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return localIP
	}

	for _, addr := range addresses {
		items := strings.Split(addr.String(), "/")
		if len(items) < 2 || items[0] == "127.0.0.1" {
			continue
		}

		if match, err := regexp.MatchString(`\d+\.\d+\.\d+\.\d+`, items[0]); err == nil && match {
			localIP = items[0]
		}
	}

	return localIP
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

var errUnexpectedRead = errors.New("unexpected read from socket")

func ConnCheck(conn net.Conn) error {
	var sysErr error

	sysConn, ok := conn.(syscall.Conn)
	if !ok {
		return nil
	}

	rawConn, err := sysConn.SyscallConn()
	if err != nil {
		return err
	}

	err = rawConn.Read(func(fd uintptr) bool {
		var buf [1]byte
		n, err := syscall.Read(int(fd), buf[:])
		switch {
		case n == 0 && err == nil:
			sysErr = io.EOF
		case n > 0:
			sysErr = errUnexpectedRead
		case err == syscall.EAGAIN || err == syscall.EWOULDBLOCK:
			sysErr = nil
		default:
			sysErr = err
		}

		return true
	})
	if err != nil {
		return err
	}

	return sysErr
}
