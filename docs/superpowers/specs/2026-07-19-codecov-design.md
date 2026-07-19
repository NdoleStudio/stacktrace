# Codecov Integration Design

## Goal

Publish repository coverage to Codecov from GitHub Actions and display the
current coverage status in `README.md`.

## CI Design

Add one dedicated `coverage` job to `.github/workflows/ci.yml`. The job runs on
Ubuntu with Go 1.26.x, independently of the existing nine-combination
compatibility matrix. This avoids duplicate uploads while keeping the existing
test job unchanged.

Grant only this job:

```yaml
permissions:
  contents: read
  id-token: write
```

Generate an atomic race-enabled Go coverage profile:

```bash
go test -race -coverprofile=coverage.out -covermode=atomic ./...
```

Upload only `coverage.out` with `codecov/codecov-action@v5`. Use GitHub OIDC
authentication instead of a repository secret, disable automatic report
search, and fail the job if the upload fails:

```yaml
with:
  use_oidc: true
  files: ./coverage.out
  disable_search: true
  fail_ci_if_error: true
```

The repository must be enabled in Codecov for OIDC uploads and badge data.

## README Design

Add the Codecov badge immediately after the CI badge:

```markdown
[![codecov](https://codecov.io/gh/NdoleStudio/stacktrace/graph/badge.svg)](https://codecov.io/gh/NdoleStudio/stacktrace)
```

## File Impact

Modify:

- `.github/workflows/ci.yml`
- `README.md`

## Acceptance Checks

1. The existing compatibility test matrix and lint job remain unchanged.
2. The coverage job has the minimum permissions needed for checkout and OIDC.
3. The coverage profile is explicit, atomic, race-enabled, and uploaded once.
4. Codecov upload failures fail CI.
5. The README badge targets `NdoleStudio/stacktrace`.
6. Existing Go tests, vet, lint, formatting, and pre-commit checks pass.
