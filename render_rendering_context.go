// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/RenderingContext.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// RenderingContext supplies information about the context in which rendering
// will take place.
//
// Author: jmarinacci, November 16, 2004
type RenderingContext struct {
	sharedContext *SharedContext
	outputDevice  OutputDevice
	fontContext   FontContext
	pageCount     int
	pageNo        int

	// page may be nil.
	page *PageBox

	// rootLayer may be nil.
	rootLayer *Layer

	initialPageNo int
}

var _ CssContext = (*RenderingContext)(nil)

// NewRenderingContext needs a new instance every run. rootLayer may be nil.
func NewRenderingContext(sharedContext *SharedContext, outputDevice OutputDevice, fontContext FontContext,
	rootLayer *Layer,
	initialPageNo int) *RenderingContext {
	return &RenderingContext{
		sharedContext: sharedContext,
		outputDevice:  outputDevice,
		fontContext:   fontContext,
		rootLayer:     rootLayer,
		initialPageNo: initialPageNo,
	}
}

func (r *RenderingContext) GetUac() UserAgentCallback {
	return r.sharedContext.GetUac()
}

func (r *RenderingContext) GetBaseURL() string {
	return r.sharedContext.GetBaseURL()
}

func (r *RenderingContext) GetDPI() float32 {
	return r.sharedContext.GetDPI()
}

func (r *RenderingContext) GetMmPerDot() float32 {
	return r.sharedContext.GetMmPerPx()
}

func (r *RenderingContext) GetDotsPerPixel() int {
	return r.sharedContext.GetDotsPerPixel()
}

func (r *RenderingContext) GetFontSize2D(font *FontSpecification) float32 {
	return r.sharedContext.GetFont(font).GetSize2D()
}

func (r *RenderingContext) GetXHeight(parentFont *FontSpecification) float32 {
	return r.sharedContext.GetXHeight(r.GetFontContext(), parentFont)
}

func (r *RenderingContext) GetTextRenderer() TextRenderer {
	return r.sharedContext.GetTextRenderer()
}

// IsPaged returns true if the currently set media type is paged. Currently,
// returns true only for print, projection, and embossed, handheld, and tv. See
// the media section (http://www.w3.org/TR/CSS21/media.html) of the CSS 2.1
// spec for more information on media types.
func (r *RenderingContext) IsPaged() bool {
	return r.sharedContext.IsPaged()
}

func (r *RenderingContext) GetFontResolver() FontResolver {
	return r.sharedContext.GetFontResolver()
}

// GetFont may return nil.
func (r *RenderingContext) GetFont(font *FontSpecification) FSFont {
	return r.sharedContext.GetFont(font)
}

// GetCanvas may return nil.
func (r *RenderingContext) GetCanvas() FSCanvas {
	return r.sharedContext.GetCanvas()
}

func (r *RenderingContext) GetFixedRectangle() *geom.Rectangle {
	var result *geom.Rectangle
	if r.IsPrint() {
		result = geom.NewRectangle(0, -r.page.GetTop(), r.page.GetContentWidth(r), r.page.GetContentHeight(r)-1)
	} else {
		result = r.sharedContext.GetFixedRectangle()
	}
	result.Translate(-1, -1)
	return result
}

func (r *RenderingContext) GetViewportRectangle() *geom.Rectangle {
	result := geom.NewRectangleFromRectangle(r.GetFixedRectangle())
	result.Y *= -1

	return result
}

func (r *RenderingContext) DebugDrawBoxes() bool {
	return r.sharedContext.DebugDrawBoxes()
}

func (r *RenderingContext) DebugDrawLineBoxes() bool {
	return r.sharedContext.DebugDrawLineBoxes()
}

func (r *RenderingContext) DebugDrawInlineBoxes() bool {
	return r.sharedContext.DebugDrawInlineBoxes()
}

func (r *RenderingContext) DebugDrawFontMetrics() bool {
	return r.sharedContext.DebugDrawFontMetrics()
}

func (r *RenderingContext) IsInteractive() bool {
	return r.sharedContext.IsInteractive()
}

func (r *RenderingContext) IsPrint() bool {
	return r.sharedContext.IsPrint()
}

func (r *RenderingContext) GetOutputDevice() OutputDevice {
	return r.outputDevice
}

func (r *RenderingContext) GetFontContext() FontContext {
	return r.fontContext
}

func (r *RenderingContext) SetPage(pageNo int, page *PageBox) {
	r.pageNo = pageNo
	r.page = page
}

func (r *RenderingContext) GetPageCount() int {
	return r.pageCount
}

func (r *RenderingContext) SetPageCount(pageCount int) {
	r.pageCount = pageCount
}

// GetPage may return nil.
func (r *RenderingContext) GetPage() *PageBox {
	return r.page
}

func (r *RenderingContext) GetPageNo() int {
	return r.pageNo
}

func (r *RenderingContext) GetCss() *StyleReference {
	return r.sharedContext.GetCss()
}

func (r *RenderingContext) GetFSFontMetrics(font FSFont) FSFontMetrics {
	return r.GetTextRenderer().GetFSFontMetrics(r.GetFontContext(), font, "")
}

// GetRootLayer may return nil.
func (r *RenderingContext) GetRootLayer() *Layer {
	return r.rootLayer
}

func (r *RenderingContext) GetInitialPageNo() int {
	return r.initialPageNo
}

// GetBoxById may return nil.
func (r *RenderingContext) GetBoxById(id string) BoxI {
	return r.sharedContext.GetBoxById(id)
}
