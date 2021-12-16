package main

// TODO grpc proxy

// 外部服务调用grpc服务的模式
// 网关模式 gateway api => 服务端直接提供restful服务
//	走http模式, 可以让所有服务直接调用
//	也可以暴露graphql模式
// 代理模式 proxy => 通过sidecar的方式访问后端服务
//	针对php服务的调用比较合适
// 集成模式 integrate => 把能力集成到上游服务当中, 和上游服务一起启动
//	原生grpc服务, 一般采用集成模式

// https://github.com/mercari/grpc-http-proxy
// https://github.com/mwitkow/grpc-proxy
// https://github.com/bradleyjkemp/grpc-tools/tree/master/grpc-proxy
