# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go CLI built around Cobra. [`main.go`](main.go) wires the root command, product registration, config loading, and BusyBox-style alias handling. Product implementations live under `products/<name>/`; shared config types live in `config/`; build automation lives in `Taskfile.yml`. Tests sit next to the code they cover, for example `products/chaitin/command_test.go`.

## Build, Test, and Development Commands

Use `task` for common workflows:

- `task build`: build `bin/chaitin-cli`.
- `task test`: run `go test ./...`.
- `task fmt`: run `go fmt ./...`.
- `task lint`: run `go vet ./...`.
- `task run:chaitin`: run the demo command locally.
- `task package GOOS=linux GOARCH=amd64`: build and archive a release artifact.

For a quick smoke test, run `go run . chaitin`.

## Coding Style & Naming Conventions

Follow standard Go formatting and keep files `gofmt`-clean. Package names are short, lowercase, and product-scoped. Exported identifiers use `CamelCase`; unexported identifiers use `mixedCase`. New product entrypoints should expose `NewCommand()`, and runtime config integration should stay inside the product package, not the root command.

## Adding a New Product

Add each product under `products/<name>/` and keep its command tree self-contained:

- Create the product package and expose `NewCommand()`.
- Import the package in [`main.go`](main.go).
- Register it in `newApp()` with `a.registerProductCommand(...)`.
- If `NewCommand()` can fail, return `(*cobra.Command, error)` and handle the error before registration, as `xray` does.
- If the product needs values from `config.yaml`, environment variables, `.env`, or root flags such as `--dry-run`, implement `ApplyRuntimeConfig(...)` in the product package and call it from `wrapProductCommand()`.
- Root-level `-c` / `--config` selects which config file is loaded before command execution. Product packages should consume the already-selected `config.Raw` via `ApplyRuntimeConfig(...)` instead of reparsing their own config path unless there is a strong reason.
- Parse product-specific config from `config.Raw` inside the product package; environment variable overrides follow `<PRODUCT>_<FIELD>` automatically via the shared config layer. Do not push product field parsing into the root command.

Keep the root command limited to shared wiring. Product behavior, flags, and config decoding should remain in the product package.

## Testing Guidelines

Write table-driven Go tests where practical and keep them in `*_test.go` files beside the implementation. Name tests with the standard `TestXxx` pattern, for example `TestNewCommand`. Run `task test` before opening a PR; when changing command wiring or config behavior, add focused coverage.

## Commit & Pull Request Guidelines

Recent history favors short, imperative, lowercase commit subjects such as `simplify root folder` and `unify args`. Keep commits narrowly scoped. Pull requests should explain user-visible CLI changes, list new commands or flags, and mention any `config.yaml`, environment variable, or `.env` impact. Include example invocations when behavior changes.

## Security & Configuration Tips

Do not commit real API keys or product endpoints. Start from `config.yaml.example` patterns in [`README.md`](README.md) and keep secrets only in local `config.yaml`, local `.env`, or shell environment variables.
