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

export const GRPC_CODE_NAMES = [
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

export function statusName(code: number): string {
  return GRPC_CODE_NAMES[code] ?? "UNKNOWN";
}

export function statusCodeFromName(name: string): number {
  const idx = GRPC_CODE_NAMES.indexOf(name.toUpperCase().replace(/[\s-]+/g, "_") as (typeof GRPC_CODE_NAMES)[number]);
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

export class GatewayError extends Error {
  readonly code: number;
  readonly grpcMessage: string;
  readonly httpStatus?: number;
  readonly trailers: Headers;
  readonly backendError?: BackendError;

  constructor(opts: {
    code: number;
    message: string;
    httpStatus?: number;
    trailers?: Headers;
    backendError?: BackendError;
  }) {
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

export function parseBackendError(json: unknown, httpStatus?: number): GatewayError {
  if (!isBackendError(json)) {
    return new GatewayError({
      code: GrpcCode.Unknown,
      message: "Unknown error format",
      httpStatus,
    });
  }
  const name = rpcCodeFromBackendError(json);
  return new GatewayError({
    code: statusCodeFromName(name),
    message: json.message || json.name || "Unknown error",
    httpStatus: httpStatus ?? json.statusCode,
    backendError: json,
  });
}

export function toRpcError(err: unknown, method?: { name: string; service: string }): RpcError {
  let rpcErr: RpcError;
  if (err instanceof GatewayError) {
    rpcErr = new RpcError(err.grpcMessage || err.message, statusName(err.code));
    (rpcErr as RpcError & { backendError?: BackendError }).backendError = err.backendError;
  } else if (err instanceof RpcError) {
    rpcErr = err;
  } else if (err instanceof Error) {
    rpcErr = new RpcError(
      err.message,
      err.name === "AbortError" ? "CANCELLED" : "INTERNAL",
    );
  } else {
    rpcErr = new RpcError(String(err), "INTERNAL");
  }
  if (method) {
    rpcErr.methodName = method.name;
    rpcErr.serviceName = method.service;
  }
  return rpcErr;
}
