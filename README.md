# ctxerrors

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/ctxerrors.svg)](https://pkg.go.dev/github.com/psyb0t/ctxerrors)
[![CI](https://github.com/psyb0t/ctxerrors/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/ctxerrors/actions/workflows/pipeline.yml)
[![coverage](https://raw.githubusercontent.com/psyb0t/ctxerrors/badges/coverage.svg)](https://github.com/psyb0t/ctxerrors/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/ctxerrors/badges/version.svg)](https://github.com/psyb0t/ctxerrors/tags)
[![license](https://raw.githubusercontent.com/psyb0t/ctxerrors/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/ctxerrors/badges/importers.svg)](https://github.com/psyb0t/ctxerrors/blob/badges/importers.md)

```
 ####  ##### #    # ###### #####  #####   ####  #####   ####  
#    #   #    #  #  #      #    # #    # #    # #    # #      
#        #     ##   #####  #    # #    # #    # #    #  ####  
#        #     ##   #      #####  #####  #    # #####       # 
#    #   #    #  #  #      #   #  #   #  #    # #   #  #    # 
 ####    #   #    # ###### #    # #    #  ####  #    #  ####  
```

fuck yeah, another Go error handling package 🖕

A Go library that wraps errors with context information (file, line, function) because debugging without context is like trying to find your dick in the dark.

## Table of Contents

- [Installation](#installation)
- [What the fuck does it do?](#what-the-fuck-does-it-do)
- [Usage](#usage)
- [Error output](#error-output)
- [Error mapping](#error-mapping)
- [Common sentinels](#common-sentinels)
- [Documentation](#documentation)
- [Sponsor](#sponsor)
- [License](#license)
- [Why?](#why)

## Installation

```bash
go get github.com/psyb0t/ctxerrors
```

Needs Go 1.26 or newer. The library itself imports only the standard library.

## What the fuck does it do?

This package automatically captures where your errors happen in your code. No more hunting through logs like some dickless detective wondering where the fuck that error came from.

- **New()** creates a new error with location context.
- **Wrap()** wraps an existing error with additional context and location.
- **Wrapf()** is Wrap() with printf-style formatting, because we're not animals.
- **Join()** squashes a pile of errors into one that still knows where they got squashed. For when one thing fans out and several bits can shit the bed independently: three log sinks, a batch of rows, whatever.
- **SetErrorMap() / MapError() / ClearErrorMap()** translate foreign sentinel errors (gorm, sql, redis...) into your own business errors at wrap time.

`New()`, `Wrap()`, `Wrapf()` and `Join()` return a plain `error` holding a `*CTXError`, which supports `errors.Unwrap()`, `errors.Is()` and `errors.As()` because Go's error handling conventions aren't completely ass-backwards. `Wrap(nil, ...)` and `Wrapf(nil, ...)` return nil and log the call site through `log/slog`, and `Join()` returns nil when everything you gave it is nil. The error map functions don't return anything; they just manage the translation map.

## Usage

```go
import "github.com/psyb0t/ctxerrors"

func processFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return ctxerrors.Wrap(err, "failed to open file")
    }
    defer file.Close()

    // do stuff...

    return nil
}

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

[Wrapping errors](https://psyb0t.github.io/ctxerrors/wrapping/) covers `New`, `Wrap`, `Wrapf` and `Join`, the nil handling and how `errors.Is`/`errors.As` see the result.

## Error output

When shit hits the fan, you get detailed context:

```
failed to connect to database at localhost:5432: dial tcp [::1]:5432: connect: connection refused [/path/to/your/file.go:42 in main.connectToDatabase]
```

Wrap it again and every layer adds its own location, so one line shows the whole path the error took. [Error output](https://psyb0t.github.io/ctxerrors/error-output/) has the exact format, longer chains, and what Go does to your closure names in the trace.

## Error mapping

Stop foreign errors from leaking through your layers. Register a translation map at startup and `Wrap`/`Wrapf` tag matching driver errors with your own business errors while wrapping:

```go
ctxerrors.SetErrorMap(map[error]error{
    gorm.ErrRecordNotFound: commerr.ErrNotFound,
    gorm.ErrDuplicatedKey:  commerr.ErrAlreadyExists,
})
```

The wrapped error matches both `commerr.ErrNotFound` and the original gorm error with `errors.Is`, and `Error()` prints the mapped sentinel so driver details stay out of logs. `CTXError.Original()` hands you the raw cause when you really need it. The matching rules are on [Error mapping](https://psyb0t.github.io/ctxerrors/error-mapping/).

## Common sentinels

`commerr` is the shared sentinel vocabulary (not-found, already-exists, timeout, the usual suspects) that every service ends up needing, parked in a subpackage so importing `ctxerrors` just to wrap shit doesn't drag the whole list in:

```go
import "github.com/psyb0t/ctxerrors/commerr"

func GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, commerr.ErrInvalidArgument
    }
    // ...
    return nil, commerr.ErrNotFound
}
```

They're plain `errors.New` sentinels, so `errors.Is(err, commerr.ErrNotFound)` keeps holding no matter how many times you `Wrap` it. All 60 of them, with their messages, are on [Common sentinels](https://psyb0t.github.io/ctxerrors/commerr/).

## Documentation

The full docs live at **[psyb0t.github.io/ctxerrors](https://psyb0t.github.io/ctxerrors/)**.

| Page | What's in it |
| --- | --- |
| [Wrapping errors](https://psyb0t.github.io/ctxerrors/wrapping/) | `New`, `Wrap`, `Wrapf`, `Join`, nil handling, unwrapping |
| [Error output](https://psyb0t.github.io/ctxerrors/error-output/) | The exact `Error()` format and a pile of stupid examples |
| [Error mapping](https://psyb0t.github.io/ctxerrors/error-mapping/) | Translating driver errors, matching rules, `Original()` and `Mapped()` |
| [Common sentinels](https://psyb0t.github.io/ctxerrors/commerr/) | Every `commerr` sentinel and its message |
| [API reference](https://psyb0t.github.io/ctxerrors/api/) | Every exported function and method |
| [Development](https://psyb0t.github.io/ctxerrors/development/) | Make targets, coverage gate, CI and releases |

## Sponsor

<a href="https://www.buymeacoffee.com/psyb0t"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me a Coffee" height="60" width="217"></a>

## License

MIT License, because lawyers are expensive and I don't want to deal with that shit

## Why?

Because Go's error handling is verbose as fuck and debugging without context is like trying to find your asshole with both hands tied behind your back. This shit makes it slightly less painful.
