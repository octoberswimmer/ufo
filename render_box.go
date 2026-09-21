// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/Box.java

package ufo

import (
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

// BoxState ports the enum Box.State.
type BoxState int

const (
	BoxStateNothing BoxState = iota
	BoxStateFlux
	BoxStateChildrenFlux
	BoxStateDone
)

// BoxDump ports the enum Box.Dump.
type BoxDump int

const (
	BoxDumpRender BoxDump = iota
	BoxDumpLayout
)

// String returns the Java enum constant name.
func (d BoxDump) String() string {
	switch d {
	case BoxDumpRender:
		return "RENDER"
	case BoxDumpLayout:
		return "LAYOUT"
	}
	return strconv.Itoa(int(d))
}

// String returns the Java enum constant name.
func (s BoxState) String() string {
	switch s {
	case BoxStateNothing:
		return "NOTHING"
	case BoxStateFlux:
		return "FLUX"
	case BoxStateChildrenFlux:
		return "CHILDREN_FLUX"
	case BoxStateDone:
		return "DONE"
	}
	return strconv.Itoa(int(s))
}

// BoxI lists every non-private method of Box. A Java value typed Box is a
// BoxI in Go. Dump and CalcCanvasLocation are abstract in Java: the Box struct
// does not define them, so a struct embedding Box must.
type BoxI interface {
	AsBox() *Box

	Dump(c *LayoutContext, indent string, which BoxDump) string
	DumpBoxes(c *LayoutContext, indent string, boxes []BoxI, which BoxDump, result *strings.Builder)
	GetWidth() int
	ToString() string
	String() string
	AppendPosition(result *strings.Builder)
	AppendSize(result *strings.Builder)
	AddChildForLayout(c *LayoutContext, child BoxI)
	AddChild(child BoxI)
	AddAllChildren(children []BoxI)
	RemoveAllChildren()
	RemoveChildBox(target BoxI)
	GetPreviousSibling() BoxI
	GetNextSibling() BoxI
	GetPrevious(child BoxI) BoxI
	GetNext(child BoxI) BoxI
	RemoveChildInt(i int)
	SetParent(box BoxI)
	GetParent() BoxI
	GetChildCount() int
	GetChild(i int) BoxI
	GetChildren() []BoxI
	GetState() BoxState
	SetState(state BoxState)
	GetStyle() CalculatedStyleI
	SetStyle(style CalculatedStyleI)
	GetContainingBlock() BoxI
	SetContainingBlock(containingBlock BoxI)
	GetMarginEdgeWithLeftTop(left int, top int, cssCtx CssContext, tx int, ty int) *geom.Rectangle
	GetMarginEdge(cssCtx CssContext, tx int, ty int) *geom.Rectangle
	GetPaintingBorderEdge(cssCtx CssContext) *geom.Rectangle
	GetPaintingPaddingEdge(cssCtx CssContext) *geom.Rectangle
	GetPaintingClipEdge(cssCtx CssContext) *geom.Rectangle
	GetChildrenClipEdge(c *RenderingContext) *geom.Rectangle
	Intersects(cssCtx CssContext, clip geom.Shape) bool
	GetBorderEdge(left int, top int, cssCtx CssContext) *geom.Rectangle
	GetPaddingEdge(left int, top int, cssCtx CssContext) *geom.Rectangle
	GetPaddingWidth(cssCtx CssContext) int
	GetContentAreaEdge(left int, top int, cssCtx CssContext) *geom.Rectangle
	GetLayer() *Layer
	SetLayer(layer *Layer)
	PositionRelative(cssCtx CssContext) *geom.Dimension
	IsInlineBlock() bool
	SetAbsY(absY int)
	GetAbsY() int
	SetAbsX(absX int)
	GetAbsX() int
	IsStyled() bool
	GetBorderSides() int
	PaintBorder(c *RenderingContext)
	PaintBackground(c *RenderingContext)
	PaintRootElementBackground(c *RenderingContext)
	GetContainingLayer() *Layer
	SetContainingLayer(containingLayer *Layer)
	InitContainingLayer(c *LayoutContext)
	ConnectChildrenToCurrentLayer(c *LayoutContext)
	GetElementBoxes(elem *dom.Element) []BoxI
	Reset(c *LayoutContext)
	Detach(c *LayoutContext)
	ResetChildrenWithStartEnd(c *LayoutContext, start int, end int)
	ResetChildren(c *LayoutContext)
	CalcCanvasLocation()
	CalcChildLocations()
	ForcePageBreakBefore(c *LayoutContext, pageBreakValue *IdentValue, pendingPageName bool) int
	ForcePageBreakAfter(c *LayoutContext, pageBreakValue *IdentValue)
	CrossesPageBreak(c *LayoutContext) bool
	GetRelativeOffset() *geom.Dimension
	Find(cssCtx CssContext, absX int, absY int, findAnonymous bool) BoxI
	IsRoot() bool
	IsBody() bool
	GetElement() *dom.Element
	SetElement(element *dom.Element)
	SetMarginTop(cssContext CssContext, marginTop int)
	SetMarginBottom(cssContext CssContext, marginBottom int)
	SetMarginLeft(cssContext CssContext, marginLeft int)
	SetMarginRight(cssContext CssContext, marginRight int)
	GetMargin(cssContext CssContext) RectPropertySetI
	GetStyleMargin(cssContext CssContext) RectPropertySetI
	GetStyleMarginNoCache(cssContext CssContext) RectPropertySetI
	GetPadding(cssCtx CssContext) RectPropertySetI
	GetBorder(cssCtx CssContext) *BorderPropertySet
	GetContainingBlockWidth() int
	ResetTopMargin(cssContext CssContext)
	ClearSelection(modified *[]BoxI)
	SelectAll()
	CalcPaintingInfo(c CssContext, useCache bool) *PaintingInfo
	CalcChildPaintingInfo(c CssContext, result *PaintingInfo, useCache bool)
	GetMarginBorderPadding(cssCtx CssContext, edge *CalculatedStyleEdge) int
	MoveIfGreater(result *geom.Dimension, test *geom.Dimension)
	Restyle(c *LayoutContext)
	RestyleChildren(c *LayoutContext)
	GetRestyleTarget() BoxI
	GetIndex() int
	SetIndex(index int)
	GetPseudoElementOrClass() string
	SetPseudoElementOrClass(pseudoElementOrClass string)
	SetX(x int)
	GetX() int
	SetY(y int)
	GetY() int
	SetTy(ty int)
	GetTy() int
	SetTx(tx int)
	GetTx() int
	SetRightMBP(rightMBP int)
	GetRightMBP() int
	SetLeftMBP(leftMBP int)
	GetLeftMBP() int
	SetHeight(height int)
	GetHeight() int
	SetContentWidth(contentWidth int)
	GetContentWidth() int
	GetPaintingInfo() *PaintingInfo
	IsAnonymous() bool
	GetBoxDimensions() *BoxDimensions
	SetBoxDimensions(dimensions *BoxDimensions)
	CollectText(c *RenderingContext, buffer *strings.Builder) error
	ExportText(c *RenderingContext, writer io.Writer) error
	ExportPageBoxTextWithYPos(c *RenderingContext, writer io.Writer, yPos int) error
	IsInDocumentFlow() bool
	AnalyzePageBreaks(c *LayoutContext, container *ContentLimitContainer)
	GetEffBackgroundColor(c *RenderingContext) FSColor
	IsMarginAreaRoot() bool
	IsContainedInMarginBox() bool
	GetEffectiveWidth() int
	IsInitialContainingBlock() bool
}

// Box is the abstract root of the box tree. Structs for its subclasses embed
// it by value and set self to the outermost object.
type Box struct {
	// self is the outermost object (a *BlockBox, *LineBox, *TableBox, ...).
	// Every call from a Box method to a method that a subclass overrides goes
	// through it, and it is the value handed out wherever Java passes this.
	self BoxI

	element *dom.Element

	x int
	y int

	absY int
	absX int

	// Box width.
	contentWidth int
	rightMBP     int
	leftMBP      int

	height int

	layer           *Layer
	containingLayer *Layer

	parent BoxI

	boxes []BoxI

	// Keeps track of the start of children's containing block.
	tx int
	ty int

	style           CalculatedStyleI
	containingBlock BoxI

	relativeOffset *geom.Dimension

	paintingInfo *PaintingInfo

	workingMargin RectPropertySetI

	index int

	// "" where Java holds null.
	pseudoElementOrClass string

	anonymous bool

	state BoxState
}

// initBox is the body of the constructor Box(Box parent, CalculatedStyle
// style), together with Box's field initializers, for a Box embedded in
// another struct. It does not set self: the constructor of the outermost
// struct calls SetSelf next, before it calls any method. Java's constructor
// calls the overridable setStyle; here the style is assigned directly because
// self is not known yet, so a subclass that overrides SetStyle (TableBox) calls
// its own SetStyle after SetSelf.
func initBox(b *Box, parent BoxI, style CalculatedStyleI) {
	b.boxes = make([]BoxI, 0, 3)
	b.state = BoxStateNothing
	b.parent = parent
	b.style = style
	b.anonymous = false
}

// initBoxWithElement is the body of the constructor
// Box(Element element, CalculatedStyle style, boolean anonymous). See initBox
// for self and setStyle.
func initBoxWithElement(b *Box, element *dom.Element, style CalculatedStyleI, anonymous bool) {
	b.boxes = make([]BoxI, 0, 3)
	b.state = BoxStateNothing
	b.element = element
	b.style = style
	b.anonymous = anonymous
}

// NewBox is the constructor Box(Box parent, CalculatedStyle style). Box is
// abstract, so the result has no self: a struct embedding the returned value
// calls SetSelf.
func NewBox(parent BoxI, style CalculatedStyleI) *Box {
	b := &Box{}
	initBox(b, parent, style)
	return b
}

// NewBoxWithElementAnonymous is the constructor
// Box(Element element, CalculatedStyle style, boolean anonymous). See NewBox
// for self.
func NewBoxWithElementAnonymous(element *dom.Element, style CalculatedStyleI, anonymous bool) *Box {
	b := &Box{}
	initBoxWithElement(b, element, style, anonymous)
	return b
}

// SetSelf records the outermost object that embeds this Box.
func (b *Box) SetSelf(self BoxI) {
	b.self = self
}

func (b *Box) AsBox() *Box {
	return b
}

// boxInt is Java's (int) cast of a float.
func boxInt(f float32) int {
	return calculatedStyleFloatToInt(f)
}

// boxJavaTrim is String.trim(): it removes leading and trailing characters
// up to and including U+0020.
func boxJavaTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= ' ' })
}

