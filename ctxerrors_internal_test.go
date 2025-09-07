package ctxerrors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	goFileExtension = ".go"
)

func TestNew(t *testing.T) {
	funcName := "TestNew"

	testCases := []struct {
		name          string
		message       string
		expectedParts []string
	}{
		{
			name:          "basic error creation",
			message:       "something went wrong",
			expectedParts: []string{"something went wrong", funcName, goFileExtension},
		},
		{
			name:          "empty message",
			message:       "",
			expectedParts: []string{funcName, goFileExtension},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := New(tc.message)
			require.NotNil(t, actual)

			var actualErr *ErrorWithContext

			require.True(t, errors.As(actual, &actualErr))
			require.Equal(t, tc.message, actualErr.message)

			// Check error string contains expected parts
			actualStr := actual.Error()
			for _, expected := range tc.expectedParts {
				require.Contains(t, actualStr, expected)
			}

			// Location checks
			require.NotEmpty(t, actualErr.file)
			require.True(t, strings.HasSuffix(actualErr.file, goFileExtension))
			require.NotZero(t, actualErr.line)
			require.Contains(t, actualErr.funcName, funcName)

			// Unwrap should return nil since this is a new error
			require.Nil(t, errors.Unwrap(actual))
		})
	}
}

func TestWrap(t *testing.T) { //nolint:funlen
	baseErr := errors.New("base error") //nolint:err113

	testCases := []struct {
		name          string
		err           error
		message       string
		expectedParts []string
		expectedBase  error
	}{
		{
			name:          "wrap existing error",
			err:           baseErr,
			message:       "additional context",
			expectedParts: []string{"additional context", "base error", "TestWrap", goFileExtension},
			expectedBase:  baseErr,
		},
		{
			name:          "wrap nil error",
			err:           nil,
			message:       "should return nil",
			expectedParts: nil,
			expectedBase:  nil,
		},
		{
			name:          "wrap with empty message",
			err:           baseErr,
			message:       "",
			expectedParts: []string{"base error", "TestWrap", goFileExtension},
			expectedBase:  baseErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Wrap(tc.err, tc.message)

			if tc.err == nil {
				require.Nil(t, actual)

				return
			}

			require.NotNil(t, actual)

			var actualErr *ErrorWithContext

			require.True(t, errors.As(actual, &actualErr))
			require.Equal(t, tc.message, actualErr.message)

			// Check error string contains expected parts
			actualStr := actual.Error()
			for _, expected := range tc.expectedParts {
				require.Contains(t, actualStr, expected)
			}

			// Location checks
			require.NotEmpty(t, actualErr.file)
			require.True(t, strings.HasSuffix(actualErr.file, goFileExtension))
			require.NotZero(t, actualErr.line)
			require.Contains(t, actualErr.funcName, "TestWrap")

			// Unwrap checks
			if tc.expectedBase != nil {
				require.Equal(t, tc.expectedBase, errors.Unwrap(actual))
			}
		})
	}
}

func TestErrorWithContextError(t *testing.T) {
	baseErr := errors.New("base error") //nolint:err113

	testCases := []struct {
		name          string
		err           *ErrorWithContext
		expectedParts []string
	}{
		{
			name: "with wrapped error",
			err: &ErrorWithContext{
				err:      baseErr,
				message:  "context message",
				file:     "test.go",
				line:     42,
				funcName: "TestFunc",
			},
			expectedParts: []string{
				"context message",
				"base error",
				"test.go",
				"42",
				"TestFunc",
			},
		},
		{
			name: "without wrapped error",
			err: &ErrorWithContext{
				message:  "standalone message",
				file:     "test.go",
				line:     42,
				funcName: "TestFunc",
			},
			expectedParts: []string{
				"standalone message",
				"test.go",
				"42",
				"TestFunc",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.err.Error()
			for _, expected := range tc.expectedParts {
				require.Contains(t, actual, expected)
			}
		})
	}
}

