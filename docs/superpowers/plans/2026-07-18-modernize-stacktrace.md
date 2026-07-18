# Stacktrace Modernization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Modernize the NdoleStudio fork as a Go 1.18+ module with standards-correct error unwrapping, current CI/lint tooling, canonical Apache-2.0 licensing, and fork-specific documentation.

**Architecture:** Keep the existing private linked `stacktrace` error model and all exported behavior. Move package identity and tooling to `github.com/NdoleStudio/stacktrace`, expose each node's immediate cause through `Unwrap`, and make tests deterministic across module checkouts and operating systems.

**Tech Stack:** Go 1.18+, `github.com/stretchr/testify` v1.11.1, GitHub Actions, golangci-lint v2.12.2, pre-commit.com

## Global Constraints

- The module path is exactly `github.com/NdoleStudio/stacktrace`.
- The `go.mod` minimum is Go 1.18.
- CI covers Go 1.18.x, 1.25.x, and 1.26.x on Ubuntu, Windows, and macOS.
- Preserve all exported names and behavior except for adding standards-correct `Unwrap`.
- `Unwrap` returns the immediate `cause`, not `RootCause`.
- Public formatting parameters use `err error`, `format string`, and `args ...any` where applicable.
- Retain Palantir copyright headers and upstream attribution.
- Remove Palantir CLA forms and contributor instructions; do not add a `NOTICE` file because upstream has none.
- Do not stage or overwrite unrelated user changes.

---

### Task 1: Migrate to modules and portable tests

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `test_helpers_test.go`
- Modify: `stacktrace.go`
- Modify: `stacktrace_test.go`
- Modify: `format_test.go`
- Modify: `cause_test.go`
- Modify: `codes_for_test.go`
- Modify: `functions_for_test.go`
- Modify: `cleanpath/gopath_test.go`

**Interfaces:**
- Consumes: Existing package API and `CleanPath` hook.
- Produces: Module `github.com/NdoleStudio/stacktrace`; deterministic fixture paths; test dependency `github.com/stretchr/testify` v1.11.1.

- [ ] **Step 1: Initialize the module and pin the test dependency**

Run:

```sh
go mod init github.com/NdoleStudio/stacktrace
go mod edit -go=1.18
go mod edit -require=github.com/stretchr/testify@v1.11.1
```

Expected: `go.mod` contains:

```go
module github.com/NdoleStudio/stacktrace

go 1.18

require github.com/stretchr/testify v1.11.1
```

`go.sum` contains checksums for testify and its transitive test dependencies.

- [ ] **Step 2: Update every package import to the fork path**

Replace `github.com/palantir/stacktrace` with `github.com/NdoleStudio/stacktrace` in Go import declarations in `stacktrace.go`, all root test files, and `cleanpath/gopath_test.go`. Do not change copyright headers in this step.

Run:

```sh
go mod tidy
go test ./...
```

Expected: tests compile under module mode; path-sensitive tests fail because captured module paths are checkout-dependent and legacy expectations hard-code `/`.

- [ ] **Step 3: Add deterministic stacktrace test paths**

Create `test_helpers_test.go`:

```go
package stacktrace_test

import (
	"path/filepath"
	"testing"

	"github.com/NdoleStudio/stacktrace"
)

func fixturePath(name string) string {
	return filepath.Join("github.com", "NdoleStudio", "stacktrace", name)
}

func useFixturePaths(t *testing.T) {
	t.Helper()

	original := stacktrace.CleanPath
	stacktrace.CleanPath = func(path string) string {
		return fixturePath(filepath.Base(path))
	}
	t.Cleanup(func() {
		stacktrace.CleanPath = original
	})
}
```

At the start of `TestMessage` and `TestFormat`, call:

```go
useFixturePaths(t)
```

In `stacktrace_test.go`, replace each literal `github.com/palantir/stacktrace/functions_for_test.go` path in `expected` with `fixturePath("functions_for_test.go")` through `fmt.Sprintf`, for example:

```go
expected := strings.Join([]string{
	"so closed",
	fmt.Sprintf(" --- at %s:51 (doClosure.func1) ---", fixturePath("functions_for_test.go")),
	"Caused by: pointedly",
	fmt.Sprintf(" --- at %s:46 (ptrObj.doPtr) ---", fixturePath("functions_for_test.go")),
	fmt.Sprintf(" --- at %s:42 (privateObj.doPrivate) ---", fixturePath("functions_for_test.go")),
	fmt.Sprintf(" --- at %s:38 (privateObj.DoPublic) ---", fixturePath("functions_for_test.go")),
	fmt.Sprintf(" --- at %s:34 (PublicObj.doPrivate) ---", fixturePath("functions_for_test.go")),
	fmt.Sprintf(" --- at %s:30 (PublicObj.DoPublic) ---", fixturePath("functions_for_test.go")),
	"Caused by: failed to start doing",
	fmt.Sprintf(" --- at %s:26 (startDoing) ---", fixturePath("functions_for_test.go")),
}, "\n")
```

