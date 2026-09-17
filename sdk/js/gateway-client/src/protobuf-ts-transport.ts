import type {
  MethodInfo,
  RpcOptions,
  RpcTransport,
  UnaryCall,
  ServerStreamingCall,
  ClientStreamingCall,
  DuplexStreamingCall,
} from "@protobuf-ts/runtime-rpc";
import { RpcError } from "@protobuf-ts/runtime-rpc";

import { GatewayError } from "./errors.js";
import { createGrpcWebClient, type GrpcWebClientOptions } from "./grpc-web.js";

export type GrpcWebTransportOptions = GrpcWebClientOptions & {
  /** Default RpcOptions merged into every call. */
  defaultOptions?: RpcOptions;
};

const GRPC_CODE_NAMES = [
  "OK",
  "CANCELLED",
  "UNKNOWN",
  "INVALID_ARGUMENT",
  "DEADLINE_EXCEEDED",
  "NOT_FOUND",
  "ALREADY_EXISTS",
  "PERMISSION_DENIED",
  "RESOURCE_EXHAUSTED",
  "FAILED_PRECONDITION",
  "ABORTED",
  "OUT_OF_RANGE",
  "UNIMPLEMENTED",
  "INTERNAL",
  "UNAVAILABLE",
  "DATA_LOSS",
  "UNAUTHENTICATED",
] as const;

function statusName(code: number): string {
  return GRPC_CODE_NAMES[code] ?? "UNKNOWN";
}

function toRpcError(err: unknown): RpcError {
  if (err instanceof GatewayError) {
    return new RpcError(err.grpcMessage || err.message, statusName(err.code));
  }
  if (err instanceof RpcError) return err;
  if (err instanceof Error) return new RpcError(err.message, "INTERNAL");
  return new RpcError(String(err), "INTERNAL");
}

function headersToMeta(h: Headers): Record<string, string> {
  const out: Record<string, string> = {};
  h.forEach((v, k) => {
    out[k] = v;
  });
  return out;
}

function metaToHeaders(meta: RpcOptions["meta"]): Headers {
  const headers = new Headers();
  if (!meta) return headers;
  for (const [k, v] of Object.entries(meta)) {
    if (v == null) continue;
    if (Array.isArray(v)) v.forEach((x) => headers.append(k, String(x)));
    else headers.set(k, String(v));
  }
  return headers;
}

class Deferred<T> {
  readonly promise: Promise<T>;
  private _resolve!: (v: T) => void;
  private _reject!: (e: unknown) => void;
  constructor() {
    this.promise = new Promise<T>((resolve, reject) => {
      this._resolve = resolve;
      this._reject = reject;
    });
  }
  resolve(v: T) {
    this._resolve(v);
  }
  reject(e: unknown) {
    this._reject(e);
  }
}

/**
 * protobuf-ts {@link RpcTransport} backed by lava gateway gRPC-Web framing.
 */
export function createGrpcWebTransport(opts: GrpcWebTransportOptions): RpcTransport {
  const client = createGrpcWebClient(opts);

  return {
    mergeOptions(options?: RpcOptions): RpcOptions {
      return { ...(opts.defaultOptions ?? {}), ...(options ?? {}) };
    },

    unary<I extends object, O extends object>(
      method: MethodInfo<I, O>,
      input: I,
      options: RpcOptions,
    ): UnaryCall<I, O> {
      const fullMethod = `/${method.service.typeName}/${method.name}`;
      const headersDef = new Deferred<Record<string, string>>();
      const trailersDef = new Deferred<Record<string, string>>();
      const statusDef = new Deferred<{ code: string; detail: string }>();
      const responseDef = new Deferred<O>();

      void (async () => {
        try {
          const result = await client.unary(fullMethod, method.I.toBinary(input), {
            headers: metaToHeaders(options.meta),
            signal: options.abort,
            compress: opts.acceptCompression,
          });
          headersDef.resolve(headersToMeta(result.headers));
          trailersDef.resolve(headersToMeta(result.trailers));
          statusDef.resolve({ code: statusName(result.grpcStatus), detail: result.grpcMessage });
          responseDef.resolve(method.O.fromBinary(result.message));
        } catch (err) {
          const rpcErr = toRpcError(err);
          headersDef.reject(rpcErr);
          trailersDef.reject(rpcErr);
          statusDef.reject(rpcErr);
          responseDef.reject(rpcErr);
        }
      })();

      const call = {
        method,
        requestHeaders: Promise.resolve(options.meta ?? {}),
        request: Promise.resolve(input),
        headers: headersDef.promise,
        response: responseDef.promise,
        status: statusDef.promise,
        trailers: trailersDef.promise,
      };
      return call as unknown as UnaryCall<I, O>;
    },

    serverStreaming<I extends object, O extends object>(
      method: MethodInfo<I, O>,
      input: I,
      options: RpcOptions,
    ): ServerStreamingCall<I, O> {
      const fullMethod = `/${method.service.typeName}/${method.name}`;
      const headersDef = new Deferred<Record<string, string>>();
      const trailersDef = new Deferred<Record<string, string>>();
      const statusDef = new Deferred<{ code: string; detail: string }>();
      const buffer: O[] = [];
      let ended = false;
      let failed: unknown;
      let wake: (() => void) | undefined;

      void (async () => {
        try {
          const result = await client.serverStream(fullMethod, method.I.toBinary(input), {
            headers: metaToHeaders(options.meta),
            signal: options.abort,
            compress: opts.acceptCompression,
          });
          headersDef.resolve(headersToMeta(result.headers));
          for (const msg of result.messages) {
            buffer.push(method.O.fromBinary(msg));
            wake?.();
          }
          trailersDef.resolve(headersToMeta(result.trailers));
          statusDef.resolve({ code: statusName(result.grpcStatus), detail: result.grpcMessage });
        } catch (err) {
          failed = toRpcError(err);
          headersDef.reject(failed);
          trailersDef.reject(failed);
          statusDef.reject(failed);
        } finally {
          ended = true;
          wake?.();
        }
      })();

      const responses: AsyncIterable<O> = {
        [Symbol.asyncIterator]() {
          let i = 0;
          return {
            async next(): Promise<IteratorResult<O>> {
              for (;;) {
                if (i < buffer.length) {
                  return { value: buffer[i++]!, done: false };
                }
                if (failed) throw failed;
                if (ended) return { value: undefined as unknown as O, done: true };
                await new Promise<void>((r) => {
                  wake = r;
                });
              }
            },
          };
        },
      };

      return {
        method,
        requestHeaders: Promise.resolve(options.meta ?? {}),
        request: Promise.resolve(input),
        headers: headersDef.promise,
        responses,
        status: statusDef.promise,
        trailers: trailersDef.promise,
      } as unknown as ServerStreamingCall<I, O>;
    },

    clientStreaming<I extends object, O extends object>(): ClientStreamingCall<I, O> {
      throw new RpcError(
        "client-streaming is not supported over HTTP/gRPC-Web; use WebSocket or native gRPC",
        "UNIMPLEMENTED",
      );
    },

    duplex<I extends object, O extends object>(): DuplexStreamingCall<I, O> {
      throw new RpcError(
        "bidi streaming is not supported over HTTP/gRPC-Web; use WebSocket or native gRPC",
        "UNIMPLEMENTED",
      );
    },
  };
}
