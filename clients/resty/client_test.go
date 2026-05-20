package resty

import (
	"net/http"
	"testing"
	"time"

	"github.com/pubgo/funk/v2/retry"
	"github.com/stretchr/testify/assert"
)

// TestRequestCreation 测试请求创建
func TestRequestCreation(t *testing.T) {
	// 创建请求规范
	reqSpec := &RequestSpec{
		Path:        "/api/users/{id}",
		Method:      http.MethodGet,
		ContentType: "application/json",
		Header:      map[string]string{"X-Test": "value"},
	}

	// 创建请求
	req := NewRequest(reqSpec)
	assert.NotNil(t, req)
	assert.Equal(t, reqSpec, req.cfg)
	assert.NotNil(t, req.header)
	assert.NotNil(t, req.query)
	assert.NotNil(t, req.params)
}

// TestRequestBuilding 测试请求构建
func TestRequestBuilding(t *testing.T) {
	// 创建请求规范
	reqSpec := &RequestSpec{
		Path:   "/api/users/{id}",
		Method: http.MethodGet,
	}

	// 创建请求并设置各种参数
	req := NewRequest(reqSpec).
		SetParam("id", "123").
		SetQuery(map[string]string{"name": "test", "age": "20"}).
		SetHeader("Authorization", "Bearer token123").
		AddHeader("X-Additional", "value").
		SetBody(map[string]string{"key": "value"}).
		SetBackoff(retry.NewConstant(5 * time.Millisecond))

	assert.NotNil(t, req)
	assert.Equal(t, "123", req.params["id"])
	assert.Equal(t, "test", req.query.Get("name"))
	assert.Equal(t, "20", req.query.Get("age"))
	assert.Equal(t, "Bearer token123", req.header.Get("Authorization"))
	assert.Equal(t, "value", req.header.Get("X-Additional"))
	assert.NotNil(t, req.body)
	assert.NotNil(t, req.backoff)
}

// TestRequestSpecCreateRequest 测试RequestSpec的CreateRequest方法
func TestRequestSpecCreateRequest(t *testing.T) {
	reqSpec := RequestSpec{
		Path:   "/api/test",
		Method: http.MethodPost,
		Header: map[string]string{"X-Test": "value"},
	}

	req := reqSpec.CreateRequest()
	assert.NotNil(t, req)
	assert.Equal(t, reqSpec.Path, req.cfg.Path)
	assert.Equal(t, reqSpec.Method, req.cfg.Method)
}
