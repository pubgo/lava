package cluster

import (
	"github.com/hashicorp/go-sockaddr"
	"github.com/pkg/errors"
	"net"
)

// OversizeMessage indicates whether or not the byte payload should be sent
// via TCP.
func OversizeMessage(b []byte) bool {
	return len(b) > MaxGossipPacketSize/2
}

////////////////////////////////////////////////////////////////

// hostToIP converts host to an IP4 address based on net.LookupIP().
func hostToIP(host string) string {
	// if host is not an IP addr, check net.LookupIP()
	if net.ParseIP(host) == nil {
		hosts, err := net.LookupIP(host)
		if err != nil {
			return host
		}
		for _, h := range hosts {
			if h.To4() != nil {
				return h.String()
			}
		}
	}
	return host
}

// discoverAdvertiseAddress will attempt to get a single IP address to use as
// the advertise address when one is not explicitly provided. It defaults to
// using a private IP address, and if not found then using a public IP if
// insecure advertising is allowed.
func discoverAdvertiseAddress(allowInsecureAdvertise bool) (net.IP, error) {
	addr, err := sockaddr.GetPrivateIP()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get private IP")
	}

	if addr == "" && !allowInsecureAdvertise {
		return nil, errors.New("no private IP found, explicit advertise addr not provided")
	}

	if addr == "" {
		addr, err = sockaddr.GetPublicIP()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get public IP")
		}
		if addr == "" {
			return nil, errors.New("no private/public IP found, explicit advertise addr not provided")
		}
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, errors.Errorf("failed to parse discovered IP '%s'", addr)
	}

	return ip, nil
}
