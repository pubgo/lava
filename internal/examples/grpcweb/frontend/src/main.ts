import {
  createGatewayTransport,
  createHttpJsonClient,
  GatewayError,
  type GatewayTransportFormat,
} from "@pubgo/lava-gateway-client";
import type { RpcMetadata, RpcOptions } from "@protobuf-ts/runtime-rpc";
import { GreeterServiceClient } from "./generated/greeter.client";

const baseUrl = window.location.origin;

function readFormat(): GatewayTransportFormat {
  const raw = localStorage.getItem("rpc-transport-format");
  if (raw === "binary" || raw === "text" || raw === "json") return raw;
  return "json";
}

function demoToken(): string {
  return (document.getElementById("demoToken") as HTMLInputElement)?.value?.trim() || "";
}

function requestId(): string {
  const el = document.getElementById("requestId") as HTMLInputElement;
  const v = el?.value?.trim();
  if (v) return v;
  const generated = `req-${Date.now().toString(36)}`;
  if (el) el.value = generated;
  return generated;
}

/** Shared request metadata for every RPC call. */
function rpcOptions(signal?: AbortSignal, extraMeta?: Record<string, string>): RpcOptions {
  const meta: Record<string, string> = {
    "x-request-id": requestId(),
    ...(extraMeta ?? {}),
  };
  const token = demoToken();
  if (token) meta["x-demo-token"] = token;
  return { meta, abort: signal };
}

const rpc = {
  get format() {
    return readFormat();
  },
  setFormat(format: GatewayTransportFormat) {
    localStorage.setItem("rpc-transport-format", format);
    location.reload();
  },
  get greeter() {
    return new GreeterServiceClient(
      createGatewayTransport({
        baseUrl,
        format: readFormat(),
        acceptCompression: true,
      }),
    );
  },
};

const http = createHttpJsonClient({ baseUrl });

const helloNameInput = document.getElementById("helloName") as HTMLInputElement;
const helloJsonBtn = document.getElementById("helloJsonBtn") as HTMLButtonElement;
const helloGrpcBtn = document.getElementById("helloGrpcBtn") as HTMLButtonElement;
const helloResult = document.getElementById("helloResult") as HTMLDivElement;

const goodbyeNameInput = document.getElementById("goodbyeName") as HTMLInputElement;
const goodbyeJsonBtn = document.getElementById("goodbyeJsonBtn") as HTMLButtonElement;
const goodbyeGrpcBtn = document.getElementById("goodbyeGrpcBtn") as HTMLButtonElement;
const goodbyeResult = document.getElementById("goodbyeResult") as HTMLDivElement;

const watchNameInput = document.getElementById("watchName") as HTMLInputElement;
const watchCountInput = document.getElementById("watchCount") as HTMLInputElement;
const watchIntervalInput = document.getElementById("watchInterval") as HTMLInputElement;
const watchBtn = document.getElementById("watchBtn") as HTMLButtonElement;
const watchAbortBtn = document.getElementById("watchAbortBtn") as HTMLButtonElement;
const watchResult = document.getElementById("watchResult") as HTMLDivElement;
const watchStatus = document.getElementById("watchStatus") as HTMLDivElement;

const chatBtn = document.getElementById("chatBtn") as HTMLButtonElement;
const chatResult = document.getElementById("chatResult") as HTMLDivElement;

const metaRequest = document.getElementById("metaRequest") as HTMLDivElement;
const metaHeaders = document.getElementById("metaHeaders") as HTMLDivElement;
const metaTrailers = document.getElementById("metaTrailers") as HTMLDivElement;
const metaStatus = document.getElementById("metaStatus") as HTMLDivElement;

const logResult = document.getElementById("logResult") as HTMLDivElement;
const formatSelect = document.getElementById("transportFormat") as HTMLSelectElement | null;

let watchAbort: AbortController | null = null;

