// Tests for the port of flying-saucer-core/src/main/java/org/xhtmlrenderer/context/StylesheetFactoryImpl.java.
// The Java suite has no test for this class.

package ufo

import (
	"strings"
	"testing"
)

// stylesheetFactoryImplTestUserAgent serves CSS from a map and counts the
// requests for each URI.
type stylesheetFactoryImplTestUserAgent struct {
	css      map[string]string
	requests map[string]int
}

func (u *stylesheetFactoryImplTestUserAgent) GetCSSResource(uri string) *CSSResource {
	u.requests[uri]++
	css, ok := u.css[uri]
	if !ok {
		return NewCSSResource(nil)
	}
	return NewCSSResource(strings.NewReader(css))
}

func (u *stylesheetFactoryImplTestUserAgent) GetImageResource(uri string) *ImageResource {
	return NewImageResource(uri, nil)
}
func (u *stylesheetFactoryImplTestUserAgent) GetXMLResource(uri string) *XMLResource { return nil }
func (u *stylesheetFactoryImplTestUserAgent) GetBinaryResource(uri string) []byte    { return nil }
func (u *stylesheetFactoryImplTestUserAgent) IsVisited(uri string) bool              { return false }
func (u *stylesheetFactoryImplTestUserAgent) SetBaseURL(url string)                  {}
func (u *stylesheetFactoryImplTestUserAgent) GetBaseURL() string                     { return "" }
func (u *stylesheetFactoryImplTestUserAgent) ResolveURI(uri string) string           { return uri }

func TestStylesheetFactoryImpl_getStylesheetLoadsThroughTheUserAgentOnce(t *testing.T) {
	userAgent := &stylesheetFactoryImplTestUserAgent{
		css:      map[string]string{"a.css": "p { color: red; }"},
		requests: map[string]int{},
	}
	factory := NewStylesheetFactoryImpl(userAgent)
	info := NewStylesheetInfo(StylesheetInfoOriginAuthor, "a.css", StylesheetInfoMediaTypes(""), nil)

	sheet := factory.GetStylesheet(info)
	if sheet == nil || len(sheet.GetContents()) != 1 {
		t.Fatalf("sheet = %v, want one ruleset", sheet)
	}
	if again := factory.GetStylesheet(info); again != sheet {
		t.Errorf("the second request returned another stylesheet")
	}
	if userAgent.requests["a.css"] != 1 {
		t.Errorf("the user agent was asked %d times, want 1", userAgent.requests["a.css"])
	}

	factory.RemoveCachedStylesheet("a.css")
	if factory.ContainsStylesheet("a.css") {
		t.Errorf("the stylesheet is cached after its removal")
	}
}

func TestStylesheetFactoryImpl_getStylesheetCachesAFailedLoad(t *testing.T) {
	userAgent := &stylesheetFactoryImplTestUserAgent{css: map[string]string{}, requests: map[string]int{}}
	factory := NewStylesheetFactoryImpl(userAgent)
	info := NewStylesheetInfo(StylesheetInfoOriginAuthor, "missing.css", StylesheetInfoMediaTypes(""), nil)

	if sheet := factory.GetStylesheet(info); sheet != nil {
		t.Errorf("sheet = %v, want nil", sheet)
	}
	if !factory.ContainsStylesheet("missing.css") {
		t.Errorf("the failed load is not cached")
	}
	factory.GetStylesheet(info)
	if userAgent.requests["missing.css"] != 1 {
		t.Errorf("the user agent was asked %d times, want 1", userAgent.requests["missing.css"])
	}
}

func TestStylesheetFactoryImpl_getStylesheetParsesInlineContentWithoutTheUserAgent(t *testing.T) {
	userAgent := &stylesheetFactoryImplTestUserAgent{css: map[string]string{}, requests: map[string]int{}}
	factory := NewStylesheetFactoryImpl(userAgent)
	content := "h1 { color: blue; } h2 { color: green; }"
	info := NewStylesheetInfo(StylesheetInfoOriginAuthor, "inline:1", StylesheetInfoMediaTypes(""), &content)

	sheet := factory.GetStylesheet(info)
	if sheet == nil || len(sheet.GetContents()) != 2 {
		t.Fatalf("sheet = %v, want two rulesets", sheet)
	}
	if len(userAgent.requests) != 0 {
		t.Errorf("the user agent was asked for %v", userAgent.requests)
	}
}

func TestStylesheetFactoryImpl_defaultStylesheetParses(t *testing.T) {
	userAgent := &stylesheetFactoryImplTestUserAgent{css: map[string]string{}, requests: map[string]int{}}
	factory := NewStylesheetFactoryImpl(userAgent)

	sheet := factory.GetStylesheet(NewXhtmlNamespaceHandler().GetDefaultStylesheet())
	if sheet == nil || len(sheet.GetContents()) == 0 {
		t.Fatalf("the default stylesheet has no rules")
	}
	if sheet.GetOrigin() != StylesheetInfoOriginUserAgent {
		t.Errorf("origin = %v, want USER_AGENT", sheet.GetOrigin())
	}
}

func TestStylesheetFactoryImpl_latin1Reader(t *testing.T) {
	reader, err := stylesheetFactoryImplNewInputStreamReader(strings.NewReader("a\xe9b"), "ISO-8859-1")
	if err != nil {
		t.Fatal(err)
	}
	var decoded strings.Builder
	buffer := make([]byte, 2)
	for {
		n, err := reader.Read(buffer)
		decoded.Write(buffer[:n])
		if err != nil {
			break
		}
	}
	if decoded.String() != "aéb" {
		t.Errorf("decoded = %q", decoded.String())
	}
	if _, err := stylesheetFactoryImplNewInputStreamReader(strings.NewReader(""), "no-such-charset"); err == nil {
		t.Errorf("an unknown charset is accepted")
	}
}
