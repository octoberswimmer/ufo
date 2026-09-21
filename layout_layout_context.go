// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/LayoutContext.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// LayoutContext tracks state which changes over the course of a layout run.
// Generally speaking, if possible, state information should be stored in the
// box tree and not here.  It also provides pass-though calls to many methods in
// SharedContext.
type LayoutContext struct {
	sharedContext *SharedContext

	rootLayer *Layer

	firstLines        *StyleTracker
	firstLetters      *StyleTracker
	currentMarkerData *MarkerData

	// blockFormattingContexts and layers are Java Deques used as stacks: the
	// last element is the current one.
	blockFormattingContexts []*BlockFormattingContext
	layers                  []*Layer

	fontContext FontContext

	contentFunctionFactory *ContentFunctionFactory

	extraSpaceTop    int
	extraSpaceBottom int

	// counterContextMap is keyed by the identity of the style, as Java's
	// HashMap is (CalculatedStyle does not override equals). The key is the
	// embedded CalculatedStyle struct, which is the same pointer whether the
	// style is reached as a *CalculatedStyle or as an *EmptyStyle.
	counterContextMap map[*CalculatedStyle]*LayoutContextCounterContext

	pendingPageName string
	pageName        string

	noPageBreak int

	rootDocumentLayer *Layer
	page              *PageBox

	mayCheckKeepTogether bool

	breakAtLineContext *BreakAtLineContext
}

func (c *LayoutContext) GetTextRenderer() TextRenderer {
	return c.sharedContext.GetTextRenderer()
}

func (c *LayoutContext) GetCss() *StyleReference {
	return c.sharedContext.GetCss()
}

func (c *LayoutContext) GetCanvas() FSCanvas {
	return c.sharedContext.GetCanvas()
}

func (c *LayoutContext) GetFixedRectangle() *geom.Rectangle {
	return c.sharedContext.GetFixedRectangle()
}

func (c *LayoutContext) GetNamespaceHandler() NamespaceHandler {
	return c.sharedContext.GetNamespaceHandler()
}

// NewLayoutContext holds the stuff that needs to have a separate instance for
// each run.
func NewLayoutContext(sharedContext *SharedContext, fontContext FontContext) *LayoutContext {
	return &LayoutContext{
		sharedContext:          sharedContext,
		firstLines:             NewStyleTracker(),
		firstLetters:           NewStyleTracker(),
		fontContext:            fontContext,
		contentFunctionFactory: NewContentFunctionFactory(),
		counterContextMap:      map[*CalculatedStyle]*LayoutContextCounterContext{},
		mayCheckKeepTogether:   true,
	}
}

func (c *LayoutContext) ReInit(keepLayers bool) {
	c.firstLines = NewStyleTracker()
	c.firstLetters = NewStyleTracker()
	c.currentMarkerData = nil

	c.blockFormattingContexts = nil

	if !keepLayers {
		c.rootLayer = nil
		c.layers = nil
	}

	c.extraSpaceTop = 0
	c.extraSpaceBottom = 0
}

func (c *LayoutContext) CaptureLayoutState() *LayoutState {
	if c.IsPrint() {
		// Java passes the bottom space in the position of the top space and
		// the reverse here, and RestoreLayoutState reads them back by name.
		return NewLayoutStateWithPageNameExtraSpaceTopExtraSpaceBottomNoPageBreak(c.firstLines, c.firstLetters, c.currentMarkerData, c.blockFormattingContexts, c.GetPageName(), c.GetExtraSpaceBottom(), c.GetExtraSpaceTop(), c.GetNoPageBreak())
	}
	return NewLayoutState(c.firstLines, c.firstLetters, c.currentMarkerData, c.blockFormattingContexts)
}

func (c *LayoutContext) RestoreLayoutState(layoutState *LayoutState) {
	c.firstLines = layoutState.GetFirstLines()
	c.firstLetters = layoutState.GetFirstLetters()

	c.currentMarkerData = layoutState.GetCurrentMarkerData()

	c.blockFormattingContexts = nil
	c.blockFormattingContexts = append(c.blockFormattingContexts, layoutState.GetBFCs()...)

	if c.IsPrint() {
		c.SetPageName(layoutState.GetPageName())
		c.SetExtraSpaceBottom(layoutState.GetExtraSpaceBottom())
		c.SetExtraSpaceTop(layoutState.GetExtraSpaceTop())
		c.SetNoPageBreak(layoutState.GetNoPageBreak())
	}
}

