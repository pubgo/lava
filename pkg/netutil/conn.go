package netutil

import (
	"crypto/tls"
	"github.com/pubgo/xerror"
	"net"
	"net/url"
	"os"
	"syscall"
)

type sockOpts struct {
	path        string
	cfgCallback func(lc *net.ListenConfig)
}

// SockOpt sets up socket file's creating option
type SockOpt func(opts *sockOpts) error

var lc net.ListenConfig

func Listen(address string, opts ...SockOpt) (net.Listener, error) {
	url, err := url.Parse(address)
	xerror.Panic(err)

	_ = url
}

// NewTCPSocket creates a TCP socket listener with the specified address and
// the specified tls configuration. If TLSConfig is set, will encapsulate the
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

// NewUnixSocketWithOpts creates a unix socket with the specified options
func NewUnixSocketWithOpts(path string, opts ...SockOpt) (net.Listener, error) {
	if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	mask := syscall.Umask(0777)
	defer syscall.Umask(mask)

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	for _, op := range opts {
		if err := op(path); err != nil {
			l.Close()
			return nil, err
		}
	}

	return l, nil
}

// NewUnixSocket creates a unix socket with the specified path and group.
func NewUnixSocket(path string, gid int) (net.Listener, error) {
	return NewUnixSocketWithOpts(path, WithChown(0, gid), WithChmod(0660))
}
