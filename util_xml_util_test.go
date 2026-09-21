// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/util/XMLUtilTest.java

package ufo

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// In Java the external entity resolves to empty content and the text is
// " Hello". dom.ParseXML reads no entity declarations, so the reference to
// the entity is a parse error; either way the content of the file does not
// reach the document.
func TestXMLUtilDocumentFromString_doesNotResolveExternalEntities(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(secretFile, []byte("top-secret-content"), 0o644); err != nil {
		t.Fatal(err)
	}
	secretUri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(secretFile)}).String()

	xml := `<?xml version="1.0"?>
<!DOCTYPE root [
  <!ENTITY xxe SYSTEM "` + secretUri + `">
]>
<root>&xxe; Hello</root>
`
	document, err := XMLUtilDocumentFromString(xml)
	if err == nil {
		text := document.GetDocumentElement().GetTextContent()
		if strings.Contains(text, "top-secret-content") {
			t.Errorf("the external entity was resolved: %q", text)
		}
		if text != " Hello" {
			t.Errorf("text %q, want %q", text, " Hello")
		}
	}
}

func TestXMLUtilDocumentFromString_parsesRegularXml(t *testing.T) {
	document, err := XMLUtilDocumentFromString("<root><child>hello</child></root>")
	if err != nil {
		t.Fatal(err)
	}
	if text := document.GetDocumentElement().GetTextContent(); text != "hello" {
		t.Errorf("text %q", text)
	}
}

func TestXMLUtilDocumentFromString_stillResolvesXhtmlNamedEntitiesFromLocalDtd(t *testing.T) {
	xml := `<?xml version="1.0"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html><body><p>a&nbsp;b&quot;c</p></body></html>
`
	document, err := XMLUtilDocumentFromString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if text := document.GetDocumentElement().GetTextContent(); text != "a\u00a0b\"c" {
		t.Errorf("text %q", text)
	}
}

// Each &lolN; expands to 10 copies of the previous one, so &lol9; alone would
// expand to 10^9 "lol"s if fully resolved. This must be rejected instead of
// exhausting memory/CPU. The Java parser rejects it for the number of entity
// expansions; dom.ParseXML rejects the reference to a declared entity.
func TestXMLUtilDocumentFromString_rejectsBillionLaughsEntityExpansion(t *testing.T) {
	xml := `<?xml version="1.0"?>
<!DOCTYPE lolz [
 <!ENTITY lol "lol">
 <!ENTITY lol2 "&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;&lol;">
 <!ENTITY lol3 "&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;&lol2;">
 <!ENTITY lol4 "&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;&lol3;">
 <!ENTITY lol5 "&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;&lol4;">
 <!ENTITY lol6 "&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;&lol5;">
 <!ENTITY lol7 "&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;&lol6;">
 <!ENTITY lol8 "&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;&lol7;">
 <!ENTITY lol9 "&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;&lol8;">
]>
<lolz>&lol9;</lolz>
`
	if _, err := XMLUtilDocumentFromString(xml); err == nil {
		t.Errorf("the document was parsed")
	}
}

// The Java test newSecureDocumentBuilderFactory_deniesExternalDtdAccessAtTheJaxpLevel
// expects the parser to refuse an external DTD. The builder here never reads
// one: the document is parsed and the server that the DOCTYPE names gets no
// request.
func TestXMLUtilNewDocumentBuilder_doesNotFetchExternalDtd(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	xml := `<?xml version="1.0"?>
<!DOCTYPE root SYSTEM "` + server.URL + `/unreachable.dtd">
<root/>
`
	document, err := XMLUtilNewDocumentBuilder().ParseInputSource(InputSourcesFromString(xml))
	if err != nil {
		t.Fatal(err)
	}
	if name := document.GetDocumentElement().GetTagName(); name != "root" {
		t.Errorf("root element %q", name)
	}
	if n := requests.Load(); n != 0 {
		t.Errorf("the external DTD was requested %d times", n)
	}
}

// The tests below are not in the Java suite.

func TestXMLUtilDocumentFromFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "doc.xml")
	if err := os.WriteFile(file, []byte("<root>from file</root>"), 0o644); err != nil {
		t.Fatal(err)
	}
	document, err := XMLUtilDocumentFromFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if text := document.GetDocumentElement().GetTextContent(); text != "from file" {
		t.Errorf("text %q", text)
	}
	if _, err := XMLUtilDocumentFromFile(filepath.Join(t.TempDir(), "missing.xml")); err == nil {
		t.Errorf("no error for a missing file")
	}
}

func TestDocumentBuilderParseString(t *testing.T) {
	file := filepath.Join(t.TempDir(), "doc.xml")
	if err := os.WriteFile(file, []byte("<root>by uri</root>"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, uri := range []string{file, (&url.URL{Scheme: "file", Path: filepath.ToSlash(file)}).String()} {
		document, err := XMLUtilNewDocumentBuilder().ParseString(uri)
		if err != nil {
			t.Fatalf("%s: %v", uri, err)
		}
		if text := document.GetDocumentElement().GetTextContent(); text != "by uri" {
			t.Errorf("%s: text %q", uri, text)
		}
	}
}
