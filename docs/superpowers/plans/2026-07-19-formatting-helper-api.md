# Formatting Helper API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `*f` formatting helpers that GoLand recognizes while preserving the existing API and exact stacktrace behavior.

**Architecture:** Add five public sibling functions beside their existing counterparts in `stacktrace.go`. Stack-capturing helpers call `create` directly to preserve `runtime.Caller(2)` attribution; tests exercise the public API and verify messages, codes, nil propagation, and caller names.

**Tech Stack:** Go 1.18+, standard library `fmt`/`errors`, `testify/assert`

## Global Constraints

- Keep every existing public function and its formatting behavior backward-compatible.
- Add exactly `NewErrorf`, `Propagatef`, `NewErrorWithCodef`, `PropagateWithCodef`, and `NewMessageWithCodef`.
- Preserve direct call depth between every stack-capturing public helper and `create`.
- Preserve nil propagation and error-code inheritance/override semantics.
- Keep tests in the external `stacktrace_test` package.

---

### Task 1: Add formatting helper APIs

**Files:**
- Modify: `functions_for_test.go`
- Modify: `stacktrace_test.go`
- Modify: `stacktrace.go`

**Interfaces:**
- Consumes: existing `create(err error, code ErrorCode, format string, args ...any) error`, `NoCode`, and private `stacktrace`
- Produces:
  - `NewErrorf(format string, args ...any) error`
  - `Propagatef(err error, format string, args ...any) error`
  - `NewErrorWithCodef(code ErrorCode, format string, args ...any) error`
  - `PropagateWithCodef(err error, code ErrorCode, format string, args ...any) error`
  - `NewMessageWithCodef(code ErrorCode, format string, args ...any) error`

- [ ] **Step 1: Add public call-site fixtures**

Append these fixtures to `functions_for_test.go` so existing exact line assertions do not move:

```go
func newErrorfAtCallSite() error {
	return stacktrace.NewErrorf("new %d", 7)
}

func propagatefAtCallSite(err error) error {
	return stacktrace.Propagatef(err, "propagate %d", 7)
}

func newErrorWithCodefAtCallSite() error {
	return stacktrace.NewErrorWithCodef(EcodeInvalidVillain, "coded %d", 7)
}

func propagateWithCodefAtCallSite(err error) error {
	return stacktrace.PropagateWithCodef(err, EcodeNotFastEnough, "coded propagate %d", 7)
}
```

- [ ] **Step 2: Write failing behavior tests**

Append these tests to `stacktrace_test.go`:

```go
func TestFormattingHelpers(t *testing.T) {
	root := errors.New("root")

	tests := []struct {
		name     string
		err      error
		expected string
		code     stacktrace.ErrorCode
	}{
		{
			name:     "new error",
			err:      stacktrace.NewErrorf("new %d", 7),
			expected: "new 7",
			code:     stacktrace.NoCode,
		},
		{
			name:     "propagate",
			err:      stacktrace.Propagatef(root, "propagate %d", 7),
			expected: "propagate 7: root",
			code:     stacktrace.NoCode,
		},
		{
			name:     "new error with code",
			err:      stacktrace.NewErrorWithCodef(EcodeInvalidVillain, "coded %d", 7),
			expected: "coded 7",
			code:     EcodeInvalidVillain,
		},
		{
			name:     "propagate with code",
			err:      stacktrace.PropagateWithCodef(root, EcodeNotFastEnough, "coded propagate %d", 7),
			expected: "coded propagate 7: root",
			code:     EcodeNotFastEnough,
		},
		{
			name:     "new message with code",
			err:      stacktrace.NewMessageWithCodef(EcodeInvalidVillain, "message %d", 7),
			expected: "message 7",
			code:     EcodeInvalidVillain,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, fmt.Sprintf("%#s", test.err))
			assert.Equal(t, test.code, stacktrace.GetCode(test.err))
		})
	}
}

func TestFormattingHelpersPreserveCode(t *testing.T) {
	err := stacktrace.NewErrorWithCodef(EcodeInvalidVillain, "inner %d", 7)
	err = stacktrace.Propagatef(err, "outer %d", 8)

	assert.Equal(t, EcodeInvalidVillain, stacktrace.GetCode(err))
	assert.Equal(t, "outer 8: inner 7", fmt.Sprintf("%#s", err))
}

func TestFormattingPropagationNil(t *testing.T) {
	assert.Nil(t, stacktrace.Propagatef(nil, "propagate %d", 7))
	assert.Nil(t, stacktrace.PropagateWithCodef(nil, EcodeInvalidVillain, "propagate %d", 7))
}

func TestFormattingHelpersCaptureCaller(t *testing.T) {
	useFixturePaths(t)
	root := errors.New("root")

	tests := []struct {
		name     string
		err      error
		function string
	}{
		{name: "new error", err: newErrorfAtCallSite(), function: "newErrorfAtCallSite"},
		{name: "propagate", err: propagatefAtCallSite(root), function: "propagatefAtCallSite"},
		{name: "new error with code", err: newErrorWithCodefAtCallSite(), function: "newErrorWithCodefAtCallSite"},
		{name: "propagate with code", err: propagateWithCodefAtCallSite(root), function: "propagateWithCodefAtCallSite"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := fmt.Sprintf("%+s", test.err)
			assert.Contains(t, trace, fixturePath("functions_for_test.go"))
			assert.Contains(t, trace, "("+test.function+")")
		})
	}
}
```

