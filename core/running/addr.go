package running

import (
	"fmt"
	"os"
	"strings"
)

const (
	// DefaultDebugPort is the default local bind port for /debug endpoints.
	DefaultDebugPort = 6060
)

// HttpListenAddr returns the local HTTP bind address.
// HTTP_ADDR overrides; otherwise the global --http-port flag is used.
func HttpListenAddr() string {
	if v := strings.TrimSpace(os.Getenv("HTTP_ADDR")); v != "" {
		return v
	}
	return ":" + HttpPort.String()
}

// DebugListenAddr returns the local bind address for /debug endpoints.
// DEBUG_ADDR overrides; otherwise DEBUG_PORT or DefaultDebugPort is used.
func DebugListenAddr() string {
	if v := strings.TrimSpace(os.Getenv("DEBUG_ADDR")); v != "" {
		return v
	}
	port := strings.TrimSpace(os.Getenv("DEBUG_PORT"))
	if port == "" {
		return fmt.Sprintf(":%d", DefaultDebugPort)
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
