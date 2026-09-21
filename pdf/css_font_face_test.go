// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/CssFontFaceTest.java

package pdf

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// Verifies that font-face declarations with the -fs-pdf-font-embed property
// actually results in the font being embedded in the pdf.
//
// The Jacquard24 font used to test is from Google Fonts
// (https://fonts.google.com/specimen/Jacquard+24) and is licenced under the
// Open Font License.
func TestCssFontFace_autoInstallationOfCssDeclaredFonts(t *testing.T) {
	pdfBytes := testUtilsFromUrl(t, "org/xhtmlrenderer/pdf/fonts/CssFontFace.html")
	cssFontFaceTestAssertEmbeddedJacquard(t, pdfBytes)
}

// Verifies that imported font-face declarations with embedFontFaces actually
// results in the font being embedded in the pdf. Also verifies that opaque
// urls with the font-face format property are correctly supported.
//
// Java sets FSEntityResolver on the DocumentBuilder, which the port does not
// have; the port's DocumentBuilder does not load external DTDs.
func TestCssFontFace_importedFontFaceWithEmbedOverride(t *testing.T) {
	htmlUrl := testUtilsResource(t, "org/xhtmlrenderer/pdf/fonts/CssImportFontFace.html")
	renderer := NewITextRenderer()
	renderer.GetFontResolver().AddEmbedFontFace("Jacquard 24", writer.BaseFontIdentityH)
	renderer.GetSharedContext().SetMedia("pdf")
	renderer.GetSharedContext().SetInteractive(false)
	renderer.GetSharedContext().GetTextRenderer().SetSmoothingThreshold(0)

	builder := ufo.XMLUtilNewDocumentBuilder()

	doc, err := builder.ParseString(htmlUrl.String())
	if err != nil {
		t.Fatal(err)
	}
	pdfBytes, err := renderer.CreatePDFDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	cssFontFaceTestAssertEmbeddedJacquard(t, pdfBytes)
}

// cssFontFaceTestUserAgent is the anonymous ITextUserAgent subclass of
// serverRelativeFontPathFromInlineStyleIsEmbedded.
type cssFontFaceTestUserAgent struct {
	*ITextUserAgent
	fontUri string
}

func (a *cssFontFaceTestUserAgent) GetBinaryResource(uri string) []byte {
	if uri != "" && strings.HasSuffix(uri, a.fontUri) {
		return a.ITextUserAgent.GetBinaryResource("classpath:fonts/Jacquard24-Regular.ttf")
	}
	return a.ITextUserAgent.GetBinaryResource(uri)
}

// Issue #695: url('/abs/path.ttf') inside an inline <style> @font-face used
// to be rewritten to inline/abs/path.ttf. Relative urls still resolve against
// the HTML document, so they never show that. A server-relative path does,
// and the font must still be embedded.
func TestCssFontFace_serverRelativeFontPathFromInlineStyleIsEmbedded(t *testing.T) {
	fontUri := "/fonts/Jacquard24-Regular.ttf"
	html := fmt.Sprintf(`<html xmlns="http://www.w3.org/1999/xhtml" lang="en">
<head>
    <style>
        @font-face {
            font-family: "Jacquard 24";
            src: url("%s");
            -fs-pdf-font-embed: embed;
        }
        .jacquard { font-family: "Jacquard 24", sans-serif; }
    </style>
</head>
<body>
    <p class="jacquard">JACQUARD FONT</p>
</body>
</html>
`, fontUri)

	renderer := NewITextRenderer()
	userAgent := &cssFontFaceTestUserAgent{
		ITextUserAgent: NewITextUserAgent(
			renderer.GetOutputDevice(),
			int(math.Floor(float64(renderer.GetOutputDevice().GetDotsPerPoint())+0.5))),
		fontUri: fontUri,
	}
	userAgent.SetSelf(userAgent)
	renderer.GetSharedContext().SetUserAgentCallback(userAgent)
	renderer.GetSharedContext().SetMedia("pdf")
	renderer.GetSharedContext().SetInteractive(false)
	renderer.GetSharedContext().GetTextRenderer().SetSmoothingThreshold(0)

	doc := ufo.XMLResourceLoadString(html).GetDocument()
	pdfBytes, err := renderer.CreatePDFDocument(doc)
	if err != nil {
		t.Fatal(err)
	}

	src := renderer.GetSharedContext().GetCss().GetFontFaceRules()[0].
		GetCalculatedStyle().
		ValueByName(ufo.CSSNameSrc).
		AsString()
	if src != fontUri {
		t.Errorf("src = %q, want %q", src, fontUri)
	}
	cssFontFaceTestAssertEmbeddedJacquard(t, pdfBytes)
}

func cssFontFaceTestAssertEmbeddedJacquard(t *testing.T, pdfBytes []byte) {
	t.Helper()
	document, err := writer.ReadPDF(pdfBytes)
	if err != nil {
		t.Fatal(err)
	}
	if document.NumPages() < 1 {
		t.Fatalf("%d pages, want at least 1", document.NumPages())
	}
	for i := 0; i < document.NumPages(); i++ {
		fontNames := testUtilsGetFontNames(document, i)
		found := false
		for _, name := range fontNames {
			if strings.Contains(name, "Jacquard24") {
				found = true
			}
		}
		if !found {
			t.Errorf("page %d: Should contain Jacquard24 font; fonts %v", i+1, fontNames)
		}
	}
}
