// Package zrpcs hosts zrpc services on NATS with Lava supervisor integration.
//
// It wires default middlewares (serviceinfo/metric/accesslog/recovery) and emits
// lifecycle logs on start/stop. Business RPC logs come from accesslog middleware.
package zrpcs
