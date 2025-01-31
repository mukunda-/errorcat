package errorcat_test

import (
	"errors"
	"fmt"

	cat "go.mukunda.com/errorcat"
)

func specialFunction(failFast bool) (rerr error) {
	defer cat.Recover(&rerr, "specialFunction failed")

	// When Catch is called with a boolean, it will break execution with the given reason
	// when the condition is true.
	shouldWeFail := failFast
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
func ExampleRecover() {
	err := specialFunction(true)
	fmt.Println(err)

	err = specialFunction(false)
	fmt.Println(err)
	// Output:
	// specialFunction failed: bad condition
	// specialFunction failed: found myerr: an error state
}
