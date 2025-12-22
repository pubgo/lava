package routertree

import (
	"fmt"
	"testing"
)

// BenchmarkAdd 测试路由注册性能
func BenchmarkAdd(b *testing.B) {
	tree := New()
	routes := []struct {
		method    string
		path      string
		operation string
	}{
		{"get", "/api/v1/users", "list_users"},
		{"get", "/api/v1/users/{id}", "get_user"},
		{"post", "/api/v1/users", "create_user"},
		{"put", "/api/v1/users/{id}", "update_user"},
		{"delete", "/api/v1/users/{id}", "delete_user"},
		{"get", "/api/v1/posts", "list_posts"},
		{"get", "/api/v1/posts/{id}", "get_post"},
		{"get", "/api/v1/posts/{id}/comments", "list_comments"},
		{"get", "/api/v1/*/search", "search"},
		{"get", "/files/{path=**}", "get_file"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			_ = tree.Add(route.method, route.path, route.operation, nil)
		}
	}
}

// BenchmarkAddMany 测试大量路由注册性能
func BenchmarkAddMany(b *testing.B) {
	tree := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			path := fmt.Sprintf("/api/v1/resource/%d", j)
			_ = tree.Add("get", path, fmt.Sprintf("op_%d", j), nil)
		}
	}
}

// BenchmarkMatchExact 测试精确匹配性能
func BenchmarkMatchExact(b *testing.B) {
	tree := New()
	// 注册一些路由
	tree.Add("get", "/api/v1/users", "list_users", nil)
	tree.Add("get", "/api/v1/users/{id}", "get_user", nil)
	tree.Add("get", "/api/v1/posts", "list_posts", nil)
	tree.Add("get", "/api/v1/posts/{id}", "get_post", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/api/v1/users")
	}
}

// BenchmarkMatchVariable 测试变量匹配性能
func BenchmarkMatchVariable(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/v1/users/{id}", "get_user", nil)
	tree.Add("get", "/api/v1/posts/{id}/comments/{comment_id}", "get_comment", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/api/v1/users/123")
	}
}

// BenchmarkMatchWildcard 测试通配符匹配性能
func BenchmarkMatchWildcard(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/v1/*/search", "search", nil)
	tree.Add("get", "/files/{path=**}", "get_file", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/api/v1/users/search")
	}
}

// BenchmarkMatchDoubleWildcard 测试双通配符匹配性能
func BenchmarkMatchDoubleWildcard(b *testing.B) {
	tree := New()
	tree.Add("get", "/files/{path=**}", "get_file", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/files/images/2023/12/photo.jpg")
	}
}

// BenchmarkMatchDeepPath 测试深层路径匹配性能
func BenchmarkMatchDeepPath(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/v1/users/{user_id}/posts/{post_id}/comments/{comment_id}", "get_comment", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/api/v1/users/u1/posts/p1/comments/c1")
	}
}

// BenchmarkMatchWithVerb 测试带动词的匹配性能
func BenchmarkMatchWithVerb(b *testing.B) {
	tree := New()
	tree.Add("post", "/users/{id}:get", "get_user", nil)
	tree.Add("post", "/users/{id}:delete", "delete_user", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("post", "/users/123:get")
	}
}

// BenchmarkMatchRootPath 测试根路径匹配性能
func BenchmarkMatchRootPath(b *testing.B) {
	tree := New()
	tree.Add("get", "/", "root", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/")
	}
}

// BenchmarkMatchComplex 测试复杂场景匹配性能
func BenchmarkMatchComplex(b *testing.B) {
	tree := New()
	// 注册多种类型的路由
	tree.Add("get", "/api/v1/users/{user_id}/posts/{post_id}/comments", "list_comments", nil)
	tree.Add("post", "/api/v1/users/{user_id}/posts/{post_id}/comments", "create_comment", nil)
	tree.Add("get", "/api/v1/users/{user_id}/posts/{post_id}/comments/{comment_id}", "get_comment", nil)
	tree.Add("get", "/api/v1/*/search", "search", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 测试精确匹配
		_, _ = tree.Match("get", "/api/v1/users/u1/posts/p1/comments")
		// 测试通配符匹配
		_, _ = tree.Match("get", "/api/v1/users/search")
	}
}

// BenchmarkMatchPrecedence 测试匹配优先级性能（精确匹配 vs 通配符）
func BenchmarkMatchPrecedence(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/users/special", "get_special_user", nil)
	tree.Add("get", "/api/users/*", "get_user_wildcard", nil)
	tree.Add("get", "/api/users/**", "get_users_double_wildcard", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 精确匹配应该优先
		_, _ = tree.Match("get", "/api/users/special")
	}
}

