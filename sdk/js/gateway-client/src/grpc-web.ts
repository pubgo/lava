import { GatewayError, GrpcCode } from "./errors.js";
import {
  Base64ByteDecoder,
  COMPRESSED_FLAG,
  decodeFrames,
  encodeFrame,
  GrpcWebFrameReader,
  gzipCompress,
  gzipDecompress,
  parseTrailerHeaders,
  type GrpcWebFrame,
} from "./frames.js";

export type GrpcWebClientOptions = {
  /** Gateway origin, e.g. https://api.example.com (no trailing slash). */
  baseUrl: string;
  /** Extra headers for every request. */
  headers?: HeadersInit;
  fetch?: typeof fetch;
  /**
   * When true, send `grpc-accept-encoding: gzip` and compress request payloads
   * with gzip when CompressionStream is available.
   */
  acceptCompression?: boolean;
  /** `binary` (default) or `text` (base64 gRPC-Web). */
  format?: "binary" | "text";
};

export type GrpcWebCallOptions = {
  headers?: HeadersInit;
  signal?: AbortSignal;
  /** Override per-call compression (defaults to client acceptCompression). */
  compress?: boolean;
};

export type GrpcWebUnaryResult = {
  message: Uint8Array;
  headers: Headers;
  trailers: Headers;
  grpcStatus: number;
  grpcMessage: string;
};

export type GrpcWebStreamResult = {
  messages: Uint8Array[];
  headers: Headers;
  trailers: Headers;
  grpcStatus: number;
  grpcMessage: string;
};

/** Live server-stream event (frames delivered as the HTTP body arrives). */
export type GrpcWebStreamEvent =
  | { type: "message"; message: Uint8Array }
  | { type: "trailer"; trailers: Headers };

