// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/CustomNamespaceHandlerTest.java

package pdf

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
)

// customNamespaceHandlerTestSpy stands for the Mockito spy of the Java test:
// it records every call of the NamespaceHandler methods and delegates to
// the NoNamespaceHandler.
type customNamespaceHandlerTestSpy struct {
	delegate *ufo.NoNamespaceHandler
	calls    []string
	docs     []*dom.Document
}

func (s *customNamespaceHandlerTestSpy) record(call string) { s.calls = append(s.calls, call) }

func (s *customNamespaceHandlerTestSpy) GetNamespace() string {
	s.record("getNamespace")
	return s.delegate.GetNamespace()
}

func (s *customNamespaceHandlerTestSpy) GetDefaultStylesheet() *ufo.StylesheetInfo {
	s.record("getDefaultStylesheet")
	return s.delegate.GetDefaultStylesheet()
}

func (s *customNamespaceHandlerTestSpy) GetDocumentTitle(doc *dom.Document) string {
	s.record("getDocumentTitle")
	return s.delegate.GetDocumentTitle(doc)
}

func (s *customNamespaceHandlerTestSpy) GetStylesheets(doc *dom.Document) []*ufo.StylesheetInfo {
	s.record("getStylesheets")
	s.docs = append(s.docs, doc)
	return s.delegate.GetStylesheets(doc)
}

func (s *customNamespaceHandlerTestSpy) GetAttributeValue(e *dom.Element, attrName string) *string {
	s.record("getAttributeValue")
	return s.delegate.GetAttributeValue(e, attrName)
}

func (s *customNamespaceHandlerTestSpy) GetAttributeValueWithNamespaceURI(e *dom.Element, namespaceURI *string, attrName string) *string {
	s.record("getAttributeValue")
	return s.delegate.GetAttributeValueWithNamespaceURI(e, namespaceURI, attrName)
}

func (s *customNamespaceHandlerTestSpy) GetClass(e *dom.Element) string {
	s.record("getClass")
	return s.delegate.GetClass(e)
}

func (s *customNamespaceHandlerTestSpy) GetID(e *dom.Element) string {
	s.record("getID")
	return s.delegate.GetID(e)
}

func (s *customNamespaceHandlerTestSpy) GetElementStyling(e *dom.Element) string {
	s.record("getElementStyling")
	return s.delegate.GetElementStyling(e)
}

func (s *customNamespaceHandlerTestSpy) GetNonCssStyling(e *dom.Element) string {
	s.record("getNonCssStyling")
	return s.delegate.GetNonCssStyling(e)
}

func (s *customNamespaceHandlerTestSpy) GetLang(e *dom.Element) string {
	s.record("getLang")
	return s.delegate.GetLang(e)
}

func (s *customNamespaceHandlerTestSpy) GetLinkUri(e *dom.Element) *string {
	s.record("getLinkUri")
	return s.delegate.GetLinkUri(e)
}

func (s *customNamespaceHandlerTestSpy) GetAnchorName(e *dom.Element) string {
	s.record("getAnchorName")
	return s.delegate.GetAnchorName(e)
}

func (s *customNamespaceHandlerTestSpy) IsImageElement(e *dom.Element) bool {
	s.record("isImageElement")
	return s.delegate.IsImageElement(e)
}

func (s *customNamespaceHandlerTestSpy) IsFormElement(e *dom.Element) bool {
	s.record("isFormElement")
	return s.delegate.IsFormElement(e)
}

func (s *customNamespaceHandlerTestSpy) GetImageSourceURI(e *dom.Element) string {
	s.record("getImageSourceURI")
	return s.delegate.GetImageSourceURI(e)
}

func TestCustomNamespaceHandler_usingCustomNamespaceHandler(t *testing.T) {
	doc := ufo.XMLResourceLoadString(`<html>
    <body><h1>Hello, world!</h1></body>
</html>
`).GetDocument()
	handler := &customNamespaceHandlerTestSpy{delegate: ufo.NewNoNamespaceHandler()}
	var os bytes.Buffer
	renderer := NewITextRenderer()
	if err := renderer.SetDocumentWithUrlNsh(doc, "", handler); err != nil {
		t.Fatal(err)
	}
	if err := renderer.CreatePDFDocumentWithOs(doc, &os); err != nil {
		t.Fatal(err)
	}
	pdf := testUtilsPrintFile(t, os.Bytes(), "custom-namespace-handler.pdf")

	pdf.containsText(t, "Hello, world!")
	// verify(handler).getDefaultStylesheet(); verify(handler).getStylesheets(doc);
	// verifyNoMoreInteractions(handler);
	if !reflect.DeepEqual(handler.calls, []string{"getDefaultStylesheet", "getStylesheets"}) {
		t.Errorf("calls = %v, want [getDefaultStylesheet getStylesheets]", handler.calls)
	}
	if len(handler.docs) != 1 || handler.docs[0] != doc {
		t.Errorf("getStylesheets was not called with the document")
	}
}