// boxSimpleClassName is getClass().getSimpleName() of the outermost object.
func boxSimpleClassName(self BoxI) string {
	t := reflect.TypeOf(self)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func (b *Box) DumpBoxes(c *LayoutContext, indent string, boxes []BoxI, which BoxDump, result *strings.Builder) {
	for i, child := range boxes {
		result.WriteString(child.Dump(c, indent+"  ", which))
		if i < len(boxes)-1 {
			result.WriteByte('\n')
		}
	}
}

func (b *Box) GetWidth() int {
	return b.self.GetContentWidth() + b.GetLeftMBP() + b.GetRightMBP()
}

func (b *Box) ToString() string {
	var result strings.Builder
	result.WriteString(boxSimpleClassName(b.self))
	result.WriteString(": ")
	b.self.AppendPosition(&result)
	b.self.AppendSize(&result)
	return boxJavaTrim(result.String())
}

func (b *Box) String() string {
	return b.self.ToString()
}

func (b *Box) AppendPosition(result *strings.Builder) {
	if b.GetAbsX() != 0 || b.GetAbsY() != 0 {
		result.WriteString("pos: (")
		result.WriteString(strconv.Itoa(b.GetAbsX()))
		result.WriteString(",")
		result.WriteString(strconv.Itoa(b.GetAbsY()))
		result.WriteString(") ")
	}
}

func (b *Box) AppendSize(result *strings.Builder) {
	if b.self.GetWidth() != 0 || b.self.GetHeight() != 0 {
		result.WriteString("size: (")
		result.WriteString(strconv.Itoa(b.self.GetWidth()))
		result.WriteString("x")
		result.WriteString(strconv.Itoa(b.self.GetHeight()))
		result.WriteString(") ")
	}
}

func (b *Box) AddChildForLayout(c *LayoutContext, child BoxI) {
	b.self.AddChild(child)

	child.InitContainingLayer(c)
}

func (b *Box) AddChild(child BoxI) {
	if child == nil {
		panic(NewXRRuntimeException("trying to add null child"))
	}
	child.SetParent(b.self)
	child.SetIndex(len(b.boxes))
	b.boxes = append(b.boxes, child)
}

func (b *Box) AddAllChildren(children []BoxI) {
	for _, box := range children {
		b.self.AddChild(box)
	}
}

// RemoveAllChildren installs a new slice rather than truncating, so a slice
// returned earlier by GetChildren is not overwritten by later AddChild calls.
func (b *Box) RemoveAllChildren() {
	b.boxes = make([]BoxI, 0, 3)
}

func (b *Box) RemoveChildBox(target BoxI) {
	found := false
	kept := make([]BoxI, 0, len(b.boxes))
	for _, child := range b.self.GetChildren() {
		if child == target {
			found = true
		} else {
			if found {
				child.SetIndex(child.GetIndex() - 1)
			}
			kept = append(kept, child)
		}
	}
	b.boxes = kept
}

// GetPreviousSibling may return nil.
func (b *Box) GetPreviousSibling() BoxI {
	parent := b.GetParent()
	if parent == nil {
		return nil
	}
	return parent.GetPrevious(b.self)
}

// GetNextSibling may return nil.
func (b *Box) GetNextSibling() BoxI {
	parent := b.GetParent()
	if parent == nil {
		return nil
	}
	return parent.GetNext(b.self)
}

// GetPrevious may return nil.
func (b *Box) GetPrevious(child BoxI) BoxI {
	if child.GetIndex() == 0 {
		return nil
	}
	return b.GetChild(child.GetIndex() - 1)
}

// GetNext may return nil.
func (b *Box) GetNext(child BoxI) BoxI {
	if child.GetIndex() == b.GetChildCount()-1 {
		return nil
	}
	return b.GetChild(child.GetIndex() + 1)
}

func (b *Box) RemoveChildInt(i int) {
	b.self.RemoveChildBox(b.GetChild(i))
}

func (b *Box) SetParent(box BoxI) {
	b.parent = box
}

// GetParent may return nil.
func (b *Box) GetParent() BoxI {
	return b.parent
}

func (b *Box) GetChildCount() int {
	return len(b.boxes)
}

func (b *Box) GetChild(i int) BoxI {
	return b.boxes[i]
}

func (b *Box) GetChildren() []BoxI {
	return b.boxes
}

func (b *Box) GetState() BoxState {
	return b.state
}

func (b *Box) SetState(state BoxState) {
	b.state = state
}

// GetStyle may return nil.
func (b *Box) GetStyle() CalculatedStyleI {
	return b.style
}

func (b *Box) SetStyle(style CalculatedStyleI) {
	b.style = style
}

// GetContainingBlock may return nil.
func (b *Box) GetContainingBlock() BoxI {
	if b.containingBlock == nil {
		return b.GetParent()
	}
	return b.containingBlock
}

func (b *Box) SetContainingBlock(containingBlock BoxI) {
	b.containingBlock = containingBlock
}

func (b *Box) GetMarginEdgeWithLeftTop(left int, top int, cssCtx CssContext, tx int, ty int) *geom.Rectangle {
	// Note that negative margins can mean this rectangle is inside the border
	// edge, but that's the way it's supposed to work...
	result := geom.NewRectangle(left, top, b.self.GetWidth(), b.self.GetHeight())
	result.Translate(tx, ty)
	return result
}

func (b *Box) GetMarginEdge(cssCtx CssContext, tx int, ty int) *geom.Rectangle {
	return b.self.GetMarginEdgeWithLeftTop(b.GetX(), b.GetY(), cssCtx, tx, ty)
}

func (b *Box) GetPaintingBorderEdge(cssCtx CssContext) *geom.Rectangle {
	return b.self.GetBorderEdge(b.GetAbsX(), b.GetAbsY(), cssCtx)
}

func (b *Box) GetPaintingPaddingEdge(cssCtx CssContext) *geom.Rectangle {
	return b.self.GetPaddingEdge(b.GetAbsX(), b.GetAbsY(), cssCtx)
}

func (b *Box) GetPaintingClipEdge(cssCtx CssContext) *geom.Rectangle {
	return b.self.GetPaintingBorderEdge(cssCtx)
}

func (b *Box) GetChildrenClipEdge(c *RenderingContext) *geom.Rectangle {
	return b.self.GetPaintingPaddingEdge(c)
}

// Intersects does not consider any children of this box. clip may be nil.
func (b *Box) Intersects(cssCtx CssContext, clip geom.Shape) bool {
	return clip == nil || clip.Intersects(b.self.GetPaintingClipEdge(cssCtx))
}

func (b *Box) GetBorderEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	margin := b.self.GetMargin(cssCtx)
	return geom.NewRectangle(left+boxInt(margin.Left()),
		top+boxInt(margin.Top()),
		b.self.GetWidth()-boxInt(margin.Left())-boxInt(margin.Right()),
		b.self.GetHeight()-boxInt(margin.Top())-boxInt(margin.Bottom()))
}

