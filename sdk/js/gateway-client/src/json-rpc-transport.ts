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

import { GatewayError, readBackendError, toRpcError } from "./errors.js";
import { joinURL } from "./http-util.js";

export type JsonRpcTransportOptions = RpcOptions & {
  /**
   * Gateway origin or path prefix, e.g. `https://api.example.com/api` or `/api`.
   * Requests go to `{baseUrl}/{package.Service}/{Method}`.
   */
  baseUrl: string;
  fetchInit?: Omit<RequestInit, "body" | "headers" | "method" | "signal">;
};

function makeUrl(baseUrl: string, method: MethodInfo): string {
  return joinURL(baseUrl, `${method.service.typeName}/${method.name}`);
}

function appendMeta(headers: Headers, meta: RpcOptions["meta"]) {
  if (!meta) return;
  for (const [k, v] of Object.entries(meta)) {
    if (typeof v === "string") headers.append(k, v);
    else if (Array.isArray(v)) for (const item of v) headers.append(k, item);
  }
}

/**
 * Recognises the gateway's in-band NDJSON error envelope. The HTTP 200 is already
 * committed when a stream fails, so the status travels as the final line:
 * `{"error":{"code":<grpc code>,"message":<string>}}`.
 *
 * Only a top-level object whose sole key is `error` counts: a legitimate message
 * that happens to nest an `error` field alongside other fields must stay data.
 */
function streamErrorFromBody(value: unknown): GatewayError | undefined {
  if (typeof value !== "object" || value === null) return undefined;
  const obj = value as Record<string, unknown>;
  const keys = Object.keys(obj);
  if (keys.length !== 1 || keys[0] !== "error") return undefined;
  const envelope = obj.error;
  if (typeof envelope !== "object" || envelope === null) return undefined;
  const { code, message } = envelope as Record<string, unknown>;
  if (typeof code !== "number") return undefined;
  return new GatewayError({
    code,
    message: typeof message === "string" ? message : "",
  });
}

/**
 * JSON-over-gRPC-full-method transport used by lava apps in development.
 *
 * - Request: `POST {baseUrl}/{package.Service}/{Method}` with `application/json` (protojson)
 * - Unary response: single JSON object
 * - Server-stream: NDJSON (`application/x-ndjson`)
 * - Errors: lava/errorpb `{ statusCode, code, name, message, details }`, the
 *   gateway's `{ code, message }`, plain-text bodies, and — on a stream whose 200
 *   was already committed — the final `{"error":{code,message}}` line
 *
 * Otherwise the same shape as agentrun's `JsonFetchTransport`.
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

        if (!res.ok) throw await readBackendError(res);

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

        if (!fetchResponse.ok) throw await readBackendError(fetchResponse);

        if (!fetchResponse.body) {
          defStatus.resolve({ code: "OK", detail: "" });
          defTrailer.resolve({});
          responses.notifyComplete();
          return;
        }

        const reader = fetchResponse.body.getReader();
        const decoder = new TextDecoder();
        let pending = "";

        const emitLine = (rawLine: string) => {
          const line = rawLine.trim();
          if (!line) return;
          const json = JSON.parse(line);
          // A failed stream ends with this line instead of a status the transport
          // could read, so it must reject the stream rather than reach the caller as
          // a message. Anything else the method's schema accepts stays data.
          const failure = streamErrorFromBody(json);
          if (failure) throw failure;
          responses.notifyMessage(method.O.fromJson(json, opt.jsonOptions));
        };

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          pending += decoder.decode(value, { stream: true });
          const lines = pending.split(/\r?\n/);
          pending = lines.pop() ?? "";

          for (const rawLine of lines) emitLine(rawLine);
        }

        emitLine(pending + decoder.decode());

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
