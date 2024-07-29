package routertree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	tree := NewRouteTree()
	assert.NoError(t, tree.Add("get", "/user/user/{id}:get1", "get_user", nil))
	assert.NoError(t, tree.Add("post", "/user/user/{id}", "post_user", nil))
	assert.NoError(t, tree.Add("post", "/user/user/{id}/send_mail", "post_mail", nil))
	opt, err := tree.Match("get", "/user/user/1:get1")
	assert.NoError(t, err)
	assert.NotNil(t, opt)
	assert.Equal(t, "get_user", opt.Operation)
	assert.Equal(t, "get1", opt.Verb)
}
