---
title: "Home"
description: "ctxerrors is a Go library that stamps every error with the file, line and function where it was created or wrapped, so you stop hunting through logs."
nav_order: 1
permalink: /
seo:
  type: WebSite
---

# ctxerrors

fuck yeah, another Go error handling package.

ctxerrors wraps errors with the file, line and function where you created or wrapped them, because debugging without context is like trying to find your dick in the dark. Every layer of a wrap chain adds its own location, so one log line tells you the whole path the error took.

```
failed to connect to database at localhost:5432: dial tcp [::1]:5432: connect: connection refused [/path/to/your/file.go:42 in main.connectToDatabase]
```

It also translates foreign errors (gorm, sql, redis, whatever) into your own sentinels at wrap time, and ships [`commerr`](commerr.md), a package of 60 general-purpose sentinel errors so you can stop declaring `ErrNotFound` in every repository.

## Install

```bash
go get github.com/psyb0t/ctxerrors
```

The module declares `go 1.26` in its [`go.mod`](https://github.com/psyb0t/ctxerrors/blob/main/go.mod), so you need Go 1.26 or newer. The library itself imports only the standard library.

## Quick look

```go
import "github.com/psyb0t/ctxerrors"

func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return ctxerrors.Wrap(err, "failed to open file")
    }
    defer file.Close()

    return nil
}
```

## Where to go next

| You want to | Read |
| --- | --- |
| Create, wrap and join errors, and check them with `errors.Is` and `errors.As` | [Wrapping errors](wrapping.md) |
| See exactly what `Error()` prints for a chain | [Error output](error-output.md) |
| Turn driver errors into your own business errors | [Error mapping](error-mapping.md) |
| Find a ready-made sentinel like `ErrNotFound` or `ErrTimeout` | [Common sentinels](commerr.md) |
| Look up a signature | [API reference](api.md) |
| Build, lint and test the library itself | [Development](development.md) |

The Go package docs are on [pkg.go.dev](https://pkg.go.dev/github.com/psyb0t/ctxerrors), and the source is on [GitHub](https://github.com/psyb0t/ctxerrors).
