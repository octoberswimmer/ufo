// Tests for the small classes of org.xhtmlrenderer.util, which have no JUnit
// tests in Flying Saucer.

package ufo

import (
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unicode"

	"github.com/octoberswimmer/ufo/dom"
)

func TestArrayUtilCloneOrEmpty(t *testing.T) {
	if got := ArrayUtilCloneOrEmptyStringArray(nil); got == nil || len(got) != 0 {
		t.Errorf("nil strings: %v", got)
	}
	if got := ArrayUtilCloneOrEmptyByteArray(nil); got == nil || len(got) != 0 {
		t.Errorf("nil bytes: %v", got)
	}
	if got := ArrayUtilCloneOrEmptyIntArray(nil); got == nil || len(got) != 0 {
		t.Errorf("nil ints: %v", got)
	}
	source := []string{"a", "b"}
	clone := ArrayUtilCloneOrEmptyStringArray(source)
	clone[0] = "changed"
	if source[0] != "a" || !reflect.DeepEqual(clone, []string{"changed", "b"}) {
		t.Errorf("source %v, clone %v", source, clone)
	}
	if got := ArrayUtilCloneOrEmptyIntArray([]int{1, 2}); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("ints: %v", got)
	}
	if got := ArrayUtilCloneOrEmptyByteArray([]byte{1, 2}); !reflect.DeepEqual(got, []byte{1, 2}) {
		t.Errorf("bytes: %v", got)
	}
}

func TestUtilReplace(t *testing.T) {
	tests := []struct{ source, target, replacement, want string }{
		{"a.b.c", ".", "::", "a::b::c"},
		{"abc", "x", "y", "abc"},
		{"aaa", "aa", "b", "ba"},
		{"", "a", "b", ""},
	}
	for _, test := range tests {
		if got := UtilReplace(test.source, test.target, test.replacement); got != test.want {
			t.Errorf("UtilReplace(%q, %q, %q) = %q, want %q", test.source, test.target, test.replacement, got, test.want)
		}
	}
	if !UtilIsNullOrEmpty("") || UtilIsNullOrEmpty(" ") {
		t.Errorf("UtilIsNullOrEmpty")
	}
}

func TestGeneralUtilParseIntRelaxed(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 0}, {"   ", 0}, {"abc", 0}, {"0", 0}, {"12", 12}, {"12px", 12}, {" width: 34 px 56", 34},
		{"-7", 7}, {"99999999999", math.MaxInt32}, {"\u0661\u0662", 12},
	}
	for _, test := range tests {
		if got := GeneralUtilParseIntRelaxed(test.s); got != test.want {
			t.Errorf("GeneralUtilParseIntRelaxed(%q) = %d, want %d", test.s, got, test.want)
		}
	}
	if got := GeneralUtilParseIntRelaxedWithDefaultValue("none", 5); got != 5 {
		t.Errorf("default value: %d", got)
	}
}

// generalUtilDigitValue relies on every range of the Nd table being a whole
// number of runs of the ten digits 0 to 9.
func TestGeneralUtilDigitRanges(t *testing.T) {
	for _, r := range unicode.Nd.R16 {
		if r.Stride != 1 || (int(r.Hi)-int(r.Lo)+1)%10 != 0 {
			t.Errorf("range %04X-%04X stride %d", r.Lo, r.Hi, r.Stride)
		}
	}
	for _, r := range unicode.Nd.R32 {
		if r.Stride != 1 || (int(r.Hi)-int(r.Lo)+1)%10 != 0 {
			t.Errorf("range %04X-%04X stride %d", r.Lo, r.Hi, r.Stride)
		}
	}
	for r, want := range map[rune]int{'0': 0, '9': 9, '\u0665': 5, '\uff17': 7, '\U0001D7D9': 1} {
		if got := generalUtilDigitValue(r); got != want {
			t.Errorf("digit value of %U = %d, want %d", r, got, want)
		}
	}
}

func TestGeneralUtilStrings(t *testing.T) {
	if !GeneralUtilCiEquals("Linear-Gradient", "linear-gradient") || GeneralUtilCiEquals("a", "b") || !GeneralUtilCiEquals("", "") {
		t.Errorf("GeneralUtilCiEquals")
	}
	if got := GeneralUtilHtmlEscapeSpace(`C:\my docs\a b.html`); got != "C:/my%20docs/a%20b.html" {
		t.Errorf("GeneralUtilHtmlEscapeSpace: %q", got)
	}
	if got := GeneralUtilEscapeHTML(`<a href="x">b & c</a>`); got != "&lt;a&nbsp;href=&quot;x&quot;&gt;b&nbsp;&amp;&nbsp;c&lt;/a&gt;" {
		t.Errorf("GeneralUtilEscapeHTML: %q", got)
	}
}

