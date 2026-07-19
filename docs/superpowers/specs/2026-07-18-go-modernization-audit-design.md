# Go Modernization Audit Design

## Goal

Perform a fresh line-by-line audit of every tracked repository file and apply
only evidence-backed, backward-compatible modernization changes while
preserving Go 1.18 support.

## Compatibility

- Keep `go 1.18` as the module minimum.
- Preserve every exported name, signature, formatting rule, stack frame,
  error-code rule, nil-propagation behavior, and error-chain behavior.
- Keep path-sensitive tests portable across Linux, Windows, and macOS.
- Do not introduce dependencies or adopt language or library features newer
  than Go 1.18 in production or test code.

## Approach

Use Go 1.26.4's official `go fix -diff ./...` modernizers as the primary source
of code changes, then verify each suggestion against the compatibility
constraints and existing tests. Review all other source, test, workflow,
configuration, documentation, dependency, and licensing files manually.
Files without a justified change remain untouched.

The audit identified four changes:

1. Apply the official modernizer's `strings.Builder` rewrite to the dynamic
   format directive assembled by `(*stacktrace).Format`.
2. Replace manual `os.Setenv` plus assertion handling in the clean-path test
   with Go 1.17's `t.Setenv`, which restores `GOPATH` automatically.
3. Implement the already-committed README formatting design so every example
   has a language-specific fence and the fork notice uses a GitHub note alert.
4. Run `go fix -diff ./...` in the Go 1.26 CI lint job so future compatible
   modernizations cannot silently accumulate.

## Testing

- Use the existing formatter table as the regression suite for the
  behavior-preserving builder rewrite.
- Run `TestRemoveGoPath` after changing environment setup.
- Run `go fix -diff ./...` and require an empty diff.
- Run `go test -cover ./...`, `go vet ./...`, golangci-lint formatting and
  analysis, and pre-commit.
- Attempt the repository's race-enabled suite. On hosts where the Go race
  detector is unsupported, rely on the existing GitHub Actions matrix for
  race execution and run the complete suite locally without `-race`.

## File Impact

Modify:

- `README.md`
- `format.go`
- `cleanpath/gopath_test.go`
- `.github/workflows/ci.yml`

Create:

- `docs/superpowers/specs/2026-07-18-go-modernization-audit-design.md`
- `docs/superpowers/plans/2026-07-18-go-modernization-audit.md`

All other tracked files were reviewed and require no change.

## Acceptance Checks

1. `go fix -diff ./...` reports no patch with Go 1.26.
2. `go test -cover ./...` and `go vet ./...` pass.
3. `golangci-lint fmt --diff` and `golangci-lint run` pass.
4. `pre-commit run --all-files` passes.
5. README fences and the fork note match the existing formatting design.
6. The final diff contains no exported API or observable formatting changes.
