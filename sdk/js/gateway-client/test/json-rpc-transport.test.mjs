import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { createJsonRpcTransport } from "../dist/json-rpc-transport.js";
import { parseBackendError } from "../dist/errors.js";

// A MethodInfo is only touched through toJsonString/fromJson here, so identity
// codecs stand in for a schemaless (google.protobuf.Struct) payload type.
const watchMethod = {
  name: "Watch",
  service: { typeName: "test.v1.Greeter" },
  I: { toJsonString: () => "{}" },
  O: { fromJson: (json) => json },
};

const echoMethod = {
  name: "Echo",
  service: { typeName: "test.v1.Greeter" },
  I: { toJsonString: () => "{}" },
  O: { fromJson: (json) => json },
};

function withFetch(handler, fn) {
  const original = globalThis.fetch;
  globalThis.fetch = handler;
  return Promise.resolve()
    .then(fn)
    .finally(() => {
      globalThis.fetch = original;
    });
}

function transport() {
  return createJsonRpcTransport({ baseUrl: "http://gateway.test" });
}

describe("JsonRpcTransport server stream errors", () => {
  it("turns the gateway's in-band error envelope into a stream error", async () => {
    // The gateway cannot retract the 200 it already committed, so a failed stream
    // ends with {"error":{code,message}}. It must surface as the real gRPC status,
    // not as a message the caller has to recognise.
    const body = `{"n":"1"}\n${JSON.stringify({ error: { code: 5, message: "gone missing" } })}\n`;

    const messages = [];
    const err = await withFetch(
      async () =>
        new Response(body, {
          status: 200,
          headers: { "Content-Type": "application/x-ndjson" },
        }),
      async () => {
        const call = transport().serverStreaming(watchMethod, {}, {});
        try {
          for await (const message of call.responses) messages.push(message);
        } catch (e) {
          return e;
        }
        return undefined;
      },
    );

    assert.ok(err, "the error envelope must reject the stream");
    assert.equal(err.code, "NOT_FOUND");
    assert.match(err.message, /gone missing/);
    assert.deepEqual(messages, [{ n: "1" }], "the envelope must not be delivered as a message");
  });

  it("keeps a data line shaped like {code,message} as data", async () => {
    // A schemaless stream may legitimately emit code/message fields; only the
    // envelope marks a failure.
    const body = `${JSON.stringify({ code: 5, message: "just data" })}\n`;
    const messages = [];
    const err = await withFetch(
      async () =>
        new Response(body, { status: 200, headers: { "Content-Type": "application/x-ndjson" } }),
      async () => {
        const call = transport().serverStreaming(watchMethod, {}, {});
        try {
          for await (const message of call.responses) messages.push(message);
        } catch (e) {
          return e;
        }
        return undefined;
      },
    );

    assert.equal(err, undefined);
    assert.deepEqual(messages, [{ code: 5, message: "just data" }]);
  });
});

describe("JsonRpcTransport error bodies", () => {
  it("keeps a plain-text error body instead of reporting an unknown format", async () => {
    // The gateway answers fiber errors (e.g. 426 for a websocket upgrade on the
    // HTTP route) as plain text with the real HTTP status.
    const text = "websocket requests must use the gateway WebSocket server";

    const err = await withFetch(
      async () => new Response(text, { status: 426 }),
      async () => {
        const call = transport().unary(echoMethod, {}, {});
        try {
          await call.response;
        } catch (e) {
          return e;
        }
        return undefined;
      },
    );

    assert.ok(err, "a failed unary call must reject");
    assert.match(err.message, /websocket requests must use/);
    assert.equal(err.code, "INTERNAL");
  });
});

describe("parseBackendError", () => {
  it("does not report a gRPC code as an HTTP status", () => {
    // `statusCode` in a lava errorpb body indexes the gRPC code space, so
    // backfilling httpStatus from it puts a 5 where an HTTP status belongs.
    assert.equal(parseBackendError({ code: 5, message: "gone" }).httpStatus, undefined);
    assert.equal(parseBackendError({ statusCode: 5, message: "gone" }).httpStatus, undefined);
    assert.equal(parseBackendError({ code: 5, message: "gone" }, 404).httpStatus, 404);
  });
});
