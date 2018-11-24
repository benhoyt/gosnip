// Tests and examples for gosnip's sniplib package

package sniplib_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/benhoyt/gosnip/sniplib"
)

func TestToProgram(t *testing.T) {
	tests := []struct {
		statements []string
		imports    []string
		output     string
	}{
		{
			[]string{`fmt.Println("Hello world")`},
			nil,
			`package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello world")
}
`},
		{
			[]string{`fmt.Println(time.Now())`},
			[]string{},
			`package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now())
}
`},
		{
			[]string{`fmt.Println("x"); fmt.Println(int(3.5))`},
			[]string{},
			`package main

import (
	"fmt"
)

func main() {
	fmt.Println("x")
	fmt.Println(int(3.5))
}
`},
		{
			[]string{`template.Must()`},
			[]string{"text/template"},
			`package main

import (
	"text/template"
)

func main() {
	template.Must()
}
`},
		{
			[]string{`foo.Bar()`},
			[]string{"github.com/user/foo"},
			`package main

import (
	"github.com/user/foo"
)

func main() {
	foo.Bar()
}
`},
		{
			[]string{`fmt.Println(`},
			[]string{},
			"ERROR: 6:1: expected operand, found '}'",
		},
		{
			[]string{`fmt.Println()`},
			[]string{"foo", "foo"},
			`ERROR: multiple "foo" packages specified`,
		},
		{
			[]string{`foo.Bar()`},
			nil,
			`ERROR: undefined name "foo", did you forget the -i flag?`,
		},
		{
			[]string{`template.Must()`},
			nil,
			`ERROR: multiple "template" packages in stdlib, use flag "-i html/template" or "-i text/template"`,
		},
	}
	for _, test := range tests {
		name := strings.Join(test.statements, "; ")
		t.Run(name, func(t *testing.T) {
			source, err := sniplib.ToProgram(test.statements, test.imports)
			if err != nil {
				errStr := "ERROR: " + err.Error()
				if errStr != test.output {
					t.Fatalf("expected %q, got %q", test.output, errStr)
				}
			} else {
				if source != test.output {
					t.Fatalf("expected:\n%sgot:\n%s", test.output, source)
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		source string
		stdout string
		stderr string
		err    string
	}{
		{
			`package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello world")
}
`,
			"Hello world\n",
			"",
			"",
		},
		{
			`package main

import (
	"fmt"
)

func main() {
	fmt.X()
}
`,
			"",
			"8:2: undefined: fmt.X\n",
			"exit status 2",
		},
		{
			`package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintf(os.Stderr, "a funky error\n")
	os.Exit(5)
}
`,
			"",
			"a funky error\nexit status 5\n",
			"exit status 1",
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			inBuf := &bytes.Buffer{}
			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			err := sniplib.Run(test.source, inBuf, outBuf, errBuf)
			if outBuf.String() != test.stdout {
				t.Errorf("expected stdout %q, got %q", test.stdout, outBuf.String())
			}
			if errBuf.String() != test.stderr {
				t.Errorf("expected stderr %q, got %q", test.stderr, errBuf.String())
			}
			if err != nil {
				if err.Error() != test.err {
					t.Errorf("expected error %q, got %q", test.err, err.Error())
				}
			}
		})
	}
}

func ExampleToProgram() {
	statements := []string{`fmt.Println("Hello world")`}
	source, err := sniplib.ToProgram(statements, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(source)
	// Output:
	// package main
	//
	// import (
	// 	"fmt"
	// )
	//
	// func main() {
	// 	fmt.Println("Hello world")
	// }
}

func ExampleToProgram_imports() {
	statements := []string{`fmt.Println(template.HTMLEscapeString("<b>"))`}
	imports := []string{"text/template"}
	source, err := sniplib.ToProgram(statements, imports)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(source)
	// Output:
	// package main
	//
	// import (
	// 	"fmt"
	// 	"text/template"
	// )
	//
	// func main() {
	// 	fmt.Println(template.HTMLEscapeString("<b>"))
	// }
}

func ExampleRun() {
	source := `
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello world")
}
`
	err := sniplib.Run(source, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Output:
	// Hello world
}
