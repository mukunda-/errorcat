package errorcat_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	cat "go.mukunda.com/errorcat"
)

var ErrWriteImaginaryFileFailed = errors.New("write to imaginary file failed")
var errTest = errors.New("test-error")
var errTest2 = errors.New("test-error2")

// All errors from Catch are warpped in CatError which
// can be unwrapped to the original error.
func TestUnwrappingCatError(t *testing.T) {

	defer func() {
		r := recover()
		assert.IsType(t, r, cat.CatError{})
		err := errors.Unwrap(r.(error))
		assert.ErrorIs(t, err, io.ErrClosedPipe)
	}()

	cat.Catch(io.ErrClosedPipe, "test")
}

// When a function panics with a non-error type, that is also caught and wrapped into
// a general error.
func TestCatchingRealPanic(t *testing.T) {

	defer cat.Guard(nil, func(err error) error {
		// The error should contain the original text.
		assert.Equal(t, "test error", err.Error())
		return err
	})

	panic("test error")

}

// Guard annotators can transform caught errors.
func TestGuardAnnotation(t *testing.T) {

	var err error
	func() {
		// String annotators wrap errors with a simple string.

		defer cat.Guard(&err, "string annotation")
		cat.Catch(true, "bad condition 1")
	}()

	assert.Equal(t, "string annotation: bad condition 1", err.Error())

	erasedError := errors.New("the error was erased")
	func() {
		// Function annotators can transform the error.
		defer cat.Guard(&err, func(err error) error {
			return erasedError
		})
		cat.Catch(true, "bad condition 2")
	}()

	assert.Equal(t, "the error was erased", err.Error())

	func() {
		// Other types should not be used, but to be safe they are formatted in a generic
		// manner, same as strings.
		defer cat.Guard(&err, 123)

		cat.Catch(true, "bad condition 3")
	}()

	assert.Equal(t, "123: bad condition 3", err.Error())

	func() {
		// Multiple annotators can be used.
		defer cat.Guard(&err, "first", "second", func(err error) error {
			return fmt.Errorf("and third: %w", err)
		})

		cat.Catch(true, "bad condition 4")
	}()

	assert.Equal(t, "and third: second: first: bad condition 4", err.Error())

	func() {
		// If an annotator function returns nil, the chain is not continued.
		defer cat.Guard(&err,
			"first",
			"second",
			func(err error) error {
				// Clear the error.
				return nil
			},
			func(err error) error {
				assert.Fail(t, "this should not be called")
				return err
			},
			"third",
		)

		cat.Catch(true, "bad condition 5")
	}()

	assert.NoError(t, err)

	func() {
		// If an error is given, that is added to the chain.
		defer cat.Guard(&err, errTest2)

		cat.Catch(true, errTest)
	}()

	assert.Equal(t, "test-error2: test-error", err.Error())
	assert.ErrorIs(t, err, errTest2)
	assert.ErrorIs(t, err, errTest)
}

func assertErrorIsCat(t *testing.T, err error) {
	if _, ok := err.(cat.CatError); !ok {
		assert.Fail(t, "error is not a CatError")
	}
}

// When catching errors, the `problem` describes the error state. It's useful for
// describing what should be done with the error.
func TestProblems(t *testing.T) {
	var err error
	var serviceError = errors.New("service error")

	// When using an error + error combination in cat, both errors are wrapped.
	func() {
		defer cat.Guard(&err)
		cat.Catch(errTest, fmt.Errorf("%w: try again later", serviceError))
	}()

	assert.Equal(t, "service error: try again later: test-error", err.Error())
	assert.ErrorIs(t, err, serviceError)
	assertErrorIsCat(t, err)

	// When using an error + nil combination, the error is wrapped and bubbled without
	// further annotation.
	func() {
		defer cat.Guard(&err)
		cat.Catch(errTest)
	}()

	assert.Equal(t, "test-error", err.Error())
	assert.ErrorIs(t, err, errTest)
	assertErrorIsCat(t, err)

	// When using an error + string combination, the error is wrapped and annotated with
	// the string.
	func() {
		defer cat.Guard(&err)
		cat.Catch(errTest, "problem")
	}()

	assert.Equal(t, "problem: test-error", err.Error())
	assert.ErrorIs(t, err, errTest)
	assertErrorIsCat(t, err)

	// When using a non-string type, it's treated the same (via fmt magic), but you
	// shouldn't be doing that.
	func() {
		defer cat.Guard(&err)
		cat.Catch(errTest, 123)
	}()

	assert.Equal(t, "123: test-error", err.Error())
	assert.ErrorIs(t, err, errTest)
	assertErrorIsCat(t, err)

	// When using a boolean + error combination, the problem is wrapped as the primary
	// error.
	func() {
		defer cat.Guard(&err)
		assert.NotPanics(t, func() {
			cat.Catch(false, fmt.Errorf("should not be thrown"))
		})
		cat.Catch(true, errTest)
	}()

	assert.Equal(t, "test-error", err.Error())
	assert.ErrorIs(t, err, errTest)
	assertErrorIsCat(t, err)

	// When using a boolean + nil combination, the problem is wrapped as an unknown error.
	// This case should not be used in practice.
	func() {
		defer cat.Guard(&err)
		assert.NotPanics(t, func() {
			cat.Catch(false)
		})
		cat.Catch(true)
	}()

	assert.ErrorIs(t, err, cat.ErrUnknown)
	assertErrorIsCat(t, err)

	// When using a boolean + string combination, the problem is wrapped as a general
	// untyped error.
	func() {
		defer cat.Guard(&err)
		cat.Catch(false, "notproblemstring")
		cat.Catch(true, "problemstring")
	}()

	assert.Equal(t, "problemstring", err.Error())
	assertErrorIsCat(t, err)

	// When using a boolean + non-string type, the problem is wrapped as a general untyped
	// error, but this case should not be used in practice.
	func() {
		defer cat.Guard(&err)
		cat.Catch(false, 123)
		cat.Catch(true, 456)
	}()

	assert.Equal(t, "456", err.Error())
	assertErrorIsCat(t, err)
}

// If a catch condition is not an error or a boolean, then Catch will wrap it into
// ErrBadCatch.
func TestInvalidCatch(t *testing.T) {
	var err error
	func() {
		defer cat.Guard(&err)
		cat.Catch(123)
	}()

	assert.ErrorIs(t, err, cat.ErrBadCatch)
	assertErrorIsCat(t, err)
}
