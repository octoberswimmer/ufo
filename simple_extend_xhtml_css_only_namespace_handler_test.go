// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/simple/extend/XhtmlCssOnlyNamespaceHandlerTest.java

package ufo

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/dom"
)

func TestXhtmlCssOnlyNamespaceHandler_readsAllCssStyles(t *testing.T) {
	handler := NewXhtmlCssOnlyNamespaceHandler()
	stylesheets := handler.GetStylesheets(xhtmlCssOnlyNamespaceHandlerTestRead(t, "/hello.css.html"))
	if len(stylesheets) != 2 {
		t.Fatalf("got %d stylesheets, want 2", len(stylesheets))
	}
	if uri := stylesheets[0].GetUri(); !regexp.MustCompile(`^inline:\d+$`).MatchString(uri) {
		t.Errorf("uri = %q, want inline:<number>", uri)
	}
	if media := stylesheets[0].GetMedia(); len(media) != 1 || media[0] != "all" {
		t.Errorf("media = %v, want [all]", media)
	}
	if origin := stylesheets[0].GetOrigin(); origin != StylesheetInfoOriginAuthor {
		t.Errorf("origin = %v, want AUTHOR", origin)
	}
	if content := stylesheets[0].GetContent(); content == nil || *content != "body {color: black;}" {
		t.Errorf("content of the first stylesheet = %v", content)
	}
	if content := stylesheets[1].GetContent(); content == nil || *content != "h1 {color: red;}" {
		t.Errorf("content of the second stylesheet = %v", content)
	}
}

func TestXhtmlCssOnlyNamespaceHandler_fileWithoutCssStyles(t *testing.T) {
	handler := NewXhtmlCssOnlyNamespaceHandler()
	stylesheets := handler.GetStylesheets(xhtmlCssOnlyNamespaceHandlerTestRead(t, "/hello.html"))
	if len(stylesheets) != 0 {
		t.Errorf("got %d stylesheets, want none", len(stylesheets))
	}
}

func TestXhtmlCssOnlyNamespaceHandler_readsDefaultCssStylesheetFromFile(t *testing.T) {
	handler := NewXhtmlCssOnlyNamespaceHandler()
	css := handler.GetDefaultStylesheet()
	if css == nil {
		t.Fatal("no default stylesheet")
	}
	if !strings.HasSuffix(css.GetUri(), "/css/XhtmlNamespaceHandler.css") {
		t.Errorf("uri = %q, want the suffix /css/XhtmlNamespaceHandler.css", css.GetUri())
	}
	if css.GetOrigin() != StylesheetInfoOriginUserAgent {
		t.Errorf("origin = %v, want USER_AGENT", css.GetOrigin())
	}
	if media := css.GetMedia(); len(media) != 1 || media[0] != "all" {
		t.Errorf("media = %v, want [all]", media)
	}
	// The Go port carries the embedded file in the StylesheetInfo.
	if content := css.GetContent(); content == nil || !strings.Contains(*content, "display: block") {
		t.Errorf("the default stylesheet has no content")
	}
}

func TestXhtmlCssOnlyNamespaceHandler_collapseWhiteSpace_samples(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{"", ""},
		{" ", " "},
		{"     ", " "},
		{" a  \t  b  \n  c  \r  d  \u000B  e    \f   f  ", " a b c d e f "},
		{"| \t  \n |  \r \u000B \u001E \f    |  \u001E   \u001F   |", "| | | |"},
	}
	for _, c := range cases {
		if got := XhtmlCssOnlyNamespaceHandlerCollapseWhiteSpace(c.text); got != c.want {
			t.Errorf("CollapseWhiteSpace(%q) = %q, want %q", c.text, got, c.want)
		}
	}
}

func xhtmlCssOnlyNamespaceHandlerTestRead(t *testing.T, name string) *dom.Document {
	t.Helper()
	file, err := os.Open("testdata" + name)
	if err != nil {
		t.Fatalf("test resource not found: %s", name)
	}
	defer file.Close()
	document, err := dom.ParseXML(file)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
