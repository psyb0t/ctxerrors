---
title: "Common sentinels"
description: "Every sentinel error in the commerr subpackage, from ErrNotFound and ErrTimeout to ErrRateLimited, with the exact message each one prints."
nav_order: 5
---

# Common sentinels

`SetErrorMap` translates foreign errors INTO something, and [`commerr`](https://github.com/psyb0t/ctxerrors/blob/main/commerr/commerr.go) is the pile of somethings. It's the shared sentinel vocabulary (not-found, already-exists, timeout, the usual suspects) that every service ends up needing, parked in a subpackage so importing `ctxerrors` just to wrap shit doesn't drag the whole list in. You only get them when you ask:

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

They're plain `errors.New` sentinels, so `errors.Is(err, commerr.ErrNotFound)` keeps holding no matter how many times you `Wrap` it, or wrap it with `fmt.Errorf("...: %w", err)`. Any of them works as a target for the [error map](error-mapping.md). `commerr` imports nothing but `errors`.

The package has 60 sentinels. The message column is exactly what `Error()` returns.

## Configuration and environment

| Sentinel | Message |
| --- | --- |
| `ErrEnvVarNotSet` | `env var is not set` |
| `ErrRequiredConfigValueNotSet` | `required config value is not set` |
| `ErrEmptyMigrationsPath` | `migrations path is empty` |

## Files and paths

| Sentinel | Message |
| --- | --- |
| `ErrFileInvalid` | `invalid file` |
| `ErrFileNotFound` | `file not found` |
| `ErrPathIsRequired` | `path is required` |
| `ErrCouldNotDownloadFiles` | `could not download files` |

## Validation and input

| Sentinel | Message |
| --- | --- |
| `ErrInvalidArgument` | `invalid argument` |
| `ErrInvalidValue` | `invalid value` |
| `ErrTargetNotPointer` | `target is not a pointer` |
| `ErrCouldNotDecode` | `could not decode` |
| `ErrValidationFailed` | `validation failed` |
| `ErrOutOfRange` | `out of range` |

## Fields and data

| Sentinel | Message |
| --- | --- |
| `ErrNilOutput` | `output is nil` |
| `ErrNilField` | `field is nil` |
| `ErrRequiredFieldNotSet` | `required field is not set` |
| `ErrRequiredLLMResponseFieldNotSet` | `required llm response field is not set` |
| `ErrAlreadyExists` | `already exists` |

## Jobs and processes

| Sentinel | Message |
| --- | --- |
| `ErrJobFailed` | `job failed` |
| `ErrUnexpectedNumberOfResults` | `unexpected number of results` |
| `ErrNotFound` | `not found` |

## Operations

| Sentinel | Message |
| --- | --- |
| `ErrFetchFailed` | `fetch failed` |
| `ErrParseFailed` | `parse failed` |
| `ErrReadFailed` | `read failed` |
| `ErrWriteFailed` | `write failed` |
| `ErrOpenFailed` | `open failed` |
| `ErrCloseFailed` | `close failed` |
| `ErrExecFailed` | `exec failed` |
| `ErrPublishFailed` | `publish failed` |
| `ErrSubscribeFailed` | `subscribe failed` |
| `ErrDownloadFailed` | `download failed` |
| `ErrUploadFailed` | `upload failed` |
| `ErrUpsertFailed` | `upsert failed` |
| `ErrDeleteFailed` | `delete failed` |
| `ErrConnectFailed` | `connect failed` |
| `ErrBrowseFailed` | `browse failed` |
| `ErrSeedFailed` | `seed failed` |
| `ErrMigrationFailed` | `migration failed` |
| `ErrUnmarshalFailed` | `unmarshal failed` |
| `ErrMarshalFailed` | `marshal failed` |

## Capability

| Sentinel | Message |
| --- | --- |
| `ErrNotImplemented` | `not implemented` |
| `ErrUnsupported` | `unsupported` |

## Process state

| Sentinel | Message |
| --- | --- |
| `ErrUnavailable` | `unavailable` |
| `ErrFailed` | `failed` |
| `ErrTimeout` | `timeout` |
| `ErrTerminated` | `terminated` |
| `ErrKilled` | `killed` |
| `ErrClosing` | `closing` |
| `ErrClosed` | `closed` |
| `ErrShuttingDown` | `shutting down` |
| `ErrCancelled` | `cancelled` |
| `ErrNotReady` | `not ready` |
| `ErrInvalidState` | `invalid state` |
| `ErrExpired` | `expired` |

## Concurrency and capacity

| Sentinel | Message |
| --- | --- |
| `ErrLockHeld` | `lock held` |
| `ErrConflict` | `conflict` |
| `ErrExhausted` | `exhausted` |

## Access and rate limits

| Sentinel | Message |
| --- | --- |
| `ErrNotAuthenticated` | `not authenticated` |
| `ErrPermissionDenied` | `permission denied` |
| `ErrRateLimited` | `rate limited` |

## Unavailable versus not found

`ErrUnavailable` is for a capability the process knows about but that was never wired: an optional dependency left unconfigured, a feature switched off, a subsystem that failed to start. It is deliberately distinct from `ErrNotFound`. A lookup that came back empty tells the caller to fix the query; an unavailable capability tells them to fix the configuration.
