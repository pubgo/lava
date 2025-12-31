package gologr

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"

	"github.com/pubgo/funk/v2/log"
)

func TestName(t *testing.T) {
	var buf bytes.Buffer
	ll := logr.New(NewSink(log.Output(&buf)))
	ll.Info("test", "hello", 123456)

	data := make(map[string]any)
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &data))
	assert.Equal(t, data["hello"], float64(123456))
	assert.Equal(t, data["level"], "info")
	assert.Equal(t, data["message"], "test")
	assert.Contains(t, data["caller"], "otellogger/log_test.go")
}
