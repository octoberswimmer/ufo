// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/Layer.java

package ufo

import (
	"sort"
	"sync"

	"github.com/octoberswimmer/ufo/geom"
)

type LayerPagedMode int

const (
	LayerPagedModePagedModeScreen LayerPagedMode = iota
	LayerPagedModePagedModePrint
)

type LayerWidth int

const (
	LayerWidthPositive LayerWidth = iota
	LayerWidthZero
	LayerWidthNegative
	LayerWidthAuto
)

// Layer: all positioned content as well as content with an overflow value
// other than visible creates a layer.  Layers which define stacking contexts
// provide the entry for rendering the box tree to an output device.  The main
// purpose of this class is to provide an implementation of Appendix E of the
// spec, but it also provides additional utility services including page
// management and mapping boxes to coordinates (for e.g. links).  When
// rendering to a paged output device, the layer is also responsible for laying
// out absolute content (which is laid out after its containing block has
// completed layout).
//
// Box identity (Java == and List.remove(Object) on boxes, none of which
// override equals) is compared through AsBox(), which gives the same pointer
// no matter which interface value holds the box.
type Layer struct {
	// parent is nil for the root layer.
	parent          *Layer
	stackingContext bool

	// childrenMutex stands for the Java methods that are synchronized on the
	// layer: addChild, getChildren and the list access in remove.
	childrenMutex sync.Mutex
	children      []*Layer
	master        BoxI

	// end may be nil.
	end BoxI

	floats []BlockBoxI

	fixedBackground bool

	inline         bool
	requiresLayout bool

	pages []*PageBox
	// lastRequestedPage may be nil.
	lastRequestedPage *PageBox

	// pageSequences is Java's HashSet<BlockBox>. pageSequenceOrder holds the
	// same boxes in insertion order; getSortedPageSequences sorts that slice,
	// so that boxes with equal absY keep a defined order (Java's order for
	// them is the iteration order of an identity-hashed set).
	pageSequences     map[*Box]struct{}
	pageSequenceOrder []BlockBoxI
	// sortedPageSequences is nil until getSortedPageSequences computes it.
	sortedPageSequences []BlockBoxI

	runningBlocks map[string][]BlockBoxI
}

func NewLayer(master BoxI) *Layer {
	return NewLayerWithParentStackingContext(nil, master, true)
}

func NewLayerWithParent(parent *Layer, master BoxI) *Layer {
	// A transformed box must be a stacking context too, otherwise collectLayers() flattens it and
	// paints its own layered descendants (position:absolute, nested transforms, ...) outside the
	// transform established by this layer's paint().
	return NewLayerWithParentStackingContext(parent, master,
		master.GetStyle().IsPositioned() && !master.GetStyle().IsAutoZIndex() || master.GetStyle().HasTransform())
}

func NewLayerWithParentStackingContext(parent *Layer, master BoxI, stackingContext bool) *Layer {
	l := &Layer{
		parent:          parent,
		master:          master,
		stackingContext: stackingContext,
	}
	master.SetLayer(l)
	master.SetContainingLayer(l)
	return l
}

// GetParent returns nil for the root layer.
func (l *Layer) GetParent() *Layer {
	return l.parent
}

func (l *Layer) IsStackingContext() bool {
	return l.stackingContext
}

func (l *Layer) GetZIndex() int {
	// A stacking context not established by z-index (e.g. a transformed box with the default
	// z-index: auto) is treated as z-index: 0 among its sibling stacking contexts.
	if l.master.GetStyle().IsAutoZIndex() {
		return 0
	}
	return calculatedStyleFloatToInt(l.master.GetStyle().AsFloat(CSSNameZIndex))
}

func (l *Layer) GetOpacity() float32 {
	return l.master.GetStyle().GetOpacity()
}

func (l *Layer) GetMaster() BoxI {
	return l.master
}

func (l *Layer) AddChild(layer *Layer) {
	l.childrenMutex.Lock()
	defer l.childrenMutex.Unlock()
	l.children = append(l.children, layer)
}

func (l *Layer) AddFloat(floater BlockBoxI) {
	l.floats = append(l.floats, floater)
	floater.GetFloatedBoxData().SetDrawingLayer(l)
}

func (l *Layer) RemoveFloat(floater BlockBoxI) {
	for i, f := range l.floats {
		if f.AsBox() == floater.AsBox() {
			l.floats = append(l.floats[:i], l.floats[i+1:]...)
			break
		}
	}
}

func (l *Layer) paintFloats(c *RenderingContext) {
	for i := len(l.floats) - 1; i >= 0; i-- {
		floater := l.floats[i]
		l.PaintAsLayer(c, floater)
	}
}

func (l *Layer) paintLayers(c *RenderingContext, layers []*Layer) {
	for _, layer := range layers {
		layer.Paint(c)
	}
}

