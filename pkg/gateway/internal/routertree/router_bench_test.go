package routertree

import (
	"fmt"
	"testing"
)

func BenchmarkRouteTree_Match(b *testing.B) {
	benchCases := []struct {
		name   string
		routes []struct {
			method    string
			pattern   string
			operation string
		}
		matchMethod string
		matchURL    string
		wantOp      string
	}{
		{
			name: "static path",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"GET", "/v1/users", "list_users"},
				{"POST", "/v1/users", "create_user"},
				{"GET", "/v1/users/{id}", "get_user"},
			},
			matchMethod: "GET",
			matchURL:    "/v1/users",
			wantOp:      "list_users",
		},
		{
			name: "path with variables",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"GET", "/v1/users/{id}", "get_user"},
				{"PUT", "/v1/users/{id}", "update_user"},
				{"DELETE", "/v1/users/{id}", "delete_user"},
			},
			matchMethod: "GET",
			matchURL:    "/v1/users/123",
			wantOp:      "get_user",
		},
		{
			name: "path with wildcards",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"GET", "/v1/{resource=*}/settings", "get_settings"},
				{"GET", "/v1/other/{path=**}", "get_resource"},
			},
			matchMethod: "GET",
			matchURL:    "/v1/test/settings",
			wantOp:      "get_settings",
		},
		{
			name: "path with verb",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"POST", "/v1/users/{name}:enable", "enable_user"},
			},
			matchMethod: "POST",
			matchURL:    "/v1/users/test:enable",
			wantOp:      "enable_user",
		},
		{
			name: "path with verb and wildcards",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"POST", "/v1/resources/{name}:activate", "activate_resource"},
			},
			matchMethod: "POST",
			matchURL:    "/v1/resources/test:activate",
			wantOp:      "activate_resource",
		},
		{
			name: "path with verb and single variable",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"POST", "/v1/resources/{name}:backup", "backup_resource"},
			},
			matchMethod: "POST",
			matchURL:    "/v1/resources/test:backup",
			wantOp:      "backup_resource",
		},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			// 设置路由树
			tree := NewRouteTree()

			// 添加路由（保持原始顺序）
			for _, r := range bc.routes {
				err := tree.Add(r.method, r.pattern, r.operation, nil)
				if err != nil {
					b.Fatalf("failed to add route %s: %v", r.pattern, err)
				}
			}

			// 验证初始匹配是否正确
			op, err := tree.Match(bc.matchMethod, bc.matchURL)
			if err != nil {
				b.Logf("Routes for %s:", bc.name)
				for _, r := range bc.routes {
					b.Logf("  %s %s -> %s", r.method, r.pattern, r.operation)
				}
				b.Logf("Trying to match: %s %s", bc.matchMethod, bc.matchURL)
				b.Fatalf("initial match failed for %s: %v", bc.name, err)
			}
			if op.Operation != bc.wantOp {
				b.Fatalf("got operation %s, want %s for %s", op.Operation, bc.wantOp, bc.name)
			}

			// 重置计时器
			b.ResetTimer()

			// 运行基准测试
			for i := 0; i < b.N; i++ {
				op, err := tree.Match(bc.matchMethod, bc.matchURL)
				if err != nil {
					b.Fatal(err)
				}
				if op == nil {
					b.Fatal("expected operation, got nil")
				}
			}
		})
	}
}

// 测试缓存性能
func BenchmarkRouteTree_MatchWithCache(b *testing.B) {
	tree := NewRouteTree()
	tree.Add("GET", "/v1/users/{id}", "get_user", nil)
	tree.Add("POST", "/v1/users", "create_user", nil)

	b.Run("cold cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree.Match("GET", "/v1/users/123")
		}
	})

	// 预热缓存
	tree.Match("GET", "/v1/users/123")

	b.Run("warm cache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree.Match("GET", "/v1/users/123")
		}
	})
}

// 测试不同大小的路由树
func BenchmarkRouteTree_MatchScale(b *testing.B) {
	sizes := []int{10, 100, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			tree := NewRouteTree()

			// 添加路由
			for i := 0; i < size; i++ {
				pattern := fmt.Sprintf("/v1/resource%d/{id}", i)
				operation := fmt.Sprintf("get_resource%d", i)
				tree.Add("GET", pattern, operation, nil)
			}

			// 测试最后一个路由的匹配
			pattern := fmt.Sprintf("/v1/resource%d/123", size-1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tree.Match("GET", pattern)
			}
		})
	}
}

// 测试内存分配
func BenchmarkRouteTree_MatchAlloc(b *testing.B) {
	tree := NewRouteTree()
	tree.Add("GET", "/v1/projects/{project}/locations/{location}/datasets/{dataset}/tables/{table}", "get_table", nil)

	b.Run("memory allocation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tree.Match("GET", "/v1/projects/p1/locations/us/datasets/d1/tables/t1")
		}
	})
}

// 添加专门的动词匹配基准测试
func BenchmarkRouteTree_MatchVerb(b *testing.B) {
	tree := NewRouteTree()

	// 添加带动词的路由
	routes := []struct {
		method    string
		pattern   string
		operation string
	}{
		{"POST", "/v1/users/{user}:enable", "enable_user"},
		{"POST", "/v1/users/{user}:disable", "disable_user"},
		{"POST", "/v1/users/{user}:reset", "reset_user"},
		{"POST", "/v1/users/{user}:delete", "delete_user"},
		{"POST", "/v1/users/{user}:update", "update_user"},
	}

	for _, r := range routes {
		if err := tree.Add(r.method, r.pattern, r.operation, nil); err != nil {
			b.Fatal(err)
		}
	}

	testCases := []struct {
		name   string
		url    string
		wantOp string
	}{
		{
			name:   "simple verb",
			url:    "/v1/users/test:enable",
			wantOp: "enable_user",
		},
		{
			name:   "complex user id",
			url:    "/v1/users/test.123-456:disable",
			wantOp: "disable_user",
		},
		{
			name:   "long verb",
			url:    "/v1/users/test:update",
			wantOp: "update_user",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// 验证初始匹配
			op, err := tree.Match("POST", tc.url)
			if err != nil {
				b.Logf("Routes:")
				for _, r := range routes {
					b.Logf("  %s %s -> %s", r.method, r.pattern, r.operation)
				}
				b.Logf("Trying to match: POST %s", tc.url)
				b.Fatalf("initial match failed: %v", err)
			}
			if op.Operation != tc.wantOp {
				b.Fatalf("got operation %s, want %s", op.Operation, tc.wantOp)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				op, err := tree.Match("POST", tc.url)
				if err != nil {
					b.Fatal(err)
				}
				if op == nil {
					b.Fatal("expected operation, got nil")
				}
			}
		})
	}
}

// 添加一个辅助函数来打印路由树（用于调试）
func printRouteTree(b *testing.B, tree *RouteTree) {
	b.Logf("Route Tree:")
	var printNode func(n *node, prefix string)
	printNode = func(n *node, prefix string) {
		if n.target != nil {
			b.Logf("%s -> %s", prefix, n.target.Operation)
		}
		for path, child := range n.children.static {
			printNode(child, prefix+"/"+path)
		}
		if n.children.wildcard != nil {
			printNode(n.children.wildcard, prefix+"/*")
		}
		if n.children.wildcard2 != nil {
			printNode(n.children.wildcard2, prefix+"/**")
		}
	}
	printNode(tree.root, "")
}
