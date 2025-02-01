/*
Request Guard - Errorcat Without Context Example

This example demonstrates a top-level error guard covering requests made by the user. Many
conventions can be made for error types and management, this is one example.

Each request has a guard at the base to catch and format errors for logging and
presentation. The error types dictate what action happens when they're caught.
*/
package main

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	cat "go.mukunda.com/errorcat"
)

// This marks an error as "OK", it's not an error, it's a control flow.
var ErrOkay = errors.New("okay")

// Control flow to exit.
var ErrExit = fmt.Errorf("%w: signal to exit", ErrOkay)

// This marks errors as "bad requests", where the user has done something wrong. The user
// should be told what's wrong, and it is not an internal error.
var ErrBadRequest = errors.New("bad request")

// Other errors will be treated as internal errors. What is shown to the user in this case
// varies by application, but for this demo, we won't show anything.

// Wrap a reason in a bad request error. The reason should be displayed to the user.
func badRequest(reason string) error {
	return fmt.Errorf("%w: %s", ErrBadRequest, reason)
}

// Log an error message to stderr. In this example, this is text that should not be
// displayed to the user.
func logError(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	for _, line := range strings.Split(message, "\n") {
		log.Printf("[ERROR] %s\n", line)
	}
}

// Execute a single request.
func startRequest() error {
	fmt.Print(`[Request Menu]
(1) Divide by zero
(2) Panic with a plain string
(3) Panic with an error
(4) Hello
(5) Catch an error
(6) Exit the program
>> Enter a request: `)

	var op int

	_, err := fmt.Scan(&op)
	// Catch can take the error directly, but `err != nil` coerces it into a boolean. When
	// it's a boolean, it will not be included in the formatted error message. This would
	// be a common scenario when you know what went wrong and don't want to preserve the
	// error for logging or displaying.
	cat.Catch(err != nil, badRequest("Invalid input. Try again."))

	switch op {
	case 1:
		// Panic with division by zero.
		a, b := 1, 0
		fmt.Printf("Division by zero: %d", a/b)
	case 2:
		// Direct panic with string.
		panic("this is a plain string panic")
	case 3:
		// Direct panic with arbitrary error type.
		panic(fmt.Errorf("this is an error panic"))
	case 4:
		// No error.
		fmt.Println("<< Hello!")
		time.Sleep(time.Second)
	case 5:
		// Errorcat catch (nearly the same as error panic).
		cat.Catch(fmt.Errorf("this is a caught error"), "with optional note")
	case 6:
		// Control flow.
		fmt.Println("<< Okay, will exit!")
		return ErrExit
	default:
		// Returning an error. Errors that are returned will be caught in the same place as
		// if they are Caught (as the upper function is returning it as-is). Returning can
		// be more performant though, as it is avoiding a panic and recover.
		return badRequest("Invalid operation number. Try again!")
	}

	return nil
}

// Errors are forwarded to here for handling. This callback is provided to `Guard` as an
// annotator.
func handleError(err error) error {
	if errors.Is(err, ErrOkay) {
		// No error, pass it through.
		return err
	}

	if errors.Is(err, ErrBadRequest) {
		// For "bad requests", show the error to the user, the error will contain a reason
		// why their input was invalid.
		//
		// Similar constructs can be made for other error types such as "forbidden",
		// "throttled", etc.
		fmt.Printf("Error: %v\n", err)

		return nil // The error is handled and will not be forwarded.
	}

	// For anything else, we log it internally and then show the user a general failure
	// message. It's normal for server software to not give much information to users for
	// security reasons
	//
	// If there -is- something the user can do to fix it themselves, you should figure out
	// a way to present that to them. Errors that the user can action can be wrapped in a
	// specific construct to contain user instructions.
	fmt.Println("Oops! An internal error occurred!")

	// Log the stack trace internally.
	logError("An error occurred: %v\ntrace: %s\n", err, string(debug.Stack()))

	return nil // The error is handled and will not be forwarded.
}

func main() {
	for {
		// Starting a new request. We aren't using the errorcat context. The errorcat
		// context is for library code, and it's okay to have a "global" guard here.
		result := cat.Guard(func(cat.Context) error {
			return startRequest()
		}, handleError)

		fmt.Print("Request finished!\n\n")
		time.Sleep(time.Second)

		if result != nil {
			if errors.Is(result, ErrExit) {
				fmt.Println("Goodbye.")
				break
			} else {
				// This never happens, but if it did, it would
				// be a bug, and it should be logged.
				logError("Unhandled request result: %v", result)
			}
		}
	}
}
