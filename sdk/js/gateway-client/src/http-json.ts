import { decodeResponseBody, parseBackendError } from "./errors.js";
import { joinURL, mergeHeaders } from "./http-util.js";

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

/**
 * HTTP/JSON client for lava gateway REST routes (`google.api.http`).
 * Failed responses are mapped by parseBackendError, the same mapping the JSON-RPC
 * transport uses, so lava/errorpb bodies, `{code,message}` bodies and plain text
 * all arrive as a GatewayError.
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

    const parsed = decodeResponseBody(await res.text());

    if (!res.ok) {
      // With no body the status text is the only reason the client gets.
      throw parseBackendError(parsed ?? res.statusText, res.status, res.headers);
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
