# Lava Project Copilot Instructions

## Project Overview

Lava is a microservice middleware integration framework abstracted from enterprise practice. It provides a comprehensive set of tools and utilities for building microservices, including service management, tunnel system, scheduler, HTTP/gRPC servers, and various middleware components.

## Project Structure

```
├── clients/        # Client libraries (gRPC and HTTP clients)
├── cmds/           # Command-line tools
├── core/           # Core modules (supervisor, tunnel, scheduler, etc.)
├── docs/           # Documentation files
├── internal/       # Internal implementation
├── lava/           # Public interfaces
├── pkg/            # Public packages (gateway, cliutil, etc.)
├── proto/          # Protobuf definitions
├── servers/        # Server implementations (HTTP and gRPC)
├── tools/          # Development tools
├── go.mod          # Go module file
├── go.sum          # Go module checksum file
├── taskfile.yml    # Task configuration
└── README.md       # Project README
```

## Core Modules

### clients/
- **grpcc/**: gRPC client implementation with load balancing and middleware support
- **resty/**: HTTP client implementation with middleware support

### cmds/
- Various command-line tools like `lavacurl`, `configcmd`, `healthcmd`, etc.

### core/
- **supervisor/**: Service lifecycle management
- **tunnel/**: Service proxy and intranet penetration
- **scheduler/**: Task scheduling system
- **logging/**: Unified logging system
- **metrics/**: Unified metrics system
- **tracing/**: Distributed tracing system
- **encoding/**: Unified encoding system
- **discovery/**: Service discovery system
- **debug/**: Debugging utilities
- **lavabuilder/**: Dependency injection setup

### internal/
- **configs/**: Configuration files
- **examples/**: Example applications
- **middlewares/**: Internal middleware implementations

### lava/
- Public interfaces and core types

### pkg/
- **gateway/**: gRPC Gateway implementation
- **cliutil/**: Command-line utility functions
- **httputil/**: HTTP utility functions
- **grpcutil/**: gRPC utility functions
- **netutil/**: Network utility functions

### servers/
- **https/**: HTTP server implementation based on Fiber
- **grpcs/**: gRPC server implementation with gateway support

## Coding Style

### Go Code Style
- Follow the standard Go code style (use `go fmt`)
- Use 4 spaces for indentation
- Keep lines under 100 characters
- Use clear, descriptive variable and function names
- Use camelCase for variables and functions
- Use PascalCase for types and interfaces
- Use ALL_CAPS for constants

### Naming Conventions
- **Packages**: Lowercase, single-word names
- **Functions**: CamelCase, descriptive names
- **Variables**: CamelCase, short but descriptive
- **Types**: PascalCase, descriptive names
- **Constants**: ALL_CAPS, with underscores

### Error Handling
- Use `github.com/pkg/errors` for error wrapping
- Return errors as the last return value
- Check and handle all errors
- Provide clear, descriptive error messages

### Documentation
- Use GoDoc format for documentation
- Document all public functions, types, and packages
- Use clear, concise language
- Include examples where appropriate

## Architecture Principles

### Dependency Injection
- Use the `dix` framework for dependency injection
- Register dependencies in the `lavabuilder`
- Use constructor functions for creating services

### Middleware Pattern
- Use middleware for cross-cutting concerns
- Implement the `lava.Middleware` interface
- Chain middleware for HTTP and gRPC requests

### Configuration Management
- Use hierarchical YAML configuration
- Support environment variable overrides
- Use structured configuration types

### Service Lifecycle
- Register services with the `supervisor`
- Implement the `lava.Service` interface
- Handle graceful shutdown

## Common Patterns

### Creating a Service
```go
func New() lava.Service {
    return &MyService{}
}

type MyService struct{}

func (s *MyService) String() string {
    return "my-service"
}

func (s *MyService) Serve(ctx context.Context) error {
    // Service implementation
    <-ctx.Done()
    return nil
}
```

### Creating Middleware
```go
func New() lava.Middleware {
    return &MyMiddleware{}
}

type MyMiddleware struct{}

func (m *MyMiddleware) Handle(ctx context.Context, req interface{}) (interface{}, error) {
    // Middleware implementation
    return req, nil
}
```

### Registering Services
```go
func main() {
    di := lavabuilder.New()
    di.Provide(func() lava.Service {
        return MyService.New()
    })
    lavabuilder.Run(di)
}
```

### HTTP Server Setup
```go
https.New(https.Params{
    Handlers: []lava.HttpRouter{
        &UserRouter{},
    },
    Middlewares: []lava.Middleware{
        middleware.New(),
    },
})
```

### gRPC Server Setup
```go
grpcs.New(grpcs.Params{
    GrpcRouters: []lava.GrpcRouter{
        &UserService{},
    },
    Middlewares: []lava.Middleware{
        middleware.New(),
    },
})
```

## Project-Specific Tools

### Task Commands
- `task test`: Run all tests
- `task lint`: Run linting
- `task build`: Build the project
- `task clean`: Clean build artifacts
- `task proto:gen`: Generate Protobuf code

### Command-Line Tools
- `lavacurl`: HTTP/gRPC client for testing APIs
- `lava config`: Manage configuration
- `lava health`: Check service health
- `lava scheduler`: Manage scheduled tasks
- `lava tunnel`: Manage tunnel connections

## Common Dependencies

### Core Dependencies
- `github.com/gofiber/fiber/v3`: HTTP server framework
- `google.golang.org/grpc`: gRPC framework
- `github.com/pubgo/dix`: Dependency injection framework
- `github.com/pubgo/redant`: Command-line framework
- `github.com/prometheus/client_golang`: Metrics collection
- `github.com/opentracing/opentracing-go`: Distributed tracing
- `github.com/pkg/errors`: Error handling
- `github.com/spf13/viper`: Configuration management

### Testing Dependencies
- `github.com/stretchr/testify`: Testing assertions
- `github.com/golang/mock`: Mock generation

## Development Workflow

1. **Branch Management**: Use `main` for stable releases, `develop` for features, and feature branches for new work
2. **Commit Messages**: Use conventional commit messages (`feat:`, `fix:`, `docs:`, etc.)
3. **Testing**: Write comprehensive tests with at least 80% coverage
4. **Linting**: Run `golangci-lint` before committing
5. **Documentation**: Update documentation for new features and changes

## Best Practices

### Performance
- Use connection pooling for HTTP/gRPC clients
- Avoid unnecessary allocations
- Use appropriate data structures
- Implement caching where appropriate

### Security
- Use HTTPS for all communications
- Implement proper authentication and authorization
- Validate all user input
- Use secure coding practices

### Reliability
- Implement graceful shutdown
- Use timeouts for all network operations
- Handle errors properly
- Implement circuit breakers for external dependencies

### Maintainability
- Write clean, readable code
- Follow Go idioms and conventions
- Document all public APIs
- Use consistent naming and formatting

## Examples

### Simple HTTP Service
```go
package main

import (
    "github.com/pubgo/lava/v2/core/lavabuilder"
    "github.com/pubgo/lava/v2/servers/https"
    "github.com/pubgo/lava/v2/lava"
    "github.com/gofiber/fiber/v3"
)

// UserRouter handles user-related routes
type UserRouter struct{}

func (r *UserRouter) Prefix() string {
    return "/users"
}

func (r *UserRouter) Middlewares() []lava.Middleware {
    return nil
}

func (r *UserRouter) Router(router fiber.Router) {
    router.Get("/", r.GetUsers)
    router.Get("/:id", r.GetUser)
}

func (r *UserRouter) GetUsers(c *fiber.Ctx) error {
    return c.JSON([]map[string]interface{}{
        {"id": "1", "name": "John"},
        {"id": "2", "name": "Jane"},
    })
}

func (r *UserRouter) GetUser(c *fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(map[string]interface{}{
        "id":   id,
        "name": "John",
    })
}

func main() {
    di := lavabuilder.New()
    di.Provide(func() lava.Service {
        return https.New(https.Params{
            Handlers: []lava.HttpRouter{
                &UserRouter{},
            },
        })
    })
    lavabuilder.Run(di)
}
```

### gRPC Service with Gateway
```go
package main

import (
    "github.com/pubgo/lava/v2/core/lavabuilder"
    "github.com/pubgo/lava/v2/servers/grpcs"
    "github.com/pubgo/lava/v2/lava"
)

// UserService handles user-related gRPC requests
type UserService struct{}

func (s *UserService) ServiceDesc() *grpc.ServiceDesc {
    return &pb.UserService_ServiceDesc
}

func (s *UserService) Register(server grpc.ServiceRegistrar) {
    pb.RegisterUserServiceServer(server, s)
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    return &pb.GetUserResponse{
        User: &pb.User{
            Id:   req.GetUserId(),
            Name: "John",
        },
    }, nil
}

func main() {
    di := lavabuilder.New()
    di.Provide(func() lava.Service {
        return grpcs.New(grpcs.Params{
            GrpcRouters: []lava.GrpcRouter{
                &UserService{},
            },
        })
    })
    lavabuilder.Run(di)
}
```

## Conclusion

Lava is a comprehensive framework for building microservices, providing a wide range of tools and utilities to simplify development. By following the guidelines and best practices outlined in this document, you can build robust, scalable, and maintainable microservices using the Lava framework.

For more detailed information, refer to the project documentation in the `docs/` directory.
