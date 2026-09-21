// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/HTMLOutlineTest.java

package pdf

import "testing"

func htmlOutlineTestAssertLevel(t *testing.T, tagName string, expected string) {
	t.Helper()
	if actual := HTMLOutlineGetOutlineLevelFromTagName(tagName); actual != expected {
		t.Errorf("HTMLOutlineGetOutlineLevelFromTagName(%q) = %q, want %q", tagName, actual, expected)
	}
}

func TestHTMLOutline_getsOutlineLevelFromTagName_header(t *testing.T) {
	htmlOutlineTestAssertLevel(t, "h1", "1")
	htmlOutlineTestAssertLevel(t, "H2", "2")
	htmlOutlineTestAssertLevel(t, "h3", "3")
	htmlOutlineTestAssertLevel(t, "H6", "6")
	htmlOutlineTestAssertLevel(t, "h7", "7")
	htmlOutlineTestAssertLevel(t, "h10", "10")
	htmlOutlineTestAssertLevel(t, "h16", "16")
	htmlOutlineTestAssertLevel(t, "h99", "99")
}

func TestHTMLOutline_getsOutlineLevelFromTagName_exclude(t *testing.T) {
	htmlOutlineTestAssertLevel(t, "blockquote", "exclude")
	htmlOutlineTestAssertLevel(t, "BLOCKQUOTE", "exclude")
	htmlOutlineTestAssertLevel(t, "details", "exclude")
	htmlOutlineTestAssertLevel(t, "fieldset", "exclude")
	htmlOutlineTestAssertLevel(t, "figure", "exclude")
	htmlOutlineTestAssertLevel(t, "td", "exclude")
	htmlOutlineTestAssertLevel(t, "TD", "exclude")
}

func TestHTMLOutline_getsOutlineLevelFromTagName_none(t *testing.T) {
	htmlOutlineTestAssertLevel(t, "div", "none")
	htmlOutlineTestAssertLevel(t, "table", "none")
	htmlOutlineTestAssertLevel(t, "span", "none")
	htmlOutlineTestAssertLevel(t, "SPAN", "none")
}
