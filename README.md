# gosnip: run small snippets of Go code from the command line

[![GoDoc](https://godoc.org/github.com/benhoyt/gosnip?status.png)](https://godoc.org/github.com/benhoyt/gosnip)
[![TravisCI Build](https://travis-ci.org/benhoyt/gosnip.svg)](https://travis-ci.org/benhoyt/gosnip)
[![AppVeyor Build](https://ci.appveyor.com/api/projects/status/github/benhoyt/gosnip?branch=master&svg=true)](https://ci.appveyor.com/project/benhoyt/gosnip)

Package gosnip is a tool that allows you to run small snippets of
Go code from the command line.

    usage: gosnip [-d] [-i import ...] statements...

For simple uses, just specify one or more Go statements on the
command line, and gosnip will roll them into a full Go program and
run the result using "go run". Standard library imports are added
automatically. Some examples:

    $ gosnip 'fmt.Println("Hello world")'
    Hello world

    $ gosnip 'fmt.Println("Current time:")' 'fmt.Println(time.Now())'
    Current time:
    2018-11-24 16:18:47.101951 -0500 EST m=+0.000419239

The -i flag specifies an explicit import, which is needed for
non-stdlib imports or to disambiguate between ambiguous stdlib
imports such as "text/template" and "html/template" (multiple -i
flags are allowed). For example, -i is needed for "rand" because
the standard library has math/rand and crypto/rand):

    $ gosnip -i math/rand 'fmt.Println(rand.Intn(100))'
    81

The -d flag turns on debug mode, which prints the full program on
stderr before running it. For example:

    $ gosnip -d 'fmt.Println(time.Now())'
    package main

    import (
        "fmt"
        "time"
    )

    func main() {
        fmt.Println(time.Now())
    }
    2018-11-24 16:33:56.681024 -0500 EST m=+0.000383308

The gosnip command-line tool is a thin wrapper around the
"sniplib" package. To run Go snippets in your Go programs, see the
[sniplib docs](https://godoc.org/github.com/benhoyt/gosnip/sniplib).

## License

gosnip is licensed under an open source [MIT license](https://github.com/benhoyt/gosnip/blob/master/LICENSE.txt).

## Contact me

Have fun, and please [contact me](https://benhoyt.com/) if you're using gosnip or have any feedback!
