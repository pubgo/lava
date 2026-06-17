// Package zrpc implements protobuf RPC over NATS.
//
// Unary calls use NATS request-reply. Streaming calls open a session with
// inbox subjects and exchange framed messages (open/ack/data/end/error).
//
// Logging and metrics are not emitted directly here. Pass lava.Middleware
// (typically accesslog/metric/recovery) to Client or Server so each request
// is recorded with request_id, subject, latency, and errors.
package zrpc