func (b *Box) GetPaddingEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	margin := b.self.GetMargin(cssCtx)
	border := b.self.GetBorder(cssCtx)
	return geom.NewRectangle(left+boxInt(margin.Left())+boxInt(border.Left()),
		top+boxInt(margin.Top())+boxInt(border.Top()),
		b.self.GetWidth()-boxInt(margin.Width())-boxInt(border.Width()),
		b.self.GetHeight()-boxInt(margin.Height())-boxInt(border.Height()))
}

func (b *Box) GetPaddingWidth(cssCtx CssContext) int {
	padding := b.self.GetPadding(cssCtx)
	return boxInt(padding.Left()) + b.self.GetContentWidth() + boxInt(padding.Right())
}

func (b *Box) GetContentAreaEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	margin := b.self.GetMargin(cssCtx)
	border := b.self.GetBorder(cssCtx)
	padding := b.self.GetPadding(cssCtx)

	return geom.NewRectangle(
		left+boxInt(margin.Left())+boxInt(border.Left())+boxInt(padding.Left()),
		top+boxInt(margin.Top())+boxInt(border.Top())+boxInt(padding.Top()),
		b.self.GetWidth()-boxInt(margin.Width())-boxInt(border.Width())-boxInt(padding.Width()),
		b.self.GetHeight()-boxInt(margin.Height())-boxInt(border.Height())-boxInt(padding.Height()))
}

