import { GatewayError, GrpcCode } from "./errors.js";

export type HttpJsonClientOptions = {
  /** Gateway origin, e.g. https://api.example.com (no trailing slash). */
  baseUrl: string;
  /** Extra headers for every request. */
  headers?: HeadersInit;
  fetch?: typeof fetch;
};

export type HttpJsonRequestOptions = {
  headers?: HeadersInit;
  signal?: AbortSignal;
};

export type HttpJsonResponse<T> = {
  data: T;
  headers: Headers;
  httpStatus: number;
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

/**
 * HTTP/JSON client for lava gateway REST routes (`google.api.http`).
 * Errors map to {@link GatewayError} using the gateway JSON error body.
 */
export function createHttpJsonClient(opts: HttpJsonClientOptions) {
  const doFetch = opts.fetch ?? globalThis.fetch.bind(globalThis);

  async function request<TReq extends object, TRes>(
    method: string,
    path: string,
    body?: TReq,
    reqOpts?: HttpJsonRequestOptions,
  ): Promise<HttpJsonResponse<TRes>> {
    const headers = mergeHeaders(
      { Accept: "application/json", "Content-Type": "application/json" },
      opts.headers,
      reqOpts?.headers,
    );
    const res = await doFetch(joinURL(opts.baseUrl, path), {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: reqOpts?.signal,
    });

    const text = await res.text();
    let parsed: unknown = undefined;
    if (text) {
      try {
        parsed = JSON.parse(text);
      } catch {
        parsed = text;
      }
    }

    if (!res.ok) {
      const obj = parsed && typeof parsed === "object" ? (parsed as Record<string, unknown>) : {};
      const code = typeof obj.code === "number" ? obj.code : GrpcCode.Unknown;
      const message =
        typeof obj.message === "string"
          ? obj.message
          : typeof parsed === "string"
            ? parsed
            : res.statusText;
      throw new GatewayError({
        code,
        message,
        httpStatus: res.status,
        trailers: res.headers,
      });
    }

    return {
      data: parsed as TRes,
      headers: res.headers,
      httpStatus: res.status,
    };
  }

  return {
    get<TRes>(path: string, reqOpts?: HttpJsonRequestOptions) {
      return request<never, TRes>("GET", path, undefined, reqOpts);
    },
    post<TReq extends object, TRes>(path: string, body: TReq, reqOpts?: HttpJsonRequestOptions) {
      return request<TReq, TRes>("POST", path, body, reqOpts);
    },
    put<TReq extends object, TRes>(path: string, body: TReq, reqOpts?: HttpJsonRequestOptions) {
      return request<TReq, TRes>("PUT", path, body, reqOpts);
    },
    patch<TReq extends object, TRes>(path: string, body: TReq, reqOpts?: HttpJsonRequestOptions) {
      return request<TReq, TRes>("PATCH", path, body, reqOpts);
    },
    delete<TRes>(path: string, reqOpts?: HttpJsonRequestOptions) {
      return request<never, TRes>("DELETE", path, undefined, reqOpts);
    },
  };
}

export type HttpJsonClient = ReturnType<typeof createHttpJsonClient>;
