# Project Guidelines for Lava

## Scope

These instructions are project-wide defaults for this repository. Keep changes focused, minimal, and aligned with existing patterns.

## Architecture

- `lava/`: core public interfaces and contracts (`Middleware`, routers, request/response abstractions).
- `core/`: runtime capabilities (supervisor, scheduler, tunnel, logging/metrics/tracing, debug, DI builder).
- `servers/`: service hosts (`gatewayserver` multi-protocol gateway, `https` on Fiber, `zrpcs` on NATS).
- `clients/`: outbound client implementations (`grpcc`, `resty`).
- `pkg/`: reusable public utilities/components (including gateway and helpers).
- `internal/`: repository-internal implementation details/examples; avoid exposing as public API.

Place new code by responsibility:
- Cross-protocol abstractions -> `lava/`
- Runtime framework capability -> `core/<module>/`
- HTTP/gRPC serving behavior -> `servers/`
- Reusable public helper/component -> `pkg/`
- Internal-only implementation -> `internal/`

## Build and Test

Run commands from repository root.

Primary local workflow (source of truth: `Taskfile.yml`):
- `task test` (Go tests: short + race + cover)
- `task lint` (golangci-lint)

When `.proto` files or `protobuf.yaml` change, run:
- `task proto:fmt`
- `task proto:lint`
- `task proto:gen`
- then `task test` and `task lint`

CI reference is `.github/workflows/lint-test.yml` (lint + gotestsum-based tests).

## Conventions

- Follow standard Go formatting and idioms (`gofmt`, package naming, error-last returns).
- Prefer wrapping errors with context (project commonly uses `github.com/pkg/errors`).
- Keep public APIs and behavior stable unless the task explicitly requires breaking changes.
- Do not edit generated protobuf files (`*.pb.go`) manually; regenerate via proto tasks.
- Prefer updating existing module patterns instead of introducing new architectural styles.

## Pitfalls to Avoid

- `Taskfile.yml` is authoritative for local commands; docs are guidance.
- Lint settings may apply automatic fixes (`.golangci.yaml` has `issues.fix: true`), so re-check diffs after lint.
- Go toolchain target is defined in `go.mod` (`go 1.25.0`).
- For protobuf generation, keep `protobuf.yaml` base module aligned with Go module path (`github.com/pubgo/lava/v2/pkg`).

## Key References

- `README.md`
- `docs/architecture-v2.md`
- `docs/design-v2.md`
- `docs/development.md`
- `Taskfile.yml`
- `.github/workflows/lint-test.yml`

## Representative Examples

- Service lifecycle and management: `core/supervisor/`
- DI registration patterns: `core/lavabuilder/`
- HTTP server composition: `servers/https/server.go`
- gRPC + gateway composition: `servers/gatewayserver/server.go`
- Gateway behavior and routing: `pkg/gateway/`
