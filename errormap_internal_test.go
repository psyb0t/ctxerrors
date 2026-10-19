package ctxerrors

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:err113,gochecknoglobals
var (
	errForeignNotFound  = errors.New("foreign: record not found")
	errForeignConflict  = errors.New("foreign: conflict")
	errBusinessNotFound = errors.New("business: not found")
	errBusinessConflict = errors.New("business: already exists")
	errUnrelated        = errors.New("unrelated boom")
)

func TestWrapWithMapping(t *testing.T) {
	testCases := []struct {
		name        string
		mappings    map[error]error
		input       error
		wantIs      []error
		wantNotIs   []error
		wantNilWrap bool
	}{
		{
			name:      "direct match adds the mapped err and keeps the original",
			mappings:  map[error]error{errForeignNotFound: errBusinessNotFound},
			input:     errForeignNotFound,
			wantIs:    []error{errBusinessNotFound, errForeignNotFound},
			wantNotIs: []error{errBusinessConflict},
		},
		{
			name:     "chained match via errors.Is still translates",
			mappings: map[error]error{errForeignNotFound: errBusinessNotFound},
			input:    fmt.Errorf("driver: %w", errForeignNotFound),
			wantIs:   []error{errBusinessNotFound, errForeignNotFound},
		},
		{
			name:      "no match preserves original err",
			mappings:  map[error]error{errForeignNotFound: errBusinessNotFound},
			input:     errUnrelated,
			wantIs:    []error{errUnrelated},
			wantNotIs: []error{errBusinessNotFound},
		},
		{
			name:     "empty map preserves original err",
			mappings: nil,
			input:    errForeignNotFound,
			wantIs:   []error{errForeignNotFound},
		},
		{
			name: "multiple mappings, picks the matching one",
			mappings: map[error]error{
				errForeignNotFound: errBusinessNotFound,
				errForeignConflict: errBusinessConflict,
			},
			input:     errForeignConflict,
			wantIs:    []error{errBusinessConflict, errForeignConflict},
			wantNotIs: []error{errBusinessNotFound, errForeignNotFound},
		},
		{
			name:        "nil input still returns nil",
			mappings:    map[error]error{errForeignNotFound: errBusinessNotFound},
			input:       nil,
			wantNilWrap: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(ClearErrorMap)

			SetErrorMap(tc.mappings)

			got := Wrap(tc.input, "ctx")

			if tc.wantNilWrap {
				require.Nil(t, got)

				return
			}

			require.NotNil(t, got)

			for _, target := range tc.wantIs {
				require.Truef(t, errors.Is(got, target), "expected errors.Is(_, %v) true", target)
			}

			for _, target := range tc.wantNotIs {
				require.Falsef(t, errors.Is(got, target), "expected errors.Is(_, %v) false", target)
			}
		})
	}
}

// typedError is an application error type that carries a foreign cause and
// matches any other *typedError with errors.Is, the shape HTTP layers detect
// with errors.As to pick a status code.
type typedError struct {
	inner error
}

func (e *typedError) Error() string { return "typed: " + e.inner.Error() }

func (e *typedError) Unwrap() error { return e.inner }

func (e *typedError) Is(target error) bool {
	_, ok := target.(*typedError)

	return ok
}

// mappedTypedError is a typed mapping target, found through errors.As.
type mappedTypedError struct{}

func (e *mappedTypedError) Error() string { return "mapped typed error" }

func TestWrapWithMapping_KeepsTypedErrors(t *testing.T) {
	t.Cleanup(ClearErrorMap)

	MapError(errForeignNotFound, errBusinessNotFound)

	input := &typedError{inner: fmt.Errorf("driver: %w", errForeignNotFound)}
	got := Wrap(input, "load user")

	require.True(t, errors.Is(got, errBusinessNotFound))
	require.True(t, errors.Is(got, errForeignNotFound))
	require.True(t, errors.Is(got, &typedError{}))

	var typed *typedError
	require.True(t, errors.As(got, &typed))
	require.Same(t, input, typed)

	require.Same(t, input, errors.Unwrap(got))

	// The message shows only the translation; the original stays reachable
	// through the *CTXError accessors.
	require.Contains(t, got.Error(), "load user: business: not found")
	require.NotContains(t, got.Error(), "foreign: record not found")
	require.NotContains(t, got.Error(), "typed:")

	var ctxErr *CTXError
	require.True(t, errors.As(got, &ctxErr))
	require.Same(t, input, ctxErr.Original())
	require.Equal(t, errBusinessNotFound, ctxErr.Mapped())
	require.Contains(t, ctxErr.Original().Error(), "foreign: record not found")

	// A second wrap keeps the typed error and the translation.
	outer := Wrap(got, "handle request")
	require.True(t, errors.Is(outer, errBusinessNotFound))
	require.True(t, errors.As(outer, &typed))
	require.Same(t, input, typed)
}