func (l *Layer) collectLayers(which LayerWidth) []*Layer {
	var result []*Layer

	if which != LayerWidthAuto {
		result = append(result, l.getStackingContextLayers(which)...)
	}

	children := l.GetChildren()
	for _, child := range children {
		if !child.IsStackingContext() {
			if which == LayerWidthAuto {
				result = append(result, child)
			}
			result = append(result, child.collectLayers(which)...)
		}
	}

	return result
}

func (l *Layer) getStackingContextLayers(which LayerWidth) []*Layer {
	var result []*Layer

	children := l.GetChildren()
	for _, target := range children {
		if target.IsStackingContext() {
			zIndex := target.GetZIndex()
			if which == LayerWidthNegative && zIndex < 0 {
				result = append(result, target)
			} else if which == LayerWidthPositive && zIndex > 0 {
				result = append(result, target)
			} else if which == LayerWidthZero && zIndex == 0 {
				result = append(result, target)
			}
		}
	}

	return result
}

func (l *Layer) getSortedLayers(which LayerWidth) []*Layer {
	result := l.collectLayers(which)
	comparator := &layerZIndexComparator{}
	// List.sort is stable.
	sort.SliceStable(result, func(i, j int) bool {
		return comparator.Compare(result[i], result[j]) < 0
	})
	return result
}

type layerZIndexComparator struct{}

func (z *layerZIndexComparator) Compare(l1 *Layer, l2 *Layer) int {
	return l1.GetZIndex() - l2.GetZIndex()
}

// paintBackgroundsAndBorders takes a nil collapsedTableBorders when no table
// in blocks paints collapsed borders.
func (l *Layer) paintBackgroundsAndBorders(
	c *RenderingContext, blocks []BoxI,
	collapsedTableBorders map[*TableCellBox][]*CollapsedBorderSide,
	rangeLists *BoxRangeLists) {
	helper := NewBoxRangeHelper(c.GetOutputDevice(), rangeLists.GetBlock())

	for i := 0; i < len(blocks); i++ {
		helper.PopClipRegions(i)

		box := blocks[i]
		box.PaintBackground(c)
		box.PaintBorder(c)
		if c.DebugDrawBoxes() {
			if blockBox, ok := box.(BlockBoxI); ok {
				blockBox.PaintDebugOutline(c)
			}
		}

		if collapsedTableBorders != nil {
			if cell, ok := box.(*TableCellBox); ok {
				if cell.HasCollapsedPaintingBorder() {
					borders, found := collapsedTableBorders[cell]
					if found {
						l.paintCollapsedTableBorders(c, borders)
					}
				}
			}
		}

		helper.PushClipRegion(c, i)
	}

	helper.PopClipRegions(len(blocks))
}

func (l *Layer) paintInlineContent(c *RenderingContext, lines []BoxI, rangeLists *BoxRangeLists) {
	helper := NewBoxRangeHelper(
		c.GetOutputDevice(), rangeLists.GetInline())

	for i := 0; i < len(lines); i++ {
		helper.PopClipRegions(i)
		helper.PushClipRegion(c, i)
		lines[i].(InlinePaintable).PaintInline(c)
	}

	helper.PopClipRegions(len(lines))
}

func (l *Layer) paintSelection(c *RenderingContext, lines []BoxI) {
	if c.GetOutputDevice().IsSupportsSelection() {
		for _, paintable := range lines {
			if inlineLayoutBox, ok := paintable.(*InlineLayoutBox); ok {
				inlineLayoutBox.PaintSelection(c)
			}
		}
	}
}

func (l *Layer) GetPaintingDimension(c *LayoutContext) *geom.Dimension {
	return l.calcPaintingDimension(c).GetOuterMarginCorner()
}

