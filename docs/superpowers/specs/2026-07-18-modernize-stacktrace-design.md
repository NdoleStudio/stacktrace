# Stacktrace Modernization Design

## Goal

Modernize the NdoleStudio fork as a Go module with current automation and conventions while preserving the existing exported API and formatting behavior. The intentional compatibility changes are the import path move to `github.com/NdoleStudio/stacktrace` and the addition of standards-correct Go error unwrapping.

## Compatibility

- Declare `go 1.18` as the minimum supported version so public variadic parameters can use `any`.
- Test the minimum version and both Go versions supported in July 2026: Go 1.18.x, 1.25.x, and 1.26.x.
- Preserve every exported function, type, constant, global, formatting verb, nil-propagation behavior, error-code rule, and `RootCause` behavior.
- Rename parameters from legacy names such as `cause`, `msg`, and `vals` to conventional `err`, `format`, and `args`. Parameter names are not part of a Go function's caller-visible signature.
- Treat the new module path as the fork's migration boundary; do not add aliases or forwarding packages for the Palantir path.

## Licensing

The existing `LICENSE` contains the complete Apache License 2.0 text, but its appendix replaces the canonical square-bracket placeholders with braces on two lines. GitHub currently classifies the file as `Other`. Replace those two lines with the exact text from <https://www.apache.org/licenses/LICENSE-2.0.txt>; do not alter the license terms or remove existing source-file copyright headers.

Delete the Palantir individual and corporate CLA PDFs and remove the README instructions that require contributors to submit them to Palantir. Apache License 2.0 redistribution requires the license and applicable copyright, patent, trademark, and attribution notices, but it does not require downstream forks to distribute the original project's contribution-agreement forms. The upstream repository has no `NOTICE` file to carry forward.

References:

- Apache License 2.0 redistribution terms: <https://www.apache.org/licenses/LICENSE-2.0>
- Apache CLA FAQ: <https://www.apache.org/licenses/cla-faq>
- GitHub repository license guidance: <https://docs.github.com/en/communities/setting-up-your-project-for-healthy-contributions/adding-a-license-to-a-repository>

## Module and API

Add a `go.mod` declaring:

```text
module github.com/NdoleStudio/stacktrace

go 1.18
```

Retain `github.com/stretchr/testify` as the test-only dependency at its current stable release and generate `go.sum` with `go mod tidy`. Update all source, test, README, and Godoc imports and links from `github.com/palantir/stacktrace` to `github.com/NdoleStudio/stacktrace`.

Use `any` in place of `interface{}` and conventional formatting names throughout the public and private call chain:

```go
func NewError(format string, args ...any) error
func Propagate(err error, format string, args ...any) error
func NewErrorWithCode(code ErrorCode, format string, args ...any) error
func PropagateWithCode(err error, code ErrorCode, format string, args ...any) error
func NewMessageWithCode(code ErrorCode, format string, args ...any) error
func create(err error, code ErrorCode, format string, args ...any) error
```

Add this method to the private error type:

```go
func (st *stacktrace) Unwrap() error {
	return st.cause
}
```

Returning the immediate cause follows the Go 1.13 error-chain contract. It improves on palantir/stacktrace PR #13, whose implementation jumps directly to `RootCause` and therefore hides intermediate wrapped errors from repeated `errors.Unwrap` and some `errors.As` searches. Existing `RootCause` behavior remains unchanged.

## Tests

Add focused coverage proving:

- `errors.Unwrap` returns the immediate cause.
- `errors.Is` finds a root sentinel through multiple stacktrace wrappers.
- `errors.As` finds an intermediate typed error in a mixed chain.
- Existing nil propagation, formatting, error codes, exit codes, and root-cause behavior remain unchanged.

Make path-sensitive tests portable without changing production path output. Build expected paths with `filepath.Join` or compare normalized paths where appropriate instead of hard-coding `/`. Keep exact function and line assertions where they protect caller attribution.

