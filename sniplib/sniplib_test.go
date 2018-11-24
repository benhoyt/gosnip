// Tests and examples for gosnip's sniplib package

package sniplib_test

import (
	"fmt"
	"os"

	"github.com/benhoyt/gosnip/sniplib"
)

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
