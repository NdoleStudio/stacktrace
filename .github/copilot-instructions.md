# Repository instructions

## Build, test, and lint

This repository is the Go 1.18+ module `github.com/NdoleStudio/stacktrace`.

```sh
go test -race -cover ./...
go test -run '^TestFormat$' .
go test -run '^TestRemoveGoPath$' ./cleanpath
go vet ./...
golangci-lint fmt --diff
golangci-lint run
pre-commit run --all-files
```

GitHub Actions runs tests on Go 1.18.x, 1.25.x, and 1.26.x across Linux,
Windows, and macOS. Keep path assertions platform-aware.

## Architecture

- The root `stacktrace` package is a small error-wrapping library. `NewError*` and `Propagate*` converge on the private `create` function, which captures the caller with `runtime.Caller(2)` and stores each wrapper as a node pointing to its cause.
- `format.go` renders that linked error chain. The private type implements both `error` and `fmt.Formatter`: `DefaultFormat` controls normal `%v`, `%s`, and `%q` output, while `%+s` always requests the full multiline trace and `%#s` requests the brief single-line chain.
- Error codes travel with stacktrace nodes. `NoCode` is `math.MaxUint16`; an ordinary `Propagate` inherits a code from a stacktrace cause, and `ExitCode` maps `NoCode` to process exit status 1.
- `RootCause` walks only the package's private `*stacktrace` chain, while `Unwrap` returns the immediate cause for `errors.Is`, `errors.As`, and `errors.Unwrap`.
- The `cleanpath` subpackage removes the longest matching Go source prefix. The root package exposes this through the mutable `CleanPath` hook; setting it to `nil` leaves captured paths unchanged.

## Repository conventions

- Keep changes backward-compatible; this remains the contribution policy stated in `README.md`.
- Stack traces are intended for logs, not user-facing CLI, web, or library output. Add wrappers at strategic boundaries, and make each `Propagate` message describe the failed action without repeating details already present in its cause.
- Preserve the direct call depth between public constructors/wrappers and `create`. Adding an intermediate frame changes `runtime.Caller(2)` attribution and therefore user-visible file, line, and function data.
- Tests are black-box tests (`package stacktrace_test` and `package cleanpath_test`) and use `testify/assert`. Add behavior coverage through the public API unless internal access is essential.
- `TestMessage` intentionally checks exact cleaned paths, shortened function names, and source line numbers produced by helpers in `functions_for_test.go`. Moving those helper calls or inserting lines above them requires updating the expected trace in `stacktrace_test.go`.
- Formatting behavior is split between the mutable `DefaultFormat`, explicit `%+s`/`%#s` overrides, and standard `fmt` flags, widths, and precisions. Keep all three paths consistent when changing formatting.
- `Propagate` and `PropagateWithCode` return `nil` for a `nil` cause, allowing callers to propagate without a preceding `if err != nil` check. Preserve this contract.
- Use the module path `github.com/NdoleStudio/stacktrace` in package links, imports, and documentation unless you are explicitly attributing the upstream `palantir/stacktrace` fork source.
