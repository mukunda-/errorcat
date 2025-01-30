package errorcat_test

import (
	"fmt"
	"io"
	"os"

	cat "go.mukunda.com/errorcat"
)

func writeLine(w io.Writer, text string) {
	_, err := w.Write([]byte(text))
	cat.Catch(err, "failed writing to file")
}

func MyFunction() (rerr error) {
	defer cat.Guard(&rerr, "myfunction failed")

	f, err := os.Open("file.txt")
	cat.Catch(err, "failed opening config file") // Annotated error reason.

	writeLine(f, "hello world")
	writeLine(f, "goodbye world")
	writeLine(f, "line number three")

	return nil
}

func ExampleMyFunction() {
	err := MyFunction()
	fmt.Println(err)
	// Output:
	// myfunction failed: failed opening config file: open file.txt: The system cannot find the file specified.
}
