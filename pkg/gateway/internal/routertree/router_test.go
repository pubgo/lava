package routertree

import (
	"testing"

	"github.com/pubgo/lava/pkg/gateway/internal/routerparser"
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

func TestRouteTree_List(t *testing.T) {
	tree := NewRouteTree()

	// 添加测试路由
	assert.NoError(t, tree.Add("GET", "/v1/users/{id}", "get_user", nil))
	assert.NoError(t, tree.Add("POST", "/v1/users", "create_user", nil))
	assert.NoError(t, tree.Add("GET", "/v1/users", "list_users", nil))
	assert.NoError(t, tree.Add("PUT", "/v1/users/{id}", "update_user", nil))
	assert.NoError(t, tree.Add("DELETE", "/v1/users/{id}", "delete_user", nil))
	assert.NoError(t, tree.Add("POST", "/v1/users/{id}:cancel", "cancel_user", nil))

	// 获取所有路由
	routes := tree.List()

	// 打印实际的路由列表，帮助调试
	t.Logf("Actual routes:")
	for _, r := range routes {
		t.Logf("Method: %s, Path: %s, Operation: %s, Verb: %s, Vars: %v",
			r.Method, r.Path, r.Operation, r.Verb, r.Vars)
	}

	// 验证结果
	assert.Len(t, routes, 6)

	// 验证具体路由信息
	expectedRoutes := map[string]struct {
		method string
		path   string
		vars   []string
		verb   string
	}{
		"get_user": {
			method: "GET",
			path:   "/v1/users/{id}",
			vars:   []string{"id"},
		},
		"create_user": {
			method: "POST",
			path:   "/v1/users",
			vars:   nil,
		},
		"list_users": {
			method: "GET",
			path:   "/v1/users",
			vars:   nil,
		},
		"update_user": {
			method: "PUT",
			path:   "/v1/users/{id}",
			vars:   []string{"id"},
		},
		"delete_user": {
			method: "DELETE",
			path:   "/v1/users/{id}",
			vars:   []string{"id"},
		},
		"cancel_user": {
			method: "POST",
			path:   "/v1/users/{id}:cancel",
			vars:   []string{"id"},
			verb:   "cancel",
		},
	}

	for _, route := range routes {
		expected, ok := expectedRoutes[route.Operation]
		assert.True(t, ok, "unexpected operation: %s", route.Operation)
		assert.Equal(t, expected.method, route.Method)
		assert.Equal(t, expected.path, route.Path)
		assert.Equal(t, expected.vars, route.Vars)
		assert.Equal(t, expected.verb, route.Verb)
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
		wantVars   []routerparser.PathFieldVar
		wantError  error
		testMethod string
	}{
		// 标准 REST 映射
		{
			name:    "get method",
			pattern: "/v1/{name=users/*}",
			method:  "GET",
			url:     "/v1/users/123",
			wantOp:  "get_user",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"name"}, Value: "123"},
			},
			testMethod: "GET",
		},
		{
			name:       "list method",
			pattern:    "/v1/users",
			method:     "GET",
			url:        "/v1/users",
			wantOp:     "list_users",
			wantVars:   nil,
			testMethod: "GET",
		},
		{
			name:       "create method",
			pattern:    "/v1/users",
			method:     "POST",
			url:        "/v1/users",
			wantOp:     "create_user",
			wantVars:   nil,
			testMethod: "POST",
		},
		{
			name:    "update method",
			pattern: "/v1/{name=users/*}",
			method:  "PUT",
			url:     "/v1/users/123",
			wantOp:  "update_user",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"name"}, Value: "123"},
			},
			testMethod: "PUT",
		},
		{
			name:    "delete method",
			pattern: "/v1/{name=users/*}",
			method:  "DELETE",
			url:     "/v1/users/123",
			wantOp:  "delete_user",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"name"}, Value: "123"},
			},
			testMethod: "DELETE",
		},

		// 自定义方法
		{
			name:     "custom verb",
			pattern:  "/v1/{name=users/*}:cancel",
			method:   "POST",
			url:      "/v1/users/123:cancel",
			wantOp:   "cancel_user",
			wantVerb: "cancel",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"name"}, Value: "123"},
			},
			testMethod: "POST",
		},

		// 嵌套资源
		{
			name:    "nested resources",
			pattern: "/v1/{parent=projects/*/locations/*}/datasets/{dataset}",
			method:  "GET",
			url:     "/v1/projects/p1/locations/us/datasets/d1",
			wantOp:  "get_dataset",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"parent"}, Value: "p1/us"},
				{Fields: []string{"dataset"}, Value: "d1"},
			},
			testMethod: "GET",
		},

		// 通配符路径
		{
			name:    "wildcard path",
			pattern: "/v1/{name=**}",
			method:  "GET",
			url:     "/v1/users/123/posts/456",
			wantOp:  "get_resource",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"name"}, Value: "users/123/posts/456"},
			},
			testMethod: "GET",
		},

		// 点分隔字段
		{
			name:    "dot separated fields",
			pattern: "/v1/{book.name=shelves/*/books/*}",
			method:  "GET",
			url:     "/v1/shelves/s1/books/b1",
			wantOp:  "get_book",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"book", "name"}, Value: "s1/b1"},
			},
			testMethod: "GET",
		},

		// 错误场景
		{
			name:       "method not allowed",
			pattern:    "/v1/{name=users/*}",
			method:     "GET",
			url:        "/v1/users/123",
			wantError:  ErrMethodNotAllowed,
			testMethod: "POST",
		},
		{
			name:       "path not found",
			pattern:    "/v1/users/{id}",
			method:     "GET",
			url:        "/v1/books/123",
			wantError:  ErrPathNodeNotFound,
			testMethod: "GET",
		},
		{
			name:       "invalid verb",
			pattern:    "/v1/{name=users/*}:cancel",
			method:     "POST",
			url:        "/v1/users/123:delete",
			wantError:  ErrVerbNotMatch,
			testMethod: "POST",
		},
		{
			name:    "list method with parent",
			pattern: "/v1/{parent=projects/*}/users",
			method:  "GET",
			url:     "/v1/projects/p1/users",
			wantOp:  "list_users",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"parent"}, Value: "p1"},
			},
			testMethod: "GET",
		},
		{
			name:    "create method with parent",
			pattern: "/v1/{parent=projects/*}/users",
			method:  "POST",
			url:     "/v1/projects/p1/users",
			wantOp:  "create_user",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"parent"}, Value: "p1"},
			},
			testMethod: "POST",
		},

		// 点分隔变量测试
		{
			name:    "simple dotted variable",
			pattern: "/v1/{resource.name}",
			method:  "GET",
			url:     "/v1/test123",
			wantOp:  "get_resource",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"resource", "name"}, Value: "test123"},
			},
			testMethod: "GET",
		},
		{
			name:    "multiple dotted variables",
			pattern: "/v1/{resource.path.name=messages/*}/items/{item.id}",
			method:  "GET",
			url:     "/v1/messages/123/items/456",
			wantOp:  "get_item",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"resource", "path", "name"}, Value: "123"},
				{Fields: []string{"item", "id"}, Value: "456"},
			},
			testMethod: "GET",
		},
		{
			name:    "nested dotted variables with wildcards",
			pattern: "/v1/{parent.resource=projects/*/locations/*}/datasets/{dataset.name}/tables/{table.id}",
			method:  "GET",
			url:     "/v1/projects/p1/locations/us/datasets/d1/tables/t1",
			wantOp:  "get_table",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"parent", "resource"}, Value: "p1/us"},
				{Fields: []string{"dataset", "name"}, Value: "d1"},
				{Fields: []string{"table", "id"}, Value: "t1"},
			},
			testMethod: "GET",
		},
		{
			name:     "dotted variables with custom verb",
			pattern:  "/v1/{database.name=projects/*/instances/*}:backup",
			method:   "POST",
			url:      "/v1/projects/p1/instances/i1:backup",
			wantOp:   "backup_database",
			wantVerb: "backup",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"database", "name"}, Value: "p1/i1"},
			},
			testMethod: "POST",
		},
		{
			name:    "complex nested dotted variables",
			pattern: "/v1/{parent.resource=orgs/*/projects/*}/locations/{location.name}/repositories/{repo.path=**}",
			method:  "GET",
			url:     "/v1/orgs/o1/projects/p1/locations/us-west1/repositories/path/to/repo",
			wantOp:  "get_repository",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"parent", "resource"}, Value: "o1/p1"},
				{Fields: []string{"location", "name"}, Value: "us-west1"},
				{Fields: []string{"repo", "path"}, Value: "path/to/repo"},
			},
			testMethod: "GET",
		},
		{
			name:       "dotted variables with validation error",
			pattern:    "/v1/{resource.path.name=messages/*}/items/{item.id}",
			method:     "GET",
			url:        "/v1/messages/123/wrong/456",
			wantError:  ErrPathNodeNotFound,
			testMethod: "GET",
		},
		{
			name:    "path with wildcard variables",
			pattern: "/v1/{project}/users/{user_id}/settings",
			method:  "GET",
			url:     "/v1/abc/users/123/settings",
			wantOp:  "get_settings",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"project"}, Value: "abc"},
				{Fields: []string{"user_id"}, Value: "123"},
			},
			testMethod: "GET",
		},
		{
			name:    "mixed variables with wildcards",
			pattern: "/v1/{project}/users/{user_id}/regions/{region=*}/clusters/{cluster=*}",
			method:  "GET",
			url:     "/v1/project1/users/u123/regions/us-west1/clusters/c1",
			wantOp:  "get_cluster",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"project"}, Value: "project1"},
				{Fields: []string{"user_id"}, Value: "u123"},
				{Fields: []string{"region"}, Value: "us-west1"},
				{Fields: []string{"cluster"}, Value: "c1"},
			},
			testMethod: "GET",
		},
		{
			name:    "path with wildcard pattern",
			pattern: "/v1/{name=*}/settings",
			method:  "GET",
			url:     "/v1/abc/settings",
			wantOp:  "get_settings",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"name"}, Value: "abc"},
			},
			testMethod: "GET",
		},
		{
			name:    "path with double wildcard pattern",
			pattern: "/v1/{path=**}",
			method:  "GET",
			url:     "/v1/abc/def/ghi",
			wantOp:  "get_resource",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"path"}, Value: "abc/def/ghi"},
			},
			testMethod: "GET",
		},
		{
			name:       "path with double wildcard and suffix",
			pattern:    "/v1/prefix/{path=**}/suffix",
			method:     "GET",
			url:        "/v1/prefix/a/b/c/suffix",
			wantError:  ErrPathNodeNotFound,
			testMethod: "GET",
		},
		{
			name:    "path with nested wildcards",
			pattern: "/v1/{parent=projects/*/locations/*}/{resource=**}",
			method:  "GET",
			url:     "/v1/projects/p1/locations/us/a/b/c",
			wantOp:  "get_nested_resource",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"parent"}, Value: "p1/us"},
				{Fields: []string{"resource"}, Value: "a/b/c"},
			},
			testMethod: "GET",
		},
		{
			name:    "path with single wildcard segments",
			pattern: "/v1/{project=*}/users/{user=*}/data",
			method:  "GET",
			url:     "/v1/p1/users/u1/data",
			wantOp:  "get_user_data",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"project"}, Value: "p1"},
				{Fields: []string{"user"}, Value: "u1"},
			},
			testMethod: "GET",
		},
		{
			name:       "method not allowed error",
			pattern:    "/v1/{resource.name}/items",
			method:     "GET",
			url:        "/v1/test/items",
			wantError:  ErrMethodNotAllowed,
			testMethod: "POST",
		},
		{
			name:       "invalid verb error",
			pattern:    "/v1/{resource.name}:process",
			method:     "POST",
			url:        "/v1/test:wrong",
			wantError:  ErrVerbNotMatch,
			testMethod: "POST",
		},
		{
			name:       "invalid path segments",
			pattern:    "/v1/{parent.resource=projects/*/locations/*}/datasets/{dataset.name}",
			method:     "GET",
			url:        "/v1/projects/p1/wrong/us/datasets/d1",
			wantError:  ErrPathNodeNotFound,
			testMethod: "GET",
		},
		{
			name:       "missing verb error",
			pattern:    "/v1/{resource.name}:process",
			method:     "POST",
			url:        "/v1/test",
			wantError:  ErrVerbNotMatch,
			testMethod: "POST",
		},
		{
			name:     "valid verb match",
			pattern:  "/v1/{resource.name}:process",
			method:   "POST",
			url:      "/v1/test:process",
			wantOp:   "process_resource",
			wantVerb: "process",
			wantVars: []routerparser.PathFieldVar{
				{Fields: []string{"resource", "name"}, Value: "test"},
			},
			testMethod: "POST",
		},
		{
			name:       "empty verb error",
			pattern:    "/v1/{resource.name}:process",
			method:     "POST",
			url:        "/v1/test:",
			wantError:  ErrVerbNotMatch,
			testMethod: "POST",
		},
		{
			name:       "invalid path error",
			pattern:    "/v1/users/{id}",
			method:     "GET",
			url:        "/v2/users/123",
			wantError:  ErrPathNodeNotFound,
			testMethod: "GET",
		},
		{
			name:       "wrong segment error",
			pattern:    "/v1/users/{id}/profile",
			method:     "GET",
			url:        "/v1/users/123/settings",
			wantError:  ErrPathNodeNotFound,
			testMethod: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewRouteTree()
			assert.NoError(t, tree.Add(tt.method, tt.pattern, tt.wantOp, nil))

			op, err := tree.Match(tt.testMethod, tt.url)
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOp, op.Operation)
			assert.Equal(t, tt.wantVerb, op.Verb)
			assert.Equal(t, tt.wantVars, op.Vars)
		})
	}
}

