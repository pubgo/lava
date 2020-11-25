// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package grpcurl

var (
	unix *bool
)

func init() {
	isUnixSocket = func() bool {
		return *unix
	}
}