In `format_test.go`, build each full-trace expectation with:

```go
fullTrace := fmt.Sprintf(
	"decorated\n --- at %s:## (TestFormat) ---\nCaused by: plain",
	fixturePath("format_test.go"),
)
```

Use `fullTrace`, `fmt.Sprintf("%q", fullTrace)`, and the existing width padding in the corresponding table rows instead of hard-coded package paths.

Specifically, use:

```go
expectedStacktrace: fullTrace,
expectedStacktrace: fmt.Sprintf("%q", fullTrace),
expectedStacktrace: fmt.Sprintf("%105s", fullTrace),
```

for the full, quoted, and width-105 cases.

- [ ] **Step 4: Make cleanpath expectations platform-aware**

In `cleanpath/gopath_test.go`, replace each matching-path expected value:

```go
expected: "pkg/prog.go",
```

with:

```go
expected: filepath.FromSlash("pkg/prog.go"),
```

Keep nonmatching absolute path expectations unchanged.

- [ ] **Step 5: Format and run the complete legacy suite**

Run:

```sh
gofmt -w *.go cleanpath/*.go
go test -race -cover ./...
go vet ./...
go mod tidy
```

Expected: all tests pass and vet reports no issues. Run `go mod tidy` a second time; `git status --short go.mod go.sum` must not change between the two runs.

- [ ] **Step 6: Commit the module migration**

```sh
git add go.mod go.sum stacktrace.go stacktrace_test.go format_test.go cause_test.go codes_for_test.go functions_for_test.go test_helpers_test.go cleanpath/gopath_test.go
git commit -m "build: migrate to Go modules"
```

---

### Task 2: Add standard unwrapping and modern signatures

**Files:**
- Modify: `stacktrace.go`
- Modify: `stacktrace_test.go`

**Interfaces:**
- Consumes: Existing private `stacktrace.cause` chain.
- Produces: `func (st *stacktrace) Unwrap() error`; source-compatible public signatures using `any` and conventional parameter names.

- [ ] **Step 1: Write a failing immediate-cause and chain test**

Add to `stacktrace_test.go`:

```go
type wrappedTestError struct {
	err error
}

func (e *wrappedTestError) Error() string {
	return "wrapped test error"
}

func (e *wrappedTestError) Unwrap() error {
	return e.err
}

func TestUnwrap(t *testing.T) {
	root := errors.New("root")
	typed := &wrappedTestError{err: root}
	inner := stacktrace.Propagate(typed, "inner")
	outer := stacktrace.Propagate(inner, "outer")

	assert.Same(t, inner, errors.Unwrap(outer))
	assert.True(t, errors.Is(outer, root))

	var target *wrappedTestError
	assert.True(t, errors.As(outer, &target))
	assert.Same(t, typed, target)
}
```

- [ ] **Step 2: Run the test and verify the missing behavior**

Run:

```sh
go test -run '^TestUnwrap$' .
```

Expected: FAIL because `*stacktrace` does not implement `Unwrap`, so the immediate cause is `nil` and chain traversal stops.

- [ ] **Step 3: Implement standards-correct unwrapping**

Add to `stacktrace.go`:

```go
// Unwrap returns the error that the stacktrace wraps.
func (st *stacktrace) Unwrap() error {
	return st.cause
}
```

Do not call `RootCause`; Go's error-chain contract requires one link at a time.

- [ ] **Step 4: Rename formatting parameters and use `any`**

Change the declarations and implementations in `stacktrace.go` to:

```go
func NewError(format string, args ...any) error {
	return create(nil, NoCode, format, args...)
}

func Propagate(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return create(err, NoCode, format, args...)
}

func NewErrorWithCode(code ErrorCode, format string, args ...any) error {
	return create(nil, code, format, args...)
}

func PropagateWithCode(err error, code ErrorCode, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return create(err, code, format, args...)
}

func NewMessageWithCode(code ErrorCode, format string, args ...any) error {
	return &stacktrace{
		message: fmt.Sprintf(format, args...),
		code:    code,
	}
}
```

