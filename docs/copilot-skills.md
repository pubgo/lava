# GitHub Copilot Skills for Lava Project

## 1. Copilot 基础技能

### 1.1 代码补全

**功能说明**：GitHub Copilot 可以根据上下文自动补全代码，包括函数、变量、表达式等。

**使用技巧**：
- 在编写代码时，只需输入函数名或变量名的前几个字符，Copilot 会自动补全
- 在编写注释时，Copilot 会根据注释内容生成相应的代码
- 使用 `Tab` 键接受建议，`Esc` 键拒绝建议

**示例**：

```go
// 输入注释
// Get user by ID
func GetUser(id string) (*User, error) {
    // Copilot 会自动生成函数体
}

// 输入函数名前缀
func GetU // Copilot 会补全为 GetUser
```

### 1.2 代码生成

**功能说明**：GitHub Copilot 可以根据自然语言描述生成完整的代码块。

**使用技巧**：
- 使用详细的注释描述你想要的功能
- 包括输入参数、返回值、处理逻辑等信息
- Copilot 会根据注释生成相应的代码

**示例**：

```go
// Generate a function that gets a user by ID from the database
// Parameters:
//   id: user ID
// Returns:
//   *User: user object
//   error: error if any
func GetUserFromDatabase(id string) (*User, error) {
    // Copilot 会生成完整的函数实现
}
```

### 1.3 代码重构

**功能说明**：GitHub Copilot 可以帮助重构代码，包括提取函数、重命名变量、优化代码结构等。

**使用技巧**：
- 选择要重构的代码块
- 使用 `Ctrl+I` (Windows/Linux) 或 `Cmd+I` (macOS) 打开 Copilot 面板
- 输入重构指令，如 "Extract this into a function"

**示例**：

```go
// 选择以下代码块
if err != nil {
    log.Printf("Error: %v", err)
    return nil, err
}

// 输入指令 "Extract this into a function"
// Copilot 会生成错误处理函数
func handleError(err error) error {
    log.Printf("Error: %v", err)
    return err
}
```

## 2. Copilot 高级技能

### 2.1 测试生成

**功能说明**：GitHub Copilot 可以根据现有代码生成相应的测试代码。

**使用技巧**：
- 打开测试文件（如 `user_test.go`）
- 输入测试函数的注释或函数名
- Copilot 会生成完整的测试代码

**示例**：

```go
// Test GetUser function
func TestGetUser(t *testing.T) {
    // Copilot 会生成测试代码
}
```

### 2.2 文档生成

**功能说明**：GitHub Copilot 可以根据代码生成相应的文档。

**使用技巧**：
- 在函数或类型定义前输入注释
- Copilot 会生成详细的文档注释

**示例**：

```go
// Copilot 会为以下函数生成文档注释
func GetUser(id string) (*User, error) {
    // 函数实现
}
```

### 2.3 代码解释

**功能说明**：GitHub Copilot 可以解释现有代码的功能和逻辑。

**使用技巧**：
- 选择要解释的代码块
- 使用 `Ctrl+I` (Windows/Linux) 或 `Cmd+I` (macOS) 打开 Copilot 面板
- 输入指令 "Explain this code"

**示例**：

```go
// 选择以下代码
func GetUser(id string) (*User, error) {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

// 输入指令 "Explain this code"
// Copilot 会解释代码功能
```

## 3. Lava 项目特定技能

### 3.1 服务创建

**功能说明**：在 Lava 项目中，Copilot 可以帮助创建符合项目架构的服务。

**使用技巧**：
- 输入服务创建的注释
- 包括服务类型、依赖项、实现逻辑等信息
- Copilot 会生成符合 Lava 架构的服务代码

**示例**：

```go
// Create a new HTTP service with user routes
// Includes:
//   - UserRouter with GetUsers and GetUser methods
//   - Service registration with lavabuilder
func NewUserService() lava.Service {
    // Copilot 会生成完整的服务实现
}
```

### 3.2 中间件创建

**功能说明**：在 Lava 项目中，Copilot 可以帮助创建符合项目架构的中间件。

**使用技巧**：
- 输入中间件创建的注释
- 包括中间件功能、实现逻辑等信息
- Copilot 会生成符合 Lava 架构的中间件代码

**示例**：

```go
// Create a logging middleware for HTTP requests
// Includes:
//   - Logs request method, path, and duration
//   - Implements lava.Middleware interface
func NewLoggingMiddleware() lava.Middleware {
    // Copilot 会生成完整的中间件实现
}
```

### 3.3 配置管理

**功能说明**：在 Lava 项目中，Copilot 可以帮助创建符合项目架构的配置管理代码。

**使用技巧**：
- 输入配置管理的注释
- 包括配置结构、加载逻辑等信息
- Copilot 会生成符合 Lava 架构的配置管理代码

**示例**：

```go
// Create a configuration struct for user service
// Includes:
//   - Database connection string
//   - API port
//   - Log level
func NewUserConfig() *UserConfig {
    // Copilot 会生成完整的配置结构和加载逻辑
}
```

