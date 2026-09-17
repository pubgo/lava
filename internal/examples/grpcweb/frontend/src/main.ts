import {
  createGatewayTransport,
  createHttpJsonClient,
  GatewayError,
  type GatewayTransportFormat,
} from "@pubgo/lava-gateway-client";
import { GreeterServiceClient } from "./generated/greeter.client";

/**
 * Demo RPC facade — same shape as agentrun `src/lib/rpc.ts`:
 * one transport factory + generated service clients.
 */
const baseUrl = window.location.origin;

function readFormat(): GatewayTransportFormat {
  const raw = localStorage.getItem("rpc-transport-format");
  if (raw === "binary" || raw === "text" || raw === "json") return raw;
  // Dev default: JSON over gRPC full-method path (readable in Network panel).
  return "json";
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

const logResult = document.getElementById("logResult") as HTMLDivElement;
const formatSelect = document.getElementById("transportFormat") as HTMLSelectElement | null;

const logs: string[] = [];
function log(message: string) {
  const timestamp = new Date().toLocaleTimeString();
  logs.push(`[${timestamp}] ${message}`);
  if (logs.length > 50) logs.shift();
  logResult.textContent = logs.join("\n");
  logResult.scrollTop = logResult.scrollHeight;
}

function showResult(element: HTMLDivElement, data: unknown, isError = false) {
  element.textContent = typeof data === "object" ? JSON.stringify(data, null, 2) : String(data);
  element.className = "result " + (isError ? "error" : "success");
}

function formatErr(error: unknown): string {
  if (error instanceof GatewayError) {
    return `GatewayError code=${error.code} http=${error.httpStatus ?? "-"} msg=${error.grpcMessage}`;
  }
  return error instanceof Error ? error.message : String(error);
}

async function testHelloREST() {
  const name = helloNameInput.value || "Anonymous";
  log(`[HTTP REST] POST /v1/greeter/hello name=${name}`);
  try {
    const { data, headers } = await http.post<{ name: string }, { message: string; timestamp?: string }>(
      "/v1/greeter/hello",
      { name },
    );
    log(`[HTTP REST] ok mw=${headers.get("x-example-mw") ?? "-"}`);
    showResult(helloResult, data);
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
  log(`[rpc.greeter] SayHello via ${rpc.format} name=${name}`);
  try {
    const call = rpc.greeter.sayHello({ name });
    const response = await call.response;
    const status = await call.status;
    log(`[rpc.greeter] status=${status.code} message=${response.message}`);
    showResult(helloResult, {
      message: response.message,
      timestamp: response.timestamp.toString(),
      status: status.code,
      transport: rpc.format,
    });
  } catch (error) {
    log(`[rpc.greeter] error: ${formatErr(error)}`);
    showResult(helloResult, formatErr(error), true);
  }
}

async function testGoodbyeRPC() {
  const name = goodbyeNameInput.value || "Anonymous";
  log(`[rpc.greeter] SayGoodbye via ${rpc.format} name=${name}`);
  try {
    const call = rpc.greeter.sayGoodbye({ name });
    const response = await call.response;
    const status = await call.status;
    log(`[rpc.greeter] status=${status.code} message=${response.message}`);
    showResult(goodbyeResult, {
      message: response.message,
      timestamp: response.timestamp.toString(),
      status: status.code,
      transport: rpc.format,
    });
  } catch (error) {
    log(`[rpc.greeter] error: ${formatErr(error)}`);
    showResult(goodbyeResult, formatErr(error), true);
  }
}

helloJsonBtn.addEventListener("click", testHelloREST);
helloGrpcBtn.addEventListener("click", testHelloRPC);
goodbyeJsonBtn.addEventListener("click", testGoodbyeREST);
goodbyeGrpcBtn.addEventListener("click", testGoodbyeRPC);

if (formatSelect) {
  formatSelect.value = rpc.format;
  formatSelect.addEventListener("change", () => {
    rpc.setFormat(formatSelect.value as GatewayTransportFormat);
  });
}

log("SDK ready: @pubgo/lava-gateway-client (agentrun-style rpc facade)");
log(`baseUrl=${baseUrl} transport=${rpc.format}`);
log("Tip: switch transport with the dropdown (json | binary | text)");

// Expose for console debugging, same spirit as agentrun rpc helpers.
(window as unknown as { rpc: typeof rpc }).rpc = rpc;
