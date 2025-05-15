package routertree

import (
	"github.com/pubgo/funk/pretty"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute(t *testing.T) {
	tree := New()
	assert.NoError(t, tree.Add("get", "/user/user/{id}:get1", "get_user", nil))
	assert.NoError(t, tree.Add("post", "/user/user/{id}", "post_user", nil))
	assert.NoError(t, tree.Add("post", "/user/user1/{id=**}", "post_user1", nil))
	assert.NoError(t, tree.Add("post", "/user/user/{id}/send_mail", "post_mail", nil))
	opt, err := tree.Match("post", "/user/user/1/send_mail")
	pretty.Println(opt)
	assert.NoError(t, err)
	assert.NotNil(t, opt)
	assert.Equal(t, "post_mail", opt.Operation)
	assert.Equal(t, "", opt.Verb)

	opt, err = tree.Match("get", "/user/user/1:get1")
	pretty.Println(opt)
	assert.NoError(t, err)
	assert.NotNil(t, opt)
	assert.Equal(t, "get_user", opt.Operation)
	assert.Equal(t, "get1", opt.Verb)

	opt, err = tree.Match("post", "/user/user1/123/123456")
	pretty.Println(opt)
	assert.NoError(t, err)
	assert.NotNil(t, opt)
	assert.Equal(t, "post_user1", opt.Operation)
	assert.Equal(t, "", opt.Verb)
}
