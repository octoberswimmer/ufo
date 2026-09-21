package pdf

import (
	"testing"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// Flying Saucer has no unit test of XHtmlMetaToPdfInfoAdapter; this checks
// what the constructor collects from the document head.
func TestXHtmlMetaToPdfInfoAdapter_collectsTitleAndMetaElements(t *testing.T) {
	doc, err := dom.ParseXMLString(`<html><head>
  <title>Element title</title>
  <meta name="DC.Title" content="Meta title"/>
  <meta name="CREATOR" content="An author"/>
  <meta name="dc.subject" content="A subject"/>
  <meta name="keywords" content="one, two"/>
  <meta name="keywords" content=""/>
  <meta name="generator" content="ignored"/>
</head><body/></html>`)
	if err != nil {
		t.Fatal(err)
	}

	adapter := NewXHtmlMetaToPdfInfoAdapter(doc)

	expected := map[writer.PdfName]string{
		writer.PdfNameTitle:    "Meta title",
		writer.PdfNameAuthor:   "An author",
		writer.PdfNameSubject:  "A subject",
		writer.PdfNameKeywords: "one, two",
	}
	if len(adapter.pdfInfoValues) != len(expected) {
		t.Errorf("collected %d values, want %d", len(adapter.pdfInfoValues), len(expected))
	}
	for name, value := range expected {
		actual := adapter.pdfInfoValues[name]
		if actual == nil || actual.String() != value {
			t.Errorf("%s = %v, want %q", name, actual, value)
		}
	}
}
