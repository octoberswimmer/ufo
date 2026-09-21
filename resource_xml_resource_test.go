// Tests for the port of flying-saucer-core/src/main/java/org/xhtmlrenderer/resource/XMLResource.java.
// The Java suite has no test for this class.

package ufo

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const xmlResourceTestDocument = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>Hello</title></head><body><p>a&nbsp;b &copy;</p></body></html>`

func TestXMLResource_loadStringResolvesNamedEntities(t *testing.T) {
	resource := XMLResourceLoadString(xmlResourceTestDocument)
	document := resource.GetDocument()
	if name := document.GetDocumentElement().GetNodeName(); name != "html" {
		t.Errorf("root element = %q, want html", name)
	}
	paragraphs := document.GetElementsByTagName("p")
	if len(paragraphs) != 1 || paragraphs[0].GetTextContent() != "a b ©" {
		t.Errorf("paragraphs = %v", paragraphs)
	}
	if resource.GetResourceInputSource() == nil || resource.GetResourceInputSource().GetCharacterStream() == nil {
		t.Errorf("the resource does not carry the character stream it was loaded from")
	}
	if resource.GetResourceLoadTimeStamp() == 0 {
		t.Errorf("the resource has no load time stamp")
	}
	if resource.GetElapsedLoadTime() < 0 {
		t.Errorf("elapsed load time = %d", resource.GetElapsedLoadTime())
	}
}

func TestXMLResource_loadInputStreamAndInputSource(t *testing.T) {
	fromStream := XMLResourceLoadInputStream(strings.NewReader(xmlResourceTestDocument))
	if fromStream.GetResourceInputSource().GetByteStream() == nil {
		t.Errorf("the resource does not carry the byte stream it was loaded from")
	}
	fromSource := XMLResourceLoadInputSource(NewInputSourceInputStream(strings.NewReader(xmlResourceTestDocument)))
	for _, resource := range []*XMLResource{fromStream, fromSource} {
		if title := resource.GetDocument().GetElementsByTagName("title"); len(title) != 1 || title[0].GetTextContent() != "Hello" {
			t.Errorf("title = %v", title)
		}
	}
}

func TestXMLResource_loadURLOpensTheSystemId(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hello.xhtml")
	if err := os.WriteFile(path, []byte(xmlResourceTestDocument), 0o600); err != nil {
		t.Fatal(err)
	}
	resource := XMLResourceLoadURL(&url.URL{Scheme: "file", Path: path})
	if title := resource.GetDocument().GetElementsByTagName("title"); len(title) != 1 || title[0].GetTextContent() != "Hello" {
		t.Errorf("title = %v", title)
	}
}

func TestXMLResource_malformedDocumentPanicsWithXRRuntimeException(t *testing.T) {
	defer func() {
		recovered := recover()
		exception, ok := recovered.(*XRRuntimeException)
		if !ok {
			t.Fatalf("recovered %v, want *XRRuntimeException", recovered)
		}
		if !strings.HasPrefix(exception.GetMessage(), "Can't load the XML resource (using TrAX transformer). ") {
			t.Errorf("message = %q", exception.GetMessage())
		}
	}()
	XMLResourceLoadString("<html><body></html>")
}
