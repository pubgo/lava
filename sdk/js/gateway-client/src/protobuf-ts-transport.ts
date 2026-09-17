import type {
  MethodInfo,
  RpcMetadata,
  RpcOptions,
  RpcStatus,
  RpcTransport,
  UnaryCall as UnaryCallType,
  ServerStreamingCall as ServerStreamingCallType,
  ClientStreamingCall,
  DuplexStreamingCall,
} from "@protobuf-ts/runtime-rpc";
import {
  Deferred,
  mergeRpcOptions,
  RpcError,
  RpcOutputStreamController,
  ServerStreamingCall,
  UnaryCall,
} from "@protobuf-ts/runtime-rpc";

import { statusName, toRpcError } from "./errors.js";
import { createGrpcWebClient, type GrpcWebClientOptions } from "./grpc-web.js";

export type GrpcWebTransportOptions = GrpcWebClientOptions & RpcOptions & {
  /** Default RpcOptions merged into every call. */
  defaultOptions?: RpcOptions;
};

function appendMeta(headers: Headers, meta: RpcOptions["meta"]) {
  if (!meta) return;
  for (const [k, v] of Object.entries(meta)) {
    if (v == null) continue;
    if (Array.isArray(v)) v.forEach((x) => headers.append(k, String(x)));
    else headers.set(k, String(v));
  }
}

function headersToMeta(h: Headers): Record<string, string> {
  const out: Record<string, string> = {};
  h.forEach((v, k) => {
    out[k] = v;
  });
  return out;
}

/**
 * protobuf-ts {@link RpcTransport} for binary (or text) gRPC-Web framing.
 */
export function createGrpcWebTransport(opts: GrpcWebTransportOptions): RpcTransport {
  const client = createGrpcWebClient(opts);
  const defaults: RpcOptions = { ...(opts.defaultOptions ?? {}), ...opts };

  return {
    mergeOptions(options?: Partial<RpcOptions>): RpcOptions {
      return mergeRpcOptions(defaults, options);
    },

    unary<I extends object, O extends object>(
      method: MethodInfo<I, O>,
      input: I,
      options: RpcOptions,
    ): UnaryCallType<I, O> {
      const opt = mergeRpcOptions(defaults, options);
      const fullMethod = `/${method.service.typeName}/${method.name}`;
      const headers = new Headers();
      appendMeta(headers, opt.meta);

      const defHeader = new Deferred<Record<string, string>>();
      const defMessage = new Deferred<O>();
      const defStatus = new Deferred<{ code: string; detail: string }>();
      const defTrailer = new Deferred<Record<string, string>>();

      void (async () => {
        try {
          const result = await client.unary(fullMethod, method.I.toBinary(input), {
            headers,
            signal: options.abort,
            compress: opts.acceptCompression,
          });
          defHeader.resolve(headersToMeta(result.headers));
          defTrailer.resolve(headersToMeta(result.trailers));
          defStatus.resolve({ code: statusName(result.grpcStatus), detail: result.grpcMessage });
          defMessage.resolve(method.O.fromBinary(result.message));
        } catch (err) {
          const rpcErr = toRpcError(err, {
            name: method.name,
            service: method.service.typeName,
          });
          defHeader.rejectPending(rpcErr);
          defMessage.rejectPending(rpcErr);
          defStatus.rejectPending(rpcErr);
          defTrailer.rejectPending(rpcErr);
        }
      })();

      return new UnaryCall<I, O>(
        method,
        opt.meta ?? {},
        input,
        defHeader.promise as Promise<RpcMetadata>,
        defMessage.promise,
        defStatus.promise as Promise<RpcStatus>,
        defTrailer.promise as Promise<RpcMetadata>,
      );
    },

    serverStreaming<I extends object, O extends object>(
      method: MethodInfo<I, O>,
      input: I,
      options: RpcOptions,
    ): ServerStreamingCallType<I, O> {
      const opt = mergeRpcOptions(defaults, options);
      const fullMethod = `/${method.service.typeName}/${method.name}`;
      const headers = new Headers();
      appendMeta(headers, opt.meta);

      const defHeader = new Deferred<Record<string, string>>();
      const defStatus = new Deferred<{ code: string; detail: string }>();
      const defTrailer = new Deferred<Record<string, string>>();
      const responses = new RpcOutputStreamController<O>();

      void (async () => {
        try {
          const result = await client.serverStream(fullMethod, method.I.toBinary(input), {
            headers,
            signal: options.abort,
            compress: opts.acceptCompression,
          });
          defHeader.resolve(headersToMeta(result.headers));
          for (const msg of result.messages) {
            responses.notifyMessage(method.O.fromBinary(msg));
          }
          responses.notifyComplete();
          defTrailer.resolve(headersToMeta(result.trailers));
          defStatus.resolve({ code: statusName(result.grpcStatus), detail: result.grpcMessage });
        } catch (err) {
          const rpcErr = toRpcError(err, {
            name: method.name,
            service: method.service.typeName,
          });
          if (!responses.closed) responses.notifyError(rpcErr);
          defHeader.rejectPending(rpcErr);
          defStatus.rejectPending(rpcErr);
          defTrailer.rejectPending(rpcErr);
        }
      })();

      return new ServerStreamingCall<I, O>(
        method,
        opt.meta ?? {},
        input,
        defHeader.promise as Promise<RpcMetadata>,
        responses,
        defStatus.promise as Promise<RpcStatus>,
        defTrailer.promise as Promise<RpcMetadata>,
      );
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
