package errorcat_test

import (
	"fmt"
	"io"

	cat "go.mukunda.com/errorcat"
)

// For this example code, any error returned from this function is annotated with a
// specific type. The annotation is not mandatory, but it can help with verbosity.
func writeToImaginaryFile(success bool) (rerr error) {
	defer cat.Recover(&rerr, ErrWriteImaginaryFileFailed)

	var err error
	if success {
		err = nil
	} else {
		// If success is false, simulate an error.
		err = io.ErrClosedPipe
	}

	// Catch will terminate execution here if the err is not nil.
	cat.Catch(err, "couldn't write to file")

	return nil
}

// Example of catching an annotated error from within a function.
func ExampleCatch() {
	err := writeToImaginaryFile(true)
	fmt.Println(err)

	err = writeToImaginaryFile(false)
	fmt.Println(err)
	// Output:
	// <nil>
	// write to imaginary file failed: couldn't write to file: io: read/write on closed pipe
}
