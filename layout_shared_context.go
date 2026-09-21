// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/SharedContext.java

package ufo

import (
	"fmt"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

var sharedContextPagedMediaTypes = map[string]struct{}{
	"print":      {},
	"projection": {},
	"embossed":   {},
	"handheld":   {},
	"tv":         {},
}

const sharedContextMmPerCm = 10

const sharedContextCmPerIn float32 = 2.54

const sharedContextDefaultDpi float32 = 72

// SharedContext is that which is kept between successive layout and render
// runs.
type SharedContext struct {
	textRenderer TextRenderer
	media        string
	uac          UserAgentCallback
	interactive  bool
	idMap        map[string]BoxI

	// used to adjust fonts, ems, points, into screen resolution
	dpi float32
	// dpi in a more usable way
	mmPerDot float32

	print bool

	dotsPerPixel int

	// styleMap is nil until the first GetStyle call after a Reset.
	styleMap map[*dom.Element]CalculatedStyleI

	replacedElementFactory ReplacedElementFactory
	temporaryCanvas        *geom.Rectangle
	lineBreakingStrategy   LineBreakingStrategy

	// unsupportedTags is a LinkedHashSet in Java: unsupportedTagsOrder keeps
	// the insertion order, which the log message of LogUnsupportedFeatures
	// shows.
	unsupportedTags      map[string]struct{}
	unsupportedTagsOrder []string

	fontResolver FontResolver

	css                  *StyleReference
	debugDrawBoxes       bool
	debugDrawLineBoxes   bool
	debugDrawInlineBoxes bool
	debugDrawFontMetrics bool

	canvas FSCanvas

	namespaceHandler NamespaceHandler
}

func newSharedContextFields() *SharedContext {
	return &SharedContext{
		interactive:          true,
		idMap:                map[string]BoxI{},
		dotsPerPixel:         1,
		lineBreakingStrategy: NewDefaultLineBreakingStrategy(),
		unsupportedTags:      map[string]struct{}{},
	}
}

// NewSharedContext is Java's no-argument constructor. Java passes a
// NaiveUserAgent to the one-argument constructor, which installs an
// AWTFontResolver, a SwingReplacedElementFactory and a Java2DTextRenderer and
// asks java.awt.Toolkit for the screen resolution. None of those Swing classes
// is ported: the user agent, font resolver, replaced element factory and text
// renderer are left nil for the caller to set, and the DPI is the value Java
// uses when the toolkit is headless.
func NewSharedContext() *SharedContext {
	s := newSharedContextFields()
	s.media = "screen"
	s.css = NewStyleReference(nil)
	XRLogRender("Using CSS implementation from: org.xhtmlrenderer.context.StyleReference")
	s.SetDPI(sharedContextDefaultDpi)
	return s
}

func NewSharedContextWithUserAgentFontResolverReplacedElementFactoryTextRendererDpiDotsPerPixel(
	userAgent UserAgentCallback, fontResolver FontResolver,
	replacedElementFactory ReplacedElementFactory,
	textRenderer TextRenderer,
	dpi float32, dotsPerPixel int) *SharedContext {
	s := newSharedContextFields()
	if userAgent == nil {
		panic(NewXRRuntimeException("userAgent may not be null"))
	}
	s.uac = userAgent
	s.css = NewStyleReference(userAgent)
	if fontResolver == nil {
		panic(NewXRRuntimeException("fontResolver may not be null"))
	}
	s.fontResolver = fontResolver
	s.replacedElementFactory = replacedElementFactory
	if textRenderer == nil {
		panic(NewXRRuntimeException("textRenderer may not be null"))
	}
	s.textRenderer = textRenderer
	s.media = "screen"
	s.SetDPI(dpi)
	s.SetDotsPerPixel(dotsPerPixel)
	s.SetPrint(true)
	s.SetInteractive(false)
	return s
}

func NewSharedContextWithUacDpiPixelsPerDot(uac UserAgentCallback, dpi float32, pixelsPerDot int) *SharedContext {
	s := NewSharedContextWithUac(uac)
	s.SetDPI(dpi)
	s.SetDotsPerPixel(pixelsPerDot)
	return s
}

// NewSharedContextWithUac is Java's SharedContext(UserAgentCallback). The
// Swing defaults (AWTFontResolver, SwingReplacedElementFactory,
// Java2DTextRenderer) are not ported, so the font resolver, the replaced
// element factory and the text renderer are nil until the caller sets them.
// The DPI is the value Java uses when java.awt.Toolkit is headless.
func NewSharedContextWithUac(uac UserAgentCallback) *SharedContext {
	s := newSharedContextFields()
	s.media = "screen"
	if uac == nil {
		panic(NewXRRuntimeException("uac may not be null"))
	}
	s.uac = uac
	s.css = NewStyleReference(uac)
	XRLogRender("Using CSS implementation from: org.xhtmlrenderer.context.StyleReference")
	s.SetDPI(sharedContextDefaultDpi)
	return s
}

func NewSharedContextWithUacFrRefTrDpi(uac UserAgentCallback, fr FontResolver, ref ReplacedElementFactory, tr TextRenderer, dpi float32) *SharedContext {
	s := newSharedContextFields()
	if fr == nil {
		panic(NewXRRuntimeException("fr may not be null"))
	}
	s.fontResolver = fr
	s.replacedElementFactory = ref
	s.media = "screen"
	s.uac = uac
	s.css = NewStyleReference(uac)
	XRLogRender("Using CSS implementation from: org.xhtmlrenderer.context.StyleReference")
	if tr == nil {
		panic(NewXRRuntimeException("tr may not be null"))
	}
	s.textRenderer = tr
	s.SetDPI(dpi)
	s.SetPrint(true)
	s.SetInteractive(false)
	return s
}

// setFormSubmissionListener(FormSubmissionListener) is not ported:
// FormSubmissionListener belongs to the Swing form support
// (simple/extend/form), and ReplacedElementFactory does not declare the
// method in Go.

func (s *SharedContext) NewLayoutContextInstance(fontContext FontContext) *LayoutContext {
	return NewLayoutContext(s, fontContext)
}

func (s *SharedContext) NewRenderingContextInstance(outputDevice OutputDevice, fontContext FontContext) *RenderingContext {
	return s.NewRenderingContextInstanceWithRootLayerInitialPageNo(outputDevice, fontContext, nil, 0)
}

// NewRenderingContextInstanceWithRootLayerInitialPageNo accepts a nil
// rootLayer.
func (s *SharedContext) NewRenderingContextInstanceWithRootLayerInitialPageNo(outputDevice OutputDevice, fontContext FontContext, rootLayer *Layer, initialPageNo int) *RenderingContext {
	return NewRenderingContext(s, outputDevice, fontContext, rootLayer, initialPageNo)
}

/*
=========== Font stuff ============== */

// GetFontResolver gets the fontResolver attribute of the Context object.
func (s *SharedContext) GetFontResolver() FontResolver {
	return s.fontResolver
}

func (s *SharedContext) FlushFonts() {
	s.fontResolver.FlushCache()
}

// GetMedia returns the media for this context.
func (s *SharedContext) GetMedia() string {
	return s.media
}

func (s *SharedContext) GetTextRenderer() TextRenderer {
	return s.textRenderer
}

func (s *SharedContext) DebugDrawBoxes() bool {
	return s.debugDrawBoxes
}

func (s *SharedContext) DebugDrawLineBoxes() bool {
	return s.debugDrawLineBoxes
}

func (s *SharedContext) DebugDrawInlineBoxes() bool {
	return s.debugDrawInlineBoxes
}

func (s *SharedContext) DebugDrawFontMetrics() bool {
	return s.debugDrawFontMetrics
}

// SetDebug_draw_boxes is Java's setDebug_draw_boxes.
func (s *SharedContext) SetDebug_draw_boxes(debugDrawBoxes bool) {
	s.debugDrawBoxes = debugDrawBoxes
}

// SetDebug_draw_line_boxes is Java's setDebug_draw_line_boxes.
func (s *SharedContext) SetDebug_draw_line_boxes(debugDrawLineBoxes bool) {
	s.debugDrawLineBoxes = debugDrawLineBoxes
}

// SetDebug_draw_inline_boxes is Java's setDebug_draw_inline_boxes.
func (s *SharedContext) SetDebug_draw_inline_boxes(debugDrawInlineBoxes bool) {
	s.debugDrawInlineBoxes = debugDrawInlineBoxes
}

// SetDebug_draw_font_metrics is Java's setDebug_draw_font_metrics.
func (s *SharedContext) SetDebug_draw_font_metrics(debugDrawFontMetrics bool) {
	s.debugDrawFontMetrics = debugDrawFontMetrics
}

/*
=========== Selection Management ============== */

func (s *SharedContext) GetCss() *StyleReference {
	return s.css
}

// Deprecated: SetCss is deprecated in Java.
func (s *SharedContext) SetCss(css *StyleReference) {
	s.css = css
}

// GetCanvas may return nil.
func (s *SharedContext) GetCanvas() FSCanvas {
	return s.canvas
}

func (s *SharedContext) SetCanvas(canvas FSCanvas) {
	s.canvas = canvas
}

func (s *SharedContext) SetTemporaryCanvas(rect *geom.Rectangle) {
	s.temporaryCanvas = rect
}

// GetFixedRectangle may return nil.
func (s *SharedContext) GetFixedRectangle() *geom.Rectangle {
	if s.GetCanvas() == nil {
		return s.temporaryCanvas
	} else {
		rect := s.GetCanvas().GetFixedRectangle()
		rect.Translate(s.GetCanvas().GetX(), s.GetCanvas().GetY())
		return rect
	}
}

func (s *SharedContext) SetNamespaceHandler(nh NamespaceHandler) {
	s.namespaceHandler = nh
}

// GetNamespaceHandler may return nil.
func (s *SharedContext) GetNamespaceHandler() NamespaceHandler {
	return s.namespaceHandler
}

func (s *SharedContext) AddBoxId(id string, box BoxI) {
	s.idMap[id] = box
}

// GetBoxById returns nil when no box has the id.
func (s *SharedContext) GetBoxById(id string) BoxI {
	box, ok := s.idMap[id]
	if !ok {
		return nil
	}
	return box
}

func (s *SharedContext) RemoveBoxId(id string) {
	delete(s.idMap, id)
}

func (s *SharedContext) GetIdMap() map[string]BoxI {
	return s.idMap
}

// SetTextRenderer sets the textRenderer attribute of the RenderingContext
// object.
//
// Deprecated: pass textRenderer to a constructor instead of using setter
func (s *SharedContext) SetTextRenderer(textRenderer TextRenderer) {
	if textRenderer == nil {
		panic(NewXRRuntimeException("textRenderer may not be null"))
	}
	s.textRenderer = textRenderer
}

// SetMedia sets the current media type. This is usually something like screen
// or print. See the media section (http://www.w3.org/TR/CSS21/media.html) of
// the CSS 2.1 spec for more information on media types.
func (s *SharedContext) SetMedia(media string) {
	s.media = media
}

// GetUac gets the uac attribute of the RenderingContext object.
func (s *SharedContext) GetUac() UserAgentCallback {
	return s.uac
}

func (s *SharedContext) GetUserAgentCallback() UserAgentCallback {
	return s.uac
}

func (s *SharedContext) SetUserAgentCallback(userAgentCallback UserAgentCallback) {
	s.GetCss().SetUserAgentCallback(userAgentCallback)
	if userAgentCallback == nil {
		panic(NewXRRuntimeException("userAgentCallback may not be null"))
	}
	s.uac = userAgentCallback
}

// GetDPI gets the dPI attribute of the RenderingContext object.
func (s *SharedContext) GetDPI() float32 {
	return s.dpi
}

// SetDPI sets the effective DPI (Dots Per Inch) of the screen. You can
// override the value if you want to scale the fonts for accessibility or
// printing purposes. Currently, the DPI setting only affects font sizing.
func (s *SharedContext) SetDPI(dpi float32) {
	s.dpi = dpi
	cmPerIn := sharedContextCmPerIn
	mmPerCm := float32(sharedContextMmPerCm)
	s.mmPerDot = cmPerIn * mmPerCm / dpi
}

// GetMmPerPx gets the dPI attribute in a more useful form of the
// RenderingContext object.
func (s *SharedContext) GetMmPerPx() float32 {
	return s.mmPerDot
}

// GetFont may return nil.
func (s *SharedContext) GetFont(spec *FontSpecification) FSFont {
	return s.fontResolver.ResolveFont(s, spec)
}

// GetXHeight: strike-through offset should always be half of the height of
// lowercase x... and it is defined even for fonts without 'x'!
func (s *SharedContext) GetXHeight(fontContext FontContext, fs *FontSpecification) float32 {
	font := s.fontResolver.ResolveFont(s, fs)
	fm := s.GetTextRenderer().GetFSFontMetrics(fontContext, font, " ")
	sto := fm.GetStrikethroughOffset()
	if sto < 0 {
		sto = -sto
	}
	return fm.GetAscent() - 2*sto + fm.GetStrikethroughThickness()
}

// GetBaseURL gets the baseURL attribute of the RenderingContext object.
func (s *SharedContext) GetBaseURL() string {
	return s.uac.GetBaseURL()
}

// SetBaseURL sets the baseURL attribute of the RenderingContext object.
func (s *SharedContext) SetBaseURL(url string) {
	s.uac.SetBaseURL(url)
}

// IsPaged returns true if the currently set media type is paged. Currently,
// returns true only for print, projection, and embossed, handheld, and tv. See
// the media section (http://www.w3.org/TR/CSS21/media.html) of the CSS 2.1
// spec for more information on media types.
func (s *SharedContext) IsPaged() bool {
	_, ok := sharedContextPagedMediaTypes[s.media]
	return ok
}

func (s *SharedContext) IsInteractive() bool {
	return s.interactive
}

func (s *SharedContext) SetInteractive(interactive bool) {
	s.interactive = interactive
}

func (s *SharedContext) IsPrint() bool {
	return s.print
}

func (s *SharedContext) SetPrint(print bool) {
	s.print = print
	if print {
		s.SetMedia("print")
	} else {
		s.SetMedia("screen")
	}
}

// setFontMapping(String, java.awt.Font) is not ported: it acts only on an
// AWTFontResolver and takes a java.awt.Font.

// SetFontResolver replaces the font resolver.
//
// Deprecated: pass resolver to a constructor instead of using setter
func (s *SharedContext) SetFontResolver(resolver FontResolver) {
	if resolver == nil {
		panic(NewXRRuntimeException("resolver may not be null"))
	}
	s.fontResolver = resolver
}

func (s *SharedContext) GetDotsPerPixel() int {
	return s.dotsPerPixel
}

func (s *SharedContext) SetDotsPerPixel(pixelsPerDot int) {
	s.dotsPerPixel = pixelsPerDot
}

func (s *SharedContext) GetStyle(e *dom.Element) CalculatedStyleI {
	return s.GetStyleWithRestyle(e, false)
}

func (s *SharedContext) GetStyleWithRestyle(e *dom.Element, restyle bool) CalculatedStyleI {
	localMap := s.styleMap

	if localMap == nil {
		localMap = make(map[*dom.Element]CalculatedStyleI, 1024)
	}

	var result CalculatedStyleI
	if !restyle {
		result = localMap[e]
	}
	if result == nil {
		parent := e.GetParentNode()
		var parentCalculatedStyle CalculatedStyleI
		switch p := parent.(type) {
		case *dom.Document:
			parentCalculatedStyle = NewEmptyStyle()
		case *dom.Element:
			// As in Java, when styleMap is still nil here the recursive call
			// fills a map of its own, which the assignment below replaces.
			parentCalculatedStyle = s.GetStyleWithRestyle(p, false)
		default:
			panic(NewXRRuntimeException(fmt.Sprintf("Unexpected parent: %T", parent)))
		}

		result = parentCalculatedStyle.DeriveStyle(s.GetCss().GetCascadedStyle(e, restyle))

		localMap[e] = result
	}

	s.styleMap = localMap

	return result
}

func (s *SharedContext) Reset() {
	//have to do this first
	if ConfigurationIsTrue("xr.cache.stylesheets", true) {
		s.css.FlushStyleSheets()
	} else {
		s.css.FlushAllStyleSheets()
	}
	s.styleMap = nil
	clear(s.idMap)
	s.replacedElementFactory.Reset()
	clear(s.unsupportedTags)
	s.unsupportedTagsOrder = nil
}

func (s *SharedContext) GetReplacedElementFactory() ReplacedElementFactory {
	return s.replacedElementFactory
}

func (s *SharedContext) SetReplacedElementFactory(ref ReplacedElementFactory) {
	if ref == nil {
		panic(NewXRRuntimeException("replacedElementFactory may not be null"))
	}

	s.replacedElementFactory.Reset()
	s.replacedElementFactory = ref
}

func (s *SharedContext) GetLineBreakingStrategy() LineBreakingStrategy {
	return s.lineBreakingStrategy
}

func (s *SharedContext) SetLineBreakingStrategy(lineBreakingStrategy LineBreakingStrategy) {
	s.lineBreakingStrategy = lineBreakingStrategy
}

func (s *SharedContext) AddUnsupportedTag(tagName string) {
	if _, ok := s.unsupportedTags[tagName]; ok {
		return
	}
	s.unsupportedTags[tagName] = struct{}{}
	s.unsupportedTagsOrder = append(s.unsupportedTagsOrder, tagName)
}

// GetUnsupportedTags returns the tags in the order they were first added
// (Java returns a LinkedHashSet).
func (s *SharedContext) GetUnsupportedTags() []string {
	return s.unsupportedTagsOrder
}

func (s *SharedContext) LogUnsupportedFeatures() {
	Html5SupportLogUnsupportedFeatures(s.unsupportedTagsOrder, s.GetCss().GetUnsupportedCssFeatures())
}