function joinURL(base: string, path: string): string {
  const b = base.replace(/\/+$/, "");
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${b}${p}`;
}

function mergeHeaders(...parts: Array<HeadersInit | undefined>): Headers {
  const out = new Headers();
  for (const part of parts) {
    if (!part) continue;
    new Headers(part).forEach((v, k) => out.set(k, v));
  }
  return out;
}

function trailerStatus(trailers: Headers): { code: number; message: string } {
  const raw = trailers.get("grpc-status") ?? trailers.get("Grpc-Status") ?? "0";
  const code = Number.parseInt(raw, 10);
  const message = trailers.get("grpc-message") ?? trailers.get("Grpc-Message") ?? "";
  return {
    code: Number.isFinite(code) ? code : GrpcCode.Unknown,
    message: decodeURIComponent(message.replace(/\+/g, " ")),
  };
}

function bytesToBase64(bytes: Uint8Array): string {
  let binary = "";
  const chunk = 0x8000;
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk));
  }
  return btoa(binary);
}

function base64ToBytes(text: string): Uint8Array {
  const cleaned = text.replace(/[^A-Za-z0-9+/=]/g, "");
  try {
    const bin = atob(cleaned);
    const out = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
    return out;
  } catch (err) {
    throw new GatewayError({
      code: GrpcCode.Internal,
      message: `invalid grpc-web-text base64 response: ${err instanceof Error ? err.message : String(err)}`,
    });
  }
}

async function maybeDecompress(framePayload: Uint8Array, compressed: boolean, responseEncoding: string | null) {
  if (!compressed) return framePayload;
  const enc = (responseEncoding ?? "gzip").toLowerCase();
  if (enc !== "gzip") {
    throw new GatewayError({
      code: GrpcCode.Unimplemented,
      message: `unsupported grpc-encoding ${enc}`,
    });
  }
  return gzipDecompress(framePayload);
}

function mergeHttpStatusTrailers(res: Response, trailers: Headers) {
  if (!trailers.has("grpc-status") && !trailers.has("Grpc-Status")) {
    const hs = res.headers.get("Grpc-Status") ?? res.headers.get("grpc-status");
    if (hs != null) trailers.set("grpc-status", hs);
    const hm = res.headers.get("Grpc-Message") ?? res.headers.get("grpc-message");
    if (hm != null) trailers.set("grpc-message", hm);
  }
}

/**
 * Low-level gRPC-Web client (binary or text proto frames).
 * Prefer {@link createGrpcWebTransport} / {@link createGatewayTransport} with protobuf-ts clients.
 */
export function createGrpcWebClient(opts: GrpcWebClientOptions) {
  const doFetch = opts.fetch ?? globalThis.fetch.bind(globalThis);
  const format = opts.format ?? "binary";

  async function prepareRequest(
    fullMethod: string,
    requestMessage: Uint8Array,
    callOpts?: GrpcWebCallOptions,
  ) {
    const compress = callOpts?.compress ?? opts.acceptCompression ?? false;
    let payload = requestMessage;
    let flags = 0;
    const contentType =
      format === "text" ? "application/grpc-web-text+proto" : "application/grpc-web+proto";
    const headers = mergeHeaders(
      {
        "Content-Type": contentType,
        Accept: contentType,
        "X-Grpc-Web": "1",
      },
      opts.headers,
      callOpts?.headers,
    );

    if (opts.acceptCompression || compress) {
      headers.set("Grpc-Accept-Encoding", "gzip");
    }
    if (compress) {
      payload = await gzipCompress(requestMessage);
      flags |= COMPRESSED_FLAG;
      headers.set("Grpc-Encoding", "gzip");
    }

    const frame = encodeFrame(payload, flags);
    const body: BodyInit =
      format === "text" ? bytesToBase64(frame) : (frame as unknown as BodyInit);

    const path = fullMethod.startsWith("/") ? fullMethod : `/${fullMethod}`;
    const res = await doFetch(joinURL(opts.baseUrl, path), {
      method: "POST",
      headers,
      body,
      signal: callOpts?.signal,
    });
    return res;
  }

  async function consumeBodyBuffered(res: Response): Promise<{
    dataFrames: Uint8Array[];
    trailers: Headers;
  }> {
    let rawBytes: Uint8Array;
    if (format === "text") {
      rawBytes = base64ToBytes(await res.text());
    } else {
      rawBytes = new Uint8Array(await res.arrayBuffer());
    }

    const frames = decodeFrames(rawBytes);
    const dataFrames: Uint8Array[] = [];
    let trailers = new Headers();
    const responseEncoding = res.headers.get("Grpc-Encoding") ?? res.headers.get("grpc-encoding");

    for (const framePart of frames) {
      if (framePart.isTrailer) {
        trailers = parseTrailerHeaders(framePart.payload);
        continue;
      }
      dataFrames.push(await maybeDecompress(framePart.payload, framePart.isCompressed, responseEncoding));
    }

    mergeHttpStatusTrailers(res, trailers);
    return { dataFrames, trailers };
  }

  async function* iterateLiveFrames(res: Response): AsyncGenerator<GrpcWebStreamEvent> {
    const responseEncoding = res.headers.get("Grpc-Encoding") ?? res.headers.get("grpc-encoding");
    const frameReader = new GrpcWebFrameReader();
    const b64 = format === "text" ? new Base64ByteDecoder() : null;
    let trailers = new Headers();
    let sawTrailer = false;

    async function emit(frames: GrpcWebFrame[]): Promise<GrpcWebStreamEvent[]> {
      const events: GrpcWebStreamEvent[] = [];
      for (const framePart of frames) {
        if (framePart.isTrailer) {
          trailers = parseTrailerHeaders(framePart.payload);
          sawTrailer = true;
          events.push({ type: "trailer", trailers });
          continue;
        }
        const message = await maybeDecompress(framePart.payload, framePart.isCompressed, responseEncoding);
        events.push({ type: "message", message });
      }
      return events;
    }

    if (!res.body) {
      // Fallback: no streaming body (some test mocks).
      const buffered = await consumeBodyBuffered(res);
      for (const message of buffered.dataFrames) {
        yield { type: "message", message };
      }
      yield { type: "trailer", trailers: buffered.trailers };
      return;
    }

    const reader = res.body.getReader();
    const textDecoder = format === "text" ? new TextDecoder() : null;

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        if (!value || value.length === 0) continue;

        let bytes: Uint8Array;
        if (b64 && textDecoder) {
          bytes = b64.push(textDecoder.decode(value, { stream: true }));
        } else {
          bytes = value;
        }
        for (const ev of await emit(frameReader.push(bytes))) {
          yield ev;
        }
      }

      if (b64) {
        const tail = b64.finish();
        if (tail.length) {
          for (const ev of await emit(frameReader.push(tail))) {
            yield ev;
          }
        }
      }
    } finally {
      reader.releaseLock();
    }

    if (!sawTrailer) {
      mergeHttpStatusTrailers(res, trailers);
      yield { type: "trailer", trailers };
    }
  }

  const api = {
    async unary(fullMethod: string, requestMessage: Uint8Array, callOpts?: GrpcWebCallOptions): Promise<GrpcWebUnaryResult> {
      const res = await prepareRequest(fullMethod, requestMessage, callOpts);
      const { dataFrames, trailers } = await consumeBodyBuffered(res);

      if (!res.ok && dataFrames.length === 0) {
        const st = trailerStatus(trailers);
        throw new GatewayError({
          code: st.code || GrpcCode.Unknown,
          message: st.message || res.statusText,
          httpStatus: res.status,
          trailers,
        });
      }

      const st = trailerStatus(trailers);
      if (st.code !== GrpcCode.OK) {
        throw new GatewayError({
          code: st.code,
          message: st.message,
          trailers,
        });
      }
      if (dataFrames.length !== 1) {
        throw new GatewayError({
          code: GrpcCode.Internal,
          message: `expected 1 response message, got ${dataFrames.length}`,
          trailers,
        });
      }
      return {
        message: dataFrames[0]!,
        headers: res.headers,
        trailers,
        grpcStatus: st.code,
        grpcMessage: st.message,
      };
    },

    /**
     * Live server-stream: yields messages as HTTP chunks arrive (pub/sub style).
     * First event is always `{ type: "headers" }`.
     */
    async *openServerStream(
      fullMethod: string,
      requestMessage: Uint8Array,
      callOpts?: GrpcWebCallOptions,
    ): AsyncGenerator<
      | { type: "headers"; headers: Headers }
      | GrpcWebStreamEvent
    > {
      const res = await prepareRequest(fullMethod, requestMessage, callOpts);
      yield { type: "headers", headers: res.headers };

      if (!res.ok) {
        const { dataFrames, trailers } = await consumeBodyBuffered(res);
        if (dataFrames.length === 0) {
          const st = trailerStatus(trailers);
          throw new GatewayError({
            code: st.code || GrpcCode.Unknown,
            message: st.message || res.statusText,
            httpStatus: res.status,
            trailers,
          });
        }
      }

      for await (const ev of iterateLiveFrames(res)) {
        yield ev;
      }
    },

    /**
     * Buffered server-stream (waits until the stream completes). Prefer
     * {@link openServerStream} / protobuf-ts transport for live delivery.
     */
    async serverStream(
      fullMethod: string,
      requestMessage: Uint8Array,
      callOpts?: GrpcWebCallOptions,
    ): Promise<GrpcWebStreamResult> {
      const messages: Uint8Array[] = [];
      let headers = new Headers();
      let trailers = new Headers();
      let grpcStatus: number = GrpcCode.OK;
      let grpcMessage = "";

      for await (const ev of api.openServerStream(fullMethod, requestMessage, callOpts)) {
        if (ev.type === "headers") {
          headers = ev.headers;
          continue;
        }
        if (ev.type === "message") {
          messages.push(ev.message);
          continue;
        }
        trailers = ev.trailers;
        const st = trailerStatus(trailers);
        grpcStatus = st.code;
        grpcMessage = st.message;
      }

      if (grpcStatus !== GrpcCode.OK) {
        throw new GatewayError({
          code: grpcStatus,
          message: grpcMessage,
          trailers,
        });
      }
      return { messages, headers, trailers, grpcStatus, grpcMessage };
    },
  };

  return api;
}

export type GrpcWebClient = ReturnType<typeof createGrpcWebClient>;
