package resty

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParams(t *testing.T) {
	assert.True(t, regParam.MatchString("/a/b/c/{a_b.c222}"))
	assert.True(t, regParam.MatchString("/a/b/c{a_b.c222}"))
	assert.True(t, regParam.MatchString("/a/b/c{ a_b.c222 }"))
	assert.True(t, regParam.MatchString("/a/b/c{ a_b:c222/123 }"))

	assert.False(t, regParam.MatchString("/a/b/c"))
}
