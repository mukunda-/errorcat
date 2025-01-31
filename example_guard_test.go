package errorcat_test

import (
	"fmt"
	"os"

	"go.mukunda.com/errorcat"
)

// Example of creating a guard that catches all errors within.
func ExampleGuard() {

	err := errorcat.Guard(func(cat errorcat.Context) error {
		f, err := os.Open("nonexistant-config-file.txt")
		cat.Catch(err, "couldn't open configuration file")
		defer f.Close()
		fmt.Println("opened file!")
		return nil
	})

	fmt.Println(err)
	// Output:
	// couldn't open configuration file: open nonexistant-config-file.txt: The system cannot find the file specified.
}
