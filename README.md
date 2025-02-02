## Errorcat

#### Go programmers HATE it!

*Use This One Weird Trick To Clean Up Your Go Code!*

This package introduces "exception" handling using panic and recover. Panics are a super
convenient way to handle *exceptional* conditions in your code, and this package
encourages broadening the scope from fatal errors to all *poor execution conditions*.

The idea isn't unheard of in practice. For example, re-read [Defer, Panic, and
Recover](https://go.dev/blog/defer-panic-and-recover) from 2010. It mentions that the
standard library (citing the [json package](https://go.dev/pkg/encoding/json/)) uses panic
and recover internally to pass errors around before returning them.

Examples:
 * https://go.dev/src/encoding/json/encode.go
 * https://go.dev/src/fmt/scan.go

Using Errorcat employs the same pattern where you can pass certain error cases around
easily. Take this code for example:

```
import (
	"io"
	"os"
)

func writeLine(w io.Writer, text string) error {
	_, err := w.Write([]byte(text + "\n"))
	if err != nil {
		return err
	}
	return nil
}

func MyFunction() error {
	f, err := os.Open("file.txt")
	if err != nil {
		return err
	}

	if err := writeLine(f, "Hallo welt!"); err != nil {
		return err
	}

	if err := writeLine(f, "Goodbye!"); err != nil {
		return err
	}

	if err := writeLine(f, "Level 3"); err != nil {
		return err
	}

	return nil
}
```

I/O has a lot of annoying error conditions that you can't do much about. Here is using
Errorcat:

```
import (
	"fmt"
	"io"
	"os"

	cat "go.mukunda.com/errorcat"
)

func writeLine(w io.Writer, text string) {
	_, err := w.Write([]byte(text + "\n"))
	cat.Catch(err, "failed writing to file")
}

func MyFunction() (rerr error) {
	return cat.Guard(func(_ cat.Context) error {

		f, err := os.Open("file.txt")
		cat.Catch(err, "failed opening config file") // Annotated error reason.

		writeLine(f, "Hallo welt!")
		writeLine(f, "Goodbye!")
		writeLine(f, "Level 3")

		return nil

	}, "myfunction failed")
}
```

The code is instantly cleansed of error-passing clutter, and it's also easier to annotate
errors and get stack traces. Many functions don't need to return an error to the caller.
Instead, they jump to the recovery point, so your logic can focus on actionable errors
only.

Errorcat is meant to handle errors that you don't expect to recover from. It's not meant
for common errors that often affect execution paths. Those should still be handled the
*normal* way, i.e., returned and checked. When writing code with Errorcat, you can filter
out rare errors, "catching" them with Errorcat, and then only return and check errors that
are of interest to your application.

While you can have nested recovery points with Errorcat, it's more meant to have only one
recovery point at the start of a *request*. This is unlike traditional try-catch where
blocks are nested freely. Go still has the normal error practice for normal conditions.
Using Errorcat for exceptional conditions comes with a number of benefits:

* Reducing error-passing noise in your codebase.
* Centralized handling of rare errors.
* Enforces consistency with exceptional error handling. Unrecoverable errors usually
  require intervention, so it's critical that they are logged properly.
* Easy annotation of errors to increase verbosity to assist debugging.
* Stack traces of failures can also be logged easily, something that is often lost with
  the normal Go error patterns if you aren't careful.

Errorcat eases the burden of error annotation, letting you add more details to errors when
they are thrown and when they are recovered from. The annotation greatly assists engineers
in debugging production issues via log files. In addition, since stack traces are
accessible from panic recovery, you have access to the full stack trace of the error
source when handling uncommon errors.

Sure, panics are more costly than passing around errors, but I'm sure you don't need a
lesson on where "performance" lies with most real application priorities.

## Usage

There are two ways to use Errorcat, with and without *context*. Context is optional and
helps to avoid programmer errors when you do not or can not have a global panic guard.

### Errorcat Without Context

This is more useful for application-level code, where you don't need to be careful about
panics leaking as you have a central panic-handler to catch all error states. For example
in a middleware function for a server, you could have a recover handler there which
translates errors into server responses.

First you set up a recovery point, like so:

	func OnRequest(req Request) error {
		return cat.Guard(func(_ cat.Context) error {

			return handleRequest(req)
		
		}, "request failed")
	}

When handling errors in your subfunctions, you use cat.Catch. I like to shorten the
package name to `cat` for brevity.

	func handleRequest() {
		err := someLibraryFunction()
		cat.Catch(err, "someLibraryFunction didn't work")
	}

If it catches an error, it will bubble to the recovery point and annotate it with the 
messages provided, e.g., `request failed: someLibraryFunction didn't work: (error text)`.
Simple, right? A bulk of error handling is just that, annotating the error for the user.

What's more useful for HTTP servers is wrapping an error with an HTTP status. For
example:

	func handlePostUser(user string) {
		cat.Catch(user == "", BadRequest("user cannot be empty"))
	}

`BadRequest` isn't a provided function, but it's easy enough to implement yourself. In the
recover area, you would check if the error is a "bad request" type, and then extract the
message to display to the user or API consumer. You can also create a higher level package
that provides more flavor for your errors directly. For example:

	func handlePostUser(user string) {
		mycat.BadIf(user == "", "user cannot be empty")
	}

But be mindful that any control flow conventions you introduce will be foreign to new
developers to your project.

### Errorcat With Context

Okay, now pretend you're writing a library, where you *cannot* have panics accidentally
leaking past your package code. How do you ensure that 100%? When writing library code,
you don't have the convenience of a central recovery area. Each function that uses the
error-panic pattern must recover on its own.

If you guard every entry point properly, no context is fine, but the context pattern makes
it programmatically impossible to miss setting up a guard.

With an Errorcat context, you use the context object to throw errors rather than the
global functions. That way, you *know* that you are within a guarded context when calling
Catch (otherwise, there is no object with which to call Catch!).

	func MyLibraryFunction() error {
		return errorcat.Guard(func(cat errorcat.Context) error {

			someSubfunction(cat)
			return nil

		}, "mylibraryfunction failed")
	}

When calling subfunctions, you will *know* that you need to guard the upper function if it
requires a context to be passed in. Without context, it's not easy to tell which of your
library functions you need to guard, and the pattern could spread needlessly.

For code that already has context passing, you can merge the two contexts together, so
long as it implements the errorcat.Context interface. A basic wrapper will do.

errorcat.Context will cause real panics if it detects misuse of the guard. For example, if
you call Catch after recover was called, it will panic, indicating that you forgot to
defer the recover or you're incorrectly reusing a context. Misuse should not be possible
if you use the `Guard` function instead of deferring a `Recover` directly.

### Crossing goroutines

As you know, panics are troublesome when it comes to goroutines. The original recovery
point can't be shared, as the goroutine's stack is separated.

A convenience "Go" function is provided to call a goroutine with a new protected context,
returning the error result in a channel.

The main thing you must avoid is passing an Errorcat context between goroutines. You just
can't do that.

### Panicking safely

Using the `Guard` function is recommended over deferring a `Recover` yourself. If you do
use `Recover` directly, I recommend having a linter rule to make sure that you never
forget a deferred recover. It is easy enough to forget to call `Recover`, just as it is
easy to forget the `defer` keyword, both of which will silently cause hidden fatal panics
later on.

## Additional Details

Package documentation: https://pkg.go.dev/go.mukunda.com/errorcat

Errorcat is licensed under MIT.
