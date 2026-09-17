const TRAILER_FLAG = 0x80;
const COMPRESSED_FLAG = 0x01;

export type GrpcWebFrame = {
  flags: number;
  payload: Uint8Array;
  isTrailer: boolean;
  isCompressed: boolean;
};

export function encodeFrame(payload: Uint8Array, flags = 0): Uint8Array {
  const out = new Uint8Array(5 + payload.length);
  out[0] = flags & 0xff;
  const view = new DataView(out.buffer);
  view.setUint32(1, payload.length, false);
  out.set(payload, 5);
  return out;
}

export function decodeFrames(buf: Uint8Array): GrpcWebFrame[] {
  const frames: GrpcWebFrame[] = [];
  let off = 0;
  while (off + 5 <= buf.length) {
    const flags = buf[off]!;
    const length =
      ((buf[off + 1]! << 24) |
        (buf[off + 2]! << 16) |
        (buf[off + 3]! << 8) |
        buf[off + 4]!) >>>
      0;
    off += 5;
    if (off + length > buf.length) {
      break;
    }
    const payload = buf.subarray(off, off + length);
    off += length;
    frames.push({
      flags,
      payload,
      isTrailer: (flags & TRAILER_FLAG) !== 0,
      isCompressed: (flags & COMPRESSED_FLAG) !== 0,
    });
  }
  return frames;
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