Change `create` to this signature:

```go
func create(err error, code ErrorCode, format string, args ...any) error
```

Within its existing body, replace `cause` with `err`, `msg` with `format`, and `vals` with `args`. The inheritance and struct initialization must be:

```go
if code == NoCode {
	code = GetCode(err)
}

st := &stacktrace{
	message: fmt.Sprintf(format, args...),
	cause:   err,
	code:    code,
}
```

- [ ] **Step 5: Verify behavior and static format analysis**

Run:

```sh
gofmt -w stacktrace.go stacktrace_test.go
go test -race -cover ./...
go vet ./...
```

Expected: all tests pass; vet recognizes the formatting wrappers and reports no format/argument mismatches.

- [ ] **Step 6: Commit unwrapping and signatures**

```sh
git add stacktrace.go stacktrace_test.go
git commit -m "feat: support standard error unwrapping"
```

---

### Task 3: Normalize Apache licensing

**Files:**
- Modify: `LICENSE`
- Delete: `Palantir_Corporate_Contributor_License_Agreement.pdf`
- Delete: `Palantir_Individual_Contributor_License_Agreement.pdf`

**Interfaces:**
- Consumes: Apache License 2.0 canonical text.
- Produces: GitHub-detectable Apache-2.0 license file without downstream Palantir CLA forms.

- [ ] **Step 1: Restore the two canonical Apache placeholders**

In `LICENSE`, change only:

```text
      boilerplate notice, with the fields enclosed by brackets "{}" replaced
```

to:

```text
      boilerplate notice, with the fields enclosed by brackets "[]" replaced
```

and:

```text
   Copyright {yyyy} {name of copyright owner}
```

to:

```text
   Copyright [yyyy] [name of copyright owner]
```

Ensure the file ends with a newline.

- [ ] **Step 2: Verify the complete canonical license text**

Run in a Unix shell:

```sh
curl -fsSL https://www.apache.org/licenses/LICENSE-2.0.txt | diff -u - LICENSE
```

Expected: no diff.

- [ ] **Step 3: Remove the original project's CLA forms**

```sh
git rm Palantir_Corporate_Contributor_License_Agreement.pdf Palantir_Individual_Contributor_License_Agreement.pdf
```

Do not remove Palantir copyright headers from source files.

- [ ] **Step 4: Commit the licensing cleanup**

```sh
git add LICENSE
git commit -m "chore: normalize Apache licensing"
```

---

### Task 4: Add CI, lint, and pre-commit automation

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.golangci.yml`
- Create: `.pre-commit-config.yaml`
- Delete: `.travis.yml`

**Interfaces:**
- Consumes: Go module and portable suite from Tasks 1-2.
- Produces: Nine-combination test matrix, one lint job, and reproducible local hooks.

- [ ] **Step 1: Add the golangci-lint v2 configuration**

Create `.golangci.yml`:

```yaml
version: "2"

linters:
  default: standard

formatters:
  enable:
    - gofmt
    - goimports
```

- [ ] **Step 2: Add the cross-platform GitHub Actions workflow**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
  pull_request:

permissions:
  contents: read

concurrency:
  group: ci-${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  test:
    name: Test (${{ matrix.os }}, Go ${{ matrix.go-version }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os:
          - ubuntu-latest
          - windows-latest
          - macos-latest
        go-version:
          - 1.18.x
          - 1.25.x
          - 1.26.x
    steps:
      - name: Check out repository
        uses: actions/checkout@v7
      - name: Set up Go
        uses: actions/setup-go@v7
        with:
          go-version: ${{ matrix.go-version }}
          cache: true
      - name: Run tests
        run: go test -race -cover ./...

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@v7
      - name: Set up Go
        uses: actions/setup-go@v7
        with:
          go-version: 1.26.x
          cache: true
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v9
        with:
          version: v2.12.2
```

- [ ] **Step 3: Add pre-commit hooks**

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v2.12.2
    hooks:
      - id: golangci-lint-fmt
      - id: golangci-lint-full
  - repo: local
    hooks:
      - id: go-vet
        name: go vet
        entry: go vet ./...
        language: system
        pass_filenames: false
        types: [go]
      - id: go-test
        name: go test
        entry: go test ./...
        language: system
        pass_filenames: false
        types: [go]
```

- [ ] **Step 4: Remove Travis CI**

```sh
git rm .travis.yml
```

- [ ] **Step 5: Validate automation locally**

Install the pinned linter:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
golangci-lint config verify
golangci-lint fmt --diff
golangci-lint run
```