const logs: string[] = [];
function log(message: string) {
  const timestamp = new Date().toLocaleTimeString();
  logs.push(`[${timestamp}] ${message}`);
  if (logs.length > 80) logs.shift();
  logResult.textContent = logs.join("\n");
  logResult.scrollTop = logResult.scrollHeight;
}

function showResult(element: HTMLDivElement, data: unknown, isError = false) {
  element.textContent = typeof data === "object" ? JSON.stringify(data, null, 2) : String(data);
  element.className = "result " + (isError ? "error" : "success");
}

function showMeta(panel: HTMLDivElement, data: unknown) {
  panel.textContent = typeof data === "object" ? JSON.stringify(data, null, 2) : String(data ?? "-");
}

function formatErr(error: unknown): string {
  if (error instanceof GatewayError) {
    return `GatewayError code=${error.code} http=${error.httpStatus ?? "-"} msg=${error.grpcMessage}`;
  }
  if (error && typeof error === "object" && "code" in error && "message" in error) {
    const e = error as { code: string; message: string };
    return `${e.code}: ${e.message}`;
  }
  return error instanceof Error ? error.message : String(error);
}

function metaToRecord(meta: RpcMetadata | Record<string, string> | undefined): Record<string, string> {
  if (!meta) return {};
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(meta)) {
    if (v == null) continue;
    out[k] = Array.isArray(v) ? v.join(", ") : String(v);
  }
  return out;
}

function pickInteresting(meta: RpcMetadata | Record<string, string> | undefined): Record<string, string> {
  const flat = metaToRecord(meta);
  const keys = [
    "x-demo-echo",
    "x-demo-method",
    "x-demo-trailer",
    "x-demo-stream-count",
    "x-demo-stream-mode",
    "x-demo-interval-ms",
    "x-example-mw",
    "x-request-id",
    "grpc-status",
    "grpc-message",
    "grpc-encoding",
    "content-type",
  ];
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(flat)) {
    const lower = k.toLowerCase();
    if (keys.includes(lower) || lower.startsWith("x-demo") || lower.startsWith("x-example")) {
      out[k] = v;
    }
  }
  return out;
}

async function captureUnaryMeta(
  call: {
    headers: Promise<RpcMetadata>;
    trailers: Promise<RpcMetadata>;
    status: Promise<{ code: string; detail: string }>;
  },
  requestMeta: Record<string, string>,
) {
  showMeta(metaRequest, requestMeta);
  try {
    const [headers, trailers, status] = await Promise.all([call.headers, call.trailers, call.status]);
    showMeta(metaHeaders, pickInteresting(headers));
    showMeta(metaTrailers, pickInteresting(trailers));
    showMeta(metaStatus, status);
    return { headers: metaToRecord(headers), trailers: metaToRecord(trailers), status };
  } catch (err) {
    showMeta(metaStatus, formatErr(err));
    throw err;
  }
}

async function testHelloREST() {
  const name = helloNameInput.value || "Anonymous";
  log(`[HTTP REST] POST /v1/greeter/hello name=${name}`);
  try {
    const { data, headers } = await http.post<{ name: string }, { message: string; timestamp?: string }>(
      "/v1/greeter/hello",
      { name },
      { headers: { "x-demo-token": demoToken(), "x-request-id": requestId() } },
    );
    log(`[HTTP REST] ok mw=${headers.get("x-example-mw") ?? "-"} echo=${headers.get("x-demo-echo") ?? "-"}`);
    showResult(helloResult, data);
    showMeta(metaRequest, { "x-demo-token": demoToken(), "x-request-id": requestId() });
    showMeta(metaHeaders, Object.fromEntries(headers.entries()));
    showMeta(metaTrailers, {});
    showMeta(metaStatus, { code: "OK", via: "HTTP" });
  } catch (error) {
    log(`[HTTP REST] error: ${formatErr(error)}`);
    showResult(helloResult, formatErr(error), true);
  }
}