func TestGeneralUtilClasspath(t *testing.T) {
	stream := GeneralUtilOpenStreamFromClasspath(nil, "resources/conf/xhtmlrenderer.conf")
	if stream == nil {
		t.Fatalf("the default configuration file is not embedded")
	}
	stream.Close()
	if stream := GeneralUtilOpenStreamFromClasspath(nil, "resources/conf/missing.conf"); stream != nil {
		t.Errorf("a stream for a missing resource")
	}
	if u := GeneralUtilGetURLFromClasspath(nil, "resources/conf/missing.conf"); u != nil {
		t.Errorf("a URL for a missing resource: %v", u)
	}
	u := GeneralUtilGetURLFromClasspath(nil, "/resources/conf/xhtmlrenderer.conf")
	if u == nil || u.String() != "classpath:///resources/conf/xhtmlrenderer.conf" {
		t.Fatalf("URL %v", u)
	}
	fromUrl := IOUtilReadBytesString(u.String())
	fromFile, err := os.ReadFile("resources/conf/xhtmlrenderer.conf")
	if err != nil {
		t.Fatal(err)
	}
	if string(fromUrl) != string(fromFile) {
		t.Errorf("the classpath URL yields %d bytes, the file has %d", len(fromUrl), len(fromFile))
	}
}

func TestIOUtilFiles(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "page.html")
	if err := os.WriteFile(page, []byte("<html/>"), 0o644); err != nil {
		t.Fatal(err)
	}
	outputDir := filepath.Join(dir, "out")
	if err := os.Mkdir(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := IOUtilCopyFile(page, outputDir); err != nil {
		t.Fatal(err)
	}
	copied, err := IOUtilReadBytesPath(filepath.Join(outputDir, "page.html"))
	if err != nil || string(copied) != "<html/>" {
		t.Errorf("copied %q, %v", copied, err)
	}

	fileUri := (&url.URL{Scheme: "file", Path: filepath.ToSlash(page)}).String()
	if got := IOUtilReadBytesString(fileUri); string(got) != "<html/>" {
		t.Errorf("IOUtilReadBytesString(%q) = %q", fileUri, got)
	}
	stream := IOUtilOpenStreamAtUrl(fileUri)
	if stream == nil {
		t.Fatalf("IOUtilOpenStreamAtUrl(%q) = nil", fileUri)
	}
	IOUtilClose(stream)

	if err := IOUtilDeleteAllFiles(outputDir); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(outputDir); len(entries) != 0 {
		t.Errorf("%d files left", len(entries))
	}
	if err := IOUtilDeleteAllFiles(filepath.Join(dir, "missing")); err != nil {
		t.Errorf("missing directory: %v", err)
	}
}

func TestIOUtilFailures(t *testing.T) {
	if stream := IOUtilGetInputStream(""); stream != nil {
		t.Errorf("a stream for an empty URI")
	}
	missing := (&url.URL{Scheme: "file", Path: filepath.ToSlash(filepath.Join(t.TempDir(), "missing"))}).String()
	for _, uri := range []string{"no-scheme.html", "unknown-scheme://x/y", missing} {
		if stream := IOUtilGetInputStream(uri); stream != nil {
			t.Errorf("IOUtilGetInputStream(%q) returned a stream", uri)
		}
		if stream := IOUtilOpenStreamAtUrl(uri); stream != nil {
			t.Errorf("IOUtilOpenStreamAtUrl(%q) returned a stream", uri)
		}
		if data := IOUtilReadBytesString(uri); data != nil {
			t.Errorf("IOUtilReadBytesString(%q) returned data", uri)
		}
	}
	IOUtilClose(nil)
}

func TestIOUtilHttp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/found" {
			io.WriteString(w, "accept="+r.Header.Get("Accept"))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	stream, err := IOUtilStreamAtUrl(server.URL + "/found")
	if err != nil {
		t.Fatal(err)
	}
	data, err := IOUtilReadBytesInputStream(stream)
	stream.Close()
	if err != nil || string(data) != "accept=*/*" {
		t.Errorf("data %q, %v", data, err)
	}
	if stream := IOUtilOpenStreamAtUrl(server.URL + "/missing"); stream != nil {
		t.Errorf("a stream for a 404 response")
	}
}

