// Package commerr holds the common sentinel errors shared across services: the
// vocabulary ctxerrors' error map translates foreign driver errors into, and
// that callers match with errors.Is. Each is a plain comparable error, so
// wrapping one with ctxerrors.Wrap keeps errors.Is(err, commerr.ErrX) true
// through every layer.
package commerr

import "errors"

var (
	// Configuration & Environment errors
	ErrEnvVarNotSet              = errors.New("env var is not set")
	ErrRequiredConfigValueNotSet = errors.New("required config value is not set")
	ErrEmptyMigrationsPath       = errors.New("migrations path is empty")

	// File & Path errors
	ErrFileInvalid           = errors.New("invalid file")
	ErrFileNotFound          = errors.New("file not found")
	ErrPathIsRequired        = errors.New("path is required")
	ErrCouldNotDownloadFiles = errors.New("could not download files")

	// Validation & Input errors
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrInvalidValue     = errors.New("invalid value")
	ErrTargetNotPointer = errors.New("target is not a pointer")
	ErrCouldNotDecode   = errors.New("could not decode")

	// Field & Data errors
	ErrNilOutput                      = errors.New("output is nil")
	ErrNilField                       = errors.New("field is nil")
	ErrRequiredFieldNotSet            = errors.New("required field is not set")
	ErrRequiredLLMResponseFieldNotSet = errors.New("required llm response field is not set")
	ErrAlreadyExists                  = errors.New("already exists")

	// Job & Process errors
	ErrJobFailed                 = errors.New("job failed")
	ErrUnexpectedNumberOfResults = errors.New("unexpected number of results")
	ErrNotFound                  = errors.New("not found")

	// Operation errors
	ErrFetchFailed     = errors.New("fetch failed")
	ErrParseFailed     = errors.New("parse failed")
	ErrWriteFailed     = errors.New("write failed")
	ErrPublishFailed   = errors.New("publish failed")
	ErrSubscribeFailed = errors.New("subscribe failed")
	ErrDownloadFailed  = errors.New("download failed")
	ErrUploadFailed    = errors.New("upload failed")
	ErrUpsertFailed    = errors.New("upsert failed")
	ErrDeleteFailed    = errors.New("delete failed")
	ErrConnectFailed   = errors.New("connect failed")
	ErrBrowseFailed    = errors.New("browse failed")
	ErrSeedFailed      = errors.New("seed failed")
	ErrMigrationFailed = errors.New("migration failed")
	ErrUnmarshalFailed = errors.New("unmarshal failed")
	ErrMarshalFailed   = errors.New("marshal failed")

	// Process State errors
	//
	// ErrUnavailable is for a capability the process knows about but that was
	// never wired: an optional dependency left unconfigured, a feature switched
	// off, a subsystem that failed to start. Deliberately distinct from
	// ErrNotFound — a lookup that came back empty tells the caller to fix the
	// query, an unavailable capability tells them to fix the configuration.
	ErrUnavailable  = errors.New("unavailable")
	ErrFailed       = errors.New("failed")
	ErrTimeout      = errors.New("timeout")
	ErrTerminated   = errors.New("terminated")
	ErrKilled       = errors.New("killed")
	ErrClosing      = errors.New("closing")
	ErrShuttingDown = errors.New("shutting down")
	ErrCancelled    = errors.New("cancelled")

	// API & HTTP errors
	ErrUnexpectedHTTPStatusCode = errors.New("unexpected http status code")
	ErrNotAuthenticated         = errors.New("not authenticated")
	ErrRateLimited              = errors.New("rate limited")
)