async function testGoodbyeREST() {
  const name = goodbyeNameInput.value || "Anonymous";
  log(`[HTTP REST] POST /v1/greeter/goodbye name=${name}`);
  try {
    const { data, headers } = await http.post<{ name: string }, { message: string; timestamp?: string }>(
      "/v1/greeter/goodbye",
      { name },
      { headers: { "x-demo-token": demoToken(), "x-request-id": requestId() } },
    );
    log(`[HTTP REST] ok mw=${headers.get("x-example-mw") ?? "-"}`);
    showResult(goodbyeResult, data);
  } catch (error) {
    log(`[HTTP REST] error: ${formatErr(error)}`);
    showResult(goodbyeResult, formatErr(error), true);
  }
}

async function testHelloRPC() {
  const name = helloNameInput.value || "Anonymous";
  const opts = rpcOptions();
  log(`[rpc.greeter] SayHello via ${rpc.format} name=${name} meta=${JSON.stringify(opts.meta)}`);
  try {
    const call = rpc.greeter.sayHello({ name }, opts);
    const meta = await captureUnaryMeta(call, opts.meta as Record<string, string>);
    const response = await call.response;
    log(`[rpc.greeter] status=${meta.status.code} echo=${meta.headers["x-demo-echo"] ?? meta.headers["X-Demo-Echo"] ?? "-"}`);
    showResult(helloResult, {
      message: response.message,
      timestamp: response.timestamp.toString(),
      transport: rpc.format,
      headers: pickInteresting(meta.headers),
      trailers: pickInteresting(meta.trailers),
      status: meta.status,
    });
  } catch (error) {
    log(`[rpc.greeter] error: ${formatErr(error)}`);
    showResult(helloResult, formatErr(error), true);
  }
}

async function testGoodbyeRPC() {
  const name = goodbyeNameInput.value || "Anonymous";
  const opts = rpcOptions();
  log(`[rpc.greeter] SayGoodbye via ${rpc.format} name=${name}`);
  try {
    const call = rpc.greeter.sayGoodbye({ name }, opts);
    const meta = await captureUnaryMeta(call, opts.meta as Record<string, string>);
    const response = await call.response;
    log(`[rpc.greeter] status=${meta.status.code}`);
    showResult(goodbyeResult, {
      message: response.message,
      timestamp: response.timestamp.toString(),
      transport: rpc.format,
      headers: pickInteresting(meta.headers),
      trailers: pickInteresting(meta.trailers),
      status: meta.status,
    });
  } catch (error) {
    log(`[rpc.greeter] error: ${formatErr(error)}`);
    showResult(goodbyeResult, formatErr(error), true);
  }
}

