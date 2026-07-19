# Formatting Helper API Design

## Goal

Add GoLand-recognizable formatting APIs without changing the behavior of the
existing public API or the stack frame recorded for callers.

## Public API

Add these functions:

```go
func NewErrorf(format string, args ...any) error
func Propagatef(err error, format string, args ...any) error
func NewErrorWithCodef(code ErrorCode, format string, args ...any) error
func PropagateWithCodef(err error, code ErrorCode, format string, args ...any) error
func NewMessageWithCodef(code ErrorCode, format string, args ...any) error
```

Each function has the same semantics as its existing counterpart. Existing
functions remain unchanged and continue to accept formatting arguments for
backward compatibility.

## Implementation

The four stack-capturing functions call `create` directly. They must not
delegate through another public function because `create` uses
`runtime.Caller(2)`, and an extra wrapper frame would change the recorded file,
function, and line.

`Propagatef` and `PropagateWithCodef` return `nil` for a nil cause.
`Propagatef` inherits an existing stacktrace code, while
`PropagateWithCodef` applies the supplied code. `NewMessageWithCodef` creates a
message-only coded error without caller information.

## Documentation

Add Go doc comments for all new exported functions. Update README signatures
and interpolated examples to use the `*f` variants, while documenting that the
original APIs remain supported.

## Tests

Add black-box coverage for:

1. String interpolation in all five new functions.
2. Nil behavior for both propagation functions.
3. Code inheritance and explicit code assignment.
4. Accurate caller attribution for stack-capturing functions.

## Compatibility

This is additive. No existing symbol, signature, formatting behavior, error
chain behavior, or error-code behavior changes.
