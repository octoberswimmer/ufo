// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/DocumentSplitterTest.java

package pdf

import (
	"strings"
	"testing"
)

// documentSplitterTestParse parses xml with a namespace-aware, non-validating
// SAX parser reporting to a new DocumentSplitter, as setUp and the test
// methods do. SAXParse returns the first error, as the StrictErrorHandler
// of the Java test throws on every error and warning.
func documentSplitterTestParse(t *testing.T, xml string) *DocumentSplitter {
	t.Helper()
	splitter := NewDocumentSplitter()
	if err := SAXParse(strings.NewReader(xml), splitter); err != nil {
		t.Fatal(err)
	}
	return splitter
}

func TestDocumentSplitter_splitDocumentWithoutHead(t *testing.T) {
	splitter := documentSplitterTestParse(t, "<h1>no head</h1>")
	if n := len(splitter.GetDocuments()); n != 0 {
		t.Errorf("%d documents, want 0", n)
	}
}

func TestDocumentSplitter_splitDocumentWithHead(t *testing.T) {
	splitter := documentSplitterTestParse(t, "<html>"+
		"<head><title>The head</title></head>"+
		"<body><h1>I have head</h1></body>"+
		"</html>")
	if n := len(splitter.GetDocuments()); n != 1 {
		t.Fatalf("%d documents, want 1", n)
	}
	doc := splitter.GetDocuments()[0]
	if n := len(doc.GetElementsByTagName("h1")); n != 1 {
		t.Errorf("%d h1 elements, want 1", n)
	}

	// "I just have fixed how it works de-facto. Seems to be an invalid result."
	if s := xmlTransformerSerialize(doc); s != "<body><head><title>The head</title></head><h1>I have head</h1></body>" {
		t.Errorf("serialized %q", s)
	}
}

func TestDocumentSplitter_splitDocumentWithMultipleBodies(t *testing.T) {
	splitter := documentSplitterTestParse(t, "<html>"+
		"<head><title>The head</title></head>"+
		"<body><h1>I have head</h1></body>"+
		"<body><h2>Second head</h2></body>"+
		"</html>")

	if n := len(splitter.GetDocuments()); n != 2 {
		t.Fatalf("%d documents, want 2", n)
	}
	doc1 := splitter.GetDocuments()[0]
	if n := len(doc1.GetElementsByTagName("h1")); n != 1 {
		t.Errorf("%d h1 elements, want 1", n)
	}
	doc2 := splitter.GetDocuments()[1]
	if n := len(doc2.GetElementsByTagName("h2")); n != 1 {
		t.Errorf("%d h2 elements, want 1", n)
	}

	// "I just have fixed how it works de-facto. Seems to be an invalid result."
	if s := xmlTransformerSerialize(doc1); s != "<body><head><title>The head</title></head><h1>I have head</h1></body>" {
		t.Errorf("serialized doc1 %q", s)
	}
	// Java expects "<title>2&gt;Second</title>": SAXEventRecorder keeps the
	// char[] that the JDK's SAX parser passes to characters() and replays
	// it for the second document after the parser has reused the array for
	// later text. The port's SAXContentHandler receives each piece of text
	// as a string, so the recorded title of the head is the text of the
	// document, "The head".
	if s := xmlTransformerSerialize(doc2); s != "<body><head><title>The head</title></head><h2>Second head</h2></body>" {
		t.Errorf("serialized doc2 %q", s)
	}
}
