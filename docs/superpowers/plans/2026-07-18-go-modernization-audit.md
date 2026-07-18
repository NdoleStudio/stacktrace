# Go Modernization Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply the justified results of a complete repository audit while preserving Go 1.18 compatibility and all public behavior.

**Architecture:** Keep the linked stacktrace implementation and public API unchanged. Apply the Go 1.26 official modernizer's behavior-preserving formatter rewrite, modernize test environment handling, complete the pending README formatting work, and enforce future official modernizer checks in the existing lint job.

**Tech Stack:** Go 1.18+, Go 1.26 `go fix`, `testify/assert`, GitHub Actions, golangci-lint v2.12.2, pre-commit

## Global Constraints

- Keep `go 1.18` as the module minimum.
- Preserve every exported name, signature, formatting rule, stack frame, error-code rule, nil-propagation behavior, and error-chain behavior.
- Keep path-sensitive tests portable across Linux, Windows, and macOS.
- Add no dependencies.
- Do not overwrite unrelated user changes.

---

### Task 1: Complete README formatting

**Files:**
- Modify: `README.md`
- Reference: `docs/superpowers/specs/2026-07-18-readme-formatting-design.md`

**Interfaces:**
- Consumes: Existing README prose, examples, and contributor commands.
- Produces: GitHub-renderable language-specific examples without changing their content.

- [ ] **Step 1: Record the current formatting gaps**

Run:

```bash
rg -n '```$|```sh|<pre>|</pre>|<b>|</b>|actively maintained fork' README.md
```

Expected: unlabeled or `sh` fences, HTML example tags, and the unquoted fork notice are found.

- [ ] **Step 2: Apply the approved Markdown containers**

Change the two-sentence fork notice to:

```markdown
> [!NOTE]
> This repository is an actively maintained fork of
> [palantir/stacktrace](https://github.com/palantir/stacktrace), which has been
> inactive for many years. The fork preserves the original Apache-2.0-licensed
> API while adding current Go tooling and error-chain support.
```

Use `bash` for command blocks, `text` for output, and `go` for Go snippets.
Replace each `<pre>`/`<b>` example with the same source in a fenced `go` block.

- [ ] **Step 3: Verify README structure**

Run:

```bash
rg -n '```$|```sh|<pre>|</pre>|<b>|</b>' README.md
```

Expected: no output. Every opening fence is `bash`, `text`, or `go`.

- [ ] **Step 4: Commit the documentation change**

```bash
git add README.md
git commit -m "docs: modernize README formatting"
```

---

### Task 2: Apply the official formatter modernization

**Files:**
- Modify: `format.go:60-78`
- Test: `format_test.go`

**Interfaces:**
- Consumes: `fmt.State`, the format verb rune, and the existing rendered stacktrace text.
- Produces: The exact same dynamic format directive passed to `fmt.Fprintf`.

- [ ] **Step 1: Establish formatter behavior**

Run:

```bash
go test -run '^TestFormat$' .
```

Expected: PASS.

- [ ] **Step 2: Preview the official modernization**

Run:

```bash
go fix -diff ./...
```

Expected: a patch replacing `formatString` concatenation in `format.go` with
`strings.Builder`; exit status is nonzero because a diff exists.

- [ ] **Step 3: Apply and inspect the official modernization**

Run:

```bash
go fix ./...
git diff -- format.go
```

Expected: only the dynamic format directive construction changes. It starts
with `var formatString strings.Builder`, writes the same flags, width,
precision, and verb in the same order, and passes `formatString.String()` to
`fmt.Fprintf`.

- [ ] **Step 4: Verify formatter behavior and modernizer cleanliness**

Run:

```bash
go test -run '^TestFormat$' .
go fix -diff ./...
```

Expected: the test passes and `go fix` prints no patch.

- [ ] **Step 5: Commit the formatter modernization**

```bash
git add format.go
git commit -m "perf: modernize format construction"
```

---

### Task 3: Modernize clean-path test environment handling

**Files:**
- Modify: `cleanpath/gopath_test.go:3-52`
- Test: `cleanpath/gopath_test.go`

**Interfaces:**
- Consumes: Each table row's synthetic `GOPATH`.
- Produces: The same `RemoveGoPath` assertions with automatic environment restoration.

- [ ] **Step 1: Establish clean-path behavior**

Run:

```bash
go test -run '^TestRemoveGoPath$' ./cleanpath
```

Expected: PASS.

- [ ] **Step 2: Replace manual environment mutation**

Remove the `os` import and replace:

```go
err := os.Setenv("GOPATH", gopath)
assert.NoError(t, err, "error setting gopath")
```

with:

```go
t.Setenv("GOPATH", gopath)
```

- [ ] **Step 3: Format and verify the focused test**

Run:

```bash
gofmt -w cleanpath/gopath_test.go
go test -run '^TestRemoveGoPath$' ./cleanpath
```

Expected: PASS.

- [ ] **Step 4: Commit the test modernization**

```bash
git add cleanpath/gopath_test.go
git commit -m "test: restore GOPATH automatically"
```

---

### Task 4: Enforce official modernizers in CI

**Files:**
- Modify: `.github/workflows/ci.yml:39-49`

**Interfaces:**
- Consumes: The existing Go 1.26 lint environment.
- Produces: A CI failure whenever `go fix -diff ./...` finds a compatible modernization.

- [ ] **Step 1: Add the modernizer check**

After the lint job's `Set up Go` step, add:

```yaml
      - name: Check Go modernizations
        run: go fix -diff ./...
```

Keep the existing golangci-lint action unchanged.

- [ ] **Step 2: Verify the command and workflow diff**

Run:

```bash
go fix -diff ./...
git diff --check
git diff -- .github/workflows/ci.yml
```

Expected: `go fix` prints no patch, `git diff --check` is clean, and the
workflow diff contains only the new step.

- [ ] **Step 3: Commit the CI guard**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: enforce Go modernizers"
```

---

### Task 5: Validate and review the complete audit

**Files:**
- Review: All tracked files
- Modify only if a validation failure is directly caused by Tasks 1-4

**Interfaces:**
- Consumes: The complete branch diff and repository quality gates.
- Produces: A clean, review-ready branch.

- [ ] **Step 1: Verify module and formatting stability**

Run:

```bash
go mod tidy
git diff --exit-code -- go.mod go.sum
golangci-lint fmt --diff
go fix -diff ./...
```

Expected: all commands succeed without changes.

- [ ] **Step 2: Run Go validation**

Run:

```bash
go test -cover ./...
go vet ./...
golangci-lint run
```

Expected: all commands pass.

- [ ] **Step 3: Run the repository hooks**

Run:

```bash
pre-commit run --all-files
```

Expected: every hook passes.

- [ ] **Step 4: Attempt race validation**

Run:

```bash
go test -race -cover ./...
```

Expected on supported hosts: PASS. On Windows/ARM64, record the Go tool's
documented `-race is not supported on windows/arm64` limitation; the existing
GitHub Actions matrix runs this command on supported hosted runners.

- [ ] **Step 5: Review the final diff**

Run:

```bash
git diff --check origin/main...HEAD
git status --short
git diff --stat origin/main...HEAD
```

Expected: no whitespace errors, a clean worktree, and only the planned files.

- [ ] **Step 6: Push and open the pull request**

```bash
git push -u origin copilot/go-modernization-audit
gh pr create --base main --head copilot/go-modernization-audit
```

Expected: GitHub returns the new pull request URL.