func (l *Layer) Paint(c *RenderingContext) {
	if l.GetMaster().GetStyle().IsFixed() {
		l.positionFixedLayer(c)
	}

	if l.IsRootLayer() {
		l.GetMaster().PaintRootElementBackground(c)
	}

	// Apply clip from parent TableCellBox if this layer is inside a paginated table cell
	var originalClip geom.Shape
	needsClipRestore := false
	if c.IsPrint() && !l.IsRootLayer() {
		master := l.GetMaster()
		parent := master.GetParent()
		// Find the containing TableCellBox if any
		for parent != nil {
			if _, ok := parent.(*TableCellBox); ok {
				break
			}
			parent = parent.GetParent()
		}
		if cell, ok := parent.(*TableCellBox); ok {
			if cell.GetTable().GetStyle().IsPaginateTable() && cell.IsNeedsClipOnPaint(c) {
				clipEdge := cell.GetChildrenClipEdge(c)
				if clipEdge != nil {
					originalClip = c.GetOutputDevice().GetClip()
					c.GetOutputDevice().Clip(clipEdge)
					needsClipRestore = true
				}
			}
		}
	}

	c.GetOutputDevice().PushTransform(c, l.GetMaster())
	// The deferred function is the Java finally block.
	defer func() {
		c.GetOutputDevice().PopTransform()

		// Restore original clip if we applied table cell clipping
		if needsClipRestore && originalClip != nil {
			c.GetOutputDevice().SetClip(originalClip)
		}
	}()

	if !l.IsInline() && l.GetMaster().(BlockBoxI).IsReplaced() {
		l.paintLayerBackgroundAndBorder(c)
		l.paintReplacedElement(c, l.GetMaster().(BlockBoxI))
	} else {
		rangeLists := NewBoxRangeLists()

		var blocks []BoxI
		var lines []BoxI

		collector := NewBoxCollector()
		collector.Collect(c, c.GetOutputDevice().GetClip(), l, &blocks, &lines, rangeLists)

		if !l.IsInline() {
			l.paintLayerBackgroundAndBorder(c)
			if c.DebugDrawBoxes() {
				l.GetMaster().(BlockBoxI).PaintDebugOutline(c)
			}
		}

		if l.IsRootLayer() || l.IsStackingContext() {
			l.paintLayers(c, l.getSortedLayers(LayerWidthNegative))
		}

		collapsedTableBorders := l.collectCollapsedTableBorders(blocks)

		l.paintBackgroundsAndBorders(c, blocks, collapsedTableBorders, rangeLists)
		l.paintFloats(c)
		l.paintListMarkers(c, blocks, rangeLists)
		l.paintInlineContent(c, lines, rangeLists)
		l.paintReplacedElements(c, blocks, rangeLists)
		l.paintSelection(c, lines) // XXX do only when there is a selection

		if l.IsRootLayer() || l.IsStackingContext() {
			l.paintLayers(c, l.collectLayers(LayerWidthAuto))
			// TODO z-index: 0 layers should be painted atomically
			l.paintLayers(c, l.getSortedLayers(LayerWidthZero))
			l.paintLayers(c, l.getSortedLayers(LayerWidthPositive))
		}
	}
}

func (l *Layer) getFloats() []BlockBoxI {
	return l.floats
}

// Find may return nil.
func (l *Layer) Find(cssCtx CssContext, absX int, absY int, findAnonymous bool) BoxI {
	if l.IsRootLayer() || l.IsStackingContext() {
		result := l.findWithLayers(cssCtx, absX, absY, l.getSortedLayers(LayerWidthPositive), findAnonymous)
		if result != nil {
			return result
		}

		result = l.findWithLayers(cssCtx, absX, absY, l.getSortedLayers(LayerWidthZero), findAnonymous)
		if result != nil {
			return result
		}

		result = l.findWithLayers(cssCtx, absX, absY, l.collectLayers(LayerWidthAuto), findAnonymous)
		if result != nil {
			return result
		}
	}

	for i := 0; i < len(l.getFloats()); i++ {
		floater := l.getFloats()[i]
		result := floater.Find(cssCtx, absX, absY, findAnonymous)
		if result != nil {
			return result
		}
	}

	result := l.GetMaster().Find(cssCtx, absX, absY, findAnonymous)
	if result != nil {
		return result
	}

	if l.IsRootLayer() || l.IsStackingContext() {
		result = l.findWithLayers(cssCtx, absX, absY, l.getSortedLayers(LayerWidthNegative), findAnonymous)
		return result
	}

	return nil
}

// findWithLayers may return nil.
func (l *Layer) findWithLayers(cssCtx CssContext, absX int, absY int, layers []*Layer, findAnonymous bool) BoxI {
	// Work backwards since layers are painted forwards and we're looking
	// for the top-most box
	for i := len(layers) - 1; i >= 0; i-- {
		layer := layers[i]
		result := layer.Find(cssCtx, absX, absY, findAnonymous)
		if result != nil {
			return result
		}
	}
	return nil
}

// A bit of a kludge here.  We need to paint collapsed table borders according
// to priority so (for example) wider borders float to the top and aren't
// over-painted by thinner borders.  This method scans the block boxes
// we're about to draw and returns a map with the last cell in a given table
// we'll paint as a key and a sorted list of borders as values.  These are
// then painted after we've drawn the background for this cell.
//
// The result is nil when no cell in blocks paints collapsed borders.
func (l *Layer) collectCollapsedTableBorders(blocks []BoxI) map[*TableCellBox][]*CollapsedBorderSide {
	cellBordersByTable := make(map[*TableBox][]*CollapsedBorderSide)
	triggerCellsByTable := make(map[*TableBox]*TableCellBox)

	all := make(map[*CollapsedBorderValue]struct{})
	for _, b := range blocks {
		if cell, ok := b.(*TableCellBox); ok {
			if cell.HasCollapsedPaintingBorder() {
				borders := cellBordersByTable[cell.GetTable()]
				triggerCellsByTable[cell.GetTable()] = cell
				// AddCollapsedBorders returns the list with the cell's
				// borders appended.
				cellBordersByTable[cell.GetTable()] = cell.AddCollapsedBorders(all, borders)
			}
		}
	}

	if len(triggerCellsByTable) == 0 {
		return nil
	} else {
		result := make(map[*TableCellBox][]*CollapsedBorderSide)

		// The iteration order is not observed: each pass only fills one
		// entry of the result map.
		for _, cell := range triggerCellsByTable {
			borders := cellBordersByTable[cell.GetTable()]
			// Collections.sort is stable.
			sort.SliceStable(borders, func(i, j int) bool {
				return borders[i].CompareTo(borders[j]) < 0
			})
			result[cell] = borders
		}

		return result
	}
}

