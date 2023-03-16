// Package sniplib converts Go code snippets to full programs and
// runs them.
package sniplib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	importspkg "golang.org/x/tools/imports"
)

// ToProgram converts a slice of Go statements to a full Go program.
// References to standard library functions and functions from
// packages in GOPATH are imported automatically (using the same
// logic as the "goimports" tool).
//
// The "imports" arg is an explicit list of imports, only needed for
// selecting between ambiguous stdlib import names like
// "text/template" and "html/template". Returns the formatted source
// text of the full Go program and any error that occurred.
func ToProgram(statements, imports []string) (string, error) {
	importStrs := make([]string, len(imports))
	for i, imp := range imports {
		importStrs[i] = fmt.Sprintf("import %q", imp)
	}
	source := fmt.Sprintf(`
package main

%s

func main() {
	%s
}`, strings.Join(importStrs, "\n"), strings.Join(statements, "; "))
	processed, err := importspkg.Process("", []byte(source), nil)
	if err != nil {
		return "", err
	}
	return string(processed), nil
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
	if err != nil {
		// Ideally we'd only do this filtering on compile error, not program
		// error, but it's hard to tell the difference ("go run" used to
		// return exit code 2 for this, but since Go 1.20, it doesn't).
		filterStderr(errBuf.Bytes(), stderr)
		return err
	}
	_, _ = errBuf.WriteTo(stderr)
	return err
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
	if !bytes.Contains(data, []byte("# command-line-arguments")) {
		writer.Write(data)
		return
	}
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
