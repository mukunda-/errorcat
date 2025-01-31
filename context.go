// errorcat - error catching utilities
// (C) 2025 Mukunda Johnson (mukunda.com)

package errorcat

import "runtime"

/*
For library code, you need to ensure that you aren't passing panics past your package
boundaries. Errorcat contexts are for avoiding that. When writing library code, you should
be using a context to call [Catch] and not use [Catch] directly. The goal of the context
is to make it impossible to call [Catch] without being inside of a guarded area, as the
context originates from the top of the guard.

This interface can be mixed with other context interfaces to not clutter code that already
uses context passing.
*/
type Context interface {
	// This is called inside of Recover.
	OnRecover()

	// Wrapper for Catch.
	Catch(condition any, problem ...any)

	// Returns a reference to the top-level error that was captured when creating the
	// context.
	ErrorRef() *error
}

// Default context implementation.
type context struct {
	errorRef      *error
	recoverCalled bool
}

// A callback function issued when [Recover] is called.
func (c *context) OnRecover() {
	if c.recoverCalled {
		panic("[errorcat] Duplicate call to Recover")
	}
	c.recoverCalled = true
}

// Create a new guarded context. `defer Recover(...)` must be used on the created context,
// ideally directly afterward. It should be easy to search through all files to make sure
// that the convention is respected.
func NewContext(errorRef *error) Context {
	ct := &context{errorRef: errorRef}
	runtime.SetFinalizer(ct, func(c *context) {
		if !c.recoverCalled {
			// This could execute anywhere, so it's not really safe. Better to have the panic
			// bubble through which would be a more predictable crash.
			//
			// Is there a good way to inform the user of the problem without being too
			// invasive?

			// panic("[errorcat] Recover was not called for a context. Make sure that all created contexts have a deferred recover.")
		}
	})
	return ct
}

// Context-based wrapper for [Catch].
func (c *context) Catch(condition any, problem ...any) {
	if c.recoverCalled {
		// The user likely forgot to defer the recover. Additional catch calls should not be
		// made with the context after Recover is called.
		panic("[errorcat] Catch was called after recovery.")
	}
	Catch(condition, problem...)
}

// Returns a reference to the top-level error that was captured when creating this
// context. This can be nil.
func (c *context) ErrorRef() *error {
	return c.errorRef
}