## GitHub Actions and Linting

Remove `.travis.yml` and replace it with `.github/workflows/ci.yml`, triggered by pushes and pull requests. Give the workflow only `contents: read` permission and cancel superseded runs on the same ref.

The test job uses `actions/checkout@v7` and `actions/setup-go@v7`, with the full matrix:

- Operating systems: Ubuntu, Windows, macOS.
- Go versions: 1.18.x, 1.25.x, 1.26.x.
- Command: `go test -race -cover ./...`.

The lint job runs once on Ubuntu with Go 1.26.x. It uses `golangci/golangci-lint-action@v9` pinned to golangci-lint v2.12.2.

Add `.golangci.yml` using configuration version 2. Use the standard linter set, which includes `govet`, `staticcheck`, `errcheck`, `ineffassign`, and `unused`, and enable formatting checks. Add only narrowly justified exclusions if the unchanged legacy behavior produces a verified false positive.

## Pre-commit

Add `.pre-commit-config.yaml` pinned to golangci-lint v2.12.2. Use the official:

- `golangci-lint-fmt` hook to format Go files.
- `golangci-lint-full` hook to analyze the whole module.

Add local, filename-independent hooks for `go vet ./...` and `go test ./...`. Document `pre-commit install` and `pre-commit run --all-files`; do not commit generated `.git/hooks` content.

## README and Package Documentation

Rewrite the README header to remove CircleCI and Travis references and add badges for the GitHub Actions workflow, pkg.go.dev, and the Apache-2.0 license.

Identify the repository as an actively maintained NdoleStudio fork of `palantir/stacktrace`, retained under Apache-2.0 because the upstream project has been inactive for years. Preserve appropriate upstream attribution without implying Palantir maintains or endorses the fork.

Add a checked Markdown feature list:

- [x] Context-rich traces captured at intentional wrapping boundaries
- [x] Full multiline and brief single-line formatting
- [x] Error codes and process exit codes
- [x] Go 1.13-compatible `errors.Is`, `errors.As`, and `errors.Unwrap`
- [x] Go modules with Go 1.18+ support
- [x] Cross-platform CI and local pre-commit checks

Update installation to `go get github.com/NdoleStudio/stacktrace@latest`, modernize all displayed signatures to `format string, args ...any`, and update contribution instructions to the module, lint, test, and pre-commit commands. Preserve the package's central guidance: stack traces are for diagnostic logs rather than user-facing output, and propagation messages should add nonredundant context about the failed action.

Update `.github/copilot-instructions.md` so future sessions use module-mode commands, the new import path, portable tests, GitHub Actions, golangci-lint, and pre-commit instead of the legacy GOPATH/Travis workflow.

## File Impact

Create:

- `go.mod`
- `go.sum`
- `.github/workflows/ci.yml`
- `.golangci.yml`
- `.pre-commit-config.yaml`

Modify:

- `LICENSE`
- `README.md`
- `doc.go`
- `stacktrace.go`
- Tests and fixtures that import the old module path or assume Unix separators
- `.github/copilot-instructions.md`

Delete:

- `.travis.yml`
- `Palantir_Corporate_Contributor_License_Agreement.pdf`
- `Palantir_Individual_Contributor_License_Agreement.pdf`

## Acceptance Checks

The implementation is complete when:

1. `go mod tidy` leaves no diff.
2. `go test -race -cover ./...` passes on the nine-version/OS matrix.
3. `go vet ./...` passes.
4. `golangci-lint run` and `golangci-lint fmt --diff` pass with v2.12.2.
5. `pre-commit run --all-files` passes.
6. `go list -m` reports `github.com/NdoleStudio/stacktrace`.
7. Repository searches find no Palantir import/package links, CLA instructions, Travis references, or CircleCI references except the explicit upstream attribution link.
8. GitHub identifies `LICENSE` as Apache-2.0 after the change is pushed.
