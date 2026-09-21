// Command html2xhtml parses an HTML document with the HTML5 parser ufo uses
// (dom.ParseHTML) and writes it as well-formed XHTML. The Java Flying Saucer
// build in reference/ reads only XML, so this is how the same document is
// given to both implementations when comparing them:
//
//	html2xhtml in.html > out.xhtml
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: html2xhtml <in.html>")
		os.Exit(2)
	}
	in, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	doc, err := dom.ParseHTML(in)
	in.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out := bufio.NewWriter(os.Stdout)
	writeNode(out, doc.GetDocumentElement(), true)
	out.WriteString("\n")
	out.Flush()
}

var escaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")

func writeNode(w *bufio.Writer, n dom.Node, root bool) {
	switch node := n.(type) {
	case *dom.Element:
		w.WriteString("<" + node.GetLocalName())
		if root {
			w.WriteString(` xmlns="` + dom.XHTMLNamespace + `"`)
		}
		for _, attr := range node.GetAttributes() {
			if strings.Contains(attr.Name, ":") || attr.Name == "xmlns" {
				continue
			}
			w.WriteString(" " + attr.Name + `="` + escaper.Replace(attr.Value) + `"`)
		}
		w.WriteString(">")
		for _, child := range node.GetChildNodes() {
			writeNode(w, child, false)
		}
		w.WriteString("</" + node.GetLocalName() + ">")
	case *dom.Text:
		if parent, ok := node.GetParentNode().(*dom.Element); ok && (parent.GetLocalName() == "style" || parent.GetLocalName() == "script") {
			w.WriteString("<![CDATA[" + strings.ReplaceAll(node.GetData(), "]]>", "]]]]><![CDATA[>") + "]]>")
			return
		}
		w.WriteString(escaper.Replace(node.GetData()))
	}
}
