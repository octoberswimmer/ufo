package ufo

import (
	"math"
	"reflect"
	"testing"

	"github.com/octoberswimmer/ufo/geom"
)

type sharedContextTestFont struct{}

func (sharedContextTestFont) GetSize2D() float32 { return 12 }

type sharedContextTestFontResolver struct {
	flushed int
}

func (r *sharedContextTestFontResolver) ResolveFont(renderingContext *SharedContext, spec *FontSpecification) FSFont {
	return sharedContextTestFont{}
}

func (r *sharedContextTestFontResolver) FlushCache() { r.flushed++ }

type sharedContextTestFontMetrics struct {
	FSFontMetrics
}

func (sharedContextTestFontMetrics) GetAscent() float32                 { return 10 }
func (sharedContextTestFontMetrics) GetStrikethroughOffset() float32    { return -3 }
func (sharedContextTestFontMetrics) GetStrikethroughThickness() float32 { return 0.5 }

// sharedContextTestTextRenderer implements only GetFSFontMetrics; the embedded
// nil interface supplies the rest of the method set.
type sharedContextTestTextRenderer struct {
	TextRenderer
	lastString string
}

func (r *sharedContextTestTextRenderer) GetFSFontMetrics(context FontContext, font FSFont, str string) FSFontMetrics {
	r.lastString = str
	return sharedContextTestFontMetrics{}
}

type sharedContextTestReplacedElementFactory struct {
	ReplacedElementFactory
	resets int
}

func (f *sharedContextTestReplacedElementFactory) Reset() { f.resets++ }

type sharedContextTestCanvas struct{}

func (sharedContextTestCanvas) GetFixedRectangle() *geom.Rectangle {
	return geom.NewRectangle(1, 2, 30, 40)
}
func (sharedContextTestCanvas) GetX() int { return 100 }
func (sharedContextTestCanvas) GetY() int { return 200 }

func TestSharedContext_defaults(t *testing.T) {
	for name, sc := range map[string]*SharedContext{
		"NewSharedContext":        NewSharedContext(),
		"NewSharedContextWithUac": NewSharedContextWithUac(&layoutContextTestUserAgent{}),
	} {
		// The values Java reports for new SharedContext(new NaiveUserAgent())
		// when java.awt.Toolkit is headless.
		if sc.GetDPI() != 72 || sc.GetMmPerPx() != 0.35277778 || sc.GetMedia() != "screen" || sc.IsPaged() ||
			sc.IsPrint() || !sc.IsInteractive() || sc.GetDotsPerPixel() != 1 {
			t.Errorf("%s: dpi=%v mm=%v media=%q paged=%v print=%v interactive=%v dotsPerPixel=%d", name,
				sc.GetDPI(), sc.GetMmPerPx(), sc.GetMedia(), sc.IsPaged(), sc.IsPrint(), sc.IsInteractive(), sc.GetDotsPerPixel())
		}
		if sc.GetCss() == nil || sc.GetLineBreakingStrategy() == nil {
			t.Errorf("%s: css and line breaking strategy must be set", name)
		}
		if sc.GetFontResolver() != nil || sc.GetTextRenderer() != nil || sc.GetReplacedElementFactory() != nil {
			t.Errorf("%s: the Swing defaults are not ported, so these collaborators must be nil", name)
		}
	}
}

func TestSharedContext_setDPI(t *testing.T) {
	// Float.floatToIntBits(getMmPerPx()) from the Java class.
	expected := []struct {
		dpi  float32
		bits uint32
	}{
		{72, 1052024650},
		{96, 1049065335},
		{1440, 1016102766},
		{1600, 1015155786},
		{300, 1034773943},
		{7, 1080572547},
	}
	sc := NewSharedContext()
	for _, e := range expected {
		sc.SetDPI(e.dpi)
		if got := math.Float32bits(sc.GetMmPerPx()); got != e.bits || sc.GetDPI() != e.dpi {
			t.Errorf("SetDPI(%v): mmPerPx bits %d, want %d", e.dpi, got, e.bits)
		}
	}
}

