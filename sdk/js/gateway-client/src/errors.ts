import { RpcError } from "@protobuf-ts/runtime-rpc";

/** Numeric gRPC status codes. */
export const GrpcCode = {
  OK: 0,
  Cancelled: 1,
  Unknown: 2,
  InvalidArgument: 3,
  DeadlineExceeded: 4,
  NotFound: 5,
  AlreadyExists: 6,
  PermissionDenied: 7,
  ResourceExhausted: 8,
  FailedPrecondition: 9,
  Aborted: 10,
  OutOfRange: 11,
  Unimplemented: 12,
  Internal: 13,
  Unavailable: 14,
  DataLoss: 15,
  Unauthenticated: 16,
} as const;

/**
 * protobuf's SCREAMING_SNAKE spelling of each {@link GrpcCode}, indexed by its
 * number. Derived rather than hand-written: a second table can be updated out
 * of step with the codes and still type-check, and the drift only shows up as a
 * wrong status name at runtime.
 */
export const GRPC_CODE_NAMES: readonly string[] = Object.entries(GrpcCode)
  .sort(([, a], [, b]) => a - b)
  .map(([name]) => name.replace(/([a-z0-9])([A-Z])/g, "$1_$2").toUpperCase());

export function statusName(code: number): string {
  return GRPC_CODE_NAMES[code] ?? "UNKNOWN";
}

export function statusCodeFromName(name: string): number {
  const idx = GRPC_CODE_NAMES.indexOf(name.toUpperCase().replace(/[\s-]+/g, "_"));
  return idx >= 0 ? idx : GrpcCode.Unknown;
}

/**
 * lava / errorpb.ErrCode shaped error body (JSON transports).
 * Also accepts the simpler gateway `{ code, message }` surface.
 */
export type BackendError = {
  statusCode?: number;
  code?: number;
  name?: string;
  message?: string;
  details?: unknown[];
  id?: string;
};

/**
 * What an HTTP gateway response knows about a failure beyond the gRPC status
 * name. GatewayError stores these as fields; toRpcError re-attaches them to the
 * RpcError handed to protobuf-ts callers.
 */
export type RpcErrorDetail = {
  httpStatus?: number;
  trailers?: Headers;
  backendError?: BackendError;
};

export class GatewayError extends Error implements RpcErrorDetail {
  readonly code: number;
  readonly grpcMessage: string;
  readonly httpStatus?: number;
  readonly trailers: Headers;
  readonly backendError?: BackendError;

  constructor(opts: RpcErrorDetail & { code: number; message: string }) {
    super(opts.message || `gateway error code=${opts.code}`);
    this.name = "GatewayError";
    this.code = opts.code;
    this.grpcMessage = opts.message;
    this.httpStatus = opts.httpStatus;
    this.trailers = opts.trailers ?? new Headers();
    this.backendError = opts.backendError;
  }
}

function isBackendError(obj: unknown): obj is BackendError {
  return (
    typeof obj === "object" &&
    obj !== null &&
    ("statusCode" in obj || "code" in obj || "name" in obj || "message" in obj)
  );
}

function mapHttpStatusToRpcCode(statusCode: number): string | undefined {
  switch (statusCode) {
    case 400:
      return "INVALID_ARGUMENT";
    case 401:
      return "UNAUTHENTICATED";
    case 403:
      return "PERMISSION_DENIED";
    case 404:
      return "NOT_FOUND";
    case 408:
      return "DEADLINE_EXCEEDED";
    case 409:
      return "ABORTED";
    case 429:
      return "RESOURCE_EXHAUSTED";
    case 499:
      return "CANCELLED";
    case 500:
      return "INTERNAL";
    case 501:
      return "UNIMPLEMENTED";
    case 503:
      return "UNAVAILABLE";
    case 504:
      return "DEADLINE_EXCEEDED";
    default:
      return undefined;
  }
}

/** Resolve a symbolic gRPC status name from a lava/backend JSON error body. */
export function rpcCodeFromBackendError(err: BackendError): string {
  const statusCode = typeof err.statusCode === "number" ? err.statusCode : undefined;
  if (statusCode !== undefined) {
    if (GRPC_CODE_NAMES[statusCode]) return GRPC_CODE_NAMES[statusCode]!;
    const httpMapped = mapHttpStatusToRpcCode(statusCode);
    if (httpMapped) return httpMapped;
  }

  // Simple gateway JSON error: { code: <grpc number>, message }
  if (typeof err.code === "number" && GRPC_CODE_NAMES[err.code]) {
    return GRPC_CODE_NAMES[err.code]!;
  }

  if (err.name) {
    const normalized = err.name.trim().toUpperCase().replace(/[\s-]+/g, "_");
    if (normalized) return normalized;
  }

  return "INTERNAL";
}

export function parseBackendError(body: unknown, httpStatus?: number, trailers?: Headers): GatewayError {
  // The gateway answers some rejections (fiber errors, plain HTTP routes) as text
  // rather than a JSON error body. That text is the only reason the client gets.
  if (typeof body === "string" && body.trim() !== "") {
    const mapped = httpStatus === undefined ? undefined : mapHttpStatusToRpcCode(httpStatus);
    return new GatewayError({
      code: statusCodeFromName(mapped ?? "INTERNAL"),
      message: body,
      httpStatus,
      trailers,
    });
  }

  if (!isBackendError(body)) {
    return new GatewayError({
      code: GrpcCode.Unknown,
      message: "Unknown error format",
      httpStatus,
      trailers,
    });
  }
  const name = rpcCodeFromBackendError(body);
  return new GatewayError({
    code: statusCodeFromName(name),
    message: body.message || body.name || "Unknown error",
    httpStatus,
    trailers,
    backendError: body,
  });
}

/** Decode a gateway response body: JSON when it parses, otherwise the raw text. */
export function decodeResponseBody(text: string): unknown {
  if (!text) return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

/**
 * Read a failed response body without assuming it is JSON. Some gateway
 * rejections are plain text, and the text is the only reason the client gets.
 */
export async function readBackendError(res: Response): Promise<GatewayError> {
  const text = await res.text().catch(() => "");
  return parseBackendError(decodeResponseBody(text), res.status, res.headers);
}

/**
 * An {@link RpcError} that still carries what the HTTP response said: protobuf-ts
 * only knows about the status name, so without this the caller cannot see the
 * HTTP status, the headers, or the parsed backend body.
 */
export type DetailedRpcError = RpcError & RpcErrorDetail;

export function toRpcError(err: unknown, method?: { name: string; service: string }): DetailedRpcError {
  let rpcErr: DetailedRpcError;
  if (err instanceof GatewayError) {
    rpcErr = new RpcError(err.grpcMessage || err.message, statusName(err.code));
    rpcErr.httpStatus = err.httpStatus;
    rpcErr.trailers = err.trailers;
    rpcErr.backendError = err.backendError;
    rpcErr.cause = err;
  } else if (err instanceof RpcError) {
    rpcErr = err;
  } else if (err instanceof Error) {
    rpcErr = new RpcError(
      err.message,
      err.name === "AbortError" ? "CANCELLED" : "INTERNAL",
    );
    rpcErr.cause = err;
  } else {
    rpcErr = new RpcError(String(err), "INTERNAL");
  }
  if (method) {
    rpcErr.methodName = method.name;
    rpcErr.serviceName = method.service;
  }
  return rpcErr;
}
