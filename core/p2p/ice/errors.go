package ice

import "errors"

var (
	ErrFailed       = errors.New("ice: connection failed")
	errNilConn      = errors.New("ice: nil connection")
	errNoRemoteAddr = errors.New("ice: no remote address")
	errAddrMismatch = errors.New("ice: address mismatch")
)
