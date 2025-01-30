package errorcat_test

import (
	"errors"
	"fmt"

	cat "go.mukunda.com/errorcat"
)

func guardedFunction() (rerr error) {
	defer cat.Guard(&rerr, "guarded function failed")

	// When Catch is called with a boolean, it will break execution with the given reason
	// when the condition is true.
	shouldWeFail := true
	cat.Catch(shouldWeFail, "bad condition")

	// Catch can also be called with an error. The reason part is optional but helps to
	// describe where the error is from. Note that this is also combined with the
	// function-level annotation from cat.Guard.
	myerr := errors.New("an error state")
	cat.Catch(myerr, "found myerr")

	// If all goes well, Catch will let execution pass through.

	return nil
}

// Example of catching an annotated error from within a function.
func ExampleGuard() {
	err := guardedFunction()
	fmt.Println(err)
	// Output:
	// guarded function failed: bad condition
}
