// errorcat - error catching utilities
// (C) 2025 Mukunda Johnson (mukunda.com)

// This package provides error handling and propagation utilities. Errors can be wrapped
// or annotated easily. Panics are used to propagate them and reduce error passing
// boilerplate in the code. Recovery scopes are used to catch and handle errors.
//
// The package name is errorcat, but aliasing it to something shorter like `cat` is
// encouraged for convenience.
//
// The basic usage pattern for Errorcat is like so:
//
//	func MyFunction() (rerr error) {
//	   // Guard this function from panicking past this point.
//	   defer cat.Guard(&rerr, "myfunction failed")
//
//	   err := someOtherFunction()
//	   cat.Catch(err, "someOtherFunction didn't work")
//
//	   return nil
//	}
//
// And if, for example, someOtherFunction used [Catch] to throw errors instead of
// returning them, then you can omit the error checking in the upper level.
//
//	func MyFunction() (rerr error) {
//	   defer cat.Guard(&rerr, "myfunction failed")
//
//	   someOtherFunction()
//
//	   return nil
//	}
//
// And then, if you don't need to stop panics at your given function, for example, if you
// are writing app-level code, then you don't need to guard there – you can guard in a
// central location, e.g., server at an upper level to log errors and continue.
//
//	func MyFunction() {
//	   someOtherFunction()
//	}
//
//	// Elsewhere...
//	func GuardMiddleware(next func()) {
//	   defer cat.Guard(nil, func(err error) error {
//	      // Log the error.
//	      log.Println(err)
//	      return err
//	   })
//
//	   next()
//	}
package errorcat

import (
	"errors"
	"fmt"
)

// This type implements the error interface and wraps any error originating from Catch.
type CatError struct {
	err error
}

// Read the error message.
func (e CatError) Error() string {
	return e.err.Error()
}

// Get the wrapped catch error.
func (e CatError) Unwrap() error {
	return e.err
}

// An annotator accepts a caught error and transforms it.
type Annotator = func(err error) error

// Callback for Scope.
type ScopedFunction = func() error

// This is a general error condition that is ideally not used. It is propagated when a
// `problem` is not provided for a boolean Catch condition.
var ErrUnknown = errors.New("unknown error")

// This error is used when the catch function is called with invalid arguments.
var ErrBadCatch = errors.New("bad catch usage")

// This is used to set a function boundary for recovery. As you know, panics are taboo if
// you are propagating them to consumers. When writing code to be used by others, you
// should always guard your public functions to not panic outside of your scope.
//
// This function captures any panics into the given error reference for returning.
//
// If `annonate` arguments are given, the error is annotated with each one. These can be
// strings, errors, or a callback Annotator function. Annotator functions also act as
// error handlers, to log or transform the error into a service response. Returning nil
// from a handler will prevent further annotators in the chain from being used.
func Guard(rerr *error, annotate ...any) {
	// Recover from panic and capture the error.
	var captured error
	if rerr != nil {
		captured = *rerr
	}

	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			captured = e
		} else {
			captured = fmt.Errorf("%v", r)
		}
	}

	// Annotate the error.
	if captured != nil {
		for _, annotator := range annotate {
			switch a := annotator.(type) {
			case Annotator:
				captured = a(captured)
			case error:
				captured = fmt.Errorf("%w: %w", a, captured)
			case string:
				captured = fmt.Errorf("%s: %w", a, captured)
			default:
				// Unknown!
				captured = fmt.Errorf("%v: %w", a, captured)
			}

			if captured == nil {
				// Break the chain if it was handled by an annotator.
				break
			}
		}
	}

	if rerr != nil {
		*rerr = captured
	}
}

// Execute the given function with a guarded scope. Any errors that are captured will be
// returned. `annotate` parameters can be used the same way as in `Guard`.
func Scope(fn ScopedFunction, annotate ...any) (rerr error) {
	defer Guard(&rerr, annotate...)
	return fn()
}

// Catch is for catching errors. In other words, it is "panic on error condition". The
// panic is recovered from by [Guard].
//
// `condition` is the condition to trigger an error state; it can be a boolean or error.
// `problem` is a description of the error.
//
// `problem` can be a string or another error. When `condition` is an error, the
// propagated error will contain both the condition and the problem. When `condition` is a
// boolean, the propagated error will contain only the problem.
//
// If the `problem` is a string, it will be wrapped into an anonymous error type.
// `problem` is optional, but it is bad practice to not provide a problem if the condition
// is not an error.
func Catch(condition any, problem ...any) {
	if condition == nil {
		return
	}

	var problem1 any
	if len(problem) > 0 {
		problem1 = problem[0]
	}

	switch cond := condition.(type) {
	case error:
		if cond != nil {
			switch p := problem1.(type) {
			case error:
				// Annotate condition with problem.
				// Wrap both errors.
				panic(CatError{fmt.Errorf("%w: %w", p, cond)})
			case nil:
				// Bubble error condition without annotation.
				panic(CatError{cond})
			default:
				// Annotate condition with problem.
				panic(CatError{fmt.Errorf("%v: %w", p, cond)})
			}
		}

	case bool:
		if cond {
			switch p := problem1.(type) {
			case error:
				// Wrap the given error.
				panic(CatError{fmt.Errorf("%w", p)})
			case nil:
				// Bad practice. A problem should be specified.
				panic(CatError{ErrUnknown})
			default:
				// Create a general error.
				panic(CatError{fmt.Errorf("%v", p)})
			}
		}

	default:
		panic(CatError{fmt.Errorf("%w: unknown catch condition type: %v", ErrBadCatch, condition)})
	}
}