func TestSharedContext_rendererConstructor(t *testing.T) {
	fontResolver := &sharedContextTestFontResolver{}
	textRenderer := &sharedContextTestTextRenderer{}
	factory := &sharedContextTestReplacedElementFactory{}
	sc := NewSharedContextWithUserAgentFontResolverReplacedElementFactoryTextRendererDpiDotsPerPixel(
		&layoutContextTestUserAgent{}, fontResolver, factory, textRenderer, 72*20, 20)
	if !sc.IsPrint() || sc.GetMedia() != "print" || !sc.IsPaged() || sc.IsInteractive() ||
		sc.GetDPI() != 1440 || sc.GetDotsPerPixel() != 20 {
		t.Errorf("print=%v media=%q paged=%v interactive=%v dpi=%v dotsPerPixel=%d",
			sc.IsPrint(), sc.GetMedia(), sc.IsPaged(), sc.IsInteractive(), sc.GetDPI(), sc.GetDotsPerPixel())
	}
	sc.SetMedia("pdf")
	if sc.IsPaged() {
		t.Error(`media "pdf" is not a paged media type`)
	}
	sc.SetPrint(false)
	if sc.GetMedia() != "screen" {
		t.Errorf("SetPrint(false) must set the media to screen, got %q", sc.GetMedia())
	}

	// ascent - 2*|strikethrough offset| + strikethrough thickness
	if got := sc.GetXHeight(nil, nil); got != 4.5 || textRenderer.lastString != " " {
		t.Errorf("GetXHeight = %v (metrics asked for %q), want 4.5 for \" \"", got, textRenderer.lastString)
	}
	c := sc.NewLayoutContextInstance(nil)
	if got := c.GetFontSize2D(nil); got != 12 {
		t.Errorf("GetFontSize2D = %v, want 12", got)
	}
	if c.GetFSFontMetrics(sharedContextTestFont{}); textRenderer.lastString != "" {
		t.Errorf("LayoutContext.GetFSFontMetrics asked for %q, want the empty string", textRenderer.lastString)
	}

	sc.FlushFonts()
	if fontResolver.flushed != 1 {
		t.Errorf("FlushFonts flushed the resolver %d times", fontResolver.flushed)
	}

	replacement := &sharedContextTestReplacedElementFactory{}
	sc.SetReplacedElementFactory(replacement)
	if factory.resets != 1 || sc.GetReplacedElementFactory() != ReplacedElementFactory(replacement) {
		t.Errorf("SetReplacedElementFactory must reset the old factory once, got %d", factory.resets)
	}

	func() {
		defer func() {
			if recover() == nil {
				t.Error("a nil font resolver must panic")
			}
		}()
		NewSharedContextWithUserAgentFontResolverReplacedElementFactoryTextRendererDpiDotsPerPixel(
			&layoutContextTestUserAgent{}, nil, factory, textRenderer, 72, 1)
	}()
}

func TestSharedContext_fixedRectangle(t *testing.T) {
	sc := NewSharedContext()
	if sc.GetFixedRectangle() != nil {
		t.Error("without a canvas and a temporary canvas the fixed rectangle is nil")
	}
	temporary := geom.NewRectangle(0, 0, 10, 20)
	sc.SetTemporaryCanvas(temporary)
	if sc.GetFixedRectangle() != temporary {
		t.Error("without a canvas the fixed rectangle is the temporary canvas")
	}
	sc.SetCanvas(sharedContextTestCanvas{})
	if got := sc.GetFixedRectangle(); !got.Equals(geom.NewRectangle(101, 202, 30, 40)) {
		t.Errorf("fixed rectangle %v, want the canvas rectangle moved by the canvas location", got)
	}
}

func TestSharedContext_unsupportedTagsKeepInsertionOrder(t *testing.T) {
	sc := NewSharedContext()
	for _, tag := range []string{"video", "audio", "video", "canvas", "audio"} {
		sc.AddUnsupportedTag(tag)
	}
	if got := sc.GetUnsupportedTags(); !reflect.DeepEqual(got, []string{"video", "audio", "canvas"}) {
		t.Errorf("unsupported tags %v", got)
	}
}

func TestSharedContext_getStyleAndReset(t *testing.T) {
	sc, doc := layoutContextTestSharedContext(t, layoutContextTestCounterDocument)
	factory := &sharedContextTestReplacedElementFactory{}
	sc.replacedElementFactory = factory

	body := doc.GetElementsByTagName("body")[0]
	h1 := doc.GetElementsByTagName("h1")[0]
	bodyStyle := sc.GetStyle(body)
	h1Style := sc.GetStyle(h1)
	if sc.GetStyle(h1) != h1Style {
		t.Error("GetStyle must return the cached style")
	}
	if h1Style.GetParent() != bodyStyle {
		t.Error("the parent of an element's style is the style of its parent element")
	}
	htmlStyle := sc.GetStyle(doc.GetDocumentElement())
	if _, ok := htmlStyle.GetParent().(*EmptyStyle); !ok {
		t.Errorf("the root element's style derives from an EmptyStyle, got %T", htmlStyle.GetParent())
	}
	if !h1Style.IsIdent(CSSNameDisplay, IdentValueBlock) {
		t.Error("h1 must be display: block from the default stylesheet")
	}

	restyled := sc.GetStyleWithRestyle(h1, true)
	if sc.GetStyle(h1) != restyled {
		t.Error("a restyle replaces the cached style")
	}

	sc.AddUnsupportedTag("video")
	sc.Reset()
	if factory.resets != 1 || len(sc.GetUnsupportedTags()) != 0 || len(sc.GetIdMap()) != 0 || sc.styleMap != nil {
		t.Errorf("Reset left state behind: resets=%d tags=%v", factory.resets, sc.GetUnsupportedTags())
	}
}
