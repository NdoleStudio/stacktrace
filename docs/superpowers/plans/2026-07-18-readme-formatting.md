# README Formatting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add language-aware syntax highlighting to every README example and display the fork notice in a GitHub note alert.

**Architecture:** This is a Markdown-only transformation in `README.md`. Preserve all technical content while replacing HTML example containers and untyped fences with GitHub-renderable, language-specific Markdown.

**Tech Stack:** GitHub Flavored Markdown

## Global Constraints

- Use `bash` for shell commands.
- Use `text` for plain error and stacktrace output.
- Use `go` for all Go source examples.
- Replace every `<pre>`, `</pre>`, `<b>`, and `</b>` tag.
- Preserve example code and prose apart from container markup.
- Wrap the fork notice in a GitHub `[!NOTE]` alert.

---

### Task 1: Format README examples and fork notice

**Files:**
- Modify: `README.md`
- Create: `docs/superpowers/plans/2026-07-18-readme-formatting.md`

**Interfaces:**
- Consumes: Existing README prose, commands, output, and Go examples.
- Produces: GitHub-renderable language fences and note alert.

- [ ] **Step 1: Prove the README does not yet satisfy the formatting rules**

Run:

```powershell
$content = Get-Content README.md -Raw
if ($content -match '</?(pre|b)>' -or $content -notmatch '> \[!NOTE\]') {
    throw 'README formatting is not yet compliant'
}

$insideFence = $false
foreach ($line in Get-Content README.md) {
    if ($line -match '^```(.*)$') {
        if (-not $insideFence) {
            if ($Matches[1] -notin @('go', 'bash', 'text')) {
                throw "Unsupported opening fence: $($Matches[1])"
            }
            $insideFence = $true
        } else {
            $insideFence = $false
        }
    }
}
```

Expected: FAIL because the README has untyped fences, HTML `<pre>/<b>` markup, and no note alert.

- [ ] **Step 2: Wrap the fork notice in a note alert**

Replace:

```markdown
This repository is an actively maintained fork of
[palantir/stacktrace](https://github.com/palantir/stacktrace), which has been
inactive for many years. The fork preserves the original Apache-2.0-licensed
API while adding current Go tooling and error-chain support.
```

with:

```markdown
> [!NOTE]
> This repository is an actively maintained fork of
> [palantir/stacktrace](https://github.com/palantir/stacktrace), which has been
> inactive for many years. The fork preserves the original Apache-2.0-licensed
> API while adding current Go tooling and error-chain support.
```

- [ ] **Step 3: Add languages to existing fenced blocks**

Apply these exact mappings:

```text
Installation command:             sh -> bash
Short error output:                untyped -> text
Full stacktrace output:            untyped -> text
ErrorCode constant declaration:    untyped -> go
Contributor commands:              sh -> bash
```

- [ ] **Step 4: Convert every HTML Go example**

For each `<pre>...</pre>` example, replace the tags with ` ```go ` and ` ``` ` fences, and remove `<b>`/`</b>` while retaining their enclosed Go expression unchanged.

This applies to:

1. `WriteAll`
2. The canonical `Propagate` call
3. `Something`
4. The canonical `NewError` call
5. `PropagateWithCode`
6. `NewMessageWithCode`
7. `GetCode`

For example:

````markdown
```go
result, err := process(arg)
if err != nil {
    return nil, stacktrace.Propagate(err, "Failed to process %v", arg)
}
```
````

- [ ] **Step 5: Verify fence languages, HTML removal, alert text, and balance**

Run:

```powershell
$content = Get-Content README.md -Raw
if ($content -match '</?(pre|b)>') { throw 'Legacy HTML code markup remains' }
if ($content -notmatch '(?m)^> \[!NOTE\]$') { throw 'NOTE alert is missing' }

$open = $null
foreach ($line in Get-Content README.md) {
    if ($line -match '^```(.*)$') {
        if ($null -eq $open) {
            if ($Matches[1] -notin @('go', 'bash', 'text')) {
                throw "Unsupported fence language: $($Matches[1])"
            }
            $open = $Matches[1]
        } else {
            if ($Matches[1] -ne '') { throw 'Closing fence has a language' }
            $open = $null
        }
    }
}
if ($null -ne $open) { throw 'Unclosed Markdown fence' }
```

Expected: PASS with no output.

- [ ] **Step 6: Run repository checks**

Run:

```powershell
go test ./...
pre-commit run --all-files
git diff --check
```

Expected: tests and all pre-commit hooks pass; Git reports no whitespace errors.

- [ ] **Step 7: Commit and push the README formatting**

```powershell
git add README.md docs/superpowers/plans/2026-07-18-readme-formatting.md
git commit -m "docs: add README syntax highlighting"
git push origin main
```

Expected: `main` is pushed with the README and approved plan changes.
