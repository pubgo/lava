package resty

// Package resty 提供了一个基于 fasthttp 的 HTTP 客户端实现，支持中间件、重试、认证等功能。
//
// 主要特性：
// - 基于 fasthttp 的高性能 HTTP 客户端
// - 支持中间件（访问日志、指标、恢复等）
// - 支持重试机制
// - 支持代理设置
// - 支持 Basic Token 和 JWT Token 认证
// - 丰富的配置选项（超时、连接池等）
//
// 参考项目：
// - https://github.com/go-resty/resty
// - https://github.com/imroc/req
// - https://github.com/sony/gobreaker
//
// 示例：
//
// 基本用法：
//
//  client := resty.New(&resty.Config{
//      BaseUrl:     "https://api.example.com",
//      ServiceName: "example-api",
//  }, resty.Params{
//      Log:    log.Default(),
//      Metric: metrics.Default(),
//  })
//
//  req := resty.NewRequest(&resty.RequestSpec{
//      Path:   "/users/{id}",
//      Method: http.MethodGet,
//  }).SetParam("id", "123").SetQuery(map[string]string{
//      "name": "test",
//  })
//
//  resp, err := client.Do(context.Background(), req).Unwrap()
//  if err != nil {
//      log.Fatal(err)
//  }
//
// 高级用法（带中间件）：
//
//  customMiddleware := func(next lava.HandlerFunc) lava.HandlerFunc {
//      return func(ctx context.Context, req lava.Request) (lava.Response, error) {
//          // 前置处理
//          resp, err := next(ctx, req)
//          // 后置处理
//          return resp, err
//      }
//  }
//
//  client := resty.New(&resty.Config{
//      BaseUrl:     "https://api.example.com",
//      ServiceName: "example-api",
//  }, resty.Params{
//      Log:    log.Default(),
//      Metric: metrics.Default(),
//  }, customMiddleware)

