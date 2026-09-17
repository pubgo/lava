import { GatewayError, GrpcCode } from "./errors.js";
import {
  COMPRESSED_FLAG,
  decodeFrames,
  encodeFrame,
  gzipCompress,
  gzipDecompress,
  parseTrailerHeaders,
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

/**
 * Low-level gRPC-Web client (binary proto frames).
 * Prefer {@link createGrpcWebTransport} with protobuf-ts generated clients.
 */
export function createGrpcWebClient(opts: GrpcWebClientOptions) {
  const doFetch = opts.fetch ?? globalThis.fetch.bind(globalThis);

  async function call(
    fullMethod: string,
    requestMessage: Uint8Array,
    callOpts?: GrpcWebCallOptions,
  ): Promise<{
    dataFrames: Uint8Array[];
    headers: Headers;
    trailers: Headers;
  }> {
    const compress = callOpts?.compress ?? opts.acceptCompression ?? false;
    let payload = requestMessage;
    let flags = 0;
    const headers = mergeHeaders(
      {
        "Content-Type": "application/grpc-web+proto",
        Accept: "application/grpc-web+proto",
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

    const path = fullMethod.startsWith("/") ? fullMethod : `/${fullMethod}`;
    const res = await doFetch(joinURL(opts.baseUrl, path), {
      method: "POST",
      headers,
      body: encodeFrame(payload, flags) as unknown as BodyInit,
      signal: callOpts?.signal,
    });

    const raw = new Uint8Array(await res.arrayBuffer());
    const frames = decodeFrames(raw);
    const dataFrames: Uint8Array[] = [];
    let trailers = new Headers();
    const responseEncoding = res.headers.get("Grpc-Encoding") ?? res.headers.get("grpc-encoding");

    for (const frame of frames) {
      if (frame.isTrailer) {
        trailers = parseTrailerHeaders(frame.payload);
        continue;
      }
      dataFrames.push(await maybeDecompress(frame.payload, frame.isCompressed, responseEncoding));
    }

    // Some gateways may also put grpc-status on HTTP headers for errors.
    if (!trailers.has("grpc-status") && !trailers.has("Grpc-Status")) {
      const hs = res.headers.get("Grpc-Status") ?? res.headers.get("grpc-status");
      if (hs != null) trailers.set("grpc-status", hs);
      const hm = res.headers.get("Grpc-Message") ?? res.headers.get("grpc-message");
      if (hm != null) trailers.set("grpc-message", hm);
    }

    if (!res.ok && dataFrames.length === 0) {
      const st = trailerStatus(trailers);
      throw new GatewayError({
        code: st.code || GrpcCode.Unknown,
        message: st.message || res.statusText,
        httpStatus: res.status,
        trailers,
      });
    }

    return { dataFrames, headers: res.headers, trailers };
  }

  return {
    async unary(fullMethod: string, requestMessage: Uint8Array, callOpts?: GrpcWebCallOptions): Promise<GrpcWebUnaryResult> {
      const { dataFrames, headers, trailers } = await call(fullMethod, requestMessage, callOpts);
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
        headers,
        trailers,
        grpcStatus: st.code,
        grpcMessage: st.message,
      };
    },

    async serverStream(
      fullMethod: string,
      requestMessage: Uint8Array,
      callOpts?: GrpcWebCallOptions,
    ): Promise<GrpcWebStreamResult> {
      const { dataFrames, headers, trailers } = await call(fullMethod, requestMessage, callOpts);
      const st = trailerStatus(trailers);
      if (st.code !== GrpcCode.OK) {
        throw new GatewayError({
          code: st.code,
          message: st.message,
          trailers,
        });
      }
      return {
        messages: dataFrames,
        headers,
        trailers,
        grpcStatus: st.code,
        grpcMessage: st.message,
      };
    },
  };
}

export type GrpcWebClient = ReturnType<typeof createGrpcWebClient>;
