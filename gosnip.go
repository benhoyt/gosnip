// Package gosnip is a tool that allows you to run small snippets of
// Go code from the command line.
//
// TODO: more details
//
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/benhoyt/gosnip/sniplib"
)

func main() {
	var imports multiString
	flag.Var(&imports, "i", "import `package` explicitly; multiple -i flags allowed\n(usually used for non-stdlib packages)")
	debug := flag.Bool("d", false, "debug mode (print full source to stderr)")
	flag.Parse()
	args := flag.Args()

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
