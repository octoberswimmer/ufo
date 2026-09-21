// Command ufo renders an HTML or XHTML document to PDF:
//
//	ufo <input.html|url> <output.pdf>
//
// The input is parsed as XML when it is well-formed and with the HTML5 parser
// otherwise.
//
//	ufo pipe
//
// renders one document for another program over standard input and output,
// with that program supplying the resources the document refers to; see
// package github.com/octoberswimmer/ufo/pipe for the protocol.
//
//	ufo version
//
// prints the version the program was built as.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/pdf"
	"github.com/octoberswimmer/ufo/pipe"
)

// version is set when a release is built (see the Makefile).
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 1 && args[0] == "pipe" {
		return pipe.Serve(os.Stdin, os.Stdout)
	}
	if len(args) == 1 && (args[0] == "version" || args[0] == "--version") {
		fmt.Println(version)
		return nil
	}
	if len(args) != 2 {
		return fmt.Errorf("Usage: ufo <input.html|url> <output.pdf>\n       ufo pipe\n       ufo version")
	}
	input, output := args[0], args[1]

	renderer := pdf.NewITextRenderer()

	var source []byte
	baseURL := input
	if strings.Contains(input, "://") {
		source = renderer.GetSharedContext().GetUac().GetBinaryResource(input)
		if source == nil {
			return fmt.Errorf("can't read %s", input)
		}
	} else {
		var err error
		if source, err = os.ReadFile(input); err != nil {
			return err
		}
		abs, err := filepath.Abs(input)
		if err != nil {
			return err
		}
		baseURL = "file:" + filepath.ToSlash(abs)
	}

	if doc, err := dom.ParseXML(bytes.NewReader(source)); err == nil {
		if err := renderer.SetDocumentWithUrl(doc, baseURL); err != nil {
			return err
		}
	} else if err := renderer.SetDocumentFromHTMLString(string(source), baseURL); err != nil {
		return err
	}

	if err := renderer.Layout(); err != nil {
		return err
	}
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	if err := renderer.CreatePDF(out); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
