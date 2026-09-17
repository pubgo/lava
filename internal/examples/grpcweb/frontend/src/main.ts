import {
  createHttpJsonClient,
  createGrpcWebTransport,
  GatewayError,
} from "@pubgo/lava-gateway-client";
import { GreeterServiceClient } from "./generated/greeter.client";

const baseUrl = window.location.origin;

const http = createHttpJsonClient({ baseUrl });
const transport = createGrpcWebTransport({
  baseUrl,
  acceptCompression: true,
});
const grpc = new GreeterServiceClient(transport);

const helloNameInput = document.getElementById("helloName") as HTMLInputElement;
const helloJsonBtn = document.getElementById("helloJsonBtn") as HTMLButtonElement;
const helloGrpcBtn = document.getElementById("helloGrpcBtn") as HTMLButtonElement;
const helloResult = document.getElementById("helloResult") as HTMLDivElement;

const goodbyeNameInput = document.getElementById("goodbyeName") as HTMLInputElement;
const goodbyeJsonBtn = document.getElementById("goodbyeJsonBtn") as HTMLButtonElement;
const goodbyeGrpcBtn = document.getElementById("goodbyeGrpcBtn") as HTMLButtonElement;
const goodbyeResult = document.getElementById("goodbyeResult") as HTMLDivElement;

const logResult = document.getElementById("logResult") as HTMLDivElement;

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

async function testHelloJSON() {
  const name = helloNameInput.value || "Anonymous";
  log(`[HTTP/JSON SDK] POST /v1/greeter/hello name=${name}`);
  try {
    const { data, headers } = await http.post<{ name: string }, { message: string; timestamp?: string }>(
      "/v1/greeter/hello",
      { name },
    );
    log(`[HTTP/JSON SDK] ok mw=${headers.get("x-example-mw") ?? "-"}`);
    showResult(helloResult, data);
  } catch (error) {
    log(`[HTTP/JSON SDK] error: ${formatErr(error)}`);
    showResult(helloResult, formatErr(error), true);
  }
}

async function testGoodbyeJSON() {
  const name = goodbyeNameInput.value || "Anonymous";
  log(`[HTTP/JSON SDK] POST /v1/greeter/goodbye name=${name}`);
  try {
    const { data, headers } = await http.post<{ name: string }, { message: string; timestamp?: string }>(
      "/v1/greeter/goodbye",
      { name },
    );
    log(`[HTTP/JSON SDK] ok mw=${headers.get("x-example-mw") ?? "-"}`);
    showResult(goodbyeResult, data);
  } catch (error) {
    log(`[HTTP/JSON SDK] error: ${formatErr(error)}`);
    showResult(goodbyeResult, formatErr(error), true);
  }
}

async function testHelloGRPC() {
  const name = helloNameInput.value || "Anonymous";
  log(`[gRPC-Web SDK] GreeterService.SayHello name=${name}`);
  try {
    const call = grpc.sayHello({ name });
    const response = await call.response;
    const status = await call.status;
    log(`[gRPC-Web SDK] status=${status.code} message=${response.message}`);
    showResult(helloResult, {
      message: response.message,
      timestamp: response.timestamp.toString(),
      status: status.code,
    });
  } catch (error) {
    log(`[gRPC-Web SDK] error: ${formatErr(error)}`);
    showResult(helloResult, formatErr(error), true);
  }
}

async function testGoodbyeGRPC() {
  const name = goodbyeNameInput.value || "Anonymous";
  log(`[gRPC-Web SDK] GreeterService.SayGoodbye name=${name}`);
  try {
    const call = grpc.sayGoodbye({ name });
    const response = await call.response;
    const status = await call.status;
    log(`[gRPC-Web SDK] status=${status.code} message=${response.message}`);
    showResult(goodbyeResult, {
      message: response.message,
      timestamp: response.timestamp.toString(),
      status: status.code,
    });
  } catch (error) {
    log(`[gRPC-Web SDK] error: ${formatErr(error)}`);
    showResult(goodbyeResult, formatErr(error), true);
  }
}

helloJsonBtn.addEventListener("click", testHelloJSON);
helloGrpcBtn.addEventListener("click", testHelloGRPC);
goodbyeJsonBtn.addEventListener("click", testGoodbyeJSON);
goodbyeGrpcBtn.addEventListener("click", testGoodbyeGRPC);

log("SDK ready: @pubgo/lava-gateway-client");
log(`baseUrl=${baseUrl} (HTTP/JSON + gRPC-Web transport, gzip accepted)`);
