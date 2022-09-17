# [lava文档](https://www.yuque.com/pubgo/lava/readme)|[example](pkg/example/services)

> lava 是一个经过企业实践而抽象出来的微服务中台集成框架

1. 配置管理, 配置驱动开发, 本地配置, 配置中心, 配置变更

2. 对象管理, 通过依赖注入框架dix管理对象, 可结合配置中心动态变更, 对象无感变更

3. runtime抽象, cli, gin, grpc, task等服务抽象统一的entry, 统一使用习惯, 多服务子命令运行

4. plugin抽象, 结合运行顺序, 启动初始化, 配置变更, 对象管理, 对所有资源进行管理

5. 便捷protobuf生成管理, lava集成protoc命令去管理protobuf依赖,插件,版本和编译, 只需要一个protobuf.yaml配置文件

6. grpc自动注册, 通过自定义protoc-gen-lava管理grpc注册函数, 并自动识别服务的handler

7. 调试友好, 便捷的系统日志和业务日志集成, 丰富且详细的debug api可以查看系统的配置参数和细节

8. 自动生成swagger和http rest client文档, 方便测试和集成

9. tracing和metric自动集成, 通过reqId打通业务日志和tracing日志, 便于系统异常排查和跟踪

10. 统一抽象的middleware, gin和grpc的server以及client共享一套middleware抽象, 定义一套middleware作用于所有组件

11. 统一protobuf定义grpc和http服务, 便于生成swagger和sdk, 方便第三方调用
