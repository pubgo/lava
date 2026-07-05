package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pubgo/funk/v2/log"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/clients/resty"
	"github.com/pubgo/lava/v2/core/metrics"
	"github.com/pubgo/lava/v2/pkg/lava"
)

func main() {
	// 初始化日志和指标
	logger := log.GetLogger()
	metric := tally.NoopScope

	// 示例 1: 基本用法
	basicExample(logger, metric)

	// 示例 2: 带中间件的高级用法
	advancedExample(logger, metric)

	// 示例 3: 带认证的用法
	authExample(logger, metric)

	// 示例 4: 带重试的用法
	retryExample(logger, metric)
}

// 基本用法示例
func basicExample(logger log.Logger, metric metrics.Metric) {
	fmt.Println("=== 基本用法示例 ===")

	// 创建客户端
	client := resty.New(&resty.Config{
		BaseUrl:     "https://api.example.com",
		ServiceName: "example-api",
	}, resty.Params{
		Log:    logger,
		Metric: metric,
	})

	// 创建请求
	req := resty.NewRequest(&resty.RequestSpec{
		Path:   "/users/{id}",
		Method: http.MethodGet,
	}).SetParam("id", "123").SetQuery(map[string]string{
		"name": "test",
	})

	// 发送请求
	resp := client.Do(context.Background(), req).Unwrap()
	logger.Info().Int("status", resp.StatusCode()).Msg("请求成功")

	// 示例：使用 JSON 方法反序列化响应
	type User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var user User
	if err := resp.JSON(&user); err != nil {
		logger.Error().Err(err).Msg("JSON 反序列化失败")
	} else {
		logger.Info().Str("user", user.Name).Msg("JSON 反序列化成功")
	}
}

// 高级用法示例（带中间件）
func advancedExample(logger log.Logger, metric metrics.Metric) {
	fmt.Println("\n=== 高级用法示例（带中间件）===")

	// 自定义中间件
	customMiddleware := lava.WithMiddleware("custom", func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (lava.Response, error) {
			// 前置处理
			logger.Info().Str("endpoint", req.Endpoint()).Msg("请求开始")

			// 调用下一个中间件
			resp, err := next(ctx, req)

			// 后置处理
			if err != nil {
				logger.Error().Err(err).Msg("请求失败")
			} else {
				logger.Info().Msg("请求结束")
			}

			return resp, err
		}
	})

	// 创建带自定义中间件的客户端
	client := resty.New(&resty.Config{
		BaseUrl:     "https://api.example.com",
		ServiceName: "example-api",
	}, resty.Params{
		Log:    logger,
		Metric: metric,
	}, customMiddleware)

	// 创建请求
	req := resty.NewRequest(&resty.RequestSpec{
		Path:   "/users",
		Method: http.MethodPost,
	})

	// 发送请求
	resp := client.Do(context.Background(), req).Unwrap()
	logger.Info().Int("status", resp.StatusCode()).Msg("请求成功")
}

// 带认证的用法示例
func authExample(logger log.Logger, metric metrics.Metric) {
	fmt.Println("\n=== 带认证的用法示例 ===")

	// 创建带认证的客户端
	client := resty.New(&resty.Config{
		BaseUrl:     "https://api.example.com",
		ServiceName: "example-api",
		EnableAuth:  true,
		JwtToken:    "your-jwt-token",
	}, resty.Params{
		Log:    logger,
		Metric: metric,
	})

	// 创建请求
	req := resty.NewRequest(&resty.RequestSpec{
		Path:   "/protected",
		Method: http.MethodGet,
	})

	// 发送请求
	resp := client.Do(context.Background(), req).Unwrap()
	logger.Info().Int("status", resp.StatusCode()).Msg("请求成功")
}

// 带重试的用法示例
func retryExample(logger log.Logger, metric metrics.Metric) {
	fmt.Println("\n=== 带重试的用法示例 ===")

	// 创建带重试的客户端
	client := resty.New(&resty.Config{
		BaseUrl:              "https://api.example.com",
		ServiceName:          "example-api",
		DefaultRetryCount:    3,
		DefaultRetryInterval: 100 * time.Millisecond,
	}, resty.Params{
		Log:    logger,
		Metric: metric,
	})

	// 创建请求
	req := resty.NewRequest(&resty.RequestSpec{
		Path:   "/flaky",
		Method: http.MethodGet,
	})

	// 发送请求
	resp := client.Do(context.Background(), req).Unwrap()
	logger.Info().Int("status", resp.StatusCode()).Msg("请求成功")
}
