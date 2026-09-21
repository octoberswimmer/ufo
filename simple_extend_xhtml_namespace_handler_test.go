// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/simple/extend/XhtmlNamespaceHandlerTest.java

package ufo

import (
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

func TestXhtmlNamespaceHandler_looksLikeAMangledColor_otherThan6Chars(t *testing.T) {
	handler := NewXhtmlNamespaceHandler()
	for _, input := range []string{"", "12345", "1234567"} {
		if handler.LooksLikeAMangledColor(input) {
			t.Errorf("LooksLikeAMangledColor(%q) = true, want false", input)
		}
	}
}

func TestXhtmlNamespaceHandler_looksLikeAMangledColor(t *testing.T) {
	handler := NewXhtmlNamespaceHandler()
	cases := []struct {
		input          string
		expectedResult bool
		explanation    string
	}{
		{"123456", true, "6 digits"},
		{"012345", true, "6 digits"},
		{"987654", true, "6 digits"},
		{"abcdef", true, "letters a..f"},
		{"abcdeg", false, "letter g"},
		{"1bdd3f", true, "letters a..f and digits"},
		{"_12345", false, "character _"},
	}
	for _, c := range cases {
		if got := handler.LooksLikeAMangledColor(c.input); got != c.expectedResult {
			t.Errorf("looksLikeAMangledColor('%s') should be %v: %s", c.input, c.expectedResult, c.explanation)
		}
	}
}

func TestXhtmlNamespaceHandler_findsTable(t *testing.T) {
	handler := NewXhtmlNamespaceHandler()
	html := newXhtmlNamespaceHandlerTestHtml(t, `<html>
    <head>
        <title>Hello</title>
    </head>
    <body>
    <table id="table1">
        <tr>
            <td id="td1">111</td>
            <td id="td2">222</td>
        </tr>
        <tbody>
            <tr>
                <td id="td3">333</td>
                <td id="td4">444</td>
            </tr>
        </tbody>
    </table>
    </body>
</html>
`)
	table := html.find("table", 0)
	for index := 0; index < 4; index++ {
		if got := handler.FindTable(html.find("td", index)); got != table {
			t.Errorf("FindTable(td %d) = %v, want the table", index, got)
		}
	}

	if got := handler.Ancestor(html.find("title", 0), "head", 1000); got != html.find("head", 0) {
		t.Errorf("Ancestor(title, head) = %v, want the head", got)
	}
	if got := handler.Ancestor(html.find("td", 0), "tbody", 1000); got != nil {
		t.Errorf("Ancestor(td 0, tbody) = %v, want nil", got)
	}
	if got := handler.Ancestor(html.find("td", 3), "tbody", 1000); got != html.find("tbody", 0) {
		t.Errorf("Ancestor(td 3, tbody) = %v, want the tbody", got)
	}
}

type xhtmlNamespaceHandlerTestHtml struct {
	t        *testing.T
	document *dom.Document
}

func newXhtmlNamespaceHandlerTestHtml(t *testing.T, html string) *xhtmlNamespaceHandlerTestHtml {
	t.Helper()
	document, err := dom.ParseXMLString(html)
	if err != nil {
		t.Fatal(err)
	}
	return &xhtmlNamespaceHandlerTestHtml{t: t, document: document}
}

func (h *xhtmlNamespaceHandlerTestHtml) find(tagName string, index int) *dom.Element {
	h.t.Helper()
	elements := h.document.GetElementsByTagName(tagName)
	if index >= len(elements) {
		h.t.Fatalf("no <%s> element at index %d", tagName, index)
	}
	return elements[index]
}
