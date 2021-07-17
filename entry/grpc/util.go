package grpc

import (
	"os"
)

func getHostname() string {
	if name, err := os.Hostname(); err != nil {
		return "unknown"
	} else {
		return name
	}
}
