// Tests for the port of flying-saucer-core/src/main/java/org/xhtmlrenderer/simple/NoNamespaceHandler.java.
// The Java suite has no test for this class.

package ufo

import (
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

func TestNoNamespaceHandler_getStylesheetsReadsXmlStylesheetProcessingInstructions(t *testing.T) {
	document, err := dom.ParseXMLString(`<?xml version="1.0"?>
<?xml-stylesheet type="text/css" href="print.css" media="print, Screen"?>
<?xml-stylesheet type = 'text/css' href = 'plain.css'?>
<?xml-stylesheet type="text/xsl" href="transform.xsl"?>
<?other type="text/css" href="other.css"?>
<doc/>`)
	if err != nil {
		t.Fatal(err)
	}
	stylesheets := NewNoNamespaceHandler().GetStylesheets(document)
	if len(stylesheets) != 2 {
		t.Fatalf("got %d stylesheets, want 2", len(stylesheets))
	}
	first := stylesheets[0]
	if first.GetUri() != "print.css" || first.GetOrigin() != StylesheetInfoOriginAuthor || first.GetContent() != nil {
		t.Errorf("first stylesheet = %v", first)
	}
	if media := first.GetMedia(); len(media) != 2 || media[0] != "print" || media[1] != "screen" {
		t.Errorf("media of the first stylesheet = %v", media)
	}
	second := stylesheets[1]
	if second.GetUri() != "plain.css" {
		t.Errorf("uri of the second stylesheet = %q", second.GetUri())
	}
	if media := second.GetMedia(); len(media) != 1 || media[0] != "screen" {
		t.Errorf("media of the second stylesheet = %v", media)
	}
}

func TestNoNamespaceHandler_getAttributeValue(t *testing.T) {
	document, err := dom.ParseXMLString(`<doc xmlns:x="urn:x" a="plain" x:b="qualified"/>`)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewNoNamespaceHandler()
	root := document.GetDocumentElement()
	noNamespace := TreeResolverNoNamespace
	namespace := "urn:x"
	cases := []struct {
		name string
		got  *string
		want string
	}{
		{"by name", handler.GetAttributeValue(root, "a"), "plain"},
		{"absent", handler.GetAttributeValue(root, "missing"), ""},
		{"no namespace", handler.GetAttributeValueWithNamespaceURI(root, &noNamespace, "a"), "plain"},
		{"any namespace", handler.GetAttributeValueWithNamespaceURI(root, nil, "b"), "qualified"},
		{"any namespace, absent", handler.GetAttributeValueWithNamespaceURI(root, nil, "missing"), ""},
		{"namespace", handler.GetAttributeValueWithNamespaceURI(root, &namespace, "b"), "qualified"},
	}
	for _, c := range cases {
		if c.got == nil || *c.got != c.want {
			t.Errorf("%s: got %v, want %q", c.name, c.got, c.want)
		}
	}
}
