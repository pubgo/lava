# resty

基于 fasthttp 的 HTTP 客户端实现，支持中间件、重试、认证等功能。

## 主要特性

- 基于 fasthttp 的高性能 HTTP 客户端
- 支持中间件（访问日志、指标、恢复等）
- 支持重试机制
- 支持代理设置
- 支持 Basic Token 和 JWT Token 认证
- 丰富的配置选项（超时、连接池等）

## 参考项目

- [https://github.com/go-resty/resty](https://github.com/go-resty/resty)
- [https://github.com/imroc/req](https://github.com/imroc/req)
- [https://github.com/sony/gobreaker](https://github.com/sony/gobreaker)

## 安装

```bash
go get github.com/pubgo/lava/v2/clients/resty
```

## 使用示例

### 基本用法

```go
import (
    "context"
    "net/http"
    "github.com/pubgo/funk/v2/log"
    "github.com/pubgo/funk/v2/metrics"
    "github.com/pubgo/lava/v2/clients/resty"
)

func main() {
    client := resty.New(&resty.Config{
        BaseUrl:     "https://api.example.com",
        ServiceName: "example-api",
    }, resty.Params{
        Log:    log.Default(),
        Metric: metrics.Default(),
    })

    req := resty.NewRequest(&resty.RequestSpec{
        Path:   "/users/{id}",
        Method: http.MethodGet,
    }).SetParam("id", "123").SetQuery(map[string]string{
        "name": "test",
    })

    resp, err := client.Do(context.Background(), req).Unwrap()
    if err != nil {
        log.Fatal(err)
    }
}
```

### 高级用法（带中间件）

```go
import (
    "context"
    "github.com/pubgo/funk/v2/log"
    "github.com/pubgo/funk/v2/metrics"
    "github.com/pubgo/lava/v2/pkg/lava"
    "github.com/pubgo/lava/v2/clients/resty"
)

func main() {
    customMiddleware := func(next lava.HandlerFunc) lava.HandlerFunc {
        return func(ctx context.Context, req lava.Request) (lava.Response, error) {
            // 前置处理
            resp, err := next(ctx, req)
            // 后置处理
            return resp, err
        }
    }

    client := resty.New(&resty.Config{
        BaseUrl:     "https://api.example.com",
        ServiceName: "example-api",
    }, resty.Params{
        Log:    log.Default(),
        Metric: metrics.Default(),
    }, customMiddleware)
}
```

### 带认证的用法

```go
import (
    "context"
    "net/http"
    "github.com/pubgo/funk/v2/log"
    "github.com/pubgo/funk/v2/metrics"
    "github.com/pubgo/lava/v2/clients/resty"
)

func main() {
    client := resty.New(&resty.Config{
        BaseUrl:     "https://api.example.com",
        ServiceName: "example-api",
        EnableAuth:  true,
        JwtToken:    "your-jwt-token",
    }, resty.Params{
        Log:    log.Default(),
        Metric: metrics.Default(),
    })

    req := resty.NewRequest(&resty.RequestSpec{
        Path:   "/protected",
        Method: http.MethodGet,
    })

    resp, err := client.Do(context.Background(), req).Unwrap()
    if err != nil {
        log.Fatal(err)
    }
}
```

### 带重试的用法

```go
import (
    "context"
    "net/http"
    "time"
    "github.com/pubgo/funk/v2/log"
    "github.com/pubgo/funk/v2/metrics"
    "github.com/pubgo/lava/v2/clients/resty"
)

func main() {
    client := resty.New(&resty.Config{
        BaseUrl:              "https://api.example.com",
        ServiceName:          "example-api",
        DefaultRetryCount:    3,
        DefaultRetryInterval: 100 * time.Millisecond,
    }, resty.Params{
        Log:    log.Default(),
        Metric: metrics.Default(),
    })

    req := resty.NewRequest(&resty.RequestSpec{
        Path:   "/flaky",
        Method: http.MethodGet,
    })

    resp, err := client.Do(context.Background(), req).Unwrap()
    if err != nil {
        log.Fatal(err)
    }
}
```

## 配置选项

| 配置项 | 类型 | 说明 | 默认值 |
|-------|------|------|--------|
| BaseUrl | string | 基础 URL | - |
| ServiceName | string | 服务名称 | - |
| DefaultHeader | map[string]string | 默认请求头 | - |
| DefaultContentType | string | 默认内容类型 | - |
| DefaultRetryCount | uint32 | 默认重试次数 | 0 |
| DefaultRetryInterval | time.Duration | 默认重试间隔 | 0 |
| BasicToken | string | Basic 认证令牌 | - |
| JwtToken | string | JWT 认证令牌 | - |
| EnableProxy | bool | 是否启用代理 | false |
| EnableAuth | bool | 是否启用认证 | false |
| DialTimeout | time.Duration | 拨号超时 | 5s |
| ReadTimeout | time.Duration | 读取超时 | 10s |
| WriteTimeout | time.Duration | 写入超时 | 10s |
| MaxConnsPerHost | int | 每个主机的最大连接数 | 512 |
| MaxIdleConnDuration | time.Duration | 最大空闲连接时长 | 10s |
| MaxIdemponentCallAttempts | int | 最大幂等调用尝试次数 | 5 |
| ReadBufferSize | int | 读取缓冲区大小 | 4096 |
| WriteBufferSize | int | 写入缓冲区大小 | 4096 |
| MaxResponseBodySize | int | 最大响应体大小 | 2MB |

## 更多示例

查看 [examples](./examples) 目录获取更多使用示例。