func (l *Layer) paintCollapsedTableBorders(c *RenderingContext, borders []*CollapsedBorderSide) {
	for _, border := range borders {
		border.GetCell().PaintCollapsedBorder(c, border.GetSide())
	}
}

func (l *Layer) PaintAsLayer(c *RenderingContext, startingPoint BlockBoxI) {
	rangeLists := NewBoxRangeLists()

	var blocks []BoxI
	var lines []BoxI

	collector := NewBoxCollector()
	collector.CollectWithContainer(c, c.GetOutputDevice().GetClip(),
		l, startingPoint, &blocks, &lines, rangeLists)

	collapsedTableBorders := l.collectCollapsedTableBorders(blocks)

	l.paintBackgroundsAndBorders(c, blocks, collapsedTableBorders, rangeLists)
	l.paintListMarkers(c, blocks, rangeLists)
	l.paintInlineContent(c, lines, rangeLists)
	l.paintSelection(c, lines) // XXX only do when there is a selection
	l.paintReplacedElements(c, blocks, rangeLists)
}

func (l *Layer) paintListMarkers(c *RenderingContext, blocks []BoxI, rangeLists *BoxRangeLists) {
	helper := NewBoxRangeHelper(c.GetOutputDevice(), rangeLists.GetBlock())

	for i := 0; i < len(blocks); i++ {
		helper.PopClipRegions(i)

		box := blocks[i].(BlockBoxI)
		box.PaintListMarker(c)

		helper.PushClipRegion(c, i)
	}

	helper.PopClipRegions(len(blocks))
}

func (l *Layer) paintReplacedElements(c *RenderingContext, blocks []BoxI, rangeLists *BoxRangeLists) {
	helper := NewBoxRangeHelper(c.GetOutputDevice(), rangeLists.GetBlock())

	for i := 0; i < len(blocks); i++ {
		helper.PopClipRegions(i)

		box := blocks[i].(BlockBoxI)
		if box.IsReplaced() {
			l.paintReplacedElement(c, box)
		}

		helper.PushClipRegion(c, i)
	}

	helper.PopClipRegions(len(blocks))
}

func (l *Layer) positionFixedLayer(c *RenderingContext) {
	rect := c.GetFixedRectangle()

	fixed := l.GetMaster()

	fixed.SetX(0)
	fixed.SetY(0)
	fixed.SetAbsX(0)
	fixed.SetAbsY(0)

	fixed.SetContainingBlock(NewViewportBox(rect))
	fixed.(BlockBoxI).PositionAbsolute(c, BlockBoxPositionBoth)

	fixed.CalcPaintingInfo(c, false)
}

func (l *Layer) paintLayerBackgroundAndBorder(c *RenderingContext) {
	if box, ok := l.GetMaster().(BlockBoxI); ok {
		box.PaintBackground(c)
		box.PaintBorder(c)
	}
}

func (l *Layer) paintReplacedElement(c *RenderingContext, replaced BlockBoxI) {
	contentBounds := replaced.GetContentAreaEdge(
		replaced.GetAbsX(), replaced.GetAbsY(), c)
	// Minor hack:  It's inconvenient to adjust for margins, border, padding during
	// layout so just do it here.
	loc := replaced.GetReplacedElement().GetLocation()
	if contentBounds.X != loc.X || contentBounds.Y != loc.Y {
		replaced.GetReplacedElement().SetLocation(contentBounds.X, contentBounds.Y)
	}
	if !c.IsInteractive() || replaced.GetReplacedElement().IsRequiresInteractivePaint() {
		c.GetOutputDevice().PaintReplacedElement(c, replaced)
	}
}

func (l *Layer) IsRootLayer() bool {
	return l.GetParent() == nil && l.IsStackingContext()
}

func (l *Layer) moveIfGreater(result *geom.Dimension, test *geom.Dimension) {
	if test.Width > result.Width {
		result.Width = test.Width
	}
	if test.Height > result.Height {
		result.Height = test.Height
	}
}

func (l *Layer) calcPaintingDimension(c *LayoutContext) *PaintingInfo {
	l.GetMaster().CalcPaintingInfo(c, true)
	result := l.GetMaster().GetPaintingInfo().CopyOf()

	children := l.GetChildren()
	for _, child := range children {
		masterStyle := child.GetMaster().GetStyle()
		if !masterStyle.IsFixed() && masterStyle.IsAbsolute() {
			info := child.calcPaintingDimension(c)
			l.moveIfGreater(result.GetOuterMarginCorner(), info.GetOuterMarginCorner())
		}
	}

	return result
}