// BenchmarkMatchManyRoutes 测试在大量路由中匹配的性能
func BenchmarkMatchManyRoutes(b *testing.B) {
	tree := New()
	// 注册大量路由
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/api/v1/resource/%d", i)
		tree.Add("get", path, fmt.Sprintf("op_%d", i), nil)
	}
	// 添加一些通配符路由
	tree.Add("get", "/api/v1/*/search", "search", nil)
	tree.Add("get", "/api/v1/**", "catch_all", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 测试精确匹配（应该很快）
		_, _ = tree.Match("get", "/api/v1/resource/500")
		// 测试通配符匹配
		_, _ = tree.Match("get", "/api/v1/unknown/search")
	}
}

// BenchmarkMatchNotFound 测试匹配失败的性能
func BenchmarkMatchNotFound(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/v1/users", "list_users", nil)
	tree.Add("get", "/api/v1/posts", "list_posts", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 测试不存在的路径
		_, _ = tree.Match("get", "/api/v1/nonexistent")
	}
}

// BenchmarkAddAndMatch 测试注册和匹配的混合性能
func BenchmarkAddAndMatch(b *testing.B) {
	tree := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 注册路由
		path := fmt.Sprintf("/api/v1/resource/%d", i%100)
		tree.Add("get", path, fmt.Sprintf("op_%d", i), nil)

		// 匹配路由
		_, _ = tree.Match("get", path)
	}
}

// BenchmarkMatchPathDepth1 测试1层路径深度匹配性能
func BenchmarkMatchPathDepth1(b *testing.B) {
	tree := New()
	tree.Add("get", "/users", "list_users", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/users")
	}
}

// BenchmarkMatchPathDepth3 测试3层路径深度匹配性能
func BenchmarkMatchPathDepth3(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/v1/users", "list_users", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/api/v1/users")
	}
}

// BenchmarkMatchPathDepth5 测试5层路径深度匹配性能
func BenchmarkMatchPathDepth5(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/v1/users/{id}/posts", "list_posts", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/api/v1/users/123/posts")
	}
}

// BenchmarkMatchPathDepth7 测试7层路径深度匹配性能
func BenchmarkMatchPathDepth7(b *testing.B) {
	tree := New()
	tree.Add("get", "/api/v1/users/{id}/posts/{pid}/comments", "list_comments", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tree.Match("get", "/api/v1/users/u1/posts/p1/comments")
	}
}

// BenchmarkMatchWithFallback 测试需要回退到通配符的匹配性能
func BenchmarkMatchWithFallback(b *testing.B) {
	tree := New()
	// 注册精确路由和通配符路由
	tree.Add("get", "/api/v1/users/{user_id}/posts/{post_id}/comments", "list_comments", nil)
	tree.Add("get", "/api/v1/*/search", "search", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 这个路径会先尝试精确匹配 users，失败后回退到通配符匹配
		_, _ = tree.Match("get", "/api/v1/users/search")
	}
}

// BenchmarkParseURL 测试 URL 解析性能
func BenchmarkParseURL(b *testing.B) {
	urls := []string{
		"/api/v1/users",
		"/api/v1/users/123",
		"/api/v1/users/123:get",
		"/api/v1/users/123/posts/456",
		"/",
		"/files/images/2023/12/photo.jpg",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := urls[i%len(urls)]
		_, _, _ = parseURL(url)
	}
}

// BenchmarkExtractPathVars 测试路径变量提取性能
func BenchmarkExtractPathVars(b *testing.B) {
	vars := []*pathVariable{
		{fields: []string{"user_id"}, start: 3, end: 3},
		{fields: []string{"post_id"}, start: 5, end: 5},
		{fields: []string{"comment_id"}, start: 7, end: 7},
	}
	paths := []string{"api", "v1", "users", "u1", "posts", "p1", "comments", "c1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractPathVars(vars, paths)
	}
}

// BenchmarkBuildMatchOperation 测试构建匹配结果性能
func BenchmarkBuildMatchOperation(b *testing.B) {
	target := &routeTarget{
		Method:    "__GET__",
		Path:      "/api/v1/users/{id}",
		Operation: "get_user",
		Verb:      nil,
		Vars: []*pathVariable{
			{fields: []string{"id"}, start: 3, end: 3},
		},
		extras: nil,
	}
	pathNodes := []string{"api", "v1", "users", "123"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildMatchOperation(target, "", pathNodes)
	}
}
