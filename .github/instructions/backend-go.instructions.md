---
name: backend-go
description: "Use when: editing Go backend code in Lava (core/servers/clients/pkg/lava/cmds/internal), including refactors, bug fixes, and tests."
applyTo: "{core,servers,clients,pkg,lava,cmds,internal}/**/*.go"
---

# Lava Backend Go Instructions

- Keep changes minimal and consistent with existing patterns in the target module.
- Place code by responsibility:
  - Cross-protocol contracts -> `lava/`
  - Runtime framework capabilities -> `core/<module>/`
  - Service hosting behavior -> `servers/`
  - Outbound client behavior -> `clients/`
  - Reusable public helpers/components -> `pkg/`
  - Internal-only implementation -> `internal/`
- Preserve public APIs unless the task explicitly requires a breaking change.
- Prefer contextual error wrapping; this repository commonly uses `github.com/pkg/errors`.
- Never hand-edit generated protobuf files (`*.pb.go`); regenerate via proto tasks.
- Follow Go idioms (`gofmt`, package naming, error-last returns).
- Reuse existing module patterns before introducing new architectural styles.

## Validation

- Run from repository root.
- Standard verification flow:
  - `task test`
  - `task lint`
- If `.proto` or `protobuf.yaml` changes are involved, run first:
  - `task proto:fmt`
  - `task proto:lint`
  - `task proto:gen`

## Pitfalls

- Treat `Taskfile.yml` as the local source of truth.
- Lint may auto-fix files (`.golangci.yaml` sets `issues.fix: true`); re-check diffs after lint.
- Keep Go toolchain aligned with `go.mod` (`go 1.25.0`).