func (l *Layer) PositionChildren(c *LayoutContext) {
	for _, child := range l.GetChildren() {
		child.position(c)
	}
}

func (l *Layer) position(c *LayoutContext) {
	if l.GetMaster().GetStyle().IsAbsolute() && !c.IsPrint() {
		l.GetMaster().(BlockBoxI).PositionAbsolute(c, BlockBoxPositionBoth)
	} else if l.GetMaster().GetStyle().IsRelative() &&
		(l.IsInline() || l.GetMaster().(BlockBoxI).IsInline()) {
		l.GetMaster().PositionRelative(c)
		if !l.IsInline() {
			l.GetMaster().CalcCanvasLocation()
			l.GetMaster().CalcChildLocations()
		}

	}
}

func (l *Layer) containsFixedLayer() bool {
	for _, child := range l.GetChildren() {
		if child.GetMaster().GetStyle().IsFixed() || child.containsFixedLayer() {
			return true
		}
	}
	return false
}

func (l *Layer) ContainsFixedContent() bool {
	return l.fixedBackground || l.containsFixedLayer()
}

func (l *Layer) SetFixedBackground(b bool) {
	l.fixedBackground = b
}

// GetChildren returns the child layers present at the time of the call;
// callers must not modify the result (Java returns an unmodifiable list).
func (l *Layer) GetChildren() []*Layer {
	l.childrenMutex.Lock()
	defer l.childrenMutex.Unlock()
	return l.children[:len(l.children):len(l.children)]
}

func (l *Layer) remove(layer *Layer) {
	removed := false

	// access to children is synchronized
	func() {
		l.childrenMutex.Lock()
		defer l.childrenMutex.Unlock()
		for i, child := range l.children {
			if child == layer {
				removed = true
				// A new backing array, so that a slice returned earlier by
				// GetChildren keeps its contents.
				children := make([]*Layer, 0, len(l.children)-1)
				children = append(children, l.children[:i]...)
				children = append(children, l.children[i+1:]...)
				l.children = children
				break
			}
		}
	}()

	if !removed {
		panic(NewXRRuntimeException("Could not find layer to remove"))
	}
}

func (l *Layer) Detach() {
	if l.GetParent() != nil {
		l.GetParent().remove(l)
	}
}

func (l *Layer) IsInline() bool {
	return l.inline
}

func (l *Layer) SetInline(inline bool) {
	l.inline = inline
}

// GetEnd may return nil.
func (l *Layer) GetEnd() BoxI {
	return l.end
}

func (l *Layer) SetEnd(end BoxI) {
	l.end = end
}

func (l *Layer) IsRequiresLayout() bool {
	return l.requiresLayout
}

func (l *Layer) SetRequiresLayout(requiresLayout bool) {
	l.requiresLayout = requiresLayout
}

func (l *Layer) Finish(c *LayoutContext) {
	if c.IsPrint() {
		l.layoutAbsoluteChildren(c)
	}
	if !l.IsInline() {
		l.PositionChildren(c)
	}
}

func (l *Layer) layoutAbsoluteChildren(c *LayoutContext) {
	children := append([]*Layer(nil), l.GetChildren()...)
	if len(children) != 0 {
		state := c.CaptureLayoutState()
		for _, layer := range children {
			if layer.IsRequiresLayout() {
				l.layoutAbsoluteChild(c, layer)
				if layer.GetMaster().GetStyle().IsAvoidPageBreakInside() &&
					layer.GetMaster().CrossesPageBreak(c) {
					layer.GetMaster().Reset(c)
					layer.GetMaster().(BlockBoxI).SetNeedPageClear(true)
					l.layoutAbsoluteChild(c, layer)
					if layer.GetMaster().CrossesPageBreak(c) {
						layer.GetMaster().Reset(c)
						l.layoutAbsoluteChild(c, layer)
					}
				}
				layer.SetRequiresLayout(false)
				layer.Finish(c)
				c.GetRootLayer().EnsureHasPage(c, layer.GetMaster())
			}
		}
		c.RestoreLayoutState(state)
	}
}

func (l *Layer) layoutAbsoluteChild(c *LayoutContext, child *Layer) {
	master := child.GetMaster().(BlockBoxI)
	if child.GetMaster().GetStyle().IsBottomAuto() {
		// Set top, left
		master.PositionAbsolute(c, BlockBoxPositionBoth)
		master.PositionAbsoluteOnPage(c)
		c.ReInit(true)
		child.GetMaster().(BlockBoxI).Layout(c)
		// Set right
		master.PositionAbsolute(c, BlockBoxPositionHorizontally)
	} else {
		// FIXME Not right in the face of pagination, but what
		// to do?  Not sure if just laying out and positioning
		// repeatedly will converge on the correct position,
		// so just guess for now
		c.ReInit(true)
		master.Layout(c)

		before := master.GetBoxDimensions()
		master.Reset(c)
		after := master.GetBoxDimensions()
		master.SetBoxDimensions(before)
		master.PositionAbsolute(c, BlockBoxPositionBoth)
		master.PositionAbsoluteOnPage(c)
		master.SetBoxDimensions(after)

		c.ReInit(true)
		child.GetMaster().(BlockBoxI).Layout(c)
	}
}

