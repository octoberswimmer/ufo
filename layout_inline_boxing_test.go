// Tests InlineBoxing and BlockBoxing
// (flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/InlineBoxing.java, BlockBoxing.java)
// against box trees laid out by the Java implementation. Flying Saucer has no
// JUnit test for these classes in the core module.

package ufo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

type inlineBoxingTestUserAgent struct {
	baseURL string
}

func (u *inlineBoxingTestUserAgent) GetCSSResource(uri string) *CSSResource     { return nil }
func (u *inlineBoxingTestUserAgent) GetImageResource(uri string) *ImageResource { return nil }
func (u *inlineBoxingTestUserAgent) GetXMLResource(uri string) *XMLResource     { return nil }
func (u *inlineBoxingTestUserAgent) GetBinaryResource(uri string) []byte        { return nil }
func (u *inlineBoxingTestUserAgent) IsVisited(uri string) bool                  { return false }
func (u *inlineBoxingTestUserAgent) SetBaseURL(url string)                      { u.baseURL = url }
func (u *inlineBoxingTestUserAgent) GetBaseURL() string                         { return u.baseURL }
func (u *inlineBoxingTestUserAgent) ResolveURI(uri string) string               { return uri }

// inlineBoxingTestFont is a font of which only the size is known.
type inlineBoxingTestFont struct{ size float32 }

func (f inlineBoxingTestFont) GetSize2D() float32 { return f.size }

type inlineBoxingTestFontResolver struct{}

func (inlineBoxingTestFontResolver) ResolveFont(renderingContext *SharedContext, spec *FontSpecification) FSFont {
	return inlineBoxingTestFont{size: spec.Size()}
}
func (inlineBoxingTestFontResolver) FlushCache() {}

// inlineBoxingTestFontMetrics derives every metric from the font size, with
// the same float32 operations as the Java harness.
type inlineBoxingTestFontMetrics struct{ size float32 }

func (m inlineBoxingTestFontMetrics) GetAscent() float32                 { return m.size * 0.8 }
func (m inlineBoxingTestFontMetrics) GetDescent() float32                { return m.size * 0.2 }
func (m inlineBoxingTestFontMetrics) GetStrikethroughOffset() float32    { return -(m.size * 0.3) }
func (m inlineBoxingTestFontMetrics) GetStrikethroughThickness() float32 { return m.size * 0.05 }
func (m inlineBoxingTestFontMetrics) GetUnderlineOffset() float32        { return m.size * 0.1 }
func (m inlineBoxingTestFontMetrics) GetUnderlineThickness() float32     { return m.size * 0.05 }

// inlineBoxingTestTextRenderer makes every character half the font size wide.
type inlineBoxingTestTextRenderer struct{ breakerTestTextRenderer }

func (inlineBoxingTestTextRenderer) GetFSFontMetrics(context FontContext, font FSFont, str string) FSFontMetrics {
	return inlineBoxingTestFontMetrics{size: font.GetSize2D()}
}

func (inlineBoxingTestTextRenderer) GetWidth(context FontContext, font FSFont, str string) int {
	return calculatedStyleFloatToInt(font.GetSize2D()*0.5) * utf8.RuneCountInString(str)
}

type inlineBoxingTestReplacedElementFactory struct{}

func (inlineBoxingTestReplacedElementFactory) CreateReplacedElement(
	c *LayoutContext, box BlockBoxI,
	uac UserAgentCallback, cssWidth int, cssHeight int) ReplacedElement {
	return nil
}
func (inlineBoxingTestReplacedElementFactory) Reset()                {}
func (inlineBoxingTestReplacedElementFactory) Remove(e *dom.Element) {}

func inlineBoxingTestClassName(b BoxI) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", b), "*ufo.")
}

