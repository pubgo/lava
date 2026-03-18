package resty

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParams(t *testing.T) {
	assert.True(t, IsPathTemplate("/a/b/c/{a_b.c222}"))
	assert.True(t, IsPathTemplate("/a/b/c{a_b.c222}"))
	assert.True(t, IsPathTemplate("/a/b/c{ a_b.c222 }"))
	assert.True(t, IsPathTemplate("/a/b/c{ a_b:c222/123 }"))

	assert.False(t, IsPathTemplate("/a/b/c"))
}