// GetPages returns the pages present at the time of the call. Java returns
// the list itself; every modification of it is made inside Layer.
func (l *Layer) GetPages() []*PageBox {
	return l.pages
}

func (l *Layer) IsLastPage(pageBox *PageBox) bool {
	return l.pages[len(l.pages)-1] == pageBox
}

func (l *Layer) AddPage(c CssContext) {
	pages := l.GetPages()
	pagesCount := len(pages)
	pseudoPage := layerPseudoPage(pagesCount)
	var pageBox *PageBox
	if len(pages) == 0 {
		pageBox = LayerCreatePageBoxWithTopPageNo(c, pseudoPage, 0, pagesCount)
	} else {
		pageBox = LayerCreatePageBoxWithTopPageNo(c, pseudoPage, pages[pagesCount-1].GetBottom(), pagesCount)
	}
	l.pages = append(l.pages, pageBox)
}

func layerPseudoPage(size int) string {
	if size == 0 {
		return "first"
	} else if size%2 == 0 {
		return "right"
	} else {
		return "left"
	}
}

func (l *Layer) RemoveLastPage() {
	pageBox := l.pages[len(l.pages)-1]
	l.pages = l.pages[:len(l.pages)-1]
	if pageBox == l.getLastRequestedPage() {
		l.setLastRequestedPage(nil)
	}
}

func LayerCreatePageBox(c CssContext, pseudoPage string) *PageBox {
	return LayerCreatePageBoxWithTopPageNo(c, pseudoPage, 0, 0)
}

func LayerCreatePageBoxWithTopPageNo(c CssContext, pseudoPage string, top int, pageNo int) *PageBox {
	pageName := ""
	// HACK We only create pages during layout, but the OutputDevice
	// queries page positions and since pages are created lazily, changing
	// this method to use LayoutContext is tricky
	if layoutContext, ok := c.(*LayoutContext); ok {
		pageName = layoutContext.GetPageName()
	}

	pageInfo := c.GetCss().GetPageStyle(pageName, pseudoPage)
	cs := NewEmptyStyle().DeriveStyle(pageInfo.GetPageStyle())
	return NewPageBox(pageInfo, c, cs, top, pageNo)
}

// GetFirstPage may return nil.
func (l *Layer) GetFirstPage(c CssContext, box BoxI) *PageBox {
	return l.GetPage(c, box.GetAbsY())
}

// GetLastPageWithCBox may return nil.
func (l *Layer) GetLastPageWithCBox(c CssContext, box BoxI) *PageBox {
	return l.GetPage(c, box.GetAbsY()+box.GetHeight()-1)
}

func (l *Layer) EnsureHasPage(c CssContext, box BoxI) {
	l.GetLastPageWithCBox(c, box)
}

// GetPage may return nil.
func (l *Layer) GetPage(c CssContext, yOffset int) *PageBox {
	if yOffset < 0 {
		return nil
	} else {
		pages := l.GetPages()
		lastRequested := l.getLastRequestedPage()
		if lastRequested != nil {
			if yOffset >= lastRequested.GetTop() && yOffset < lastRequested.GetBottom() {
				return lastRequested
			}
		}
		last := pages[len(pages)-1]
		if yOffset < last.GetBottom() {
			// The page we're looking for is probably at the end of the
			// document so do a linear search for the first few pages
			// and then fall back to a binary search if that doesn't work
			// out
			count := len(pages)
			for i := count - 1; i >= 0 && i >= count-5; i-- {
				pageBox := pages[i]
				if yOffset >= pageBox.GetTop() && yOffset < pageBox.GetBottom() {
					l.setLastRequestedPage(pageBox)
					return pageBox
				}
			}

			low := 0
			high := count - 6

			for low <= high {
				mid := (low + high) >> 1
				pageBox := pages[mid]

				if yOffset >= pageBox.GetTop() && yOffset < pageBox.GetBottom() {
					l.setLastRequestedPage(pageBox)
					return pageBox
				}

				if pageBox.GetTop() < yOffset {
					low = mid + 1
				} else {
					high = mid - 1
				}
			}
		} else {
			l.addPagesUntilPosition(c, yOffset)
			pages = l.GetPages()
			result := pages[len(pages)-1]
			l.setLastRequestedPage(result)
			return result
		}
	}

	panic(NewXRRuntimeException("internal error"))
}

func (l *Layer) addPagesUntilPosition(c CssContext, position int) {
	pages := l.GetPages()
	last := pages[len(pages)-1]
	for position >= last.GetBottom() {
		l.AddPage(c)
		pages = l.GetPages()
		last = pages[len(pages)-1]
	}
}