func TestWrapWithMapping_TypedTarget(t *testing.T) {
	t.Cleanup(ClearErrorMap)

	target := &mappedTypedError{}
	MapError(errForeignNotFound, target)

	got := Wrap(errForeignNotFound, "ctx")

	var mapped *mappedTypedError
	require.True(t, errors.As(got, &mapped))
	require.Same(t, target, mapped)

	unmapped := Wrap(errUnrelated, "ctx")
	require.False(t, errors.As(unmapped, &mapped))
}

func TestCTXError_OriginalAndMapped(t *testing.T) {
	t.Cleanup(ClearErrorMap)

	MapError(errForeignNotFound, errBusinessNotFound)

	var ctxErr *CTXError

	// Unmapped: the message and Original are the wrapped error, Mapped is nil.
	require.True(t, errors.As(Wrap(errUnrelated, "ctx"), &ctxErr))
	require.Same(t, errUnrelated, ctxErr.Original())
	require.NoError(t, ctxErr.Mapped())
	require.Contains(t, ctxErr.Error(), "ctx: unrelated boom")

	// Mapped: the message shows the translation.
	require.True(t, errors.As(Wrap(errForeignNotFound, "ctx"), &ctxErr))
	require.Same(t, errForeignNotFound, ctxErr.Original())
	require.Equal(t, errBusinessNotFound, ctxErr.Mapped())
	require.Contains(t, ctxErr.Error(), "ctx: business: not found")
	require.NotContains(t, ctxErr.Error(), "foreign")

	// New wraps nothing.
	require.True(t, errors.As(New("boom"), &ctxErr))
	require.NoError(t, ctxErr.Original())
	require.NoError(t, ctxErr.Mapped())

	// A nil *CTXError is safe.
	var nilErr *CTXError
	require.NoError(t, nilErr.Original())
	require.NoError(t, nilErr.Mapped())
}

func TestMapError(t *testing.T) {
	testCases := []struct {
		name      string
		from      error
		to        error
		input     error
		wantIs    []error
		wantNotIs []error
	}{
		{
			name:      "registers a working mapping",
			from:      errForeignNotFound,
			to:        errBusinessNotFound,
			input:     errForeignNotFound,
			wantIs:    []error{errBusinessNotFound, errForeignNotFound},
			wantNotIs: []error{errBusinessConflict},
		},
		{
			name:   "nil from is no-op",
			from:   nil,
			to:     errBusinessNotFound,
			input:  errForeignNotFound,
			wantIs: []error{errForeignNotFound},
		},
		{
			name:   "nil to is no-op",
			from:   errForeignNotFound,
			to:     nil,
			input:  errForeignNotFound,
			wantIs: []error{errForeignNotFound},
		},
		{
			name:   "both nil is no-op",
			from:   nil,
			to:     nil,
			input:  errForeignNotFound,
			wantIs: []error{errForeignNotFound},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(ClearErrorMap)

			MapError(tc.from, tc.to)

			got := Wrap(tc.input, "ctx")
			for _, target := range tc.wantIs {
				require.Truef(t, errors.Is(got, target), "expected errors.Is(_, %v) true", target)
			}

			for _, target := range tc.wantNotIs {
				require.Falsef(t, errors.Is(got, target), "expected errors.Is(_, %v) false", target)
			}
		})
	}
}

func TestSetErrorMap_ReplacesPreviousAndSkipsNils(t *testing.T) {
	t.Cleanup(ClearErrorMap)

	MapError(errForeignNotFound, errBusinessNotFound)

	SetErrorMap(map[error]error{
		errForeignConflict: errBusinessConflict,
		nil:                errBusinessNotFound,
		errForeignNotFound: nil,
	})

	got := Wrap(errForeignNotFound, "ctx")
	require.True(t, errors.Is(got, errForeignNotFound))
	require.False(t, errors.Is(got, errBusinessNotFound))

	got = Wrap(errForeignConflict, "ctx")
	require.True(t, errors.Is(got, errBusinessConflict))
}

func TestClearErrorMap(t *testing.T) {
	MapError(errForeignNotFound, errBusinessNotFound)
	ClearErrorMap()

	got := Wrap(errForeignNotFound, "ctx")
	require.True(t, errors.Is(got, errForeignNotFound))
	require.False(t, errors.Is(got, errBusinessNotFound))
}

func TestMapError_Concurrent(t *testing.T) {
	t.Cleanup(ClearErrorMap)

	MapError(errForeignNotFound, errBusinessNotFound)

	var wg sync.WaitGroup

	const workers = 32

	wg.Add(workers)

	for i := range workers {
		go func(i int) {
			defer wg.Done()

			if i%2 == 0 {
				MapError(errForeignConflict, errBusinessConflict)

				return
			}

			got := Wrap(errForeignNotFound, "ctx")
			require.True(t, errors.Is(got, errBusinessNotFound))
		}(i)
	}

	wg.Wait()
}
