# AGENTS.md

Guidance for coding agents working in `github.com/olive-io/gflow`.
Use this as the default operating playbook for edits, tests, and style.

## Repository Overview

- Language: Go (`go 1.24.5`).
- Main entrypoints:
  - `cmd/gflow-server/main.go`
  - `cmd/gflow-runner/main.go`
- Core code:
  - `server/`
  - `runner/`
  - `pkg/`
  - `plugins/`
- Protobuf/OpenAPI sources:
  - `api/types/*.proto`
  - `api/rpc/*.proto`
- Generated outputs:
  - `api/**/*.pb.go`
  - `api/**/*.pb.gw.go`
  - `api/**/*.pb.validate.go`

## Cursor/Copilot Rules Check

Requested rule locations were checked:

- `.cursorrules`: not present
- `.cursor/rules/`: not present
- `.github/copilot-instructions.md`: not present

If these files are added later, treat them as high-priority constraints and update this file.

## Build, Lint, and Test Commands

Run commands from repository root.

### Setup

- Download dependencies: `go mod download`
- Vendor dependencies: `make vendor`
- Tidy dependencies (only when needed): `go mod tidy`

### Build

- Build all packages: `go build ./...`
- Build server binary: `go build ./cmd/gflow-server`
- Build runner binary: `go build ./cmd/gflow-runner`

Note: `Makefile` declares `all: build`, but no `build` target is defined.
Prefer the direct `go build` commands above unless a build target is added.

### Lint and Static Analysis

- Vet all packages: `make vet` (runs `go vet ./...`)
- Lint with golint: `make lint` (runs `golint -set_exit_status ./..`)

Notes:

- `golint` is legacy and may not be preinstalled.
- Preserve existing Makefile behavior unless explicitly asked to fix it.

### Test Commands

- Full test suite: `make test`
- Full verbose suite: `go test -v ./...`
- Coverage/bench target: `make test-coverage`

### Single Test Recipes (important)

- Run one package:
  - `go test -v ./server/...`
- Run one exact test:
  - `go test -v ./server -run '^Test_parseDoc$'`
- Run one test in runner package:
  - `go test -v ./runner -run '^TestExtractTask$'`
- Run one subtest:
  - `go test -v ./some/pkg -run 'TestName/Subcase'`
- Disable cache while iterating:
  - `go test -v ./server -run '^Test_parseDoc$' -count=1`

### Protobuf/API Generation

- Install generator tools: `make install`
- Regenerate API artifacts: `make apis`

`make apis` generates gRPC, gateway, validation, tags, and OpenAPI assets.

## Code Style Guidelines

Follow existing code conventions first. These are patterns observed in this repository.

### Formatting and Imports

- Run `gofmt` on every changed `.go` file.
- Use standard import grouping:
  1) stdlib
  2) third-party
  3) internal `github.com/olive-io/gflow/...`
- Keep imports explicit; no dot imports.
- Use aliases only when needed for clarity/conflicts (e.g. `urlpkg`, `gwrt`).
- Remove unused imports and keep blocks stable.

### Types and Function Signatures

- Prefer concrete types; use `any` only where dynamic input is required.
- Keep `context.Context` as the first parameter for request-scoped operations.
- Use generics consistently in generic layers (see `server/dao/dao.go`).
- Return pointers for mutable/domain structs where existing code does so.

### Naming

- Exported symbols: `CamelCase`.
- Internal symbols: `camelCase`.
- Packages: short lowercase names.
- Test functions: `TestXxx` with descriptive names.
- Keep acronyms readable and consistent (`RPC`, `URL`, `ID`).

### Error Handling

- Return errors; avoid panic for normal paths.
- Add context and wrap with `%w`:
  - `fmt.Errorf("load config: %w", err)`
- Prefer early returns over deeply nested conditionals.
- Map internal errors at transport boundaries (e.g., gRPC status errors).

### Logging

- Logging stack is zap/otelzap.
- Prefer informative, operation-level log messages.
- Keep startup/shutdown and subsystem initialization logs explicit.

### Concurrency and Lifecycle

- Respect context cancellation in long-running code.
- For goroutines/services, surface errors via channels and shut down cleanly.
- Use bounded shutdown timeouts for server close paths.

### Testing

- Use table-driven tests for branch-heavy logic.
- `testing` is primary; `testify/assert` is used in parts of the repo.
- Keep tests deterministic unless randomness is intentional.
- Keep fixtures under package-local `testdata/`.

### Generated and Vendor Code

- Do not hand-edit generated protobuf/gateway/validate files.
- Edit `.proto` sources, then run `make apis`.
- Do not modify `vendor/` manually; regenerate with `make vendor` when needed.

## Agent Change Checklist

Before finalizing:

- Run `gofmt` on touched files.
- Run focused tests on changed package(s).
- Run `go test ./...` for broad-impact refactors.
- Regenerate artifacts when schema/proto changes.
- Keep diffs scoped; avoid opportunistic unrelated cleanup.

If user instructions conflict with this file, follow the user.
