---
title: "Wrapping errors"
description: "How New, Wrap, Wrapf and Join capture the caller's file, line and function, what they do with nil, and how errors.Is, errors.As and errors.Unwrap see the result."
nav_order: 2
---

# Wrapping errors

This package captures where your errors happen. No more hunting through logs like some dickless detective wondering where the fuck that error came from. Four functions do the work, and each one records the file, line and function of whoever called it.

| Function | What it does |
| --- | --- |
| `New(message)` | Creates a new error with location context. |
| `Wrap(err, message)` | Wraps an existing error with a message and the location. |
| `Wrapf(err, format, args...)` | Like `Wrap()`, with printf-style formatting, because we're not animals. |
| `Join(errs...)` | Squashes a pile of errors into one that still knows where they got squashed. |

All four return a plain `error`. When it is not nil, it holds a `*CTXError`, which supports `errors.Unwrap()`, `errors.Is()` and `errors.As()`. The full signatures are in the [API reference](api.md).

## Creating new errors

```go
import "github.com/psyb0t/ctxerrors"

func doSomething() error {
    return ctxerrors.New("shit went sideways")
}
```

## Wrapping existing errors

```go
func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return ctxerrors.Wrap(err, "failed to open file")
    }
    defer file.Close()

    // do stuff...

    return nil
}
```

## Formatted wrapping

```go
func connectToDatabase(host string, port int) error {
    conn, err := sql.Open(
        "postgres",
        fmt.Sprintf("host=%s port=%d", host, port),
    )
    if err != nil {
        return ctxerrors.Wrapf(
            err,
            "failed to connect to database at %s:%d",
            host,
            port,
        )
    }
    defer conn.Close()

    return nil
}
```

## Wrapping nil

`Wrap(nil, ...)` and `Wrapf(nil, ...)` return nil, so a wrap on a happy path never invents an error. They also log it, because wrapping nil is almost always a bug in the caller. The log goes through the default `log/slog` logger at error level, with the message `Trying to wrap a nil error` and the file, line and function of the three frames above the call. With the default handler, the line starts with a timestamp:

```
2026/01/02 15:04:05 ERROR Trying to wrap a nil error sourceFile1=/path/to/main.go:57 sourceFunc1=main.main sourceFile2=... sourceFunc2=... sourceFile3=... sourceFunc3=...
```

Point `slog.SetDefault()` at your own handler if you want that line in your log pipeline.

## Joining errors

Sometimes one operation fans out and more than one branch can fail on its own: three log sinks, a batch of rows, whatever. Bailing on the first failure means the rest never even get attempted, and you find out about exactly one problem when there were three. Do all the work, collect what broke, hand it back as one error:

```go
func writeToAllSinks(sinks []Sink, record Record) error {
    var errs []error

    for i, sink := range sinks {
        if err := sink.Write(record); err != nil {
            errs = append(errs, ctxerrors.Wrapf(err, "sink %d failed", i))
        }
    }

    return ctxerrors.Join(errs...)
}
```

`Join()` ignores nil errors and returns nil when every one of them is nil, so you can hand it the slice without checking whether anything actually went wrong. The message is the number of non-nil errors, followed by what the standard library's `errors.Join` prints, one error per line:

```
2 errors: sink 0 failed: disk full [/path/to/sinks.go:12 in main.writeToAllSinks]
sink 2 failed: connection reset [/path/to/sinks.go:12 in main.writeToAllSinks] [/path/to/sinks.go:16 in main.writeToAllSinks]
```

The result unwraps to the `errors.Join` value, so `errors.Is()` and `errors.As()` still find any of the individual errors you put in. Unlike the standard library version, the result knows the line where you joined them. `Join()` does not run the [error map](error-mapping.md); the errors inside it keep whatever translation their own `Wrap` gave them.

## Unwrapping errors

You can unwrap the chain to get to the original error. This uses `startServer` from the [error chaining example](error-output.md#error-chaining):

```go
err := startServer()
originalErr := errors.Unwrap(errors.Unwrap(err))
// originalErr is now the "config file missing" error

// Or use errors.Is() to check if a specific error is in the chain
if errors.Is(err, someSpecificError) {
    // handle it
}

// Or use errors.As() to check if it's a CTXError
var ctxErr *ctxerrors.CTXError
if errors.As(err, &ctxErr) {
    fmt.Println("Got a CTXError:", ctxErr.Error())
}
```

`errors.Unwrap` always returns the error you passed to `Wrap`, even when the [error map](error-mapping.md) translated it. `errors.Is` and `errors.As` match both the translation and everything in the original chain.

No more guessing where the fuck everything went tits up. To see what the chain prints, read [Error output](error-output.md).