// GetLayer may return nil.
func (b *Box) GetLayer() *Layer {
	return b.layer
}

func (b *Box) SetLayer(layer *Layer) {
	b.layer = layer
}

func (b *Box) PositionRelative(cssCtx CssContext) *geom.Dimension {
	initialX := b.GetX()
	initialY := b.GetY()

	style := b.GetStyle()
	if !style.IsIdent(CSSNameLeft, IdentValueAuto) {
		b.SetX(b.GetX() + boxInt(style.GetFloatPropertyProportionalWidth(
			CSSNameLeft, float32(b.GetContainingBlock().GetContentWidth()), cssCtx)))
	} else if !style.IsIdent(CSSNameRight, IdentValueAuto) {
		b.SetX(b.GetX() - boxInt(style.GetFloatPropertyProportionalWidth(
			CSSNameRight, float32(b.GetContainingBlock().GetContentWidth()), cssCtx)))
	}

	cbContentHeight := 0
	if !b.GetContainingBlock().GetStyle().IsAutoHeight() {
		cbStyle := b.GetContainingBlock().GetStyle()
		cbContentHeight = boxInt(cbStyle.GetFloatPropertyProportionalHeight(
			CSSNameHeight, 0, cssCtx))
	} else if b.self.IsInlineBlock() {
		// FIXME Should be content height, not overall height
		cbContentHeight = b.GetContainingBlock().GetHeight()
	}

	if !style.IsIdent(CSSNameTop, IdentValueAuto) {
		b.SetY(b.GetY() + boxInt(style.GetFloatPropertyProportionalHeight(
			CSSNameTop, float32(cbContentHeight), cssCtx)))
	} else if !style.IsIdent(CSSNameBottom, IdentValueAuto) {
		b.SetY(b.GetY() - boxInt(style.GetFloatPropertyProportionalHeight(
			CSSNameBottom, float32(cbContentHeight), cssCtx)))
	}

	b.relativeOffset = geom.NewDimension(b.GetX()-initialX, b.GetY()-initialY)
	return b.GetRelativeOffset()
}

func (b *Box) IsInlineBlock() bool {
	return false
}

func (b *Box) SetAbsY(absY int) {
	b.absY = absY
}

func (b *Box) GetAbsY() int {
	return b.absY
}

func (b *Box) SetAbsX(absX int) {
	b.absX = absX
}

