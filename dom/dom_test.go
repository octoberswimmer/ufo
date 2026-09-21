package dom

import (
	"fmt"
	"testing"
)

func TestParseXML_builds_a_namespaced_tree(t *testing.T) {
	doc, err := ParseXMLString(`<?xml version="1.0"?>
<?xml-stylesheet href="a.css" type="text/css"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:x="urn:x"><body class="main" x:role="r">Hello&nbsp;<b>world</b><![CDATA[ <raw> ]]><!-- note --><x:widget/></body></html>`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if pi, ok := doc.GetFirstChild().(*ProcessingInstruction); !ok || pi.GetTarget() != "xml-stylesheet" || pi.GetData() != `href="a.css" type="text/css"` {
		t.Errorf("first child = %#v, want the xml-stylesheet processing instruction", doc.GetFirstChild())
	}
	html := doc.GetDocumentElement()
	if html.GetLocalName() != "html" || html.GetNamespaceURI() != XHTMLNamespace {
		t.Errorf("root = {%s}%s", html.GetNamespaceURI(), html.GetLocalName())
	}
	body := html.GetFirstChild().(*Element)
	if body.GetAttribute("class") != "main" || body.GetAttributeNS("urn:x", "role") != "r" || body.GetAttribute("x:role") != "r" {
		t.Errorf("attributes = %#v", body.GetAttributes())
	}
	if body.HasAttribute("missing") || body.GetAttribute("missing") != "" {
		t.Error("a missing attribute reads as absent and empty")
	}
	if got, want := body.GetTextContent(), "Hello world <raw> "; got != want {
		t.Errorf("text content = %q, want %q", got, want)
	}
	var kinds []NodeType
	for _, child := range body.GetChildNodes() {
		kinds = append(kinds, child.GetNodeType())
	}
	// Text, <b>, the CDATA section as its own node, the comment, <x:widget>.
	if want := []NodeType{TextNode, ElementNode, CDATASectionNode, CommentNode, ElementNode}; fmt.Sprint(kinds) != fmt.Sprint(want) {
		t.Errorf("body children = %v, want %v", kinds, want)
	}
	widget := body.GetLastChild().(*Element)
	if widget.GetTagName() != "x:widget" || widget.GetLocalName() != "widget" || widget.GetNamespaceURI() != "urn:x" || widget.GetPrefix() != "x" {
		t.Errorf("widget = %s / %s / %s", widget.GetTagName(), widget.GetLocalName(), widget.GetNamespaceURI())
	}
	if widget.GetPreviousSibling().GetNodeType() != CommentNode || widget.GetNextSibling() != nil {
		t.Error("sibling navigation is wrong around the last child")
	}
	if widget.GetParentNode() != Node(body) || widget.GetOwnerDocument() != doc {
		t.Error("parent or owner document is wrong")
	}
}

func TestParseXML_rejects_markup_that_is_not_well_formed(t *testing.T) {
	for _, source := range []string{`<a><b></a>`, `<a>`, ``, `<a></a><b></b></c>`, `<a/><b/>`, `<a/>text`} {
		if _, err := ParseXMLString(source); err == nil {
			t.Errorf("expected %q to be rejected", source)
		}
	}
}

func TestParseHTML_accepts_markup_that_is_not_xml(t *testing.T) {
	doc, err := ParseHTMLString(`<!DOCTYPE html><title>T</title><p class=intro>One<br>Two<p>Three`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	paragraphs := doc.GetElementsByTagName("p")
	if len(paragraphs) != 2 || paragraphs[0].GetAttribute("class") != "intro" || paragraphs[0].GetNamespaceURI() != XHTMLNamespace {
		t.Fatalf("paragraphs = %v", paragraphs)
	}
	if got := paragraphs[0].GetTextContent(); got != "OneTwo" {
		t.Errorf("first paragraph text = %q", got)
	}
	if len(doc.GetElementsByTagName("br")) != 1 || len(doc.GetElementsByTagName("body")) != 1 {
		t.Error("expected the HTML parser to supply body and keep br")
	}
}

func TestTreeMutation(t *testing.T) {
	doc := NewDocument()
	root := doc.CreateElement("root")
	doc.AppendChild(root)
	a, b, c := doc.CreateElement("a"), doc.CreateElement("b"), doc.CreateElement("c")
	root.AppendChild(a)
	root.AppendChild(c)
	root.InsertBefore(b, c)
	if names := childNames(root); names != "a b c" {
		t.Errorf("children = %s", names)
	}
	// Appending a node that already has a parent moves it.
	root.AppendChild(a)
	if names := childNames(root); names != "b c a" {
		t.Errorf("children after move = %s", names)
	}
	root.RemoveChild(c)
	if names := childNames(root); names != "b a" || c.GetParentNode() != nil {
		t.Errorf("children after remove = %s", names)
	}
	a.SetAttribute("id", "x")
	if doc.GetElementByID("x") != a {
		t.Error("GetElementByID did not find the element")
	}
}

func childNames(n Node) string {
	names := ""
	for i, child := range n.GetChildNodes() {
		if i > 0 {
			names += " "
		}
		names += child.GetNodeName()
	}
	return names
}
