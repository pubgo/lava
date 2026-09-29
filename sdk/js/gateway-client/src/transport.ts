import type { RpcOptions, RpcTransport } from "@protobuf-ts/runtime-rpc";

import { createGrpcWebTransport, type GrpcWebTransportOptions } from "./protobuf-ts-transport.js";
import { createJsonRpcTransport, type JsonRpcTransportOptions } from "./json-rpc-transport.js";

/**
 * Transport formats aligned with real lava frontends (see agentrun `rpc.ts`):
 *
 * - `json`   — protojson over `POST {baseUrl}/{package.Service}/{Method}` (dev-friendly)
 * - `binary` — `application/grpc-web+proto` frames (production)
 * - `text`   — `application/grpc-web-text+proto` (base64 framed body)
 */
export type GatewayTransportFormat = "json" | "binary" | "text";

export type CreateGatewayTransportOptions = {
  /** e.g. `https://api.example.com` or `/api` */
  baseUrl: string;
  format?: GatewayTransportFormat;
  /** Enable gzip accept/compress for binary/text gRPC-Web. */
  acceptCompression?: boolean;
  headers?: HeadersInit;
  fetch?: typeof fetch;
  defaultOptions?: RpcOptions;
  fetchInit?: Omit<RequestInit, "body" | "headers" | "method" | "signal">;
};

/**
 * Single factory used by apps that switch transports the way agentrun does.
 */
export function createGatewayTransport(opts: CreateGatewayTransportOptions): RpcTransport {
  const format = opts.format ?? "json";
  switch (format) {
    case "json":
      return createJsonRpcTransport({
        baseUrl: opts.baseUrl,
        fetchInit: opts.fetchInit,
        ...(opts.defaultOptions ?? {}),
      } satisfies JsonRpcTransportOptions);
    case "text":
      return createGrpcWebTransport({
        baseUrl: opts.baseUrl,
        acceptCompression: opts.acceptCompression,
        headers: opts.headers,
        fetch: opts.fetch,
        defaultOptions: opts.defaultOptions,
        format: "text",
      } satisfies GrpcWebTransportOptions);
    case "binary":
    default:
      return createGrpcWebTransport({
        baseUrl: opts.baseUrl,
        acceptCompression: opts.acceptCompression,
        headers: opts.headers,
        fetch: opts.fetch,
        defaultOptions: opts.defaultOptions,
        format: "binary",
      } satisfies GrpcWebTransportOptions);
  }
}
