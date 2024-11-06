package routertree

import (
	"fmt"
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

func BenchmarkRouteTree_Performance(b *testing.B) {
	tree := NewRouteTree()

	// Add test routes
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/api/v1/users/%d/posts/%d", i, i)
		assert.NoError(b, tree.Add("GET", path, fmt.Sprintf("op_%d", i), nil))
	}

	// Benchmark matching
	b.Run("Match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := fmt.Sprintf("/api/v1/users/%d/posts/%d", i%1000, i%1000)
			_, err := tree.Match("GET", path)
			assert.NoError(b, err)
		}
	})
}

func TestRouteTree_List(t *testing.T) {
	tree := NewRouteTree()

	// 添加一些测试路由
	assert.NoError(t, tree.Add("GET", "/users/{id}", "get_user", nil))
	assert.NoError(t, tree.Add("POST", "/users", "create_user", nil))
	assert.NoError(t, tree.Add("GET", "/users/{id}/posts/{post.id}:list", "list_posts", nil))

	// 获取所有路由
	routes := tree.List()

	// 验证结果
	assert.Len(t, routes, 3)

	// 验证具体路由信息
	for _, route := range routes {
		switch route.Operation {
		case "get_user":
			assert.Equal(t, "GET", route.Method)
			assert.Equal(t, "/users/{id}", route.Path)
			assert.Equal(t, []string{"id"}, route.Vars)

		case "create_user":
			assert.Equal(t, "POST", route.Method)
			assert.Equal(t, "/users", route.Path)
			assert.Empty(t, route.Vars)

		case "list_posts":
			assert.Equal(t, "GET", route.Method)
			assert.Equal(t, "/users/{id}/posts/{post.id}:list", route.Path)
			assert.Equal(t, []string{"id", "post.id"}, route.Vars)
			assert.Equal(t, "list", route.Verb)

		default:
			t.Errorf("unexpected operation: %s", route.Operation)
		}
	}
}

func TestRouteTree_Match(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		method     string
		url        string
		wantOp     string
		wantVerb   string
		wantVars   []PathFieldVar
		wantError  bool
		testMethod string
	}{
		{
			name:    "simple match",
			pattern: "/users/{id}",
			method:  "GET",
			url:     "/users/123",
			wantOp:  "get_user",
			wantVars: []PathFieldVar{
				{Fields: []string{"id"}, Value: "123"},
			},
			testMethod: "GET",
		},
		{
			name:     "match with verb",
			pattern:  "/users/{id}:list",
			method:   "GET",
			url:      "/users/123:list",
			wantOp:   "list_user",
			wantVerb: "list",
			wantVars: []PathFieldVar{
				{Fields: []string{"id"}, Value: "123"},
			},
			testMethod: "GET",
		},
		{
			name:    "match with double wildcard",
			pattern: "/users/{path=**}/details",
			method:  "GET",
			url:     "/users/a/b/c/details",
			wantOp:  "get_details",
			wantVars: []PathFieldVar{
				{Fields: []string{"path"}, Value: "a/b/c"},
			},
			testMethod: "GET",
		},
		{
			name:       "no match - wrong method",
			pattern:    "/users/{id}",
			method:     "GET",
			url:        "/users/123",
			wantError:  true,
			testMethod: "POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewRouteTree()
			assert.NoError(t, tree.Add(tt.method, tt.pattern, tt.wantOp, nil))

			op, err := tree.Match(tt.testMethod, tt.url)
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOp, op.Operation)
			assert.Equal(t, tt.wantVerb, op.Verb)
			assert.Equal(t, tt.wantVars, op.Vars)
		})
	}
}