func (b *Box) GetAbsX() int {
	return b.absX
}

func (b *Box) IsStyled() bool {
	return b.style != nil
}

func (b *Box) GetBorderSides() int {
	return BorderPainterAll
}

func (b *Box) PaintBorder(c *RenderingContext) {
	c.GetOutputDevice().PaintBorder(c, b.self)
}

func (b *Box) isPaintsRootElementBackground() bool {
	return b.IsRoot() && b.GetStyle().IsHasBackground() ||
		b.IsBody() && !b.GetParent().GetStyle().IsHasBackground()
}

func (b *Box) PaintBackground(c *RenderingContext) {
	if !b.isPaintsRootElementBackground() {
		c.GetOutputDevice().PaintBackground(c, b.self)
	}
}

func (b *Box) PaintRootElementBackground(c *RenderingContext) {
	pI := b.GetPaintingInfo()
	if pI != nil {
		if b.GetStyle().IsHasBackground() {
			b.paintRootElementBackgroundWithPI(c, pI)
		} else if b.GetChildCount() > 0 {
			body := b.GetChild(0)
			body.AsBox().paintRootElementBackgroundWithPI(c, pI)
		}
	}
}

func (b *Box) paintRootElementBackgroundWithPI(c *RenderingContext, pI *PaintingInfo) {
	marginCorner := pI.GetOuterMarginCorner()
	canvasBounds := geom.NewRectangle(0, 0, marginCorner.Width, marginCorner.Height)
	canvasBounds.Add(c.GetViewportRectangle())
	c.GetOutputDevice().PaintBackgroundWithStyleBoundsBgImageContainerBorder(c, b.GetStyle(), canvasBounds, canvasBounds, BorderPropertySetEmptyBorder)
}

// GetContainingLayer may return nil.
func (b *Box) GetContainingLayer() *Layer {
	return b.containingLayer
}

func (b *Box) SetContainingLayer(containingLayer *Layer) {
	b.containingLayer = containingLayer
}

func (b *Box) InitContainingLayer(c *LayoutContext) {
	if b.GetLayer() != nil {
		b.SetContainingLayer(b.GetLayer())
	} else if b.GetContainingLayer() == nil {
		if b.GetParent() == nil || b.GetParent().GetContainingLayer() == nil {
			panic(NewXRRuntimeException("internal error"))
		}
		b.SetContainingLayer(b.GetParent().GetContainingLayer())

		// FIXME Will be glacially slow for large inline relative layers.  Could
		// be much more efficient.  We're just looking for block boxes which are
		// directly wrapped by an inline relative layer (i.e. block boxes sandwiched
		// between anonymous block boxes)
		if c.GetLayer().IsInline() {
			content := c.GetLayer().GetMaster().(*InlineLayoutBox).GetElementWithContent()
			for _, candidate := range content {
				if candidate == b.self {
					b.SetContainingLayer(c.GetLayer())
					break
				}
			}
		}
	}
}

func (b *Box) ConnectChildrenToCurrentLayer(c *LayoutContext) {
	for i := 0; i < b.GetChildCount(); i++ {
		box := b.GetChild(i)
		box.SetContainingLayer(c.GetLayer())
		box.ConnectChildrenToCurrentLayer(c)
	}
}

func (b *Box) GetElementBoxes(elem *dom.Element) []BoxI {
	result := []BoxI{}
	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		if child.GetElement() == elem {
			result = append(result, child)
		}
		result = append(result, child.GetElementBoxes(elem)...)
	}
	return result
}

func (b *Box) Reset(c *LayoutContext) {
	b.self.ResetChildren(c)
	if b.layer != nil {
		b.layer.Detach()
		b.layer = nil
	}

	b.SetContainingLayer(nil)
	b.SetLayer(nil)
	b.setPaintingInfo(nil)
	b.SetContentWidth(0)

	b.workingMargin = nil

	e := b.GetElement()
	if e != nil {
		anchorName := c.GetNamespaceHandler().GetAnchorName(e)
		if anchorName != "" {
			c.RemoveBoxId(anchorName)
		}

		id := c.GetNamespaceHandler().GetID(e)
		if id != "" {
			c.RemoveBoxId(id)
		}
	}
}

func (b *Box) Detach(c *LayoutContext) {
	b.self.Reset(c)

	if b.GetParent() != nil {
		b.GetParent().RemoveChildBox(b.self)
		b.SetParent(nil)
	}
}

func (b *Box) ResetChildrenWithStartEnd(c *LayoutContext, start int, end int) {
	for i := start; i <= end; i++ {
		box := b.GetChild(i)
		box.Reset(c)
	}
}

func (b *Box) ResetChildren(c *LayoutContext) {
	remaining := b.GetChildCount()
	for i := 0; i < remaining; i++ {
		box := b.GetChild(i)
		box.Reset(c)
	}
}

func (b *Box) CalcChildLocations() {
	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		child.CalcCanvasLocation()
		child.CalcChildLocations()
	}
}