func (l *Layer) TrimEmptyPages(maxYHeight int) {
	// Empty pages may result when a "keep together" constraint
	// cannot be satisfied and is dropped
	for i := len(l.pages) - 1; i > 0; i-- {
		page := l.pages[i]
		if page.GetTop() >= maxYHeight {
			if page == l.getLastRequestedPage() {
				l.setLastRequestedPage(nil)
			}
			l.pages = append(l.pages[:i], l.pages[i+1:]...)
		} else {
			break
		}
	}
}

func (l *Layer) TrimPageCount(newPageCount int) {
	for len(l.pages) > newPageCount {
		pageBox := l.pages[len(l.pages)-1]
		l.pages = l.pages[:len(l.pages)-1]
		if pageBox == l.getLastRequestedPage() {
			l.setLastRequestedPage(nil)
		}
	}
}

func (l *Layer) AssignPagePaintingPositions(cssCtx CssContext, mode LayerPagedMode) {
	l.AssignPagePaintingPositionsWithAdditionalClearance(cssCtx, mode, 0)
}

func (l *Layer) AssignPagePaintingPositionsWithAdditionalClearance(
	cssCtx CssContext, mode LayerPagedMode, additionalClearance int) {
	pages := l.GetPages()
	paintingTop := additionalClearance
	for _, page := range pages {
		page.SetPaintingTop(paintingTop)
		switch mode {
		case LayerPagedModePagedModeScreen:
			page.SetPaintingBottom(paintingTop + page.GetHeight(cssCtx))
		case LayerPagedModePagedModePrint:
			page.SetPaintingBottom(paintingTop + page.GetContentHeight(cssCtx))
		}
		paintingTop = page.GetPaintingBottom() + additionalClearance
	}
}

func (l *Layer) GetMaxPageWidth(cssCtx CssContext, additionalClearance int) int {
	pages := l.GetPages()
	maxWidth := 0
	for _, page := range pages {
		pageWidth := page.GetWidth(cssCtx) + additionalClearance*2
		if pageWidth > maxWidth {
			maxWidth = pageWidth
		}
	}

	return maxWidth
}

// GetLastPage returns nil when there are no pages.
func (l *Layer) GetLastPage() *PageBox {
	pages := l.GetPages()
	if len(pages) == 0 {
		return nil
	}
	return pages[len(pages)-1]
}

func (l *Layer) CrossesPageBreak(c *LayoutContext, top int, bottom int) bool {
	if top < 0 {
		return false
	}
	page := l.GetPage(c, top)
	return bottom >= page.GetBottom()-c.GetExtraSpaceBottom()
}

func (l *Layer) FindRoot() *Layer {
	if l.IsRootLayer() {
		return l
	} else {
		return l.GetParent().FindRoot()
	}
}

func (l *Layer) AddRunningBlock(block BlockBoxI) {
	if l.runningBlocks == nil {
		l.runningBlocks = make(map[string][]BlockBoxI)
	}

	identifier := block.GetStyle().GetRunningName()
	blocks := append(l.runningBlocks[identifier], block)
	// List.sort is stable.
	sort.SliceStable(blocks, func(i, j int) bool {
		return blocks[i].GetAbsY() < blocks[j].GetAbsY()
	})
	l.runningBlocks[identifier] = blocks
}

func (l *Layer) RemoveRunningBlock(block BlockBoxI) {
	if l.runningBlocks == nil {
		return
	}

	identifier := block.GetStyle().GetRunningName()

	blocks, found := l.runningBlocks[identifier]
	if found {
		for i, b := range blocks {
			if b.AsBox() == block.AsBox() {
				l.runningBlocks[identifier] = append(blocks[:i], blocks[i+1:]...)
				break
			}
		}
	}
}

// GetRunningBlock may return nil.
func (l *Layer) GetRunningBlock(identifier string, page *PageBox, which *PageElementPosition) BlockBoxI {
	if l.runningBlocks == nil {
		return nil
	}

	blocks, found := l.runningBlocks[identifier]
	if !found {
		return nil
	}

	if which == PageElementPositionStart {
		var prev BlockBoxI
		for _, b := range blocks {
			if b.GetStaticEquivalent().GetAbsY() >= page.GetTop() {
				break
			}
			prev = b
		}
		return prev
	} else if which == PageElementPositionFirst {
		for _, b := range blocks {
			absY := b.GetStaticEquivalent().GetAbsY()
			if absY >= page.GetTop() && absY < page.GetBottom() {
				return b
			}
		}
		return l.GetRunningBlock(identifier, page, PageElementPositionStart)
	} else if which == PageElementPositionLast {
		var prev BlockBoxI
		for _, b := range blocks {
			if b.GetStaticEquivalent().GetAbsY() > page.GetBottom() {
				break
			}
			prev = b
		}
		return prev
	} else if which == PageElementPositionLastExcept {
		var prev BlockBoxI
		for _, b := range blocks {
			absY := b.GetStaticEquivalent().GetAbsY()
			if absY >= page.GetTop() && absY < page.GetBottom() {
				return nil
			}
			if absY > page.GetBottom() {
				break
			}
			prev = b
		}
		return prev
	}

	panic(NewXRRuntimeException("bug: internal error"))
}

