package routerparser

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    *Pattern
		wantErr bool
		errMsg  string
	}{
		// 基本场景
		{
			name:    "simple_path",
			pattern: "/v1/messages",
			want: &Pattern{
				raw:      "/v1/messages",
				Segments: []string{"v1", "messages"},
			},
		},
		{
			name:    "simple_variable",
			pattern: "/v1/{message_id}",
			want: &Pattern{
				raw:      "/v1/{message_id}",
				Segments: []string{"v1", "*"},
				Variables: []*PathVariable{{
					FieldPath: []string{"message_id"},
					StartIdx:  1,
					EndIdx:    1,
				}},
			},
		},

		// 通配符场景
		{
			name:    "single_wildcard",
			pattern: "/v1/{name=messages/*}",
			want: &Pattern{
				raw:      "/v1/{name=messages/*}",
				Segments: []string{"v1", "messages", "*"},
				Variables: []*PathVariable{{
					FieldPath: []string{"name"},
					StartIdx:  1,
					EndIdx:    2,
					Pattern:   "messages/*",
				}},
			},
		},
		{
			name:    "double_wildcard",
			pattern: "/v1/{name=messages/**}",
			want: &Pattern{
				raw:      "/v1/{name=messages/**}",
				Segments: []string{"v1", "messages", "**"},
				Variables: []*PathVariable{{
					FieldPath: []string{"name"},
					StartIdx:  1,
					EndIdx:    -1,
					Pattern:   "messages/**",
				}},
			},
		},

		// 嵌套资源场景
		{
			name:    "nested_resource",
			pattern: "/v1/{name=projects/*/locations/*}",
			want: &Pattern{
				raw:      "/v1/{name=projects/*/locations/*}",
				Segments: []string{"v1", "projects", "*", "locations", "*"},
				Variables: []*PathVariable{{
					FieldPath: []string{"name"},
					StartIdx:  1,
					EndIdx:    4,
					Pattern:   "projects/*/locations/*",
				}},
			},
		},

		// 带动词的场景
		{
			name:    "verb_suffix",
			pattern: "/v1/messages/{message_id}:cancel",
			want: &Pattern{
				raw:      "/v1/messages/{message_id}:cancel",
				Segments: []string{"v1", "messages", "*"},
				Variables: []*PathVariable{{
					FieldPath: []string{"message_id"},
					StartIdx:  2,
					EndIdx:    2,
				}},
				HttpVerb: lo.ToPtr("cancel"),
			},
		},

		// 错误场景
		{
			name:    "empty_pattern",
			pattern: "",
			wantErr: true,
			errMsg:  "empty pattern",
		},
		{
			name:    "invalid_variable_format",
			pattern: "/v1/{name",
			wantErr: true,
			errMsg:  "parse route failed",
		},
		{
			name:    "invalid_double_wildcard_position",
			pattern: "/v1/{name=**/messages}",
			wantErr: true,
			errMsg:  "** must be the last part",
		},
		{
			name:    "multiple_double_wildcards",
			pattern: "/v1/{name=**/**}",
			wantErr: true,
			errMsg:  "multiple ** patterns are not allowed",
		},

		// 变量嵌套场景
		{
			name:    "nested_variables",
			pattern: "/v1/{parent=projects/*/locations/*}/datasets/{dataset}/tables/{table}",
			want: &Pattern{
				raw:      "/v1/{parent=projects/*/locations/*}/datasets/{dataset}/tables/{table}",
				Segments: []string{"v1", "projects", "*", "locations", "*", "datasets", "*", "tables", "*"},
				Variables: []*PathVariable{
					{
						FieldPath: []string{"parent"},
						StartIdx:  1,
						EndIdx:    4,
						Pattern:   "projects/*/locations/*",
					},
					{
						FieldPath: []string{"dataset"},
						StartIdx:  6,
						EndIdx:    6,
					},
					{
						FieldPath: []string{"table"},
						StartIdx:  8,
						EndIdx:    8,
					},
				},
			},
		},
		{
			name:    "complex_nested_variables",
			pattern: "/v1/{parent=projects/*/locations/*}/models/{model}/evaluations/{evaluation}/{slice=**}",
			want: &Pattern{
				raw:      "/v1/{parent=projects/*/locations/*}/models/{model}/evaluations/{evaluation}/{slice=**}",
				Segments: []string{"v1", "projects", "*", "locations", "*", "models", "*", "evaluations", "*", "**"},
				Variables: []*PathVariable{
					{
						FieldPath: []string{"parent"},
						StartIdx:  1,
						EndIdx:    4,
						Pattern:   "projects/*/locations/*",
					},
					{
						FieldPath: []string{"model"},
						StartIdx:  6,
						EndIdx:    6,
					},
					{
						FieldPath: []string{"evaluation"},
						StartIdx:  8,
						EndIdx:    8,
					},
					{
						FieldPath: []string{"slice"},
						StartIdx:  9,
						EndIdx:    -1,
						Pattern:   "**",
					},
				},
			},
		},

		// 点分隔变量场景
		{
			name:    "dotted_variable",
			pattern: "/v1/{resource.name}",
			want: &Pattern{
				raw:      "/v1/{resource.name}",
				Segments: []string{"v1", "*"},
				Variables: []*PathVariable{
					{
						FieldPath: []string{"resource", "name"},
						StartIdx:  1,
						EndIdx:    1,
					},
				},
			},
		},
		{
			name:    "multiple_dotted_variable",
			pattern: "/v1/{resource.path.name=messages/*}/items/{item.id}",
			want: &Pattern{
				raw:      "/v1/{resource.path.name=messages/*}/items/{item.id}",
				Segments: []string{"v1", "messages", "*", "items", "*"},
				Variables: []*PathVariable{
					{
						FieldPath: []string{"resource", "path", "name"},
						StartIdx:  1,
						EndIdx:    2,
						Pattern:   "messages/*",
					},
					{
						FieldPath: []string{"item", "id"},
						StartIdx:  4,
						EndIdx:    4,
					},
				},
			},
		},
		{
			name:    "nested_dotted_variable",
			pattern: "/v1/{parent.resource=projects/*/locations/*}/datasets/{dataset.name}/tables/{table.id}",
			want: &Pattern{
				raw:      "/v1/{parent.resource=projects/*/locations/*}/datasets/{dataset.name}/tables/{table.id}",
				Segments: []string{"v1", "projects", "*", "locations", "*", "datasets", "*", "tables", "*"},
				Variables: []*PathVariable{
					{
						FieldPath: []string{"parent", "resource"},
						StartIdx:  1,
						EndIdx:    4,
						Pattern:   "projects/*/locations/*",
					},
					{
						FieldPath: []string{"dataset", "name"},
						StartIdx:  6,
						EndIdx:    6,
					},
					{
						FieldPath: []string{"table", "id"},
						StartIdx:  8,
						EndIdx:    8,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePattern(tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPattern_Match(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		urls    []string
		verb    string
		want    []PathFieldVar
		wantErr bool
		errMsg  string
	}{
		// 基本场景
		{
			name:    "simple_variable",
			pattern: "/v1/{message_id}",
			urls:    []string{"v1", "123"},
			verb:    "get",
			want: []PathFieldVar{{
				Fields: []string{"message_id"},
				Value:  "123",
			}},
		},

		// 通配符场景
		{
			name:    "single_wildcard",
			pattern: "/v1/{name=messages/*}",
			urls:    []string{"v1", "messages", "123"},
			verb:    "get",
			want: []PathFieldVar{{
				Fields: []string{"name"},
				Value:  "123",
			}},
		},
		{
			name:    "double_wildcard",
			pattern: "/v1/{name=messages/**}",
			urls:    []string{"v1", "messages", "123", "456", "789"},
			verb:    "get",
			want: []PathFieldVar{{
				Fields: []string{"name"},
				Value:  "123/456/789",
			}},
		},

		// 嵌套资源场景
		{
			name:    "nested_resource",
			pattern: "/v1/{name=projects/*/locations/*}",
			urls:    []string{"v1", "projects", "p1", "locations", "us"},
			verb:    "get",
			want: []PathFieldVar{{
				Fields: []string{"name"},
				Value:  "p1/us",
			}},
		},

		// 存储对象路径场景
		{
			name:    "storage_object_path",
			pattern: "/storage/{bucket}/objects/{object=**}",
			urls:    []string{"storage", "mybucket", "objects", "path", "to", "file.txt"},
			verb:    "get",
			want: []PathFieldVar{
				{
					Fields: []string{"bucket"},
					Value:  "mybucket",
				},
				{
					Fields: []string{"object"},
					Value:  "path/to/file.txt",
				},
			},
		},

		// 带动词的场景
		{
			name:    "verb_suffix",
			pattern: "/v1/messages/{message_id}:cancel",
			urls:    []string{"v1", "messages", "123"},
			verb:    "cancel",
			want: []PathFieldVar{{
				Fields: []string{"message_id"},
				Value:  "123",
			}},
		},
		{
			name:    "verb_not_match",
			pattern: "/v1/messages/{message_id}:cancel",
			urls:    []string{"v1", "messages", "123"},
			verb:    "delete",
			wantErr: true,
			errMsg:  "verb not match",
		},

		// 错误场景
		{
			name:    "url_too_short",
			pattern: "/v1/{name=projects/*/locations/*}",
			urls:    []string{"v1", "projects", "p1"},
			verb:    "get",
			wantErr: true,
			errMsg:  "url segments too short",
		},
		{
			name:    "segment_not_match",
			pattern: "/v1/{name=projects/*/locations/*}",
			urls:    []string{"v1", "wrong", "p1", "locations", "us"},
			verb:    "get",
			wantErr: true,
			errMsg:  "segment not match",
		},

		// 变量嵌套场景
		{
			name:    "nested_variables",
			pattern: "/v1/{parent=projects/*/locations/*}/datasets/{dataset}/tables/{table}",
			urls:    []string{"v1", "projects", "p1", "locations", "us", "datasets", "d1", "tables", "t1"},
			verb:    "get",
			want: []PathFieldVar{
				{
					Fields: []string{"parent"},
					Value:  "p1/us",
				},
				{
					Fields: []string{"dataset"},
					Value:  "d1",
				},
				{
					Fields: []string{"table"},
					Value:  "t1",
				},
			},
		},
		{
			name:    "complex_nested_variables",
			pattern: "/v1/{parent=projects/*/locations/*}/models/{model}/evaluations/{evaluation}/{slice=**}",
			urls:    []string{"v1", "projects", "p1", "locations", "us", "models", "m1", "evaluations", "e1", "s1", "s2", "s3"},
			verb:    "get",
			want: []PathFieldVar{
				{
					Fields: []string{"parent"},
					Value:  "p1/us",
				},
				{
					Fields: []string{"model"},
					Value:  "m1",
				},
				{
					Fields: []string{"evaluation"},
					Value:  "e1",
				},
				{
					Fields: []string{"slice"},
					Value:  "s1/s2/s3",
				},
			},
		},
		{
			name:    "nested_variables_too_short",
			pattern: "/v1/{parent=projects/*/locations/*}/datasets/{dataset}/tables/{table}",
			urls:    []string{"v1", "projects", "p1", "locations", "us", "datasets"},
			verb:    "get",
			wantErr: true,
			errMsg:  "url segments too short",
		},
		{
			name:    "nested_variables_segment_not_match",
			pattern: "/v1/{parent=projects/*/locations/*}/datasets/{dataset}/tables/{table}",
			urls:    []string{"v1", "projects", "p1", "locations", "us", "wrong", "d1", "tables", "t1"},
			verb:    "get",
			wantErr: true,
			errMsg:  "segment not match",
		},

		// 点分隔变量场景
		{
			name:    "dotted_variable",
			pattern: "/v1/{resource.name}",
			urls:    []string{"v1", "test"},
			verb:    "get",
			want: []PathFieldVar{
				{
					Fields: []string{"resource", "name"},
					Value:  "test",
				},
			},
		},
		{
			name:    "multiple_dotted_variable",
			pattern: "/v1/{resource.path.name=messages/*}/items/{item.id}",
			urls:    []string{"v1", "messages", "123", "items", "456"},
			verb:    "get",
			want: []PathFieldVar{
				{
					Fields: []string{"resource", "path", "name"},
					Value:  "123",
				},
				{
					Fields: []string{"item", "id"},
					Value:  "456",
				},
			},
		},
		{
			name:    "nested_dotted_variable",
			pattern: "/v1/{parent.resource=projects/*/locations/*}/datasets/{dataset.name}/tables/{table.id}",
			urls:    []string{"v1", "projects", "p1", "locations", "us", "datasets", "d1", "tables", "t1"},
			verb:    "get",
			want: []PathFieldVar{
				{
					Fields: []string{"parent", "resource"},
					Value:  "p1/us",
				},
				{
					Fields: []string{"dataset", "name"},
					Value:  "d1",
				},
				{
					Fields: []string{"table", "id"},
					Value:  "t1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern, err := ParsePattern(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}

			got, err := pattern.Match(tt.urls, tt.verb)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
