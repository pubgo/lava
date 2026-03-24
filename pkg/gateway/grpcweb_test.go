package gateway

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWebRequest(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)
			_, _, ok := isWebRequest(req)
			assert.Equal(t, tt.expected, ok)
		})
	}
}
