package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

// Flying Saucer has no unit tests of ITextOutputDevice; these check the
// coordinate conversion, clipping and metadata of the port without a rendered
// document.

// iTextOutputDeviceTestContent reads the operators written to a content
// stream after a mark.
type iTextOutputDeviceTestContent struct {
	cb   *writer.PdfContentByte
	mark int
}

// reset moves the mark to the end of the content written so far.
func (c *iTextOutputDeviceTestContent) reset() {
	c.mark = len(c.cb.ToPdf())
}

// iTextOutputDeviceTestPage returns a device at 20 dots per point that draws
// on a page 100 points high, and the content it writes after the page was
// initialized.
func iTextOutputDeviceTestPage(t *testing.T) (*ITextOutputDevice, *iTextOutputDeviceTestContent) {
	t.Helper()
	document := writer.NewDocument(writer.NewRectangleWithUrxUry(200, 100), 0, 0, 0, 0)
	w, err := writer.PdfWriterGetInstance(document, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	document.Open()
	device := NewITextOutputDevice(20)
	device.SetWriter(w)
	cb := w.GetDirectContent()
	device.InitializePage(cb, 100)
	content := &iTextOutputDeviceTestContent{cb: cb}
	content.reset()
	return device, content
}

func iTextOutputDeviceTestOperators(content *iTextOutputDeviceTestContent) string {
	return strings.Join(strings.Fields(string(content.cb.ToPdf()[content.mark:])), " ")
}

func TestITextOutputDevice_fillRectConvertsDotsToPointsAndFlipsY(t *testing.T) {
	device, cb := iTextOutputDeviceTestPage(t)

	device.SetColor(ufo.NewFSRGBColor(255, 0, 0))
	device.FillRect(200, 400, 600, 200)

	// x 200..800 dots = 10..40 pt; y 400..600 dots = 20..30 pt from the top
	// of a page 100 pt high = 80..70 pt from the bottom.
	expected := "1 0 0 rg 10 80 m 40 80 l 40 70 l 10 70 l 10 80 l h f"
	if actual := iTextOutputDeviceTestOperators(cb); actual != expected {
		t.Errorf("content = %q, want %q", actual, expected)
	}
}

func TestITextOutputDevice_fillColorIsSetOnceUntilItChanges(t *testing.T) {
	device, cb := iTextOutputDeviceTestPage(t)

	device.SetColor(ufo.NewFSRGBColor(0, 0, 255))
	device.FillRect(0, 0, 20, 20)
	device.SetColor(ufo.NewFSRGBColor(0, 0, 255))
	device.FillRect(0, 0, 20, 20)

	if count := strings.Count(iTextOutputDeviceTestOperators(cb), " rg"); count != 1 {
		t.Errorf("rg operators = %d, want 1", count)
	}
}

func TestITextOutputDevice_translateMovesLaterDrawing(t *testing.T) {
	device, cb := iTextOutputDeviceTestPage(t)

	device.Translate(100, 200)
	device.DrawLine(0, 0, 100, 0)

	actual := iTextOutputDeviceTestOperators(cb)
	if !strings.HasSuffix(actual, "5 90 m 10 90 l S") {
		t.Errorf("content = %q, want a line from 5 90 to 10 90", actual)
	}
}

func TestITextOutputDevice_clipOfSeveralShapesClipsToEachInTurn(t *testing.T) {
	device, cb := iTextOutputDeviceTestPage(t)

	device.Clip(geom.NewRectangle(0, 0, 400, 400))
	device.Clip(geom.NewEllipse2DFloat(0, 0, 400, 400))
	saved := device.GetClip()
	cb.reset()

	// SetClip with the saved area restores the state and clips to both
	// member shapes again.
	device.SetClip(saved)

	actual := iTextOutputDeviceTestOperators(cb)
	if !strings.HasPrefix(actual, "Q q ") {
		t.Errorf("content = %q, want it to start with Q q", actual)
	}
	if count := strings.Count(actual, " W n"); count != 2 {
		t.Errorf("clip operators = %d, want 2 in %q", count, actual)
	}
	if !strings.Contains(actual, "0 100 m 20 100 l 20 80 l 0 80 l") {
		t.Errorf("content = %q, want the rectangle 0..20 pt wide at the top of the page", actual)
	}
}

func TestITextOutputDevice_getClipReturnsTheClipInDots(t *testing.T) {
	device, _ := iTextOutputDeviceTestPage(t)

	if clip := device.GetClip(); clip != nil {
		t.Fatalf("GetClip() = %v before any clip, want nil", clip)
	}
	device.Translate(40, 60)
	device.Clip(geom.NewRectangle(0, 0, 400, 200))

	bounds := device.GetClip().GetBounds()
	if !bounds.Equals(geom.NewRectangle(0, 0, 400, 200)) {
		t.Errorf("GetClip().GetBounds() = %v, want the rectangle that was clipped to", bounds)
	}
}

func TestITextOutputDevice_setStrokeScalesWidthAndDashToPoints(t *testing.T) {
	device, cb := iTextOutputDeviceTestPage(t)

	stroke := geom.NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(40, geom.BasicStrokeCapButt, geom.BasicStrokeJoinMiter, 10, []float32{20, 60}, 0)
	device.SetStroke(stroke)
	device.DrawLine(0, 0, 200, 0)

	if device.GetStroke() != geom.Stroke(stroke) {
		t.Errorf("GetStroke() does not return the stroke that was set")
	}
	actual := iTextOutputDeviceTestOperators(cb)
	for _, expected := range []string{"2 w", "0 J", "[1 3 ]0 d"} {
		if !strings.Contains(actual, expected) {
			t.Errorf("content = %q, want it to contain %q", actual, expected)
		}
	}
}

func TestITextOutputDevice_pushTransformWithoutCssTransformIsRestoredByPop(t *testing.T) {
	device, _ := iTextOutputDeviceTestPage(t)
	box := &cssTransformTestBox{style: &cssTransformTestStyle{}}

	before := device.GetTransform().Clone()
	device.PushTransform(nil, box)
	device.Translate(10, 10)
	device.PopTransform()

	if !device.GetTransform().Equals(before) {
		t.Errorf("transform after PopTransform = %v, want %v", device.GetTransform(), before)
	}
}

func TestITextOutputDevice_metadata(t *testing.T) {
	device := NewITextOutputDevice(20)

	if value := device.GetMetadataByName("title"); value != nil {
		t.Errorf("GetMetadataByName on an empty device = %q, want nil", *value)
	}

	device.AddMetadata("Keywords", "a")
	device.AddMetadata("keywords", "b")
	device.AddMetadata("author", "c")

	if value := device.GetMetadataByName("KEYWORDS"); value == nil || *value != "a" {
		t.Errorf("GetMetadataByName(KEYWORDS) = %v, want a", value)
	}
	if values := device.GetMetadataListByName("keywords"); len(values) != 2 || values[0] != "a" || values[1] != "b" {
		t.Errorf("GetMetadataListByName(keywords) = %v, want [a b]", values)
	}

	replacement := "d"
	device.SetMetadata("keywords", &replacement)
	if values := device.GetMetadataListByName("keywords"); len(values) != 1 || values[0] != "d" {
		t.Errorf("after SetMetadata, GetMetadataListByName(keywords) = %v, want [d]", values)
	}

	// The slot the second keywords pair left is used for a new name.
	subject := "e"
	device.SetMetadata("subject", &subject)
	if value := device.GetMetadataByName("subject"); value == nil || *value != "e" {
		t.Errorf("GetMetadataByName(subject) = %v, want e", value)
	}

	device.SetMetadata("keywords", nil)
	if values := device.GetMetadataListByName("keywords"); len(values) != 0 {
		t.Errorf("after removal, GetMetadataListByName(keywords) = %v, want none", values)
	}
	if value := device.GetMetadataByName("author"); value == nil || *value != "c" {
		t.Errorf("GetMetadataByName(author) = %v, want c", value)
	}
}

// htmlOutlineTestBox is a box tree node with an element and children; every
// other BoxI method panics on the nil embedded interface.
type htmlOutlineTestBox struct {
	ufo.BoxI
	element  *dom.Element
	children []ufo.BoxI
}

func (b *htmlOutlineTestBox) GetElement() *dom.Element { return b.element }
func (b *htmlOutlineTestBox) GetChildCount() int       { return len(b.children) }
func (b *htmlOutlineTestBox) GetChild(i int) ufo.BoxI  { return b.children[i] }

func htmlOutlineTestBoxTree(e *dom.Element) *htmlOutlineTestBox {
	box := &htmlOutlineTestBox{element: e}
	for _, child := range e.GetChildNodes() {
		if childElement, ok := child.(*dom.Element); ok {
			box.children = append(box.children, htmlOutlineTestBoxTree(childElement))
		}
	}
	return box
}

func htmlOutlineTestDescribe(bookmarks []*ITextOutputDeviceBookmark) string {
	var b strings.Builder
	for i, bookmark := range bookmarks {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(bookmark.GetName())
		if bookmark.GetBox() != nil {
			b.WriteString("@" + bookmark.GetBox().GetElement().GetTagName())
		}
		if len(bookmark.GetChildren()) > 0 {
			b.WriteString(" { " + htmlOutlineTestDescribe(bookmark.GetChildren()) + " }")
		}
	}
	return b.String()
}

func TestHTMLOutline_generate(t *testing.T) {
	// The document of the HTMLOutlineGenerate comment, with the
	// customization attributes added.
	doc, err := dom.ParseXMLString(`<html><body>
  <h1>Foo</h1>
  <h3>  Bar
     bar </h3>
  <blockquote>
    <h5>Bla</h5>
  </blockquote>
  <p>Baz</p>
  <h2>Quux</h2>
  <section>
    <h3>Thud</h3>
  </section>
  <h4>Grunt</h4>
  <h3 data-pdf-bookmark="none">Hidden</h3>
  <strong data-pdf-bookmark="4">Strong</strong>
  <p data-pdf-bookmark="2" data-pdf-bookmark-name=" Named ">Text</p>
  <p data-pdf-bookmark="0">Zero</p>
  <p data-pdf-bookmark="x">Invalid</p>
</body></html>`)
	if err != nil {
		t.Fatal(err)
	}
	root := doc.GetDocumentElement()

	bookmarks := HTMLOutlineGenerate(root, htmlOutlineTestBoxTree(root))

	// The heading inside the blockquote is part of the outline: a
	// NodeIterator does not leave out the subtree of a rejected node.
	expected := "Foo@h1 { Bar bar@h3 { Bla@h5 }, Quux@h2 { Thud@h3 { Grunt@h4, Strong@strong } }, Named@p }"
	if actual := htmlOutlineTestDescribe(bookmarks); actual != expected {
		t.Errorf("outline = %s, want %s", actual, expected)
	}
}

func iTextOutputDeviceTestFont(t *testing.T) *ITextFSFont {
	t.Helper()
	baseFont, err := writer.CreateFont(writer.BaseFontHelvetica, writer.BaseFontCp1252, writer.BaseFontNotEmbedded)
	if err != nil {
		t.Fatal(err)
	}
	return NewITextFSFont(NewFontDescriptionWithStyleWeight(baseFont, ufo.IdentValueNormal, 400), 240)
}

func TestITextOutputDevice_drawStringPlacesTheBaselineInPoints(t *testing.T) {
	device, content := iTextOutputDeviceTestPage(t)
	device.SetFont(iTextOutputDeviceTestFont(t))

	device.DrawString("Hi", 200, 400, nil)

	// 240 dots at 20 dots per point is a 12 pt font; (200, 400) dots is
	// (10, 20) pt from the top left, 80 pt above the bottom of the page.
	actual := iTextOutputDeviceTestOperators(content)
	for _, expected := range []string{"BT", " 12 Tf", "1 0 0 1 10 80 Tm", "(Hi)Tj ET"} {
		if !strings.Contains(actual, expected) {
			t.Errorf("content = %q, want it to contain %q", actual, expected)
		}
	}
	if strings.Contains(actual, " Tr") {
		t.Errorf("content = %q, want no text rendering mode without a font specification", actual)
	}
}

func TestITextOutputDevice_drawStringSimulatesBoldAndItalic(t *testing.T) {
	device, content := iTextOutputDeviceTestPage(t)
	device.SetFont(iTextOutputDeviceTestFont(t))
	device.SetFontSpecification(ufo.NewFontSpecification(240, ufo.IdentValueBold, []string{"Helvetica"}, ufo.IdentValueItalic, ufo.IdentValueNormal))

	device.DrawString("Hi", 200, 400, nil)

	// Bold: fill and stroke with a line 4% of the font size wide. Italic: a
	// shear of 0.21256 in the text matrix.
	actual := iTextOutputDeviceTestOperators(content)
	for _, expected := range []string{"2 Tr 0.48 w", "1 0 0.21256 1 10 80 Tm", "(Hi)Tj 0 Tr 1 w ET"} {
		if !strings.Contains(actual, expected) {
			t.Errorf("content = %q, want it to contain %q", actual, expected)
		}
	}
}

func TestITextOutputDevice_drawStringJustified(t *testing.T) {
	device, content := iTextOutputDeviceTestPage(t)
	device.SetFont(iTextOutputDeviceTestFont(t))

	device.DrawString("a b", 0, 0, ufo.NewJustificationInfo(24, 48))

	// An adjustment in dots becomes thousandths of the font size: 24 dots of
	// a 240 dot font is 100, 48 dots is 200. Nothing follows the last
	// character.
	actual := iTextOutputDeviceTestOperators(content)
	if !strings.Contains(actual, "[(a)-100( )-200(b)]TJ") {
		t.Errorf("content = %q, want the adjustments -100 and -200 between the characters", actual)
	}
}
