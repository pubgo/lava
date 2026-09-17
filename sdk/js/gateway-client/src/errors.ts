export class GatewayError extends Error {
  readonly code: number;
  readonly grpcMessage: string;
  readonly httpStatus?: number;
  readonly trailers: Headers;

  constructor(opts: {
    code: number;
    message: string;
    httpStatus?: number;
    trailers?: Headers;
  }) {
    super(opts.message || `gateway error code=${opts.code}`);
    this.name = "GatewayError";
    this.code = opts.code;
    this.grpcMessage = opts.message;
    this.httpStatus = opts.httpStatus;
    this.trailers = opts.trailers ?? new Headers();
  }
}

/** gRPC status codes used by the gateway JSON error surface. */
export const GrpcCode = {
  OK: 0,
  Cancelled: 1,
  Unknown: 2,
  InvalidArgument: 3,
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
  Unauthenticated: 16,
} as const;
