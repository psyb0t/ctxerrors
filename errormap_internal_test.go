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
			name:      "direct match swaps underlying err",
			mappings:  map[error]error{errForeignNotFound: errBusinessNotFound},
			input:     errForeignNotFound,
			wantIs:    []error{errBusinessNotFound},
			wantNotIs: []error{errForeignNotFound},
		},
		{
			name:     "chained match via errors.Is still translates",
			mappings: map[error]error{errForeignNotFound: errBusinessNotFound},
			input:    fmt.Errorf("driver: %w", errForeignNotFound),
			wantIs:   []error{errBusinessNotFound},
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
			wantIs:    []error{errBusinessConflict},
			wantNotIs: []error{errBusinessNotFound, errForeignConflict},
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
			wantIs:    []error{errBusinessNotFound},
			wantNotIs: []error{errForeignNotFound},
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
