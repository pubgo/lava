package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWebRequestFromContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		method      string
		expected    bool
	}{
		{"valid grpc-web", "application/grpc-web+proto", "POST", true},
		{"valid grpc-web-text", "application/grpc-web-text+json", "POST", true},
		{"invalid method", "application/grpc-web+proto", "GET", false},
		{"invalid content-type", "application/json", "POST", false},
		{"no subtype", "application/grpc-web", "POST", true},
		{"json alias is plain JSON, not grpc-web", "application/grpc-web-json", "POST", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, ok := isWebRequestFromContentType(tt.contentType, tt.method)
			assert.Equal(t, tt.expected, ok)
		})
	}
}