func (c *LayoutContext) CopyStateForRelayout() *LayoutState {
	if c.IsPrint() {
		return NewLayoutStateWithPageNameExtraSpaceTopExtraSpaceBottomNoPageBreak(c.firstLines.CopyOf(), c.firstLetters.CopyOf(), c.currentMarkerData, nil, c.GetPageName(), 0, 0, 0)
	}
	return NewLayoutState(c.firstLines.CopyOf(), c.firstLetters.CopyOf(), c.currentMarkerData, nil)
}

func (c *LayoutContext) RestoreStateForRelayout(layoutState *LayoutState) {
	c.firstLines = layoutState.GetFirstLines()
	c.firstLetters = layoutState.GetFirstLetters()

	c.currentMarkerData = layoutState.GetCurrentMarkerData()

	if c.IsPrint() {
		c.SetPageName(layoutState.GetPageName())
	}
}

// GetBlockFormattingContext returns the innermost block formatting context.
// As Java's Deque.getLast, it panics when there is none.
func (c *LayoutContext) GetBlockFormattingContext() *BlockFormattingContext {
	if len(c.blockFormattingContexts) == 0 {
		panic(NewXRRuntimeException("NoSuchElementException: no block formatting context"))
	}
	return c.blockFormattingContexts[len(c.blockFormattingContexts)-1]
}

func (c *LayoutContext) PushBFC(bfc *BlockFormattingContext) {
	c.blockFormattingContexts = append(c.blockFormattingContexts, bfc)
}

func (c *LayoutContext) PopBFC() {
	if len(c.blockFormattingContexts) == 0 {
		panic(NewXRRuntimeException("NoSuchElementException: no block formatting context"))
	}
	c.blockFormattingContexts = c.blockFormattingContexts[:len(c.blockFormattingContexts)-1]
}

// PushLayerBox is pushLayer(Box).
func (c *LayoutContext) PushLayerBox(master BoxI) {
	var layer *Layer

	if c.rootLayer == nil {
		layer = NewLayer(master)
		c.rootLayer = layer
	} else {
		parent := c.GetLayer()
		layer = NewLayerWithParent(parent, master)
		parent.AddChild(layer)
	}

	c.PushLayerLayer(layer)
}

// PushLayerLayer is pushLayer(Layer).
func (c *LayoutContext) PushLayerLayer(layer *Layer) {
	c.layers = append(c.layers, layer)
}

func (c *LayoutContext) PopLayer() {
	layer := c.GetLayer()

	layer.Finish(c)

	c.layers = c.layers[:len(c.layers)-1]
}

// GetLayer returns the innermost layer. As Java's Deque.getLast, it panics
// when there is none.
func (c *LayoutContext) GetLayer() *Layer {
	if len(c.layers) == 0 {
		panic(NewXRRuntimeException("NoSuchElementException: no layer"))
	}
	return c.layers[len(c.layers)-1]
}

// GetRootLayer may return nil.
func (c *LayoutContext) GetRootLayer() *Layer {
	return c.rootLayer
}

func (c *LayoutContext) Translate(x int, y int) {
	c.GetBlockFormattingContext().Translate(x, y)
}

/* code to keep track of all the id'd boxes */
func (c *LayoutContext) AddBoxId(id string, box BoxI) {
	c.sharedContext.AddBoxId(id, box)
}

func (c *LayoutContext) RemoveBoxId(id string) {
	c.sharedContext.RemoveBoxId(id)
}

func (c *LayoutContext) IsInteractive() bool {
	return c.sharedContext.IsInteractive()
}

func (c *LayoutContext) GetMmPerDot() float32 {
	return c.sharedContext.GetMmPerPx()
}

func (c *LayoutContext) GetDotsPerPixel() int {
	return c.sharedContext.GetDotsPerPixel()
}

func (c *LayoutContext) GetFontSize2D(font *FontSpecification) float32 {
	return c.sharedContext.GetFont(font).GetSize2D()
}

func (c *LayoutContext) GetXHeight(parentFont *FontSpecification) float32 {
	return c.sharedContext.GetXHeight(c.GetFontContext(), parentFont)
}

// GetFont may return nil.
func (c *LayoutContext) GetFont(font *FontSpecification) FSFont {
	return c.sharedContext.GetFont(font)
}

func (c *LayoutContext) GetUac() UserAgentCallback {
	return c.sharedContext.GetUac()
}