Expected: config verification succeeds and both lint commands report no changes or findings. If formatting changes are reported, run `golangci-lint fmt`, inspect the diff, and repeat all four checks.

Install pre-commit if it is not already available, then run:

```sh
pre-commit run --all-files
```

Expected: `golangci-lint-fmt`, `golangci-lint-full`, `go vet`, and `go test` all pass.

- [ ] **Step 6: Commit automation**

```sh
git add .github/workflows/ci.yml .golangci.yml .pre-commit-config.yaml .travis.yml
git commit -m "ci: add cross-platform Go checks"
```

---

### Task 5: Rewrite fork and contributor documentation

**Files:**
- Modify: `README.md`
- Modify: `doc.go`
- Modify: `.github/copilot-instructions.md`

**Interfaces:**
- Consumes: Final module path, API signatures, commands, and workflow.
- Produces: Fork-specific user, contributor, Godoc, and Copilot guidance.

- [ ] **Step 1: Replace the README header and add fork positioning**

Replace the title and old CircleCI/Travis badges with:

````markdown
# Stacktrace

[![CI](https://github.com/NdoleStudio/stacktrace/actions/workflows/ci.yml/badge.svg)](https://github.com/NdoleStudio/stacktrace/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/NdoleStudio/stacktrace.svg)](https://pkg.go.dev/github.com/NdoleStudio/stacktrace)
[![License](https://img.shields.io/github/license/NdoleStudio/stacktrace)](LICENSE)

Stacktrace adds compact, contextual call-site information to Go errors.

This repository is an actively maintained fork of
[palantir/stacktrace](https://github.com/palantir/stacktrace), which has been
inactive for many years. The fork preserves the original Apache-2.0-licensed
API while adding current Go tooling and error-chain support.

## Fork features

- [x] Context-rich traces captured at intentional wrapping boundaries
- [x] Full multiline and brief single-line formatting
- [x] Error codes and process exit codes
- [x] Go 1.13-compatible `errors.Is`, `errors.As`, and `errors.Unwrap`
- [x] Go modules with Go 1.18+ support
- [x] Cross-platform CI and local pre-commit checks

## Installation

```sh
go get github.com/NdoleStudio/stacktrace@latest
```
````

Remove the legacy “Look at Palantir…” joke. Keep the existing material beginning with “Why would anyone want stack traces in Go code?”, plus the intent, examples, formatting descriptions, and error-code guidance.

- [ ] **Step 2: Update README package references and signatures**

Replace every package URL/import reference:

```text
github.com/palantir/stacktrace
```

with:

```text
github.com/NdoleStudio/stacktrace
```

Do not alter the explicit upstream attribution link added in Step 1.

Update displayed declarations to:

```go
stacktrace.Propagate(err error, format string, args ...any) error
stacktrace.NewError(format string, args ...any) error
stacktrace.PropagateWithCode(err error, code ErrorCode, format string, args ...any) error
stacktrace.NewErrorWithCode(code ErrorCode, format string, args ...any) error
stacktrace.NewMessageWithCode(code ErrorCode, format string, args ...any) error
```

Add an error-unwrapping section after error codes:

```markdown
### Standard error unwrapping

Stacktrace errors implement Go's standard `Unwrap() error` contract. Use
`errors.Is`, `errors.As`, and `errors.Unwrap` to inspect wrapped causes without
discarding stacktrace context. `RootCause` remains available when the original
underlying error is needed directly.
```

- [ ] **Step 3: Replace license and contribution sections**

Use:

````markdown
## License

Stacktrace is available under the [Apache License 2.0](LICENSE). The fork
retains the original Palantir copyright notices and attribution.

## Contributing

Contributions are welcome.

```sh
go test -race -cover ./...
go vet ./...
golangci-lint run
pre-commit run --all-files
```

Install repository hooks once with `pre-commit install`.
````

Remove all instructions and links for Palantir individual/corporate CLAs.

- [ ] **Step 4: Update package documentation**

In `doc.go`, retain the package overview and add this paragraph before the package declaration:

```go
Stacktrace is maintained at https://github.com/NdoleStudio/stacktrace.
```

Do not rewrite the illustrative Shield stack trace; it documents output shape rather than an import path.

- [ ] **Step 5: Replace legacy Copilot guidance**

Update `.github/copilot-instructions.md` so it states:

````markdown
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
````

Retain its accurate architecture and compatibility notes, but remove all GOPATH, Go 1.6, Travis, and legacy `golint` instructions. Add that `Unwrap` returns the immediate cause and that module/package links use the NdoleStudio path.

- [ ] **Step 6: Verify links, wording, and documentation**

Run:

```sh
go test -race -cover ./...
go vet ./...
rg -n "travis|circleci|Palantir_.*Contributor_License|opensource@palantir.com" . --glob "!docs/superpowers/**"
rg -n "github.com/palantir/stacktrace" . --glob "!README.md" --glob "!docs/superpowers/**"
```

Expected: tests and vet pass; both searches return no matches. The README may contain exactly one lowercase upstream attribution URL. Copyright headers containing `Palantir Technologies` remain.

- [ ] **Step 7: Commit documentation**

```sh
git add README.md doc.go .github/copilot-instructions.md
git commit -m "docs: document maintained NdoleStudio fork"
```

---

### Task 6: Run final repository verification

**Files:**
- Create: `docs/superpowers/plans/2026-07-18-modernize-stacktrace.md`
- Modify other files only when a verification command proves they require correction.

**Interfaces:**
- Consumes: All prior tasks.
- Produces: A clean, release-ready modernization diff.

- [ ] **Step 0: Commit the approved implementation plan**

Run:

```sh
git add docs/superpowers/plans/2026-07-18-modernize-stacktrace.md
git commit -m "docs: add modernization implementation plan"
```

Expected: only the approved plan is included in this commit.

- [ ] **Step 1: Normalize generated and formatted files**

Run:

```sh
gofmt -w *.go cleanpath/*.go
go mod tidy
golangci-lint fmt
git diff --exit-code -- go.mod go.sum
```

Expected: no unexpected semantic changes and no module-file drift.

- [ ] **Step 2: Run the complete local quality gate**

Run:

```sh
go test -race -cover ./...
go vet ./...
golangci-lint config verify
golangci-lint fmt --diff
golangci-lint run
pre-commit run --all-files
```

Expected: every command exits successfully with zero test failures, vet diagnostics, lint findings, or formatting differences.

- [ ] **Step 3: Verify module, licensing, and stale-reference invariants**

Run:

```sh
test "$(go list -m)" = "github.com/NdoleStudio/stacktrace"
python -c "import json, subprocess; repo=subprocess.check_output(['git','show','HEAD:LICENSE']).decode(); body=json.loads(subprocess.check_output(['gh','api','licenses/apache-2.0']))['body']; body += '' if body.endswith('\n') else '\n'; assert repo == body"
test ! -e .travis.yml
test ! -e Palantir_Corporate_Contributor_License_Agreement.pdf
test ! -e Palantir_Individual_Contributor_License_Agreement.pdf
rg -n "travis|circleci|opensource@palantir.com|github.com/palantir/stacktrace" . --glob "!README.md" --glob "!docs/superpowers/**"
```

Expected: module and file checks pass; the license has no diff; the search returns no matches. Manually confirm the README's sole `github.com/palantir/stacktrace` occurrence is the upstream attribution link.

- [ ] **Step 4: Inspect the final diff**

Run:

```sh
git diff --check
git status --short
git diff 2525a1c..HEAD --check
git diff --stat 2525a1c..HEAD
```

Expected: no whitespace errors; only modernization files are changed or committed. Do not include unrelated worktree changes.

- [ ] **Step 5: Verify GitHub license detection after pushing**

After the implementation branch is pushed, run:

```sh
gh api repos/NdoleStudio/stacktrace/license --jq '.license.spdx_id'
```

Expected: `Apache-2.0`.

---

### Task 7: Switch the default branch to main

**Files:**
- No repository file changes.

**Interfaces:**
- Consumes: Reviewed modernization commits on local `master`.
- Produces: Local and remote `main` branch with GitHub configured to use `main` by default.

- [ ] **Step 1: Confirm the implementation is ready to publish**

Run:

```sh
git status --short
git branch --show-current
```

Expected: no uncommitted implementation changes and current branch `master`. Preserve unrelated user files if any remain untracked.

- [ ] **Step 2: Rename and push the branch**

Run:

```sh
git branch -m master main
git push -u origin main
```

Expected: local `main` tracks `origin/main`.

- [ ] **Step 3: Change the GitHub default branch**

Run:

```sh
gh repo edit NdoleStudio/stacktrace --default-branch main
gh repo view NdoleStudio/stacktrace --json defaultBranchRef --jq '.defaultBranchRef.name'
```

Expected: `main`.

Do not delete `origin/master`; branch deletion requires a separate explicit request.
