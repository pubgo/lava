package netutil

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/pubgo/xerror"
)

type sockOpts struct {
	path        string
	cfgCallback func(lc *net.ListenConfig)
}

// SockOpt sets up socket file's creating option
type SockOpt func(opts *sockOpts) error

func Listen(address string, opts ...SockOpt) (_ net.Listener, err error) {
	defer xerror.RespErr(&err)

	if !strings.Contains(address, "//") {
		address = "tcp4://" + address
	}

	uri := xerror.PanicErr(url.Parse(address)).(*url.URL)
	if uri.Scheme == "" {
		uri.Scheme = "tcp4"
	}

	var sOpts sockOpts
	for i := range opts {
		xerror.Panic(opts[i](&sOpts))
	}

	var lc net.ListenConfig
	if cb := sOpts.cfgCallback; cb != nil {
		cb(&lc)
	}

	return lc.Listen(context.Background(), uri.Scheme, uri.Host)
}

func ListenPacket(address string, opts ...SockOpt) (_ net.PacketConn, err error) {
	defer xerror.RespErr(&err)

	uri, err := url.Parse(address)
	xerror.Panic(err)
	if uri.Scheme == "" {
		uri.Scheme = "udp4"
	}

	var sOpts sockOpts
	for i := range opts {
		xerror.Panic(opts[i](&sOpts))
	}

	var lc net.ListenConfig
	if cb := sOpts.cfgCallback; cb != nil {
		cb(&lc)
	}

	return lc.ListenPacket(context.Background(), uri.Scheme, uri.Host)
}

// NewTCPSocket creates a TCP socket listener with the specified address and
// the specified tls configuration. If tlsConfig is set, will encapsulate the
// TCP listener inside a TLS one.
func NewTCPSocket(addr string, tlsConfig *tls.Config) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		tlsConfig.NextProtos = []string{"http/1.1"}
		l = tls.NewListener(l, tlsConfig)
	}
	return l, nil
}

// WithChown modifies the socket file's uid and gid
func WithChown(uid, gid int) SockOpt {
	return func(opts *sockOpts) error {
		if err := os.Chown(opts.path, uid, gid); err != nil {
			return err
		}
		return nil
	}
}

func WithNetCfg(fn func(lc *net.ListenConfig)) SockOpt {
	return func(opts *sockOpts) error {
		opts.cfgCallback = fn
		return nil
	}
}

// WithChmod modifies socket file's access mode
func WithChmod(mask os.FileMode) SockOpt {
	return func(opts *sockOpts) error {
		if err := os.Chmod(opts.path, mask); err != nil {
			return err
		}
		return nil
	}
}

func MustGetPort(addrOrNet interface{}) int {
	return xerror.PanicErr(GetPort(addrOrNet)).(int)
}

// GetPort returns the port of an endpoint address.
func GetPort(addrOrNet interface{}) (int, error) {
	var addr string
	switch addrNet := addrOrNet.(type) {
	case net.Addr:
		addr = addrNet.String()
	case string:
		addr = addrNet
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return -1, err
	}

	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		return -1, err
	}

	return parsedPort, nil
}

// NewUnixSocketWithOpts creates a unix socket with the specified options
//func NewUnixSocketWithOpts(path string, opts ...SockOpt) (net.Listener, error) {
//	if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
//		return nil, err
//	}
//	mask := syscall.Umask(0777)
//	defer syscall.Umask(mask)
//
//	l, err := net.Listen("unix", path)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, op := range opts {
//		if err := op(path); err != nil {
//			l.Close()
//			return nil, err
//		}
//	}
//
//	return l, nil
//}

// NewUnixSocket creates a unix socket with the specified path and group.
//func NewUnixSocket(path string, gid int) (net.Listener, error) {
//	return NewUnixSocketWithOpts(path, WithChown(0, gid), WithChmod(0660))
//}