func (b *Box) ForcePageBreakBefore(c *LayoutContext, pageBreakValue *IdentValue, pendingPageName bool) int {
	page := c.GetRootLayer().GetFirstPage(c, b.self)
	if page == nil {
		XRLogLayout(LevelWarning, "Box has no page")
		return 0
	} else {
		pageBreakCount := 1
		if page.GetTop() == b.GetAbsY() {
			pageBreakCount--
			if pendingPageName && page == c.GetRootLayer().GetLastPage() {
				c.GetRootLayer().RemoveLastPage()
				c.SetPageName(c.GetPendingPageName())
				c.GetRootLayer().AddPage(c)
			}
		}
		if page.IsLeftPage() && pageBreakValue == IdentValueLeft ||
			page.IsRightPage() && pageBreakValue == IdentValueRight {
			pageBreakCount++
		}

		if pageBreakCount == 0 {
			return 0
		}

		if pageBreakCount == 1 && pendingPageName {
			c.SetPageName(c.GetPendingPageName())
		}

		delta := page.GetBottom() + c.GetExtraSpaceTop() - b.GetAbsY()
		if page == c.GetRootLayer().GetLastPage() {
			c.GetRootLayer().AddPage(c)
		}

		if pageBreakCount == 2 {
			page = c.GetRootLayer().GetPages()[page.GetPageNo()+1]
			delta += page.GetContentHeight(c)

			if pendingPageName {
				c.SetPageName(c.GetPendingPageName())
			}

			if page == c.GetRootLayer().GetLastPage() {
				c.GetRootLayer().AddPage(c)
			}
		}

		b.SetY(b.GetY() + delta)

		return delta
	}
}

func (b *Box) ForcePageBreakAfter(c *LayoutContext, pageBreakValue *IdentValue) {
	needSecondPageBreak := false
	page := c.GetRootLayer().GetLastPageWithCBox(c, b.self)

	if page != nil {
		if page.IsLeftPage() && pageBreakValue == IdentValueLeft ||
			page.IsRightPage() && pageBreakValue == IdentValueRight {
			needSecondPageBreak = true
		}

		delta := page.GetBottom() + c.GetExtraSpaceTop() - (b.GetAbsY() +
			b.GetMarginBorderPadding(c, CalculatedStyleEdgeTop) + b.self.GetHeight())

		if page == c.GetRootLayer().GetLastPage() {
			c.GetRootLayer().AddPage(c)
		}

		if needSecondPageBreak {
			page = c.GetRootLayer().GetPages()[page.GetPageNo()+1]
			delta += page.GetContentHeight(c)

			if page == c.GetRootLayer().GetLastPage() {
				c.GetRootLayer().AddPage(c)
			}
		}

		b.SetHeight(b.self.GetHeight() + delta)
	}
}

func (b *Box) CrossesPageBreak(c *LayoutContext) bool {
	if !c.IsPageBreaksAllowed() {
		return false
	}

	pageBox := c.GetRootLayer().GetFirstPage(c, b.self)
	if pageBox == nil {
		return false
	} else {
		return b.GetAbsY()+b.self.GetHeight() >= pageBox.GetBottom()-c.GetExtraSpaceBottom()
	}
}

// GetRelativeOffset may return nil.
func (b *Box) GetRelativeOffset() *geom.Dimension {
	return b.relativeOffset
}

// Find may return nil.
func (b *Box) Find(cssCtx CssContext, absX int, absY int, findAnonymous bool) BoxI {
	pI := b.GetPaintingInfo()
	if pI != nil && !pI.GetAggregateBounds().Contains(absX, absY) {
		return nil
	}

	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		result := child.Find(cssCtx, absX, absY, findAnonymous)
		if result != nil {
			return result
		}
	}

	edge := b.self.GetContentAreaEdge(b.GetAbsX(), b.GetAbsY(), cssCtx)
	if edge.Contains(absX, absY) && b.GetStyle().IsVisible() {
		return b.self
	}
	return nil
}

func (b *Box) IsRoot() bool {
	return b.GetElement() != nil && !b.IsAnonymous() && b.GetElement().GetParentNode().GetNodeType() == dom.DocumentNode
}

func (b *Box) IsBody() bool {
	return b.GetParent() != nil && b.GetParent().IsRoot()
}

// GetElement may return nil.
func (b *Box) GetElement() *dom.Element {
	return b.element
}

func (b *Box) SetElement(element *dom.Element) {
	b.element = element
}

func (b *Box) SetMarginTop(cssContext CssContext, marginTop int) {
	b.ensureWorkingMargin(cssContext).SetTop(float32(marginTop))
}

func (b *Box) SetMarginBottom(cssContext CssContext, marginBottom int) {
	b.ensureWorkingMargin(cssContext).SetBottom(float32(marginBottom))
}

func (b *Box) SetMarginLeft(cssContext CssContext, marginLeft int) {
	b.ensureWorkingMargin(cssContext).SetLeft(float32(marginLeft))
}

func (b *Box) SetMarginRight(cssContext CssContext, marginRight int) {
	b.ensureWorkingMargin(cssContext).SetRight(float32(marginRight))
}

func (b *Box) ensureWorkingMargin(cssContext CssContext) RectPropertySetI {
	if b.workingMargin == nil {
		b.workingMargin = b.GetStyleMargin(cssContext).CopyOf()
	}
	return b.workingMargin
}

func (b *Box) GetMargin(cssContext CssContext) RectPropertySetI {
	if b.workingMargin != nil {
		return b.workingMargin
	}
	return b.GetStyleMargin(cssContext)
}

func (b *Box) GetStyleMargin(cssContext CssContext) RectPropertySetI {
	return b.GetStyle().GetMarginRectWithUseCache(float32(b.self.GetContainingBlockWidth()), cssContext, true)
}