func TestRouteTree_MatchPath(t *testing.T) {
	tests := []struct {
		name   string
		routes []struct {
			method    string
			pattern   string
			operation string
		}
		matchMethod string
		matchURL    string
		wantOp      string
		wantErr     error
	}{
		{
			name: "complex path matching",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"GET", "/v1/projects/{project}/locations/{location}", "get_location"},
				{"GET", "/v1/projects/{project}/locations/{location}/datasets/{dataset}", "get_dataset"},
				{"GET", "/v1/projects/{project}/locations/{location}/datasets/{dataset}/{path=**}", "get_dataset_resource"},
			},
			matchMethod: "GET",
			matchURL:    "/v1/projects/p1/locations/us/datasets/d1/tables/t1",
			wantOp:      "get_dataset_resource",
		},
		{
			name: "wildcard priority",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"GET", "/v1/resources/{name=*}", "get_by_name"},
				{"GET", "/v1/resources/special", "get_special"},
			},
			matchMethod: "GET",
			matchURL:    "/v1/resources/special",
			wantOp:      "get_special", // 静态路径应该优先于通配符
		},
		{
			name: "nested resources with wildcards",
			routes: []struct {
				method    string
				pattern   string
				operation string
			}{
				{"GET", "/v1/{name=projects/*/locations/*}/datasets", "list_datasets"},
				{"GET", "/v1/{name=projects/*/locations/*}/datasets/{dataset}", "get_dataset"},
				{"GET", "/v1/{name=projects/*/locations/*}/datasets/{dataset}/{resource=**}", "get_dataset_resource"},
			},
			matchMethod: "GET",
			matchURL:    "/v1/projects/p1/locations/us/datasets/d1/tables/t1/rows/r1",
			wantOp:      "get_dataset_resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewRouteTree()

			// 添加路由
			for _, r := range tt.routes {
				err := tree.Add(r.method, r.pattern, r.operation, nil)
				assert.NoError(t, err)
			}

			// 执行匹配
			op, err := tree.Match(tt.matchMethod, tt.matchURL)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOp, op.Operation)
		})
	}
}