async function testWatchHello() {
  if (watchAbort) {
    log("[stream] already subscribed — cancel first");
    return;
  }
  const name = watchNameInput.value || "Anonymous";
  const count = Number.parseInt(watchCountInput.value || "0", 10);
  const intervalMs = Number.parseInt(watchIntervalInput.value || "1000", 10);
  watchAbort = new AbortController();
  const opts = rpcOptions(watchAbort.signal, {
    "x-demo-interval-ms": String(Math.max(200, intervalMs || 1000)),
    ...(count <= 0 ? { "x-demo-subscribe": "1" } : {}),
  });

  const mode = count <= 0 ? "subscribe" : `finite(${count})`;
  log(`[rpc.greeter] WatchHello via ${rpc.format} name=${name} mode=${mode} interval=${intervalMs}ms`);
  watchStatus.textContent = `订阅中 · ${mode} · ${rpc.format}`;
  watchStatus.className = "result success";
  showResult(watchResult, { live: true, received: 0, messages: [] });
  showMeta(metaRequest, opts.meta);

  const messages: Array<{ n: number; message: string; at: string }> = [];
  try {
    const call = rpc.greeter.watchHello({ name, count }, opts);
    void call.headers.then((h) => {
      const flat = metaToRecord(h);
      showMeta(metaHeaders, pickInteresting(h));
      log(
        `[stream] headers mode=${flat["x-demo-stream-mode"] ?? "-"} interval=${flat["x-demo-interval-ms"] ?? "-"}`,
      );
    });

    for await (const msg of call.responses) {
      messages.push({
        n: messages.length + 1,
        message: msg.message,
        at: new Date().toLocaleTimeString(),
      });
      // Keep a rolling window so the panel stays readable during long subscriptions.
      const view = messages.slice(-30);
      showResult(watchResult, {
        live: true,
        received: messages.length,
        showing: view.length,
        latest: view[view.length - 1],
        messages: view,
      });
      watchStatus.textContent = `订阅中 · 已收 ${messages.length} 条 · ${rpc.format}`;
      log(`[stream] #${messages.length} ${msg.message}`);
    }

    const [trailers, status] = await Promise.all([call.trailers, call.status]);
    showMeta(metaTrailers, pickInteresting(trailers));
    showMeta(metaStatus, status);
    watchStatus.textContent = `已结束 · 共 ${messages.length} 条 · ${status.code}`;
    watchStatus.className = "result success";
    showResult(watchResult, {
      live: false,
      received: messages.length,
      messages: messages.slice(-30),
      trailers: pickInteresting(trailers),
      status,
    });
    log(`[stream] done status=${status.code} frames=${messages.length}`);
  } catch (error) {
    const aborted = watchAbort?.signal.aborted || (error instanceof DOMException && error.name === "AbortError");
    if (aborted) {
      watchStatus.textContent = `已取消订阅 · 共收 ${messages.length} 条`;
      watchStatus.className = "result";
      showResult(watchResult, { live: false, cancelled: true, received: messages.length, messages: messages.slice(-30) });
      log(`[stream] unsubscribed after ${messages.length} messages`);
    } else {
      log(`[stream] error: ${formatErr(error)}`);
      watchStatus.textContent = `出错 · ${formatErr(error)}`;
      watchStatus.className = "result error";
      showResult(watchResult, { error: formatErr(error), partial: messages.slice(-30) }, true);
      showMeta(metaStatus, formatErr(error));
    }
  } finally {
    watchAbort = null;
  }
}

function abortWatch() {
  if (!watchAbort) {
    log("[stream] not subscribed");
    return;
  }
  log("[stream] unsubscribe / abort()");
  watchAbort.abort();
}

async function testChatUnsupported() {
  log(`[rpc.greeter] Chat via ${rpc.format} (expect UNIMPLEMENTED on HTTP/gRPC-Web)`);
  try {
    const call = rpc.greeter.chat(rpcOptions());
    // DuplexStreamingCall may throw synchronously from transport.
    await call.headers;
    showResult(chatResult, "unexpected success", true);
  } catch (error) {
    log(`[rpc.greeter] Chat: ${formatErr(error)}`);
    showResult(
      chatResult,
      {
        expected: true,
        error: formatErr(error),
        hint: "Use internal/examples/grpcwebsocket for bidi Chat over WebSocket",
      },
      true,
    );
    showMeta(metaStatus, formatErr(error));
  }
}

helloJsonBtn.addEventListener("click", testHelloREST);
helloGrpcBtn.addEventListener("click", testHelloRPC);
goodbyeJsonBtn.addEventListener("click", testGoodbyeREST);
goodbyeGrpcBtn.addEventListener("click", testGoodbyeRPC);
watchBtn.addEventListener("click", () => void testWatchHello());
watchAbortBtn.addEventListener("click", abortWatch);
chatBtn.addEventListener("click", () => void testChatUnsupported());

if (formatSelect) {
  formatSelect.value = rpc.format;
  formatSelect.addEventListener("change", () => {
    rpc.setFormat(formatSelect.value as GatewayTransportFormat);
  });
}

log("SDK ready: unary + server-stream + headers/trailers");
log(`baseUrl=${baseUrl} transport=${rpc.format}`);
(window as unknown as { rpc: typeof rpc }).rpc = rpc;
