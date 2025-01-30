package errorcat_test

import (
	"fmt"
	"os"

	cat "go.mukunda.com/errorcat"
)

// Example of creating a scope that catches all errors within.
func ExampleScope() {

	err := cat.Scope(func() error {
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
