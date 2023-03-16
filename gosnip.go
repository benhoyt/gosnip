// Package gosnip is a tool that allows you to run small snippets of
// Go code from the command line.
//
//	usage: gosnip [-d] [-i import ...] statements...
//
// For simple uses, just specify one or more Go statements on the
// command line, and gosnip will roll them into a full Go program and
// run the result using "go run". Standard library imports and any
// imports needed for packages in GOPATH are added automatically
// (using the same logic as the "goimports" tool). Some examples:
//
//	$ gosnip 'fmt.Println("Hello world")'
//	Hello world
//
//	$ gosnip 'fmt.Println("Current time:")' 'fmt.Println(time.Now())'
//	Current time:
//	2018-11-24 16:18:47.101951 -0500 EST m=+0.000419239
//
// The -i flag allows you to specify an import explicitly, which may be
// needed to select between ambiguous stdlib imports such as
// "text/template" and "html/template" (multiple -i flags are
// allowed). For example:
//
//	$ ./gosnip -i text/template 't, _ := template.New("w").Parse("{{ . }}\n")' \
//	                            't.Execute(os.Stdout, "<b>")'
//	<b>
//	$ ./gosnip -i html/template 't, _ := template.New("w").Parse("{{ . }}\n")' \
//	                            't.Execute(os.Stdout, "<b>")'
//	&lt;b&gt;
//
// The -d flag turns on debug mode, which prints the full program on
// stderr before running it. For example:
//
//	$ gosnip -d 'fmt.Println(time.Now())'
//	package main
//
//	import (
//	    "fmt"
//	    "time"
//	)
//
//	func main() {
//	    fmt.Println(time.Now())
//	}
//	2018-11-24 16:33:56.681024 -0500 EST m=+0.000383308
//
// The gosnip command-line tool is a thin wrapper around the
// "sniplib" package. To run Go snippets in your Go programs, see the
// sniplib docs.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/benhoyt/gosnip/sniplib"
)

const (
	version = "v1.1.1"
)

func main() {
	var imports multiString
	flag.Var(&imports, "i", "import `package` explicitly; multiple -i flags allowed\n(usually used for non-stdlib packages)")
	debug := flag.Bool("d", false, "debug mode (print full source to stderr)")
	showVersion := flag.Bool("version", false, "show gosnip version and exit")
	flag.Parse()
	args := flag.Args()

	if *showVersion {
		fmt.Printf("gosnip %s - Copyright (c) 2018 Ben Hoyt\n", version)
		return
	}

	if len(args) < 1 {
		errorExit("usage: gosnip [-d] [-i import ...] statements...\n")
	}

	source, err := sniplib.ToProgram(args, imports)
	if err != nil {
		errorExit("%v\n", err)
	}
	if *debug {
		fmt.Fprint(os.Stderr, source)
	}

	err = sniplib.Run(source, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		os.Exit(1)
	}
}

func errorExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

type multiString []string

func (m *multiString) String() string {
	return fmt.Sprintf("%v", []string(*m))
}

func (m *multiString) Set(value string) error {
	*m = append(*m, value)
	return nil
}
