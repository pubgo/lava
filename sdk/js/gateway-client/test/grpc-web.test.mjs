import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { createGrpcWebClient } from "../dist/grpc-web.js";
import { encodeFrame, TRAILER_FLAG } from "../dist/frames.js";
import { GrpcCode } from "../dist/errors.js";

function bytesToBase64(bytes) {
  let binary = "";
  for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
  return btoa(binary);
}

function buildTextResponse(messageBytes, status = "0", message = "ok") {
  const data = encodeFrame(messageBytes);
  const trailerBody = new TextEncoder().encode(`grpc-status: ${status}\r\ngrpc-message: ${message}\r\n`);
  const trailer = encodeFrame(trailerBody, TRAILER_FLAG);
  const raw = new Uint8Array(data.length + trailer.length);
  raw.set(data, 0);
  raw.set(trailer, data.length);
  // Continuous base64 stream (matches gateway streaming encoder after Close).
  return bytesToBase64(raw);
}

describe("createGrpcWebClient text format", () => {
  it("decodes continuous base64 response bodies", async () => {
    const payload = new Uint8Array([10, 5, 104, 101, 108, 108, 111]); // arbitrary protobuf-ish bytes
    const b64 = buildTextResponse(payload);

    const client = createGrpcWebClient({
      baseUrl: "http://gateway.test",
      format: "text",
      fetch: async () =>
        new Response(b64, {
          status: 200,
          headers: { "Content-Type": "application/grpc-web-text+proto" },
        }),
    });

    const res = await client.unary("/test.Echo/Echo", new Uint8Array([1]));
    assert.deepEqual(res.message, payload);
    assert.equal(res.grpcStatus, GrpcCode.OK);
  });

  it("rejects invalid base64 response bodies", async () => {
    const client = createGrpcWebClient({
      baseUrl: "http://gateway.test",
      format: "text",
      fetch: async () => new Response("A", { status: 200 }),
    });

    await assert.rejects(() => client.unary("/test.Echo/Echo", new Uint8Array([1])), (err) => {
      assert.match(String(err.message), /invalid grpc-web-text base64/i);
      return true;
    });
  });

  it("rejects mid-stream padding (legacy per-Write Encode)", async () => {
    // Two independently padded base64 chunks concatenated — invalid continuous stream.
    const broken = `${btoa("ab")}==${btoa("cdef")}`; // inject padding mid-stream
    const client = createGrpcWebClient({
      baseUrl: "http://gateway.test",
      format: "text",
      fetch: async () => new Response(broken, { status: 200 }),
    });

    await assert.rejects(() => client.unary("/test.Echo/Echo", new Uint8Array([1])), (err) => {
      assert.match(String(err.message), /invalid grpc-web-text base64|expected 1 response/i);
      return true;
    });
  });

  it("surfaces grpc-status errors from text trailers", async () => {
    const b64 = buildTextResponse(new Uint8Array(), "3", "bad%20name");
    const client = createGrpcWebClient({
      baseUrl: "http://gateway.test",
      format: "text",
      fetch: async () => new Response(b64, { status: 200 }),
    });

    await assert.rejects(() => client.unary("/test.Echo/Echo", new Uint8Array()), (err) => {
      assert.equal(err.code, GrpcCode.InvalidArgument);
      assert.match(err.message, /bad name/);
      return true;
    });
  });

  it("sends base64 request body for text format", async () => {
    let sawBody = "";
    let sawCT = "";
    const payload = new Uint8Array([7, 8, 9]);
    const b64 = buildTextResponse(new Uint8Array([1]));

    const client = createGrpcWebClient({
      baseUrl: "http://gateway.test",
      format: "text",
      fetch: async (_url, init) => {
        sawBody = String(init.body);
        sawCT = new Headers(init.headers).get("Content-Type") ?? "";
        return new Response(b64, { status: 200 });
      },
    });

    await client.unary("/svc/M", payload);
    assert.equal(sawCT, "application/grpc-web-text+proto");
    // Request is base64 of a 5+len frame wrapping payload.
    const decoded = Uint8Array.from(atob(sawBody), (c) => c.charCodeAt(0));
    assert.equal(decoded[0], 0);
    assert.equal(decoded.length, 5 + payload.length);
    assert.deepEqual(decoded.subarray(5), payload);
  });
});

describe("createGrpcWebClient binary format", () => {
  it("round-trips binary frames", async () => {
    const payload = new Uint8Array([9, 9, 9]);
    const data = encodeFrame(payload);
    const trailer = encodeFrame(new TextEncoder().encode("grpc-status: 0\r\n"), TRAILER_FLAG);
    const body = new Uint8Array(data.length + trailer.length);
    body.set(data, 0);
    body.set(trailer, data.length);

    const client = createGrpcWebClient({
      baseUrl: "http://gateway.test",
      format: "binary",
      fetch: async () =>
        new Response(body, {
          status: 200,
          headers: { "Content-Type": "application/grpc-web+proto" },
        }),
    });

    const res = await client.unary("/test.Echo/Echo", new Uint8Array([1]));
    assert.deepEqual(res.message, payload);
  });
});
