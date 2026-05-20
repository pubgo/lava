import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport';
import { GreeterServiceClient } from './generated/greeter.client';

// 创建 gRPC Web Transport
const transport = new GrpcWebFetchTransport({
    baseUrl: window.location.origin,
    format: 'binary', // 使用二进制格式 (application/grpc-web+proto)
});

// 创建 gRPC 客户端
const client = new GreeterServiceClient(transport);

// DOM 元素
const helloNameInput = document.getElementById('helloName') as HTMLInputElement;
const helloJsonBtn = document.getElementById('helloJsonBtn') as HTMLButtonElement;
const helloGrpcBtn = document.getElementById('helloGrpcBtn') as HTMLButtonElement;
const helloResult = document.getElementById('helloResult') as HTMLDivElement;

const goodbyeNameInput = document.getElementById('goodbyeName') as HTMLInputElement;
const goodbyeJsonBtn = document.getElementById('goodbyeJsonBtn') as HTMLButtonElement;
const goodbyeGrpcBtn = document.getElementById('goodbyeGrpcBtn') as HTMLButtonElement;
const goodbyeResult = document.getElementById('goodbyeResult') as HTMLDivElement;

const logResult = document.getElementById('logResult') as HTMLDivElement;

// 日志记录
const logs: string[] = [];
function log(message: string) {
    const timestamp = new Date().toLocaleTimeString();
    logs.push(`[${timestamp}] ${message}`);
    if (logs.length > 50) {
        logs.shift();
    }
    logResult.textContent = logs.join('\n');
    logResult.scrollTop = logResult.scrollHeight;
}

// 显示结果
function showResult(element: HTMLDivElement, data: unknown, isError = false) {
    element.textContent = typeof data === 'object' ? JSON.stringify(data, null, 2) : String(data);
    element.className = 'result ' + (isError ? 'error' : 'success');
}

// HTTP/JSON 请求
async function testHelloJSON() {
    const name = helloNameInput.value || 'Anonymous';
    log(`[HTTP/JSON] 发送请求到 /v1/greeter/hello，name=${name}`);
    
    try {
        const response = await fetch('/v1/greeter/hello', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name })
        });
        
        const data = await response.json();
        log(`[HTTP/JSON] 收到响应: ${JSON.stringify(data)}`);
        showResult(helloResult, data);
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        log(`[HTTP/JSON] 错误: ${message}`);
        showResult(helloResult, message, true);
    }
}

async function testGoodbyeJSON() {
    const name = goodbyeNameInput.value || 'Anonymous';
    log(`[HTTP/JSON] 发送请求到 /v1/greeter/goodbye，name=${name}`);
    
    try {
        const response = await fetch('/v1/greeter/goodbye', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name })
        });
        
        const data = await response.json();
        log(`[HTTP/JSON] 收到响应: ${JSON.stringify(data)}`);
        showResult(goodbyeResult, data);
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        log(`[HTTP/JSON] 错误: ${message}`);
        showResult(goodbyeResult, message, true);
    }
}

// gRPC Web 请求 (使用 protobuf-ts)
async function testHelloGRPC() {
    const name = helloNameInput.value || 'Anonymous';
    log(`[gRPC Web] 调用 GreeterService.SayHello，name=${name}`);
    
    try {
        const call = client.sayHello({ name });
        
        // 等待响应
        const response = await call.response;
        
        log(`[gRPC Web] 收到响应: message=${response.message}, timestamp=${response.timestamp}`);
        showResult(helloResult, {
            message: response.message,
            timestamp: response.timestamp.toString()
        });
        
        // 显示状态
        const status = await call.status;
        log(`[gRPC Web] 状态: ${status.code}`);
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        log(`[gRPC Web] 错误: ${message}`);
        showResult(helloResult, message, true);
    }
}

async function testGoodbyeGRPC() {
    const name = goodbyeNameInput.value || 'Anonymous';
    log(`[gRPC Web] 调用 GreeterService.SayGoodbye，name=${name}`);
    
    try {
        const call = client.sayGoodbye({ name });
        
        // 等待响应
        const response = await call.response;
        
        log(`[gRPC Web] 收到响应: message=${response.message}, timestamp=${response.timestamp}`);
        showResult(goodbyeResult, {
            message: response.message,
            timestamp: response.timestamp.toString()
        });
        
        // 显示元数据
        const status = await call.status;
        log(`[gRPC Web] 状态: ${status.code}`);
    } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        log(`[gRPC Web] 错误: ${message}`);
        showResult(goodbyeResult, message, true);
    }
}

// 绑定事件
helloJsonBtn.addEventListener('click', testHelloJSON);
helloGrpcBtn.addEventListener('click', testHelloGRPC);
goodbyeJsonBtn.addEventListener('click', testGoodbyeJSON);
goodbyeGrpcBtn.addEventListener('click', testGoodbyeGRPC);

// 初始化日志
log('页面加载完成，protobuf-ts gRPC Web 客户端已初始化');
log(`Transport: GrpcWebFetchTransport, baseUrl: ${window.location.origin}`);
