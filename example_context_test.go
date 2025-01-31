package errorcat_test

import (
	"fmt"

	"go.mukunda.com/errorcat"
)

func unsafeFunction(ct errorcat.Context) {
	// You can add a type alias to reduce boilerplate of typing "errorcat.Context" in any
	// signatures that need it.

	// In here, we can use the context to catch errors.

	err := fmt.Errorf("this is an error")
	ct.Catch(err, "caught the err")

	fmt.Println("this won't be printed")
}

func ExampleContext() {
	err := func() (rerr error) {
		ct := errorcat.NewContext(&rerr)
		defer errorcat.Recover(ct, "test")
		// "defer Recover" must always follow context creation.

		unsafeFunction(ct)

		// The output will contain the error annotation plus the actual error text. We could
		// also annotate it further if we added one or more annotators in the Recover call.
		return nil
	}()
	fmt.Println(err)

	// An alternate way to create the context is to use the Guard function.
	err2 := errorcat.Guard(func(ct errorcat.Context) error {
		unsafeFunction(ct)
		return nil
	}, "test")
	fmt.Println(err2)

	// Output:
	// test: caught the err: this is an error
	// test: caught the err: this is an error
}
