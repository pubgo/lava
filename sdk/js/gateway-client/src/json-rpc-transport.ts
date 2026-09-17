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

import { parseBackendError, toRpcError } from "./errors.js";

export type JsonRpcTransportOptions = RpcOptions & {
  /**
   * Gateway origin or path prefix, e.g. `https://api.example.com/api` or `/api`.
   * Requests go to `{baseUrl}/{package.Service}/{Method}`.
   */
  baseUrl: string;
  fetchInit?: Omit<RequestInit, "body" | "headers" | "method" | "signal">;
};

function makeUrl(baseUrl: string, method: MethodInfo): string {
  let base = baseUrl;
  if (base.endsWith("/")) base = base.slice(0, -1);
  return `${base}/${method.service.typeName}/${method.name}`;
}

function appendMeta(headers: Headers, meta: RpcOptions["meta"]) {
  if (!meta) return;
  for (const [k, v] of Object.entries(meta)) {
    if (typeof v === "string") headers.append(k, v);
    else if (Array.isArray(v)) for (const item of v) headers.append(k, item);
  }
}

/**
 * JSON-over-gRPC-full-method transport used by lava apps in development.
 *
 * - Request: `POST {baseUrl}/{package.Service}/{Method}` with `application/json` (protojson)
 * - Unary response: single JSON object
 * - Server-stream: NDJSON (`application/x-ndjson`)
 * - Errors: lava/errorpb `{ statusCode, code, name, message, details }` or `{ code, message }`
 *
 * This is the same shape as agentrun's `JsonFetchTransport`.
 */
export class JsonRpcTransport implements RpcTransport {
  private readonly defaultOptions: JsonRpcTransportOptions;

  constructor(options: JsonRpcTransportOptions) {
    this.defaultOptions = options;
  }

  mergeOptions(options?: Partial<RpcOptions>): RpcOptions {
    return mergeRpcOptions(this.defaultOptions, options);
  }

  unary<I extends object, O extends object>(
    method: MethodInfo<I, O>,
    input: I,
    options: RpcOptions,
  ): UnaryCallType<I, O> {
    const opt = this.mergeOptions(options) as JsonRpcTransportOptions;
    const url = makeUrl(opt.baseUrl, method);
    const fetchInit = opt.fetchInit ?? {};

    const defHeader = new Deferred<Record<string, string>>();
    const defMessage = new Deferred<O>();
    const defStatus = new Deferred<{ code: string; detail: string }>();
    const defTrailer = new Deferred<Record<string, string>>();

    const headers = new Headers();
    headers.set("Content-Type", "application/json");
    headers.set("Accept", "application/json");
    appendMeta(headers, opt.meta);

    const requestBody = method.I.toJsonString(input, opt.jsonOptions);

    globalThis
      .fetch(url, {
        ...fetchInit,
        method: "POST",
        headers,
        body: requestBody,
        signal: options.abort ?? null,
      })
      .then(async (res) => {
        const responseHeaders: Record<string, string> = {};
        res.headers.forEach((value, key) => {
          if (!["content-type", "content-length"].includes(key.toLowerCase())) {
            responseHeaders[key] = value;
          }
        });
        defHeader.resolve(responseHeaders);

        if (!res.ok) {
          const json = await res.json().catch(() => ({}));
          throw parseBackendError(json, res.status);
        }

        return method.O.fromJson(await res.json(), opt.jsonOptions);
      })
      .then((message) => {
        defMessage.resolve(message);
        defStatus.resolve({ code: "OK", detail: "" });
        defTrailer.resolve({});
      })
      .catch((reason) => {
        const error = toRpcError(reason, {
          name: method.name,
          service: method.service.typeName,
        });
        defHeader.rejectPending(error);
        defMessage.rejectPending(error);
        defStatus.rejectPending(error);
        defTrailer.rejectPending(error);
      });

    return new UnaryCall<I, O>(
      method,
      opt.meta ?? {},
      input,
      defHeader.promise as Promise<RpcMetadata>,
      defMessage.promise,
      defStatus.promise as Promise<RpcStatus>,
      defTrailer.promise as Promise<RpcMetadata>,
    );
  }