// inlineBoxingTestDump writes the box tree in the format of the Java harness.
func inlineBoxingTestDump(b BoxI, ind string, sb *strings.Builder) {
	fmt.Fprintf(sb, "%s%s %d,%d %dx%d abs %d,%d", ind, inlineBoxingTestClassName(b),
		b.GetX(), b.GetY(), b.GetWidth(), b.GetHeight(), b.GetAbsX(), b.GetAbsY())
	switch box := b.(type) {
	case *LineBox:
		fmt.Fprintf(sb, " baseline %d paint %d+%d cw %d", box.GetBaseline(),
			box.GetPaintingTop(), box.GetPaintingHeight(), box.GetContentWidth())
		if box.IsEndsOnNL() {
			sb.WriteString(" NL")
		}
		sb.WriteString("\n")
		for _, f := range box.GetNonFlowContent() {
			inlineBoxingTestDump(f, ind+"  ~", sb)
		}
		for _, ch := range box.GetChildren() {
			inlineBoxingTestDump(ch, ind+"  ", sb)
		}
	case *InlineLayoutBox:
		fmt.Fprintf(sb, " iw %d baseline %d", box.GetInlineWidth(), box.GetBaseline())
		if box.IsStartsHere() {
			sb.WriteString(" S")
		}
		if box.IsEndsHere() {
			sb.WriteString(" E")
		}
		for _, td := range box.GetTextDecorations() {
			fmt.Fprintf(sb, " td %s@%d/%d", td.GetIdentValue().String(), td.GetOffset(), td.GetThickness())
		}
		sb.WriteString("\n")
		for i := 0; i < box.GetInlineChildCount(); i++ {
			ch := box.GetInlineChild(i)
			if t, ok := ch.(*InlineText); ok {
				fmt.Fprintf(sb, "%s  T %d %d '%s'\n", ind, t.GetX(), t.GetWidth(), breakerTestEscape(t.GetSubstring()))
			} else {
				inlineBoxingTestDump(ch.(BoxI), ind+"  ", sb)
			}
		}
	default:
		sb.WriteString("\n")
		for _, ch := range b.GetChildren() {
			inlineBoxingTestDump(ch, ind+"  ", sb)
		}
	}
}

// TestInlineBoxingLayoutAgainstJava lays out the documents in
// testdata/layout/inline_boxing the way ITextRenderer.layout does, with a font
// resolver, font metrics and a text renderer that derive everything from the
// font size, and compares the box tree with the one a Java harness dumped
// from Flying Saucer with the same fakes (the .txt file next to each
// document). The documents cover line breaking, text-align, text-indent,
// first-line and first-letter styles, vertical-align, text decorations,
// floats, white-space values, word-wrap and word-break, inline blocks, list
// markers, and for BlockBoxing page breaks, page-break-inside/after: avoid,
// relative positioning and multi-column content.
func TestInlineBoxingLayoutAgainstJava(t *testing.T) {
	docs, err := filepath.Glob("testdata/layout/inline_boxing/*.xhtml")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) == 0 {
		t.Fatal("no documents")
	}
	for _, path := range docs {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(strings.TrimSuffix(path, ".xhtml") + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			doc, err := dom.ParseXMLString(string(source))
			if err != nil {
				t.Fatal(err)
			}
			var dotsPerPoint float32 = 20.0 * 4.0 / 3.0
			sc := NewSharedContextWithUserAgentFontResolverReplacedElementFactoryTextRendererDpiDotsPerPixel(
				&inlineBoxingTestUserAgent{}, inlineBoxingTestFontResolver{},
				inlineBoxingTestReplacedElementFactory{}, inlineBoxingTestTextRenderer{},
				72*dotsPerPoint, 20)
			sc.SetBaseURL("")
			sc.SetNamespaceHandler(NewXhtmlNamespaceHandler())
			sc.GetCss().SetDocumentContext(sc, sc.GetNamespaceHandler(), doc, nil)
			c := sc.NewLayoutContextInstance(nil)
			root := BoxBuilderCreateRootBox(c, doc)
			first := LayerCreatePageBox(c, "first")
			root.SetContainingBlock(NewViewportBox(geom.NewRectangle(0, 0, first.GetContentWidth(c), first.GetContentHeight(c))))
			root.Layout(c)

			var sb strings.Builder
			inlineBoxingTestDump(root, "", &sb)
			fmt.Fprintf(&sb, "pages %d\n", len(root.GetLayer().GetPages()))

			got := sb.String()
			if got != string(want) {
				gotLines := strings.Split(got, "\n")
				wantLines := strings.Split(string(want), "\n")
				for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
					var g, w string
					if i < len(gotLines) {
						g = gotLines[i]
					}
					if i < len(wantLines) {
						w = wantLines[i]
					}
					if g != w {
						t.Fatalf("line %d differs from Java:\n got  %s\n want %s", i+1, g, w)
					}
				}
			}
		})
	}
}