func (b *Box) GetStyleMarginNoCache(cssContext CssContext) RectPropertySetI {
	return b.GetStyle().GetMarginRectWithUseCache(float32(b.self.GetContainingBlockWidth()), cssContext, false)
}

func (b *Box) GetPadding(cssCtx CssContext) RectPropertySetI {
	return b.GetStyle().GetPaddingRect(float32(b.self.GetContainingBlockWidth()), cssCtx)
}

func (b *Box) GetBorder(cssCtx CssContext) *BorderPropertySet {
	return b.GetStyle().GetBorder(cssCtx)
}

func (b *Box) GetContainingBlockWidth() int {
	return b.GetContainingBlock().GetContentWidth()
}

func (b *Box) ResetTopMargin(cssContext CssContext) {
	if b.workingMargin != nil {
		styleMargin := b.GetStyleMargin(cssContext)

		b.workingMargin.SetTop(styleMargin.Top())
	}
}

// ClearSelection appends the boxes whose selection changed to modified.
func (b *Box) ClearSelection(modified *[]BoxI) {
	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		child.ClearSelection(modified)
	}
}

func (b *Box) SelectAll() {
	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		child.SelectAll()
	}
}

func (b *Box) CalcPaintingInfo(c CssContext, useCache bool) *PaintingInfo {
	cached := b.GetPaintingInfo()
	if cached != nil && useCache {
		return cached
	}

	bounds := b.self.GetMarginEdgeWithLeftTop(b.GetAbsX(), b.GetAbsY(), c, 0, 0)

	result := NewPaintingInfo(
		geom.NewDimension(bounds.X+bounds.Width, bounds.Y+bounds.Height),
		b.self.GetPaintingClipEdge(c),
	)

	if !b.GetStyle().IsOverflowApplies() || b.GetStyle().IsOverflowVisible() {
		b.self.CalcChildPaintingInfo(c, result, useCache)
	}

	b.setPaintingInfo(result)

	return result
}

func (b *Box) CalcChildPaintingInfo(c CssContext, result *PaintingInfo, useCache bool) {
	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		info := child.CalcPaintingInfo(c, useCache)
		b.MoveIfGreater(result.GetOuterMarginCorner(), info.GetOuterMarginCorner())
		result.GetAggregateBounds().Add(info.GetAggregateBounds())
	}
}

func (b *Box) GetMarginBorderPadding(cssCtx CssContext, edge *CalculatedStyleEdge) int {
	border := b.self.GetBorder(cssCtx)
	margin := b.self.GetMargin(cssCtx)
	padding := b.self.GetPadding(cssCtx)
	return edge.GetMarginBorderPadding(margin, border, padding)
}

func (b *Box) MoveIfGreater(result *geom.Dimension, test *geom.Dimension) {
	if test.Width > result.Width {
		result.Width = test.Width
	}
	if test.Height > result.Height {
		result.Height = test.Height
	}
}

func (b *Box) Restyle(c *LayoutContext) {
	e := b.GetElement()
	var style CalculatedStyleI

	pe := b.GetPseudoElementOrClass()
	if pe != "" {
		if e != nil {
			style = c.GetSharedContext().GetStyleWithRestyle(e, true)
			style = style.DeriveStyle(c.GetCss().GetPseudoElementStyle(e, pe))
		} else {
			container := b.GetParent().GetParent().(BlockBoxI)
			e = container.GetElement()
			style = c.GetSharedContext().GetStyleWithRestyle(e, true)
			style = style.DeriveStyle(c.GetCss().GetPseudoElementStyle(e, pe))
			style = style.CreateAnonymousStyle(IdentValueInline)
		}
	} else {
		if e != nil {
			style = c.GetSharedContext().GetStyleWithRestyle(e, true)
			if b.IsAnonymous() {
				style = style.CreateAnonymousStyle(b.GetStyle().GetIdent(CSSNameDisplay))
			}
		} else {
			parent := b.GetParent()
			if parent != nil {
				e = parent.GetElement()
				if e != nil {
					style = c.GetSharedContext().GetStyleWithRestyle(e, true)
					style = style.CreateAnonymousStyle(IdentValueInline)
				}
			}
		}
	}

	if style != nil {
		b.self.SetStyle(style)
	}

	b.self.RestyleChildren(c)
}

func (b *Box) RestyleChildren(c *LayoutContext) {
	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		child.Restyle(c)
	}
}

func (b *Box) GetRestyleTarget() BoxI {
	return b.self
}

func (b *Box) GetIndex() int {
	return b.index
}

func (b *Box) SetIndex(index int) {
	b.index = index
}

// GetPseudoElementOrClass returns "" where Java returns null.
func (b *Box) GetPseudoElementOrClass() string {
	return b.pseudoElementOrClass
}

func (b *Box) SetPseudoElementOrClass(pseudoElementOrClass string) {
	b.pseudoElementOrClass = pseudoElementOrClass
}

func (b *Box) SetX(x int) {
	b.x = x
}

func (b *Box) GetX() int {
	return b.x
}

func (b *Box) SetY(y int) {
	b.y = y
}

func (b *Box) GetY() int {
	return b.y
}

func (b *Box) SetTy(ty int) {
	b.ty = ty
}

func (b *Box) GetTy() int {
	return b.ty
}

func (b *Box) SetTx(tx int) {
	b.tx = tx
}

func (b *Box) GetTx() int {
	return b.tx
}