func (l *Layer) LayoutPages(c *LayoutContext) {
	c.SetRootDocumentLayer(c.GetRootLayer())
	for _, pageBox := range l.pages {
		pageBox.Layout(c)
	}
}

func (l *Layer) AddPageSequence(start BlockBoxI) {
	if l.pageSequences == nil {
		l.pageSequences = make(map[*Box]struct{})
	}

	if _, present := l.pageSequences[start.AsBox()]; !present {
		l.pageSequences[start.AsBox()] = struct{}{}
		l.pageSequenceOrder = append(l.pageSequenceOrder, start)
	}
}

// getSortedPageSequences returns ok == false where Java returns null (no page
// sequence was ever added). The sorted list is computed once and kept, as in
// Java.
func (l *Layer) getSortedPageSequences() (sequences []BlockBoxI, ok bool) {
	if l.pageSequences == nil {
		return nil, false
	}

	if l.sortedPageSequences == nil {
		result := append([]BlockBoxI(nil), l.pageSequenceOrder...)
		// List.sort is stable.
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].GetAbsY() < result[j].GetAbsY()
		})
		l.sortedPageSequences = result
	}

	return l.sortedPageSequences, true
}

func (l *Layer) GetRelativePageNoWithAbsY(c *RenderingContext, absY int) int {
	sequences, hasSequences := l.getSortedPageSequences()
	initial := 0
	if c.GetInitialPageNo() > 0 {
		initial = c.GetInitialPageNo() - 1
	}
	if !hasSequences || len(sequences) == 0 {
		return initial + l.GetPage(c, absY).GetPageNo()
	} else {
		pageSequence := l.findPageSequence(sequences, absY)
		sequenceStartAbsolutePageNo := l.GetPage(c, pageSequence.GetAbsY()).GetPageNo()
		absoluteRequiredPageNo := l.GetPage(c, absY).GetPageNo()
		return absoluteRequiredPageNo - sequenceStartAbsolutePageNo
	}
}

// findPageSequence may return nil.
func (l *Layer) findPageSequence(sequences []BlockBoxI, absY int) BlockBoxI {
	for i := 0; i < len(sequences); i++ {
		result := sequences[i]
		if i < len(sequences)-1 && sequences[i+1].GetAbsY() > absY {
			return result
		}
	}

	return nil
}

func (l *Layer) GetRelativePageNo(c *RenderingContext) int {
	sequences, hasSequences := l.getSortedPageSequences()
	initial := 0
	if c.GetInitialPageNo() > 0 {
		initial = c.GetInitialPageNo() - 1
	}
	if !hasSequences {
		return initial + c.GetPageNo()
	} else {
		sequenceStartIndex := l.getPageSequenceStart(sequences, c.GetPage())
		if sequenceStartIndex == -1 {
			return initial + c.GetPageNo()
		} else {
			block := sequences[sequenceStartIndex]
			return c.GetPageNo() - l.GetFirstPage(c, block).GetPageNo()
		}
	}
}

func (l *Layer) GetRelativePageCount(c *RenderingContext) int {
	sequences, hasSequences := l.getSortedPageSequences()
	initial := 0
	if c.GetInitialPageNo() > 0 {
		initial = c.GetInitialPageNo() - 1
	}
	if !hasSequences {
		return initial + c.GetPageCount()
	} else {
		var firstPage int
		var lastPage int

		sequenceStartIndex := l.getPageSequenceStart(sequences, c.GetPage())

		if sequenceStartIndex == -1 {
			firstPage = 0
		} else {
			block := sequences[sequenceStartIndex]
			firstPage = l.GetFirstPage(c, block).GetPageNo()
		}

		if sequenceStartIndex < len(sequences)-1 {
			block := sequences[sequenceStartIndex+1]
			lastPage = l.GetFirstPage(c, block).GetPageNo()
		} else {
			lastPage = c.GetPageCount()
		}

		sequenceLength := lastPage - firstPage
		if sequenceStartIndex == -1 {
			sequenceLength += initial
		}

		return sequenceLength
	}
}

func (l *Layer) getPageSequenceStart(sequences []BlockBoxI, page *PageBox) int {
	for i := len(sequences) - 1; i >= 0; i-- {
		start := sequences[i]
		if start.GetAbsY() < page.GetBottom()-1 {
			return i
		}
	}

	return -1
}

// getLastRequestedPage may return nil.
func (l *Layer) getLastRequestedPage() *PageBox {
	return l.lastRequestedPage
}

func (l *Layer) setLastRequestedPage(lastRequestedPage *PageBox) {
	l.lastRequestedPage = lastRequestedPage
}