### 3.4 gRPC 服务创建

**功能说明**：在 Lava 项目中，Copilot 可以帮助创建符合项目架构的 gRPC 服务。

**使用技巧**：
- 输入 gRPC 服务创建的注释
- 包括服务定义、方法实现等信息
- Copilot 会生成符合 Lava 架构的 gRPC 服务代码

**示例**：

```go
// Create a gRPC service for user management
// Includes:
//   - UserService with GetUser and ListUsers methods
//   - Service registration with grpcs server
func NewUserGrpcService() lava.GrpcRouter {
    // Copilot 会生成完整的 gRPC 服务实现
}
```

## 4. Copilot 与 Lava 工具集成

### 4.1 与 Task 集成

**功能说明**：GitHub Copilot 可以帮助创建和修改 Task 配置文件。

**使用技巧**：
- 在 `taskfile.yml` 文件中，输入任务的名称和描述
- Copilot 会生成相应的任务配置

**示例**：

```yaml
# Add a new task for running integration tests
# Includes:
#   - Runs integration tests with race detection
#   - Generates coverage report
integration-test:
  # Copilot 会生成完整的任务配置
```

### 4.2 与 Protobuf 集成

**功能说明**：GitHub Copilot 可以帮助创建和修改 Protobuf 文件。

**使用技巧**：
- 在 `.proto` 文件中，输入服务或消息的定义
- Copilot 会生成相应的 Protobuf 代码

**示例**：

```protobuf
// Define a user service with GetUser and ListUsers methods
// Includes:
//   - User message with id, name, and email fields
//   - GetUserRequest and GetUserResponse messages
//   - ListUsersRequest and ListUsersResponse messages
service UserService {
  // Copilot 会生成完整的服务定义
}
```

### 4.3 与 Docker 集成

**功能说明**：GitHub Copilot 可以帮助创建和修改 Docker 配置文件。

**使用技巧**：
- 在 `Dockerfile` 文件中，输入基础镜像和构建步骤
- Copilot 会生成相应的 Docker 配置

**示例**：

```dockerfile
# Create a Dockerfile for lava service
# Includes:
#   - Uses golang:1.25 as base image
#   - Builds the service
#   - Runs the service
FROM golang:1.25-alpine AS builder
# Copilot 会生成完整的 Dockerfile
```

## 5. Copilot 最佳实践

### 5.1 编写清晰的注释

**功能说明**：清晰的注释可以帮助 Copilot 更好地理解你的意图，生成更准确的代码。

**最佳实践**：
- 使用详细的注释描述你想要的功能
- 包括输入参数、返回值、处理逻辑等信息
- 使用自然语言，避免使用过于技术性的术语

**示例**：

```go
// Good: Detailed comment
// Get user by ID from the database
// Parameters:
//   id: user ID
// Returns:
//   *User: user object
//   error: error if any
func GetUser(id string) (*User, error) {
    // Copilot 会生成更准确的代码
}

// Bad: Vague comment
// Get user
func GetUser(id string) (*User, error) {
    // Copilot 可能生成不准确的代码
}
```

### 5.2 使用类型提示

**功能说明**：明确的类型提示可以帮助 Copilot 生成更准确的代码。

**最佳实践**：
- 为函数参数和返回值指定明确的类型
- 为变量指定明确的类型
- 使用结构体和接口定义数据模型

**示例**：

```go
// Good: With type hints
func GetUser(id string) (*User, error) {
    // Copilot 会生成更准确的代码
}

// Bad: Without type hints
func GetUser(id) {
    // Copilot 可能生成不准确的代码
}
```

### 5.3 分步骤生成代码

**功能说明**：复杂的功能可以分步骤生成，每一步生成一部分代码，然后再生成下一部分。

**最佳实践**：
- 先生成函数签名和注释
- 然后生成函数体的框架
- 最后生成具体的实现细节

**示例**：

```go
// Step 1: Generate function signature
func GetUser(id string) (*User, error) {
}

// Step 2: Generate function body framework
func GetUser(id string) (*User, error) {
    // Get user from database
    user, err := db.GetUser(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

// Step 3: Generate detailed implementation
func GetUser(id string) (*User, error) {
    // Validate ID
    if id == "" {
        return nil, errors.New("invalid user ID")
    }
    
    // Get user from database
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, errors.Wrap(err, "failed to get user")
    }
    
    // Check if user exists
    if user == nil {
        return nil, errors.New("user not found")
    }
    
    return user, nil
}
```

### 5.4 验证生成的代码

**功能说明**：生成的代码可能存在错误或不符合项目要求，需要验证和修改。

**最佳实践**：
- 检查生成的代码是否符合项目的编码风格
- 检查生成的代码是否符合项目的架构要求
- 检查生成的代码是否存在语法错误或逻辑错误
- 运行测试验证生成的代码是否正确

**示例**：

