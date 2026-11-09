---
title: "Error mapping"
description: "Translate foreign driver errors like gorm.ErrRecordNotFound or sql.ErrNoRows into your own sentinels at wrap time, while keeping the original reachable."
nav_order: 4
---

# Error mapping

Stop foreign errors from leaking through your layers. Register a translation map at startup, and `Wrap`/`Wrapf` tag matching driver errors with your own business errors while wrapping.

```go
package main

import (
    "gorm.io/gorm"

    "github.com/psyb0t/ctxerrors"
    "github.com/psyb0t/ctxerrors/commerr"
)

func init() {
    ctxerrors.SetErrorMap(map[error]error{
        gorm.ErrRecordNotFound: commerr.ErrNotFound,
        gorm.ErrDuplicatedKey:  commerr.ErrAlreadyExists,
    })
}

func GetUser(id int) error {
    err := db.First(&user, id).Error // returns gorm.ErrRecordNotFound
    if err != nil {
        // wrapped err satisfies errors.Is(err, commerr.ErrNotFound), and the
        // original gorm error is still in the chain
        return ctxerrors.Wrap(err, "get user")
    }
    return nil
}
```

The mapping targets come from [`commerr`](commerr.md), the shared sentinel package shipped alongside, so there's no need to hand-declare `ErrNotFound` yourself. Any error works as a target, your own included.

## Managing the map

| Function | What it does |
| --- | --- |
| `SetErrorMap(map[error]error)` | Replaces the whole map. Meant for startup. |
| `MapError(from, to error)` | Adds or overwrites one entry. Good for `init()` in each package. |
| `ClearErrorMap()` | Wipes it. Mostly for tests. |

There is one map per process, guarded by a read-write mutex, so registering and wrapping from different goroutines is safe. `SetErrorMap` copies the map you pass, so changing your map afterwards does nothing. Entries with a nil key or a nil value are skipped, and `MapError` with a nil argument does nothing.

## How matching works

- `Wrap` and `Wrapf` look up the error they are given. `New` and `Join` never touch the map.
- Matching uses `errors.Is`, so an already-wrapped foreign error still translates.
- Translation is single-pass. There is no chained `A→B→C`; the target is used as is.
- If no entry matches, behavior is unchanged.
- If more than one entry matches, Go's random map order decides which target wins. Keep keys that can match the same error out of the same map.
- The lookup happens at wrap time. Changing the map later does not change errors you already wrapped.

## What the wrapped error looks like

The translation is added, not swapped in. `errors.Is` and `errors.As` match the mapped sentinel and everything in the original chain, and `errors.Unwrap` returns the original error. That matters when the foreign error sits inside one of your typed errors, say an `*ErrNotFound` with an HTTP status that wraps `sql.ErrNoRows`. Wrapping it keeps the `*ErrNotFound` findable, so the HTTP layer still answers 404.

`Error()` prints the mapped sentinel, never the original message, so driver details stay out of logs and responses:

```go
ctxerrors.SetErrorMap(map[error]error{
    sql.ErrNoRows: commerr.ErrNotFound,
})

err := ctxerrors.Wrap(sql.ErrNoRows, "get user")

fmt.Println(err)                               // get user: not found [/path/to/main.go:55 in main.main]
fmt.Println(errors.Is(err, commerr.ErrNotFound)) // true
fmt.Println(errors.Is(err, sql.ErrNoRows))       // true
```

When you really need the raw cause, get the `*CTXError` and ask for it:

```go
var ctxErr *ctxerrors.CTXError
if errors.As(err, &ctxErr) {
    ctxErr.Original() // the error passed to Wrap, before translation
    ctxErr.Mapped()   // the translation, nil when no entry matched
}
```

`errors.As` stops at the outermost `*CTXError`. If the translated error sits deeper in the chain, call `errors.Unwrap` on that `*CTXError` and keep looking.
