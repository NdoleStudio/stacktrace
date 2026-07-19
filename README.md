# Stacktrace

[![CI](https://github.com/NdoleStudio/stacktrace/actions/workflows/ci.yml/badge.svg)](https://github.com/NdoleStudio/stacktrace/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/NdoleStudio/stacktrace/graph/badge.svg)](https://codecov.io/gh/NdoleStudio/stacktrace)
[![Go Reference](https://pkg.go.dev/badge/github.com/NdoleStudio/stacktrace.svg)](https://pkg.go.dev/github.com/NdoleStudio/stacktrace)
[![License](https://img.shields.io/github/license/NdoleStudio/stacktrace)](LICENSE)

Stacktrace adds compact, contextual call-site information to Go errors.

> [!IMPORTANT]
> This repository is an actively maintained fork of
> [palantir/stacktrace](https://github.com/palantir/stacktrace), which has been
> inactive for many years. The fork preserves the original Apache-2.0-licensed
> API while adding current Go tooling and error-chain support.

## Fork features

- [x] Context-rich traces captured at intentional wrapping boundaries
- [x] Full multiline and brief single-line formatting
- [x] Error codes and process exit codes
- [x] Go 1.13-compatible `errors.Is`, `errors.As`, and `errors.Unwrap`
- [x] Go modules with Go 1.18+ support
- [x] Cross-platform CI and local pre-commit checks

## Installation

```bash
go get github.com/NdoleStudio/stacktrace@latest
```

## Why would anyone want stack traces in Go code?

This is difficult to debug:

```text
Inverse tachyon pulse failed
```

This gives the full story and is easier to debug:

```text
Failed to register for villain discovery
 --- at github.com/palantir/shield/agent/discovery.go:265 (ShieldAgent.reallyRegister) ---
 --- at github.com/palantir/shield/connector/impl.go:89 (Connector.Register) ---
Caused by: Failed to load S.H.I.E.L.D. config from /opt/shield/conf/shield.yaml
 --- at github.com/palantir/shield/connector/config.go:44 (withShieldConfig) ---
Caused by: There isn't enough time (4 picoseconds required)
 --- at github.com/palantir/shield/axiom/pseudo/resource.go:46 (PseudoResource.Adjust) ---
 --- at github.com/palantir/shield/axiom/pseudo/growth.go:110 (reciprocatingPseudo.growDown) ---
 --- at github.com/palantir/shield/axiom/pseudo/growth.go:121 (reciprocatingPseudo.verify) ---
Caused by: Inverse tachyon pulse failed
 --- at github.com/palantir/shield/metaphysic/tachyon.go:72 (TryPulse) ---
```

Note that stack traces are *not designed to be user-visible*. We have found them
to be valuable in log files of server applications. Nobody wants to see these in
CLI output or a web interface or a return value from library code.

## Intent

The intent is *not* that we capture the exact state of the stack when an error
happens, including every function call. For a library that does that, see
[github.com/go-errors/errors](https://github.com/go-errors/errors). The intent
here is to attach relevant contextual information (messages, variables) at
strategic places along the call stack, keeping stack traces compact and
maximally useful.

## Example Usage

```go
func WriteAll(baseDir string, entities []Entity) error {
    err := os.MkdirAll(baseDir, 0755)
    if err != nil {
        return stacktrace.Propagate(err, "Failed to create base directory")
    }
    for _, ent := range entities {
        path := filepath.Join(baseDir, fileNameForEntity(ent))
        err = Write(path, ent)
        if err != nil {
            return stacktrace.Propagatef(err, "Failed to write %v to %s", ent, path)
        }
    }
    return nil
}
```

## Functions

#### `stacktrace.Propagatef(err error, format string, args ...any) error`

Propagatef wraps an error with a formatted message and line number
information. Use `Propagate` when the message has no formatting arguments.
Both functions return `nil` when `err` is nil.

The `format` and `args` work like `fmt.Sprintf`.

The message passed to Propagatef should describe the action that failed,
resulting in `err`. The canonical call looks like this:

```go
result, err := process(arg)
if err != nil {
    return nil, stacktrace.Propagatef(err, "Failed to process %v", arg)
}
```

To write the message, ask yourself "what does this call do?" What does
`process(arg)` do? It processes ${arg}, so the message is that we failed to
process ${arg}.

Pay attention that the message is not redundant with the one in `err`. In the
`WriteAll` example above, any error from `os.MkdirAll` will already contain the
path it failed to create, so it would be redundant to include it again in our
message. However, the error from `os.MkdirAll` will not identify that path as
corresponding to the "base directory" so we propagate with that information.

If it is not possible to add any useful contextual information beyond what is
already included in an error, `format` can be an empty string:

```go
func Something() error {
    mutex.Lock()
    defer mutex.Unlock()

    err := reallySomething()
    return stacktrace.Propagate(err, "")
}
```

The purpose of `""` as opposed to a separate function is to make you feel a
little guilty every time you do this.

This example also illustrates the behavior of Propagate when `err` is nil
&ndash; it returns nil as well. There is no need to check `if err != nil`.

#### `stacktrace.NewErrorf(format string, args ...any) error`

NewErrorf creates an error with a formatted message and line number information.
Use `NewError` when the message has no formatting arguments.

The `*f` helpers format their messages like `fmt.Sprintf`. Their counterparts
without `f` remain supported for compatibility.

> [!NOTE]
> Unlike `fmt.Errorf`, `%w` is not supported; use `Propagate` or `Propagatef` to
> wrap a cause.

The canonical call looks like this:

```go
if !IsOkay(arg) {
    return stacktrace.NewErrorf("Expected %v to be okay", arg)
}
```

### Error Codes

Occasionally it can be useful to propagate an error code while unwinding the
stack. For example, a RESTful API may use the error code to set the HTTP status
code.

The type `stacktrace.ErrorCode` is a typedef for uint16. You name the set of
error codes relevant to your application.

```go
const (
    EcodeManifestNotFound = stacktrace.ErrorCode(iota)
    EcodeBadInput
    EcodeTimeout
)
```

The special value `stacktrace.NoCode` is equal to `math.MaxUint16`, so avoid
using that. NoCode is the error code of errors with no code explicitly attached.

An ordinary `stacktrace.Propagate` preserves the error code of an error.

#### `stacktrace.PropagateWithCodef(err error, code ErrorCode, format string, args ...any) error`

#### `stacktrace.NewErrorWithCodef(code ErrorCode, format string, args ...any) error`

#### `stacktrace.NewMessageWithCodef(code ErrorCode, format string, args ...any) error`

PropagateWithCodef, NewErrorWithCodef, and NewMessageWithCodef are analogous to
Propagatef, NewErrorf, and NewMessageWithCode respectively, but also attach an
error code. Propagate and Propagatef inherit existing codes, while
PropagateWithCode and PropagateWithCodef override them with the supplied code.
Their non-`f` counterparts remain supported for compatibility. PropagateWithCodef
retains nil propagation.

```go
_, err := os.Stat(manifestPath)
if os.IsNotExist(err) {
    return stacktrace.PropagateWithCode(err, EcodeManifestNotFound, "")
}
```

The error code mechanism can be useful by itself even where stack traces with
line numbers are not required. `NewMessageWithCodef` returns an error that
formats its message like `fmt.Sprintf`, but including a code.

```go
ttl := req.URL.Query().Get("ttl")
if ttl == "" {
    return 0, stacktrace.NewMessageWithCode(EcodeBadInput, "Missing ttl query parameter")
}
```

#### `stacktrace.GetCode(err error) ErrorCode`

GetCode extracts the error code from an error.

```go
for i := 0; i < attempts; i++ {
    err := Do()
    if stacktrace.GetCode(err) != EcodeTimeout {
        return err
    }
    // try a few more times
}
return stacktrace.NewErrorf("timed out after %d attempts", attempts)
```

GetCode returns the special value `stacktrace.NoCode` if `err` is nil or if
there is no error code attached to `err`.

### Standard error unwrapping

Stacktrace errors implement Go's standard `Unwrap() error` contract. Use
`errors.Is`, `errors.As`, and `errors.Unwrap` to inspect wrapped causes without
discarding stacktrace context. `RootCause` remains available when the original
underlying error is needed directly.

## License

Stacktrace is available under the [Apache License 2.0](LICENSE). The fork
retains the original Palantir copyright notices and attribution.

## Contributing

Contributions are welcome.

```bash
go test -race -cover ./...
go vet ./...
golangci-lint run
pre-commit run --all-files
```

Install repository hooks once with `pre-commit install`.
