import assert from "node:assert/strict";
import { describe, it } from "node:test";
import {
  COMPRESSED_FLAG,
  FRAME_HEADER_SIZE,
  MAX_MESSAGE_SIZE,
  TRAILER_FLAG,
  decodeFrames,
  encodeFrame,
  parseTrailerHeaders,
} from "../dist/frames.js";

describe("encodeFrame / decodeFrames", () => {
  it("round-trips a data frame", () => {
    const payload = new TextEncoder().encode("hello-grpc-web");
    const frame = encodeFrame(payload);
    assert.equal(frame.length, FRAME_HEADER_SIZE + payload.length);
    assert.equal(frame[0], 0);

    const frames = decodeFrames(frame);
    assert.equal(frames.length, 1);
    const first = frames[0];
    assert.ok(first);
    assert.equal(first.isTrailer, false);
    assert.equal(first.isCompressed, false);
    assert.deepEqual(first.payload, payload);
  });

  it("decodes data + trailer frames", () => {
    const data = encodeFrame(new Uint8Array([1, 2, 3]));
    const trailerBody = new TextEncoder().encode("grpc-status: 0\r\ngrpc-message: ok\r\n");
    const trailer = encodeFrame(trailerBody, TRAILER_FLAG);
    const buf = new Uint8Array(data.length + trailer.length);
    buf.set(data, 0);
    buf.set(trailer, data.length);

    const frames = decodeFrames(buf);
    assert.equal(frames.length, 2);
    assert.equal(frames[0].isTrailer, false);
    assert.deepEqual(frames[0].payload, new Uint8Array([1, 2, 3]));
    assert.equal(frames[1].isTrailer, true);

    const headers = parseTrailerHeaders(frames[1].payload);
    assert.equal(headers.get("grpc-status"), "0");
    assert.equal(headers.get("grpc-message"), "ok");
  });

  it("marks compressed flag", () => {
    const frame = encodeFrame(new Uint8Array([9]), COMPRESSED_FLAG);
    const frames = decodeFrames(frame);
    assert.equal(frames[0].isCompressed, true);
  });

  it("extracts frames incrementally via GrpcWebFrameReader", async () => {
    const { GrpcWebFrameReader, encodeFrame, TRAILER_FLAG } = await import("../dist/frames.js");
    const data = encodeFrame(new Uint8Array([1, 2, 3]));
    const trailer = encodeFrame(new TextEncoder().encode("grpc-status: 0\r\n"), TRAILER_FLAG);
    const reader = new GrpcWebFrameReader();
    assert.equal(reader.push(data.subarray(0, 3)).length, 0);
    const mid = reader.push(data.subarray(3));
    assert.equal(mid.length, 1);
    assert.deepEqual(mid[0].payload, new Uint8Array([1, 2, 3]));
    const end = reader.push(trailer);
    assert.equal(end.length, 1);
    assert.equal(end[0].isTrailer, true);
  });

  it("stops on truncated frame without throwing", () => {
    const frames = decodeFrames(new Uint8Array([0, 0, 0, 0, 10, 1, 2]));
    assert.equal(frames.length, 0);
  });

  it("rejects a frame header above the gateway message cap", async () => {
    // The gateway bounds an inbound frame at 4 MiB (grpcMaxRecvMsgSize) and errors
    // past it. Without the same bound here, a bogus length prefix makes the live
    // reader wait for bytes that can never arrive: no frame, no error.
    const { GrpcWebFrameReader } = await import("../dist/frames.js");
    const header = new Uint8Array(FRAME_HEADER_SIZE);
    new DataView(header.buffer).setUint32(1, 0x7fffffff, false);
    const overCap = new RegExp(String(MAX_MESSAGE_SIZE));

    assert.throws(() => decodeFrames(header), overCap);

    const reader = new GrpcWebFrameReader();
    assert.throws(() => reader.push(header), overCap);
  });

  it("rejects gzip inflate past MAX_MESSAGE_SIZE", async () => {
    // A tiny compressed payload can expand past 4 MiB; mirror Go LimitReader.
    const { gzipCompress, gzipDecompress } = await import("../dist/frames.js");
    const compressed = await gzipCompress(new Uint8Array(MAX_MESSAGE_SIZE + 1));
    assert.ok(compressed.byteLength < MAX_MESSAGE_SIZE, "fixture must stay small on the wire");
    await assert.rejects(
      () => gzipDecompress(compressed),
      (err) => {
        assert.match(String(err?.message ?? err), /exceeds limit/);
        return true;
      },
    );
  });
});