func (c *LayoutContext) IsPrint() bool {
	return c.sharedContext.IsPrint()
}

func (c *LayoutContext) GetFirstLinesTracker() *StyleTracker {
	return c.firstLines
}

func (c *LayoutContext) GetFirstLettersTracker() *StyleTracker {
	return c.firstLetters
}

// GetCurrentMarkerData may return nil.
func (c *LayoutContext) GetCurrentMarkerData() *MarkerData {
	return c.currentMarkerData
}

func (c *LayoutContext) SetCurrentMarkerData(currentMarkerData *MarkerData) {
	c.currentMarkerData = currentMarkerData
}

func (c *LayoutContext) GetReplacedElementFactory() ReplacedElementFactory {
	return c.sharedContext.GetReplacedElementFactory()
}

func (c *LayoutContext) GetFontContext() FontContext {
	return c.fontContext
}

func (c *LayoutContext) GetContentFunctionFactory() *ContentFunctionFactory {
	return c.contentFunctionFactory
}

func (c *LayoutContext) GetSharedContext() *SharedContext {
	return c.sharedContext
}

func (c *LayoutContext) GetExtraSpaceBottom() int {
	return c.extraSpaceBottom
}

func (c *LayoutContext) SetExtraSpaceBottom(extraSpaceBottom int) {
	c.extraSpaceBottom = extraSpaceBottom
}

func (c *LayoutContext) GetExtraSpaceTop() int {
	return c.extraSpaceTop
}

func (c *LayoutContext) SetExtraSpaceTop(extraSpaceTop int) {
	c.extraSpaceTop = extraSpaceTop
}

// ResolveCountersWithStartIndex accepts a nil startIndex.
func (c *LayoutContext) ResolveCountersWithStartIndex(style CalculatedStyleI, startIndex *int) {
	//new context for child elements
	cc := newLayoutContextCounterContext(c, style, startIndex)
	c.counterContextMap[style.AsCalculatedStyle()] = cc
}

func (c *LayoutContext) ResolveCounters(style CalculatedStyleI) {
	c.ResolveCountersWithStartIndex(style, nil)
}

// GetCounterContext returns nil when ResolveCounters was not called for the
// style.
func (c *LayoutContext) GetCounterContext(style CalculatedStyleI) *LayoutContextCounterContext {
	if style == nil {
		return nil
	}
	return c.counterContextMap[style.AsCalculatedStyle()]
}

func (c *LayoutContext) GetFSFontMetrics(font FSFont) FSFontMetrics {
	return c.GetTextRenderer().GetFSFontMetrics(c.GetFontContext(), font, "")
}

// LayoutContextCounterContext is LayoutContext.CounterContext.
type LayoutContextCounterContext struct {
	counters map[string]int
	// parent is different because it needs to work even when the
	// counter-properties cascade, and it should also logically be redefined on
	// each level (think list-items within list-items)
	parent *LayoutContextCounterContext
}

// newLayoutContextCounterContext is CounterContext(CalculatedStyle, Integer);
// startIndex may be nil.
//
// A CounterContext should really be reflected in the element hierarchy, but
// CalculatedStyles reflect the ancestor hierarchy just as well and also handles
// pseudo-elements seamlessly.
func newLayoutContextCounterContext(c *LayoutContext, style CalculatedStyleI, startIndex *int) *LayoutContextCounterContext {
	cc := &LayoutContextCounterContext{counters: map[string]int{}}
	// Numbering restarted via <ol start="x">
	if startIndex != nil {
		cc.counters["list-item"] = *startIndex
	}
	cc.parent = c.GetCounterContext(style.GetParent())
	if cc.parent == nil {
		//top-level context, above root element
		cc.parent = &LayoutContextCounterContext{counters: map[string]int{}}
	}
	//first the explicitly named counters
	resets := style.GetCounterReset()
	for _, cd := range resets {
		cc.parent.resetCounter(cd)
	}

	increments := style.GetCounterIncrement()
	for _, cd := range increments {
		if !cc.parent.incrementCounter(cd) {
			cc.parent.resetCounter(NewCounterData(cd.GetName(), 0))
			cc.parent.incrementCounter(cd)
		}
	}

	// then the implicit list-item counter
	if style.IsIdent(CSSNameDisplay, IdentValueListItem) {
		// Numbering restarted via <li value="x">
		if startIndex != nil {
			cc.parent.counters["list-item"] = *startIndex
		}
		cc.parent.incrementListItemCounter(1)
	}
	return cc
}

