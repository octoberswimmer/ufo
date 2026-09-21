// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/TestUtils.java,
// with the assertions of com.codeborne:pdf-test that the Java tests use.

package pdf

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// TestMain puts testdata on the class path, as Maven puts
// src/test/resources on the test class path: the test documents load fonts
// and style sheets with "classpath:" URLs.
func TestMain(m *testing.M) {
	ufo.GeneralUtilAddClasspathEntry(os.DirFS("testdata"))
	os.Exit(m.Run())
}

// testUtilsResource is getClass().getResource(name) for a file under
// testdata: a "file:" URL.
func testUtilsResource(t testing.TB, name string) *url.URL {
	t.Helper()
	path := filepath.Join("testdata", filepath.FromSlash(name))
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Test resource not found: %s", name)
	}
	s, err := naiveUserAgentFileToURL(path)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// testUtilsFromUrl is Html2Pdf.fromUrl(getResource(name)).
func testUtilsFromUrl(t testing.TB, name string) []byte {
	t.Helper()
	data, err := Html2PdfFromUrl(testUtilsResource(t, name))
	if err != nil {
		t.Fatalf("rendering %s: %v", name, err)
	}
	return data
}

// testUtilsFromClasspathResource is Html2Pdf.fromClasspathResource(name).
func testUtilsFromClasspathResource(t testing.TB, name string) []byte {
	t.Helper()
	data, err := Html2PdfFromClasspathResource(name)
	if err != nil {
		t.Fatalf("rendering %s: %v", name, err)
	}
	return data
}

// testUtilsPDF is com.codeborne.pdftest.PDF: the document read back and its
// text.
type testUtilsPDF struct {
	doc  *writer.ReadDocument
	text string
}

// testUtilsPrintFile reads the PDF back. The Java method also writes it to
// target/<filename>; the port leaves no files behind.
func testUtilsPrintFile(t testing.TB, pdf []byte, filename string) *testUtilsPDF {
	t.Helper()
	doc, err := writer.ReadPDF(pdf)
	if err != nil {
		t.Fatalf("reading %s: %v", filename, err)
	}
	text, err := doc.Text()
	if err != nil {
		t.Fatalf("extracting text of %s: %v", filename, err)
	}
	return &testUtilsPDF{doc: doc, text: text}
}

// testUtilsPageContent is TestUtils.pageContent: the content stream of the
// first page.
func testUtilsPageContent(t testing.TB, pdf []byte) string {
	t.Helper()
	doc, err := writer.ReadPDF(pdf)
	if err != nil {
		t.Fatal(err)
	}
	content, err := doc.PageContent(0)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

// testUtilsSpaces is the pattern of com.codeborne.pdftest.Spaces.reduce,
// which removes white space (and no-break spaces) from the text and from the
// expected strings before containsText compares them.
var testUtilsSpaces = regexp.MustCompile("[\\s\\n\\r ]+")

func testUtilsReduce(s string) string {
	return strings.TrimSpace(testUtilsSpaces.ReplaceAllString(s, ""))
}

// containsText is PdfAssert.containsText.
func (p *testUtilsPDF) containsText(t testing.TB, texts ...string) {
	t.Helper()
	reduced := testUtilsReduce(p.text)
	for _, text := range texts {
		if !strings.Contains(reduced, testUtilsReduce(text)) {
			t.Errorf("PDF text does not contain %q; text:\n%s", text, p.text)
		}
	}
}

// doesNotContainText is PdfAssert.doesNotContainText.
func (p *testUtilsPDF) doesNotContainText(t testing.TB, texts ...string) {
	t.Helper()
	reduced := testUtilsReduce(p.text)
	for _, text := range texts {
		if strings.Contains(reduced, testUtilsReduce(text)) {
			t.Errorf("PDF text contains %q; text:\n%s", text, p.text)
		}
	}
}

// containsExactText is PdfAssert.containsExactText.
func (p *testUtilsPDF) containsExactText(t testing.TB, text string) {
	t.Helper()
	if !strings.Contains(p.text, text) {
		t.Errorf("PDF text does not contain exactly %q; text:\n%s", text, p.text)
	}
}

// lines is pdf.text.lines().toList().
func (p *testUtilsPDF) lines() []string {
	return strings.Split(strings.TrimRight(p.text, "\n"), "\n")
}

// testUtilsGetFontNames is TestUtils.getFontNames: the BaseFont names of the
// fonts in the resources of page i.
func testUtilsGetFontNames(doc *writer.ReadDocument, i int) []string {
	return doc.PageFonts(i)
}