  serverStreaming<I extends object, O extends object>(
    method: MethodInfo<I, O>,
    input: I,
    options: RpcOptions,
  ): ServerStreamingCallType<I, O> {
    const opt = this.mergeOptions(options) as JsonRpcTransportOptions;
    const url = makeUrl(opt.baseUrl, method);
    const fetchInit = opt.fetchInit ?? {};

    const defHeader = new Deferred<Record<string, string>>();
    const defStatus = new Deferred<{ code: string; detail: string }>();
    const defTrailer = new Deferred<Record<string, string>>();
    const responses = new RpcOutputStreamController<O>();

    const headers = new Headers();
    headers.set("Content-Type", "application/json");
    headers.set("Accept", "application/x-ndjson, application/json");
    appendMeta(headers, opt.meta);

    const requestBody = method.I.toJsonString(input, opt.jsonOptions);

    globalThis
      .fetch(url, {
        ...fetchInit,
        method: "POST",
        headers,
        body: requestBody,
        signal: options.abort ?? null,
      })
      .then(async (fetchResponse) => {
        const responseHeaders: Record<string, string> = {};
        fetchResponse.headers.forEach((value, key) => {
          if (!["content-type", "content-length"].includes(key.toLowerCase())) {
            responseHeaders[key] = value;
          }
        });
        defHeader.resolve(responseHeaders);

        if (!fetchResponse.ok) {
          const json = await fetchResponse.json().catch(() => ({}));
          throw parseBackendError(json, fetchResponse.status);
        }

        if (!fetchResponse.body) {
          defStatus.resolve({ code: "OK", detail: "" });
          defTrailer.resolve({});
          responses.notifyComplete();
          return;
        }

        const reader = fetchResponse.body.getReader();
        const decoder = new TextDecoder();
        let pending = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          pending += decoder.decode(value, { stream: true });
          const lines = pending.split(/\r?\n/);
          pending = lines.pop() ?? "";

          for (const rawLine of lines) {
            const line = rawLine.trim();
            if (!line) continue;
            const json = JSON.parse(line);
            responses.notifyMessage(method.O.fromJson(json, opt.jsonOptions));
          }
        }

        const tail = (pending + decoder.decode()).trim();
        if (tail) {
          responses.notifyMessage(method.O.fromJson(JSON.parse(tail), opt.jsonOptions));
        }

        responses.notifyComplete();
        defStatus.resolve({ code: "OK", detail: "" });
        defTrailer.resolve({});
      })
      .catch((reason) => {
        const error = toRpcError(reason, {
          name: method.name,
          service: method.service.typeName,
        });
        if (!responses.closed) responses.notifyError(error);
        defHeader.rejectPending(error);
        defStatus.rejectPending(error);
        defTrailer.rejectPending(error);
      });

    return new ServerStreamingCall<I, O>(
      method,
      opt.meta ?? {},
      input,
      defHeader.promise as Promise<RpcMetadata>,
      responses,
      defStatus.promise as Promise<RpcStatus>,
      defTrailer.promise as Promise<RpcMetadata>,
    );
  }

  clientStreaming<I extends object, O extends object>(): ClientStreamingCall<I, O> {
    throw new RpcError("Client streaming is not supported by JSON transport", "UNIMPLEMENTED");
  }

  duplex<I extends object, O extends object>(): DuplexStreamingCall<I, O> {
    throw new RpcError("Duplex streaming is not supported by JSON transport", "UNIMPLEMENTED");
  }
}

export function createJsonRpcTransport(options: JsonRpcTransportOptions): RpcTransport {
  return new JsonRpcTransport(options);
}
