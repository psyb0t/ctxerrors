---
title: "API reference"
description: "Every exported function and CTXError method in ctxerrors, with its signature, return values and nil behavior."
nav_order: 6
---

# API reference

Everything `github.com/psyb0t/ctxerrors` exports. The sentinel errors live in a separate package, listed on [Common sentinels](commerr.md). The same reference, generated from the doc comments, is on [pkg.go.dev](https://pkg.go.dev/github.com/psyb0t/ctxerrors).

## Creating errors

| Signature | Returns |
| --- | --- |
| `func New(message string) error` | A `*CTXError` with the message and the caller's location. Wraps nothing. |
| `func Wrap(err error, message string) error` | A `*CTXError` around `err` with the message and the caller's location, translated through the [error map](error-mapping.md). Returns nil and logs through `log/slog` when `err` is nil. |
| `func Wrapf(err error, format string, args ...any) error` | Same as `Wrap`, with the message built by `fmt.Sprintf(format, args...)`. |
| `func Join(errs ...error) error` | A `*CTXError` around `errors.Join(errs...)` with the message `<n> errors` and the caller's location. Returns nil when every error is nil. |

The behavior of each one, with examples, is on [Wrapping errors](wrapping.md).

## Error map

| Signature | Effect |
| --- | --- |
| `func SetErrorMap(m map[error]error)` | Replaces the process-wide map with a copy of `m`, skipping nil keys and values. |
| `func MapError(from, to error)` | Adds or overwrites one entry. Does nothing when either is nil. |
| `func ClearErrorMap()` | Removes every entry. |

None of them return anything. How matching works is on [Error mapping](error-mapping.md).

## CTXError

```go
type CTXError struct {
    // unexported fields
}
```

You never build one yourself; `New`, `Wrap`, `Wrapf` and `Join` do. Get at one with `errors.As`:

```go
var ctxErr *ctxerrors.CTXError
if errors.As(err, &ctxErr) {
    fmt.Println(ctxErr.Original())
}
```

Every method is safe to call on a nil `*CTXError`.

| Method | Returns |
| --- | --- |
| `Error() string` | The formatted message described on [Error output](error-output.md). When the error map translated the wrapped error, the translation is printed instead of the original. Empty string on a nil receiver. |
| `Unwrap() error` | The original wrapped error, never the translation. Nil for an error from `New`. |
| `Is(target error) bool` | Whether the error map's translation matches `target` through `errors.Is`. False when nothing was translated. `errors.Is` then keeps going through `Unwrap`, so the original chain matches too. |
| `As(target any) bool` | Whether the translation, or something in its chain, matches `target` through `errors.As`. False when nothing was translated. `errors.As` falls back to the original chain through `Unwrap`. |
| `Original() error` | The error passed to `Wrap` or `Wrapf`, before translation. Nil for an error from `New`. |
| `Mapped() error` | The error map's translation, or nil when no entry matched. |
