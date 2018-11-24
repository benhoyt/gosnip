// Package sniplib converts Go code snippets to full programs and
// runs them.
package sniplib

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// ToProgram converts a slice of Go statements to a full Go program.
// Standard library imports are added automatically. The imports arg
// is a list of explicit imports for importing non-stdlib packages or
// for selecting between ambiguous stdlib import names like
// "text/template" and "html/template". Returns the formatted source
// text of the full Go program and any error that occurred.
func ToProgram(statements, imports []string) (string, error) {
	snippet := strings.Join(statements, "; ")
	fset, file, err := parse(snippet)
	if err != nil {
		return "", err
	}
	err = addImports(file, imports)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = printer.Fprint(buf, fset, file)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Parses a Go source snippet and returns the file set and AST.
func parse(snippet string) (*token.FileSet, *ast.File, error) {
	fset := token.NewFileSet()
	source := fmt.Sprintf(`
package main

func main() {
    %s
}
`, snippet)
	file, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		return nil, nil, err
	}
	return fset, file, nil
}

// Add stdlib imports or explicit imports necessary to resolve any
// unresolved names in the parsed file.
func addImports(file *ast.File, imports []string) error {
	importDecl := &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: 1, // made-up nonzero position for '(' so it's a multiline "import ( ... )"
	}
	// Add import declaration first in the declarations list
	file.Decls = append([]ast.Decl{importDecl}, file.Decls...)

	// Convert explicit imports list to a map of imported name to
	// full package name
	importsMap := make(map[string]string)
	for _, imp := range imports {
		parts := strings.Split(imp, "/")
		name := parts[len(parts)-1]
		if _, ok := importsMap[name]; ok {
			return fmt.Errorf("multiple %q packages specified", name)
		}
		importsMap[name] = imp
	}

	seen := make(map[string]bool)
	for _, u := range file.Unresolved {
		if builtins[u.Name] {
			// For whatever reason, builtins like "int" are considered
			// unresolved names in the AST
			continue
		}
		if seen[u.Name] {
			// We've already seen/handled this unresolved name
			continue
		}
		seen[u.Name] = true

		// First look it up in the explicit imports list, if it's not
		// there try the standard library
		importPath := importsMap[u.Name]
		if importPath == "" {
			switch len(stdlib[u.Name]) {
			case 0:
				return fmt.Errorf("undefined name %q, did you forget the -i flag?", u.Name)
			case 1:
				importPath = stdlib[u.Name][0]
			default:
				iflag := []string{}
				for _, p := range stdlib[u.Name] {
					iflag = append(iflag, fmt.Sprintf(`"-i %s"`, p))
				}
				iflagStr := strings.Join(iflag, " or ")
				return fmt.Errorf("multiple %q packages in stdlib, use flag %s", u.Name, iflagStr)
			}
		}

		// Add the import to the import declaration
		spec := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(importPath),
			},
		}
		importDecl.Specs = append(importDecl.Specs, spec)
	}
	return nil
}

// Run runs the given Go program source. Use the provided stdin
// reader and stdout/stderr writers for I/O.
func Run(source string, stdin io.Reader, stdout, stderr io.Writer) error {
	// First write source to a temporary file (deleted afterwards)
	file, err := ioutil.TempFile("", "gosnip_*.go")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = io.WriteString(file, source)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	// Then use "go run" to run it
	cmd := exec.Command("go", "run", file.Name())
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	errBuf := &bytes.Buffer{}
	cmd.Stderr = errBuf
	err = cmd.Run()
	if exitStatus(err) == 2 {
		// "go run" exit status 2 means compile error, so filter the
		// funky temp filename and extraneous "go run" comments
		filterStderr(errBuf.Bytes(), stderr)
		return err
	}
	_, _ = errBuf.WriteTo(stderr)
	return err
}

// Return exit status from given exec error (there'll be a better way
// to do this in Go 1.12!).
func exitStatus(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 1
}

// Filter out extraneous output and temp file name from "go run"
// output in case of compile error. Example "go run" output:
//
//     # command-line-arguments
//     /var/folders/sz/thh6m7316b3gvvvmjp8qpdrm0000gp/T/gosnip_615300750.go:8:2: undefined: fmt.X
//
// Output after filtering:
//
//     8:2: undefined: fmt.X
//
func filterStderr(data []byte, writer io.Writer) {
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("# ")) {
			continue
		}
		pos := bytes.Index(line, []byte(".go:"))
		if pos < 0 {
			continue
		}
		writer.Write(line[pos+4:])
		writer.Write([]byte("\n"))
	}
}