- [ ] **Step 3: Run the targeted tests and verify they fail**

Run:

```bash
go test -run '^TestFormatting' .
```

Expected: build failure reporting undefined `stacktrace.NewErrorf`, `stacktrace.Propagatef`, `stacktrace.NewErrorWithCodef`, `stacktrace.PropagateWithCodef`, and `stacktrace.NewMessageWithCodef`.

- [ ] **Step 4: Implement direct formatting siblings**

Add each function immediately after its existing counterpart in `stacktrace.go`:

```go
// NewErrorf is the formatting variant of NewError.
func NewErrorf(format string, args ...any) error {
	return create(nil, NoCode, format, args...)
}

// Propagatef is the formatting variant of Propagate.
func Propagatef(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return create(err, NoCode, format, args...)
}

// NewErrorWithCodef is the formatting variant of NewErrorWithCode.
func NewErrorWithCodef(code ErrorCode, format string, args ...any) error {
	return create(nil, code, format, args...)
}

// PropagateWithCodef is the formatting variant of PropagateWithCode.
func PropagateWithCodef(err error, code ErrorCode, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return create(err, code, format, args...)
}

// NewMessageWithCodef is the formatting variant of NewMessageWithCode.
func NewMessageWithCodef(code ErrorCode, format string, args ...any) error {
	return &stacktrace{
		message: fmt.Sprintf(format, args...),
		code:    code,
	}
}
```

- [ ] **Step 5: Format and run targeted tests**

Run:

```bash
gofmt -w stacktrace.go stacktrace_test.go functions_for_test.go
go test -run '^TestFormatting' .
```

Expected: all `TestFormatting*` tests pass.

- [ ] **Step 6: Run package regression tests**

Run:

```bash
go test -race -cover ./...
```

Expected: all packages pass under the race detector.

- [ ] **Step 7: Commit the API and tests**

```bash
git add stacktrace.go stacktrace_test.go functions_for_test.go
git commit -m "feat: add formatting helper APIs" \
  -m "Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>" \
  -m "Copilot-Session: 1601a53e-b461-4419-af08-50a1839599b5"
```

### Task 2: Document formatting helper usage

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: the five `*f` functions added in Task 1
- Produces: user guidance that distinguishes formatting calls from static-message calls

- [ ] **Step 1: Update interpolated README examples**

Change calls with interpolation arguments to their formatting variants:

```go
return stacktrace.Propagatef(err, "Failed to write %v to %s", ent, path)
```

```go
return nil, stacktrace.Propagatef(err, "Failed to process %v", arg)
```

```go
return stacktrace.NewErrorf("Expected %v to be okay", arg)
```

```go
return stacktrace.NewErrorf("timed out after %d attempts", attempts)
```

Keep static-message calls such as `Propagate(err, "")` and
`Propagate(err, "Failed to create base directory")` unchanged.

- [ ] **Step 2: Document all five signatures**

Change the primary formatted signatures and code variants to:

```markdown
#### `stacktrace.Propagatef(err error, format string, args ...any) error`

`Propagatef` wraps an error with a formatted message and line number
information. Use `Propagate` when the message has no formatting arguments.
Both functions return `nil` when `err` is nil.
```

```markdown
#### `stacktrace.NewErrorf(format string, args ...any) error`

`NewErrorf` is a drop-in replacement for `fmt.Errorf` that includes line
number information. Use `NewError` when the message has no formatting
arguments.
```

```markdown
#### `stacktrace.PropagateWithCodef(err error, code ErrorCode, format string, args ...any) error`

#### `stacktrace.NewErrorWithCodef(code ErrorCode, format string, args ...any) error`

#### `stacktrace.NewMessageWithCodef(code ErrorCode, format string, args ...any) error`
```

State that the `*f` code variants format their messages like `fmt.Sprintf`,
that their counterparts without `f` remain supported for compatibility, and
that `PropagateWithCodef` retains nil propagation.

- [ ] **Step 3: Verify README API names and fences**

Run:

```bash
rg -n 'Propagatef|NewErrorf|WithCodef' README.md
pre-commit run --all-files
```

Expected: every interpolated stacktrace example uses a `*f` function, all five
new names appear in the function documentation, and all pre-commit hooks pass.

- [ ] **Step 4: Run the final quality gate**

Run:

```bash
go test -race -cover ./...
go vet ./...
golangci-lint fmt --diff
golangci-lint run
```

Expected: every command exits successfully with no findings.

- [ ] **Step 5: Commit the documentation**

```bash
git add README.md
git commit -m "docs: explain formatting helper APIs" \
  -m "Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>" \
  -m "Copilot-Session: 1601a53e-b461-4419-af08-50a1839599b5"
```
