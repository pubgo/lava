package routerparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  bool
		validate func(*testing.T, *RoutePattern)
	}{
		{
			name: "simple path",
			path: "/hello",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"hello"}, r.Segments)
				assert.Nil(t, r.HttpVerb)
				assert.Empty(t, r.Variables)
			},
		},
		{
			name: "nested path",
			path: "/hello/world",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"hello", "world"}, r.Segments)
			},
		},
		{
			name: "path with hyphen",
			path: "/hello-world",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"hello-world"}, r.Segments)
			},
		},
		{
			name: "path with underscore",
			path: "/hello_world",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"hello_world"}, r.Segments)
			},
		},
		{
			name: "path with dot",
			path: "/hello.world",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"hello.world"}, r.Segments)
			},
		},
		{
			name: "path with verb",
			path: "/users:get",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users"}, r.Segments)
				assert.Equal(t, "get", *r.HttpVerb)
			},
		},
		{
			name: "path with simple variable",
			path: "/users/{id}",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users", "*"}, r.Segments)
				assert.Len(t, r.Variables, 1)
				assert.Equal(t, []string{"id"}, r.Variables[0].FieldPath)
			},
		},
		{
			name: "path with nested variable",
			path: "/users/{user.id}/posts",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users", "*", "posts"}, r.Segments)
				assert.Len(t, r.Variables, 1)
				assert.Equal(t, []string{"user", "id"}, r.Variables[0].FieldPath)
				assert.Equal(t, 1, r.Variables[0].StartIdx)
				assert.Equal(t, 1, r.Variables[0].EndIdx)
			},
		},
		{
			name: "path with wildcard",
			path: "/users/*/posts",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users", "*", "posts"}, r.Segments)
			},
		},
		{
			name: "path with double wildcard",
			path: "/users/**",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users", "**"}, r.Segments)
			},
		},
		{
			name: "complex path with variable and verb",
			path: "/api/v1/users/{user.id}/posts/{post.id}:get",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"api", "v1", "users", "*", "posts", "*"}, r.Segments)
				assert.Equal(t, "get", *r.HttpVerb)
				assert.Len(t, r.Variables, 2)
				assert.Equal(t, []string{"user", "id"}, r.Variables[0].FieldPath)
				assert.Equal(t, []string{"post", "id"}, r.Variables[1].FieldPath)
			},
		},
		{
			name: "path with variable and custom matching",
			path: "/users/{id=**}/posts",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users", "**", "posts"}, r.Segments)
				assert.Len(t, r.Variables, 1)
				assert.Equal(t, []string{"id"}, r.Variables[0].FieldPath)
				assert.Equal(t, 1, r.Variables[0].StartIdx)
				assert.Equal(t, -1, r.Variables[0].EndIdx)
			},
		},
		{
			name:    "invalid path - no leading slash",
			path:    "users/{id}",
			wantErr: true,
		},
		{
			name:    "invalid path - empty variable",
			path:    "/users/{}/posts",
			wantErr: true,
		},
		{
			name:    "invalid path - unclosed variable",
			path:    "/users/{id/posts",
			wantErr: true,
		},
		{
			name: "multiple nested variables",
			path: "/users/{user.id}/posts/{post.title}/comments",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users", "*", "posts", "*", "comments"}, r.Segments)
				assert.Len(t, r.Variables, 2)
				// 第一个变量
				assert.Equal(t, []string{"user", "id"}, r.Variables[0].FieldPath)
				assert.Equal(t, 1, r.Variables[0].StartIdx)
				assert.Equal(t, 1, r.Variables[0].EndIdx)
				// 第二个变量
				assert.Equal(t, []string{"post", "title"}, r.Variables[1].FieldPath)
				assert.Equal(t, 3, r.Variables[1].StartIdx)
				assert.Equal(t, 3, r.Variables[1].EndIdx)
			},
		},
		{
			name: "nested variable with double wildcard",
			path: "/users/{user.profile=**}/details",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"users", "**", "details"}, r.Segments)
				assert.Len(t, r.Variables, 1)
				assert.Equal(t, []string{"user", "profile"}, r.Variables[0].FieldPath)
				assert.Equal(t, 1, r.Variables[0].StartIdx)
				assert.Equal(t, -1, r.Variables[0].EndIdx)
			},
		},
		{
			name: "complex nested variables",
			path: "/api/{service.version}/users/{user.profile.id}/posts/{post.data.title}:get",
			validate: func(t *testing.T, r *RoutePattern) {
				assert.Equal(t, []string{"api", "*", "users", "*", "posts", "*"}, r.Segments)
				assert.Len(t, r.Variables, 3)
				// 服务版本变量
				assert.Equal(t, []string{"service", "version"}, r.Variables[0].FieldPath)
				assert.Equal(t, 1, r.Variables[0].StartIdx)
				assert.Equal(t, 1, r.Variables[0].EndIdx)
				// 用户配置变量
				assert.Equal(t, []string{"user", "profile", "id"}, r.Variables[1].FieldPath)
				assert.Equal(t, 3, r.Variables[1].StartIdx)
				assert.Equal(t, 3, r.Variables[1].EndIdx)
				// 文章数据变量
				assert.Equal(t, []string{"post", "data", "title"}, r.Variables[2].FieldPath)
				assert.Equal(t, 5, r.Variables[2].StartIdx)
				assert.Equal(t, 5, r.Variables[2].EndIdx)
				// 验证 HTTP 动词
				assert.Equal(t, "get", *r.HttpVerb)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := ParseRoutePattern(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Test String() method
			rr, err := ParseRoutePattern(pattern.String())
			assert.NoError(t, err)
			assert.True(t, pattern.Equal(rr))

			if tt.validate != nil {
				tt.validate(t, pattern)
			}
		})
	}
}

