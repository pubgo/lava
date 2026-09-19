const TRAILER_FLAG = 0x80;
const COMPRESSED_FLAG = 0x01;

/** 1 flag byte + 4 big-endian length bytes. */
export const FRAME_HEADER_SIZE = 5;

/**
 * Mirrors `grpcMaxRecvMsgSize` in pkg/gateway/stream.http.go. A 4-byte length
 * prefix is caller-controlled: without the same bound here, a bogus length makes
 * the live reader wait forever for bytes that can never arrive.
 */
export const MAX_MESSAGE_SIZE = 4 << 20;

export type GrpcWebFrame = {
  flags: number;
  payload: Uint8Array;
  isTrailer: boolean;
  isCompressed: boolean;
};

export function encodeFrame(payload: Uint8Array, flags = 0): Uint8Array {
  const out = new Uint8Array(FRAME_HEADER_SIZE + payload.length);
  out[0] = flags & 0xff;
  const view = new DataView(out.buffer);
  view.setUint32(1, payload.length, false);
  out.set(payload, FRAME_HEADER_SIZE);
  return out;
}

export function decodeFrames(buf: Uint8Array): GrpcWebFrame[] {
  const { frames } = extractFrames(buf);
  return frames;
}

/** Incremental gRPC-Web frame parser for live server-streaming. */
export class GrpcWebFrameReader {
  private buf: Uint8Array = new Uint8Array(0);

  push(chunk: Uint8Array): GrpcWebFrame[] {
    if (chunk.length === 0) return [];
    const next = new Uint8Array(this.buf.length + chunk.length);
    next.set(this.buf, 0);
    next.set(chunk, this.buf.length);
    const { frames, rest } = extractFrames(next);
    this.buf = new Uint8Array(rest);
    return frames;
  }

  /** Bytes not yet forming a complete frame. */
  get pending(): number {
    return this.buf.length;
  }
}

function extractFrames(buf: Uint8Array): { frames: GrpcWebFrame[]; rest: Uint8Array } {
  const frames: GrpcWebFrame[] = [];
  let off = 0;
  while (off + FRAME_HEADER_SIZE <= buf.length) {
    const flags = buf[off]!;
    const length =
      ((buf[off + 1]! << 24) |
        (buf[off + 2]! << 16) |
        (buf[off + 3]! << 8) |
        buf[off + 4]!) >>>
      0;
    if (length > MAX_MESSAGE_SIZE) {
      throw new Error(
        `grpc-web frame exceeds message limit: ${length} bytes > ${MAX_MESSAGE_SIZE}`,
      );
    }
    if (off + FRAME_HEADER_SIZE + length > buf.length) {
      break;
    }
    const payload = buf.subarray(off + FRAME_HEADER_SIZE, off + FRAME_HEADER_SIZE + length);
    off += FRAME_HEADER_SIZE + length;
    frames.push({
      flags,
      payload,
      isTrailer: (flags & TRAILER_FLAG) !== 0,
      isCompressed: (flags & COMPRESSED_FLAG) !== 0,
    });
  }
  return { frames, rest: buf.subarray(off) };
}

/**
 * Streaming base64 → bytes decoder for grpc-web-text bodies.
 * Decodes complete 4-char groups as they arrive; call {@link finish} at EOF.
 */
export class Base64ByteDecoder {
  private pending = "";

  push(text: string): Uint8Array {
    this.pending += text.replace(/[^A-Za-z0-9+/=]/g, "");
    const complete = this.pending.length - (this.pending.length % 4);
    if (complete === 0) return new Uint8Array(0);
    const chunk = this.pending.slice(0, complete);
    this.pending = this.pending.slice(complete);
    return base64ChunkToBytes(chunk);
  }

  finish(): Uint8Array {
    if (!this.pending) return new Uint8Array(0);
    const out = base64ChunkToBytes(this.pending);
    this.pending = "";
    return out;
  }
}

function base64ChunkToBytes(text: string): Uint8Array {
  try {
    const bin = atob(text);
    const out = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
    return out;
  } catch (err) {
    throw new Error(`invalid grpc-web-text base64: ${err instanceof Error ? err.message : String(err)}`);
  }
}

export function parseTrailerHeaders(payload: Uint8Array): Headers {
  const text = new TextDecoder().decode(payload);
  const headers = new Headers();
  for (const line of text.split(/\r?\n/)) {
    if (!line.trim()) continue;
    const idx = line.indexOf(":");
    if (idx <= 0) continue;
    headers.append(line.slice(0, idx).trim(), line.slice(idx + 1).trim());
  }
  return headers;
}

export async function gzipCompress(data: Uint8Array): Promise<Uint8Array> {
  if (typeof CompressionStream === "undefined") {
    throw new Error("gzip compression requires CompressionStream (modern browsers / Node 18+)");
  }
  const stream = new Blob([data as BlobPart]).stream().pipeThrough(new CompressionStream("gzip"));
  const ab = await new Response(stream).arrayBuffer();
  return new Uint8Array(ab);
}

export async function gzipDecompress(data: Uint8Array): Promise<Uint8Array> {
  if (typeof DecompressionStream === "undefined") {
    throw new Error("gzip decompression requires DecompressionStream (modern browsers / Node 18+)");
  }
  const stream = new Blob([data as BlobPart]).stream().pipeThrough(new DecompressionStream("gzip"));
  const ab = await new Response(stream).arrayBuffer();
  return new Uint8Array(ab);
}

export { TRAILER_FLAG, COMPRESSED_FLAG };