```go
// Generated code
func GetUser(id string) (*User, error) {
    user, err := db.GetUser(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

// After validation and modification
func GetUser(id string) (*User, error) {
    if id == "" {
        return nil, errors.New("invalid user ID")
    }
    
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, errors.Wrap(err, "failed to get user")
    }
    
    if user == nil {
        return nil, errors.New("user not found")
    }
    
    return user, nil
}
```

## 6. Copilot 快捷键

### 6.1 基本快捷键

| 快捷键 | 功能 |
|--------|------|
| `Tab` | 接受建议 |
| `Esc` | 拒绝建议 |
| `Ctrl+Enter` (Windows/Linux) 或 `Cmd+Enter` (macOS) | 打开 Copilot 面板 |
| `Ctrl+I` (Windows/Linux) 或 `Cmd+I` (macOS) | 显示内联建议 |

### 6.2 高级快捷键

| 快捷键 | 功能 |
|--------|------|
| `Ctrl+Shift+P` (Windows/Linux) 或 `Cmd+Shift+P` (macOS) | 打开命令面板 |
| `Ctrl+K Ctrl+I` (Windows/Linux) 或 `Cmd+K Cmd+I` (macOS) | 显示代码提示 |
| `Alt+Shift+F` (Windows/Linux) 或 `Option+Shift+F` (macOS) | 格式化代码 |

## 7. Copilot 常见问题

### 7.1 建议质量不高

**问题**：Copilot 生成的建议质量不高，不符合项目要求。

**解决方案**：
- 提供更详细的注释和上下文
- 使用更具体的变量名和函数名
- 分步骤生成代码
- 参考项目中已有的代码风格和架构

### 7.2 建议与现有代码冲突

**问题**：Copilot 生成的建议与项目中已有的代码冲突。

**解决方案**：
- 检查项目中已有的代码和架构
- 提供更具体的注释说明你的需求
- 手动修改生成的代码以适应项目要求

### 7.3 建议速度慢

**问题**：Copilot 生成建议的速度很慢。

**解决方案**：
- 确保网络连接稳定
- 减少同时打开的文件数量
- 减少代码文件的大小
- 分步骤生成代码，每一步生成一部分

### 7.4 建议不相关

**问题**：Copilot 生成的建议与当前上下文不相关。

**解决方案**：
- 提供更详细的注释和上下文
- 明确指定你想要的功能和实现方式
- 参考项目中已有的代码风格和架构
- 手动修改生成的代码以适应项目要求

## 8. Copilot 与 Lava 项目集成

### 8.1 了解项目结构

**功能说明**：了解 Lava 项目的结构可以帮助 Copilot 生成更符合项目要求的代码。

**最佳实践**：
- 熟悉项目的目录结构和文件组织
- 了解项目的核心模块和功能
- 参考项目中已有的代码风格和架构

**项目结构**：

```
├── clients/        # Client libraries
├── cmds/           # Command-line tools
├── core/           # Core modules
├── docs/           # Documentation
├── internal/       # Internal implementation
├── lava/           # Public interfaces
├── pkg/            # Public packages
├── proto/          # Protobuf definitions
├── servers/        # Server implementations
└── tools/          # Development tools
```

### 8.2 参考现有代码

**功能说明**：参考项目中已有的代码可以帮助 Copilot 生成更符合项目要求的代码。

**最佳实践**：
- 查看项目中已有的类似功能的代码
- 了解项目中使用的设计模式和架构
- 参考项目中已有的命名约定和编码风格

**示例**：

```go
// Reference existing code in the project
// For example, look at how other services are implemented
// Then ask Copilot to generate similar code
```

### 8.3 遵循项目规范

**功能说明**：遵循项目的规范可以帮助 Copilot 生成更符合项目要求的代码。

**最佳实践**：
- 遵循项目的命名约定和编码风格
- 遵循项目的架构要求和设计模式
- 使用项目中已有的工具和库

**示例**：

```go
// Follow project conventions
// For example, use the same error handling pattern as other parts of the project
func GetUser(id string) (*User, error) {
    user, err := db.GetUser(id)
    if err != nil {
        return nil, errors.Wrap(err, "failed to get user")
    }
    return user, nil
}
```

## 9. 总结

GitHub Copilot 是一个强大的 AI 辅助开发工具，可以帮助你在 Lava 项目中更高效地编写代码。通过使用本文档中介绍的技能和技巧，你可以：

1. **提高开发效率**：自动补全和生成代码，减少重复工作
2. **改善代码质量**：生成符合项目规范的代码，减少错误
3. **学习项目架构**：通过参考现有代码和生成的建议，了解项目的架构和设计模式
4. **探索新功能**：快速原型设计和探索新功能，加速开发过程

记住，GitHub Copilot 是一个辅助工具，不是替代品。你仍然需要：

1. **验证生成的代码**：检查生成的代码是否符合项目要求和最佳实践
2. **理解生成的代码**：确保你理解生成的代码的功能和逻辑
3. **修改生成的代码**：根据项目要求和具体情况修改生成的代码
4. **测试生成的代码**：运行测试验证生成的代码是否正确

通过合理使用 GitHub Copilot，你可以在 Lava 项目中更高效地开发，同时保持代码的质量和可维护性。