func TestRoutePattern_Match(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		urls      []string
		verb      string
		wantVars  []PathFieldVar
		wantMatch bool
	}{
		{
			name:      "simple match",
			pattern:   "/users/{id}",
			urls:      []string{"users", "123"},
			verb:      "",
			wantVars:  []PathFieldVar{{Fields: []string{"id"}, Value: "123"}},
			wantMatch: true,
		},
		{
			name:      "match with verb",
			pattern:   "/users/{id}:get",
			urls:      []string{"users", "123"},
			verb:      "get",
			wantVars:  []PathFieldVar{{Fields: []string{"id"}, Value: "123"}},
			wantMatch: true,
		},
		{
			name:      "no match - wrong verb",
			pattern:   "/users/{id}:get",
			urls:      []string{"users", "123"},
			verb:      "post",
			wantMatch: false,
		},
		{
			name:      "match with wildcard",
			pattern:   "/users/*/posts",
			urls:      []string{"users", "anything", "posts"},
			verb:      "",
			wantMatch: true,
		},
		{
			name:      "match with double wildcard",
			pattern:   "/users/**",
			urls:      []string{"users", "anything", "else", "here"},
			verb:      "",
			wantMatch: true,
		},
		{
			name:      "no match - wrong path",
			pattern:   "/users/{id}",
			urls:      []string{"posts", "123"},
			verb:      "",
			wantMatch: false,
		},
		{
			name:    "match nested variable",
			pattern: "/users/{user.id}/posts",
			urls:    []string{"users", "123", "posts"},
			verb:    "",
			wantVars: []PathFieldVar{
				{Fields: []string{"user", "id"}, Value: "123"},
			},
			wantMatch: true,
		},
		{
			name:    "match multiple nested variables",
			pattern: "/users/{user.id}/posts/{post.title}",
			urls:    []string{"users", "123", "posts", "my-post"},
			verb:    "",
			wantVars: []PathFieldVar{
				{Fields: []string{"user", "id"}, Value: "123"},
				{Fields: []string{"post", "title"}, Value: "my-post"},
			},
			wantMatch: true,
		},
		{
			name:    "match nested variable with double wildcard",
			pattern: "/users/{user.profile=**}/details",
			urls:    []string{"users", "123", "extra", "path", "details"},
			verb:    "",
			wantVars: []PathFieldVar{
				{Fields: []string{"user", "profile"}, Value: "123/extra/path"},
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := ParseRoutePattern(tt.pattern)
			assert.NoError(t, err)

			vars, err := pattern.Match(tt.urls, tt.verb)

			if !tt.wantMatch {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantVars, vars)
		})
	}
}

func BenchmarkParser(b *testing.B) {
	patterns := []string{
		"/simple/path",
		"/users/{id}",
		"/api/v1/users/{user.id}/posts/{post.id}:get",
		"/users/{id=**}/posts",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pattern := range patterns {
			_, err := parse(pattern)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkRoutePattern_Match(b *testing.B) {
	pattern, _ := ParseRoutePattern("/api/v1/users/{user.id}/posts/{post.id}:get")
	urls := []string{"api", "v1", "users", "123", "posts", "456"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pattern.Match(urls, "get")
		if err != nil {
			b.Fatal(err)
		}
	}
}
