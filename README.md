## Errorcat

#### Go programmers HATE it!

*Use This One Weird Trick To Clean Up Your Go Code!*

This package introduces error handling concepts using panic and recover. The *panic-pattern*
I call it. The *anti-pattern* others might call it, but panics are a super convenient way
to handle *exceptional* conditions in your code.

Take this code for example:

```
import (
   "io"
   "os"
)

func writeLine(w io.Writer, text string) error {
   _, err := w.Write([]byte(text))
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

   if err := writeLine(f, "hello world"); err != nil {
      return err
   }

   if err := writeLine(f, "goodbye world"); err != nil {
      return err
   }

   if err := writeLine(f, "line number three"); err != nil {
      return err
   }

   return nil
}
```

I/O has a lot of error conditions that you can't do much about. Here is using Errorcat:

```
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

   // Telling the user what file operation failed
   cat.Catch(err, "failed opening config file") 

   writeLine(f, "hello world")
   writeLine(f, "goodbye world")
   writeLine(f, "line number three")

   return nil
}
```

The code instantly becomes cleaner without clutter of error-passing boilerplate, and it's
easier to better describe errors.

Errorcat is meant to handle errors that you don't expect to recover from. It's not meant
for common errors that affect execution paths - I would say that those should still be
handled manually. Using Errorcat for exceptional conditions comes with a number of
benefits:

* Centralized handling of errors.
* Easy to log errors that are caught.
* Easy annotation of errors to increase verbosity.
* Stack traces that can also be easily logged.

A major importance of error annotation is to assist engineers in debugging production
issues. Since stack traces are accessible from panic recovery, you can easily log the
stack trace when handling uncommon errors, providing a powerful point of information to
find the source of errors.

Sure, panics are more costly than passing around errors, but I'm sure you don't need a
lesson on what is truly important in software engineering, and where "performance" lies
with most application priorities.

### Documentation

https://pkg.go.dev/go.mukunda.com/errorcat

Errorcat is licensed under MIT.