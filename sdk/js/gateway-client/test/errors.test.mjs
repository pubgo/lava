import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { createHttpJsonClient } from "../dist/http-json.js";
import {
  GrpcCode,
  GRPC_CODE_NAMES,
  GatewayError,
  parseBackendError,
  readBackendError,
  statusName,
  statusCodeFromName,
  toRpcError,
} from "../dist/errors.js";

describe("gRPC code tables", () => {
  it("keeps the numeric codes and their protobuf spellings in step", () => {
    // Two hand-written tables drift silently: adding a code to one and not the
    // other still type-checks and only shows up as a wrong status name at runtime.
    for (const [name, code] of Object.entries(GrpcCode)) {
      assert.equal(statusCodeFromName(statusName(code)), code, `round trip for ${name}`);
    }
    assert.equal(GRPC_CODE_NAMES[GrpcCode.OK], "OK");
    assert.equal(GRPC_CODE_NAMES[GrpcCode.DeadlineExceeded], "DEADLINE_EXCEEDED");
    assert.equal(GRPC_CODE_NAMES[GrpcCode.DataLoss], "DATA_LOSS");
    assert.equal(GRPC_CODE_NAMES[GrpcCode.Unauthenticated], "UNAUTHENTICATED");
  });
});

describe("toRpcError", () => {
  it("keeps what the HTTP response knew about the failure", () => {
    // A transport that collapses a GatewayError into a bare RpcError leaves the
    // caller with no way to see the status code or the headers that produced it.
    const trailers = new Headers({ "grpc-status": "5" });
    const gatewayErr = new GatewayError({
      code: GrpcCode.NotFound,
      message: "no such widget",
      httpStatus: 404,
      trailers,
      backendError: { code: 5, message: "no such widget" },
    });

    const rpcErr = toRpcError(gatewayErr, { name: "GetWidget", service: "test.v1.Widget" });

    assert.equal(rpcErr.message, "no such widget");
    assert.equal(rpcErr.code, "NOT_FOUND");
    assert.equal(rpcErr.methodName, "GetWidget");
    assert.equal(rpcErr.serviceName, "test.v1.Widget");
    assert.equal(rpcErr.httpStatus, 404);
    assert.equal(rpcErr.trailers.get("grpc-status"), "5");
    assert.equal(rpcErr.cause, gatewayErr);
    assert.deepEqual(rpcErr.backendError, { code: 5, message: "no such widget" });
  });
});

describe("parseBackendError", () => {
  it("carries the response headers through to the error", () => {
    const err = parseBackendError(
      { statusCode: 8, message: "slow down" },
      429,
      new Headers({ "retry-after": "30" }),
    );
    assert.equal(err.code, GrpcCode.ResourceExhausted);
    assert.equal(err.httpStatus, 429);
    assert.equal(err.trailers.get("retry-after"), "30");
  });
});

describe("readBackendError", () => {
  it("keeps a plain-text rejection and the headers that came with it", async () => {
    // Some rejections are fiber text with an HTTP status and no JSON body. The
    // text is the only reason, and the headers are the only request handle.
    const err = await readBackendError(
      new Response("unsupported websocket upgrade request", {
        status: 426,
        headers: { "x-request-id": "req-7" },
      }),
    );

    assert.ok(err instanceof GatewayError, `got ${String(err)}`);
    assert.equal(err.message, "unsupported websocket upgrade request");
    assert.equal(err.httpStatus, 426);
    assert.equal(err.trailers.get("x-request-id"), "req-7");
  });
});

describe("createHttpJsonClient errors", () => {
  function client(response) {
    return createHttpJsonClient({
      baseUrl: "http://gateway.test",
      fetch: async () => response,
    });
  }

  it("reads the lava/errorpb shape, not just {code,message}", async () => {
    // The gateway's rich errors arrive as {statusCode, code, name, message}; a
    // client that only looks at "code" reports every one of them as Unknown.
    const err = await client(
      new Response(JSON.stringify({ statusCode: 7, message: "denied" }), {
        status: 403,
        headers: { "Content-Type": "application/json", "x-request-id": "abc" },
      }),
    )
      .post("/v1/widgets", {})
      .catch((e) => e);

    assert.ok(err instanceof GatewayError, `got ${String(err)}`);
    assert.equal(err.code, GrpcCode.PermissionDenied);
    assert.equal(err.message, "denied");
    assert.equal(err.httpStatus, 403);
    assert.equal(err.trailers.get("x-request-id"), "abc");
  });

  it("keeps a plain-text body as the message", async () => {
    const err = await client(new Response("no route for POST /v1/widgets", { status: 404 }))
      .post("/v1/widgets", {})
      .catch((e) => e);

    assert.ok(err instanceof GatewayError, `got ${String(err)}`);
    assert.equal(err.code, GrpcCode.NotFound);
    assert.equal(err.message, "no route for POST /v1/widgets");
    assert.equal(err.httpStatus, 404);
  });

  it("falls back to the status text when there is no body at all", async () => {
    // Bodyless rejections happen at the proxy layer; the message must not become
    // "Unknown error format" while the HTTP status still says Unavailable.
    const err = await client(
      new Response("", { status: 503, statusText: "Service Unavailable" }),
    )
      .post("/v1/widgets", {})
      .catch((e) => e);

    assert.ok(err instanceof GatewayError, `got ${String(err)}`);
    assert.equal(err.code, GrpcCode.Unavailable);
    assert.equal(err.message, "Service Unavailable");
    assert.equal(err.httpStatus, 503);
  });
});