func TestLocationFunctions(t *testing.T) {
	t.Run("getCallerInfo", func(t *testing.T) {
		file, line, funcName := getCallerInfo(0)
		require.NotEmpty(t, file)
		require.True(t, strings.HasSuffix(file, goFileExtension))
		require.NotZero(t, line)
		require.NotEmpty(t, funcName)
		require.Contains(t, funcName, "TestLocationFunctions")

		// Invalid skip should return empty values
		file, line, funcName = getCallerInfo(9999)
		require.Empty(t, file)
		require.Zero(t, line)
		require.Empty(t, funcName)
	})
}

func TestErrorUnwrapping(t *testing.T) {
	baseErr := errors.New("base error") //nolint:err113
	wrapped := Wrap(baseErr, "first wrap")
	doubleWrapped := Wrap(wrapped, "second wrap")

	// Test unwrapping chain
	actual := errors.Unwrap(errors.Unwrap(doubleWrapped))
	require.Equal(t, baseErr, actual)

	// Test Is functionality
	require.True(t, errors.Is(doubleWrapped, baseErr))
	require.True(t, errors.Is(wrapped, baseErr))

	// Test As functionality
	var contextErr *ErrorWithContext

	require.True(t, errors.As(doubleWrapped, &contextErr))
	require.Equal(t, "second wrap", contextErr.message)
}

func TestWrapf(t *testing.T) { //nolint:funlen
	baseErr := errors.New("base error") //nolint:err113

	testCases := []struct {
		name          string
		err           error
		format        string
		args          []any
		expectedParts []string
		expectedBase  error
	}{
		{
			name:          "wrap with formatted message",
			err:           baseErr,
			format:        "error occurred: %s (code=%d)",
			args:          []any{"invalid input", 400},
			expectedParts: []string{"error occurred: invalid input (code=400)", "base error", "TestWrapf", goFileExtension},
			expectedBase:  baseErr,
		},
		{
			name:          "wrap nil error",
			err:           nil,
			format:        "should return nil: %s",
			args:          []any{"unused"},
			expectedParts: nil,
			expectedBase:  nil,
		},
		{
			name:          "wrap with empty format",
			err:           baseErr,
			format:        "",
			args:          nil,
			expectedParts: []string{"base error", "TestWrapf", goFileExtension},
			expectedBase:  baseErr,
		},
		{
			name:          "wrap with format but no args",
			err:           baseErr,
			format:        "simple message with no formatting",
			args:          nil,
			expectedParts: []string{"simple message with no formatting", "base error", "TestWrapf", goFileExtension},
			expectedBase:  baseErr,
		},
		{
			name:          "wrap with multiple format args",
			err:           baseErr,
			format:        "error in %s: value=%d, status=%s",
			args:          []any{"processor", 42, "failed"},
			expectedParts: []string{"error in processor: value=42, status=failed", "base error", "TestWrapf", goFileExtension},
			expectedBase:  baseErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Wrapf(tc.err, tc.format, tc.args...)

			if tc.err == nil {
				require.Nil(t, actual)

				return
			}

			require.NotNil(t, actual)

			var actualErr *ErrorWithContext

			require.True(t, errors.As(actual, &actualErr))

			// Verify formatted message
			expectedMessage := fmt.Sprintf(tc.format, tc.args...)
			require.Equal(t, expectedMessage, actualErr.message)

			// Check error string contains expected parts
			actualStr := actual.Error()
			for _, expected := range tc.expectedParts {
				require.Contains(t, actualStr, expected)
			}

			// Location checks
			require.NotEmpty(t, actualErr.file)
			require.True(t, strings.HasSuffix(actualErr.file, goFileExtension))
			require.NotZero(t, actualErr.line)
			require.Contains(t, actualErr.funcName, "TestWrapf")

			// Unwrap checks
			if tc.expectedBase != nil {
				require.Equal(t, tc.expectedBase, errors.Unwrap(actual))
			}
		})
	}
}