func TestFontUtil(t *testing.T) {
	if !FontUtilIsEmbeddedBase64Font("data:font/ttf;base64,AAEC") || FontUtilIsEmbeddedBase64Font("data:image/png;base64,AAEC") || FontUtilIsEmbeddedBase64Font("") {
		t.Errorf("FontUtilIsEmbeddedBase64Font")
	}
	// The bytes 0xfb 0xff 0xbf are "+/+/" in base 64; the second URI writes
	// the characters as percent escapes, and the third has no padding.
	for uri, want := range map[string][]byte{
		"data:font/ttf;base64,+/+/":             {0xfb, 0xff, 0xbf},
		"data:font/ttf;base64,%2B%2F%2B%2F":     {0xfb, 0xff, 0xbf},
		"data:font/otf;charset=x;base64,AAECAw": {0, 1, 2, 3},
		"data:font/otf;base64,AAECAw==":         {0, 1, 2, 3},
	} {
		reader := FontUtilGetEmbeddedBase64Data(uri)
		if reader == nil {
			t.Fatalf("%s: no data", uri)
		}
		got, _ := io.ReadAll(reader)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, want %v", uri, got, want)
		}
	}
	if reader := FontUtilGetEmbeddedBase64Data("data:font/ttf;utf8,abc"); reader != nil {
		t.Errorf("data for a URI that is not base 64 encoded")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("malformed base 64 data did not panic")
			}
		}()
		FontUtilGetEmbeddedBase64Data("data:font/ttf;base64,!!!!")
	}()
}

func TestSupportedEmbeddedFontTypes(t *testing.T) {
	if !SupportedEmbeddedFontTypesIsSupported("data:font/otf;base64,AAEC") || SupportedEmbeddedFontTypesIsSupported("data:font/woff;base64,AAEC") {
		t.Errorf("SupportedEmbeddedFontTypesIsSupported")
	}
	if got := SupportedEmbeddedFontTypesGetExtension("data:font/ttf;base64,AAEC"); got != ".ttf" {
		t.Errorf("extension %q", got)
	}
	if got := SupportedEmbeddedFontTypesGetExtension("data:font/woff;base64,AAEC"); got != "" {
		t.Errorf("extension %q", got)
	}
	if SupportedEmbeddedFontTypesOtf.TypeString != "font/otf" || SupportedEmbeddedFontTypesOtf.Name() != "OTF" {
		t.Errorf("OTF constant: %+v", SupportedEmbeddedFontTypesOtf)
	}
}

func TestTextUtilReadTextContent(t *testing.T) {
	document, err := dom.ParseXMLString("<p>one<b>bold</b> two<![CDATA[ <three>]]><!-- comment --></p>")
	if err != nil {
		t.Fatal(err)
	}
	if got := TextUtilReadTextContent(document.GetDocumentElement()); got != "one two <three>" {
		t.Errorf("text %q", got)
	}
	empty, err := dom.ParseXMLString("<p><b>bold</b></p>")
	if err != nil {
		t.Fatal(err)
	}
	if got := TextUtilReadTextContentOrNull(empty.GetDocumentElement()); got != "" {
		t.Errorf("text %q", got)
	}
}

func TestLazyEvaluated(t *testing.T) {
	calls := 0
	lazy := LazyEvaluatedLazy(func() string {
		calls++
		return "value"
	})
	if calls != 0 {
		t.Errorf("the supplier ran before Get")
	}
	if lazy.Get() != "value" || lazy.Get() != "value" || calls != 1 {
		t.Errorf("value %q after %d calls of the supplier", lazy.Get(), calls)
	}
}

func TestInputSources(t *testing.T) {
	if InputSourcesFromStream(nil) != nil {
		t.Errorf("an input source for a nil stream")
	}
	source := InputSourcesFromString("<a/>")
	data, _ := io.ReadAll(source.GetCharacterStream())
	if string(data) != "<a/>" || source.GetByteStream() != nil || source.GetSystemId() != "" {
		t.Errorf("from string: %q", data)
	}
	u, _ := url.Parse("http://example.com/a.xhtml")
	if got := InputSourcesFromURL(u).GetSystemId(); got != "http://example.com/a.xhtml" {
		t.Errorf("system id %q", got)
	}
}