func (b *Box) SetRightMBP(rightMBP int) {
	b.rightMBP = rightMBP
}

func (b *Box) GetRightMBP() int {
	return b.rightMBP
}

func (b *Box) SetLeftMBP(leftMBP int) {
	b.leftMBP = leftMBP
}

func (b *Box) GetLeftMBP() int {
	return b.leftMBP
}

func (b *Box) SetHeight(height int) {
	b.height = height
}

func (b *Box) GetHeight() int {
	return b.height
}

func (b *Box) SetContentWidth(contentWidth int) {
	b.contentWidth = max(contentWidth, 0)
}

func (b *Box) GetContentWidth() int {
	return b.contentWidth
}

// GetPaintingInfo may return nil.
func (b *Box) GetPaintingInfo() *PaintingInfo {
	return b.paintingInfo
}

func (b *Box) setPaintingInfo(paintingInfo *PaintingInfo) {
	b.paintingInfo = paintingInfo
}

func (b *Box) IsAnonymous() bool {
	return b.anonymous
}

func (b *Box) GetBoxDimensions() *BoxDimensions {
	return NewBoxDimensions(b.GetLeftMBP(), b.GetRightMBP(), b.self.GetContentWidth(), b.self.GetHeight())
}

func (b *Box) SetBoxDimensions(dimensions *BoxDimensions) {
	b.SetLeftMBP(dimensions.GetLeftMBP())
	b.SetRightMBP(dimensions.GetRightMBP())
	b.SetContentWidth(dimensions.GetContentWidth())
	b.SetHeight(dimensions.GetHeight())
}

func (b *Box) CollectText(c *RenderingContext, buffer *strings.Builder) error {
	for _, child := range b.GetChildren() {
		if err := child.CollectText(c, buffer); err != nil {
			return err
		}
	}
	return nil
}

func (b *Box) ExportText(c *RenderingContext, writer io.Writer) error {
	if c.IsPrint() && b.IsRoot() {
		c.SetPage(0, c.GetRootLayer().GetPages()[0])
		if err := c.GetPage().ExportLeadingText(c, writer); err != nil {
			return err
		}
	}
	for _, child := range b.GetChildren() {
		if err := child.ExportText(c, writer); err != nil {
			return err
		}
	}
	if c.IsPrint() && b.IsRoot() {
		if err := b.exportPageBoxText(c, writer); err != nil {
			return err
		}
	}
	return nil
}

func (b *Box) exportPageBoxText(c *RenderingContext, writer io.Writer) error {
	if err := c.GetPage().ExportTrailingText(c, writer); err != nil {
		return err
	}
	if c.GetPage() != c.GetRootLayer().GetLastPage() {
		pages := c.GetRootLayer().GetPages()
		for {
			next := pages[c.GetPageNo()+1]
			c.SetPage(next.GetPageNo(), next)
			if err := next.ExportLeadingText(c, writer); err != nil {
				return err
			}
			if err := next.ExportTrailingText(c, writer); err != nil {
				return err
			}
			if c.GetPage() == c.GetRootLayer().GetLastPage() {
				break
			}
		}
	}
	return nil
}

func (b *Box) ExportPageBoxTextWithYPos(c *RenderingContext, writer io.Writer, yPos int) error {
	if err := c.GetPage().ExportTrailingText(c, writer); err != nil {
		return err
	}
	pages := c.GetRootLayer().GetPages()
	next := pages[c.GetPageNo()+1]
	c.SetPage(next.GetPageNo(), next)
	for next.GetBottom() < yPos {
		if err := next.ExportLeadingText(c, writer); err != nil {
			return err
		}
		if err := next.ExportTrailingText(c, writer); err != nil {
			return err
		}
		next = pages[c.GetPageNo()+1]
		c.SetPage(next.GetPageNo(), next)
	}
	return next.ExportLeadingText(c, writer)
}

func (b *Box) IsInDocumentFlow() bool {
	flowRoot := b.self
	for {
		parent := flowRoot.GetParent()
		if parent == nil {
			break
		} else {
			flowRoot = parent
		}
	}

	return flowRoot.IsRoot()
}

func (b *Box) AnalyzePageBreaks(c *LayoutContext, container *ContentLimitContainer) {
	container.UpdateTop(c, b.GetAbsY())
	for _, child := range b.GetChildren() {
		child.AnalyzePageBreaks(c, container)
	}
	container.UpdateBottom(c, b.GetAbsY()+b.self.GetHeight())
}

func (b *Box) GetEffBackgroundColor(c *RenderingContext) FSColor {
	var result FSColor
	current := b.self
	for current != nil {
		result = current.GetStyle().GetBackgroundColor()
		if result != nil {
			return result
		}

		current = current.GetContainingBlock()
	}

	page := c.GetPage()
	result = page.GetStyle().GetBackgroundColor()
	if result != nil {
		return result
	}
	return NewFSRGBColor(255, 255, 255)
}

func (b *Box) IsMarginAreaRoot() bool {
	return false
}

func (b *Box) IsContainedInMarginBox() bool {
	current := b.self
	for {
		parent := current.GetParent()
		if parent == nil {
			break
		} else {
			current = parent
		}
	}

	return current.IsMarginAreaRoot()
}

func (b *Box) GetEffectiveWidth() int {
	return b.self.GetWidth()
}

func (b *Box) IsInitialContainingBlock() bool {
	return false
}