// incrementCounter returns true if a counter was found and incremented.
func (cc *LayoutContextCounterContext) incrementCounter(cd *CounterData) bool {
	if "list-item" == cd.GetName() { //reserved name for list-item counter in CSS3
		cc.incrementListItemCounter(cd.GetValue())
		return true
	} else {
		currentValue, ok := cc.counters[cd.GetName()]
		if !ok {
			if cc.parent == nil {
				return false
			}
			return cc.parent.incrementCounter(cd)
		} else {
			cc.counters[cd.GetName()] = currentValue + cd.GetValue()
			return true
		}
	}
}

func (cc *LayoutContextCounterContext) incrementListItemCounter(increment int) {
	currentValue, ok := cc.counters["list-item"]
	if !ok {
		currentValue = 0
	}
	cc.counters["list-item"] = currentValue + increment
}

func (cc *LayoutContextCounterContext) resetCounter(cd *CounterData) {
	cc.counters[cd.GetName()] = cd.GetValue()
}

func (cc *LayoutContextCounterContext) GetCurrentCounterValue(name string) int {
	//only the counters of the parent are in scope
	//_parent is never null for a publicly accessible CounterContext
	value, ok := cc.parent.getCounter(name)
	if !ok {
		cc.parent.resetCounter(NewCounterData(name, 0))
		return 0
	} else {
		return value
	}
}

// getCounter reports false when no context in the chain has the counter
// (Java returns a null Integer).
func (cc *LayoutContextCounterContext) getCounter(name string) (int, bool) {
	value, ok := cc.counters[name]
	if ok {
		return value, true
	}
	if cc.parent == nil {
		return 0, false
	}
	return cc.parent.getCounter(name)
}

func (cc *LayoutContextCounterContext) GetCurrentCounterValues(name string) []int {
	//only the counters of the parent are in scope
	//_parent is never null for a publicly accessible CounterContext
	var values []int
	values = cc.parent.getCounterValues(name, values)
	if len(values) == 0 {
		cc.parent.resetCounter(NewCounterData(name, 0))
		values = append(values, 0)
	}
	return values
}

// getCounterValues appends to values and returns the extended slice (Java
// adds to the list it is given).
func (cc *LayoutContextCounterContext) getCounterValues(name string, values []int) []int {
	if cc.parent != nil {
		values = cc.parent.getCounterValues(name, values)
	}
	value, ok := cc.counters[name]
	if ok {
		values = append(values, value)
	}
	return values
}

// GetPageName returns "" for Java's null.
func (c *LayoutContext) GetPageName() string {
	return c.pageName
}

func (c *LayoutContext) SetPageName(currentPageName string) {
	c.pageName = currentPageName
}

func (c *LayoutContext) GetNoPageBreak() int {
	return c.noPageBreak
}

func (c *LayoutContext) SetNoPageBreak(noPageBreak int) {
	c.noPageBreak = noPageBreak
}

func (c *LayoutContext) IsPageBreaksAllowed() bool {
	return c.noPageBreak == 0
}

// GetPendingPageName returns "" for Java's null.
func (c *LayoutContext) GetPendingPageName() string {
	return c.pendingPageName
}

func (c *LayoutContext) SetPendingPageName(pendingPageName string) {
	c.pendingPageName = pendingPageName
}

// GetRootDocumentLayer may return nil.
func (c *LayoutContext) GetRootDocumentLayer() *Layer {
	return c.rootDocumentLayer
}

func (c *LayoutContext) SetRootDocumentLayer(rootDocumentLayer *Layer) {
	c.rootDocumentLayer = rootDocumentLayer
}

// GetPage may return nil.
func (c *LayoutContext) GetPage() *PageBox {
	return c.page
}

func (c *LayoutContext) SetPage(page *PageBox) {
	c.page = page
}

func (c *LayoutContext) IsMayCheckKeepTogether() bool {
	return c.mayCheckKeepTogether
}

func (c *LayoutContext) SetMayCheckKeepTogether(mayKeepTogether bool) {
	c.mayCheckKeepTogether = mayKeepTogether
}

// GetBreakAtLineContext may return nil.
func (c *LayoutContext) GetBreakAtLineContext() *BreakAtLineContext {
	return c.breakAtLineContext
}

func (c *LayoutContext) SetBreakAtLineContext(breakAtLineContext *BreakAtLineContext) {
	c.breakAtLineContext = breakAtLineContext
}
