// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/InlineLayoutBox.java

package ufo

import (
	"fmt"
	"math"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

// InlineLayoutBox is a Box which contains the portion of an inline element
// laid out on a single line. It may contain content from several InlineBox
// objects if the original inline element was interrupted by nested content.
// Unlike other boxes, its children may be either Box objects
// (for example, a box with display: inline-block) or
// InlineText objects. For this reason, its children are not
// stored in the children property, but instead stored in the
// inlineChildren property.
type InlineLayoutBox struct {
	Box

	baseline int

	startsHere bool
	endsHere   bool

	inlineChildren []InlineChild

	pending bool

	inlineWidth int

	// textDecorations is nil where Java has null.
	textDecorations []*TextDecoration

	containingBlockWidth int
}

// newInlineLayoutBoxFromElementStyle is the private constructor
// InlineLayoutBox(Element, CalculatedStyle), which CopyOf uses.
func newInlineLayoutBoxFromElementStyle(elem *dom.Element, style CalculatedStyleI) *InlineLayoutBox {
	b := &InlineLayoutBox{}
	b.initFromElementStyle(elem, style)
	b.SetState(BoxStateDone)
	b.containingBlockWidth = 0
	return b
}

// initFromElementStyle is the call super(elem, style, false) of both Java
// constructors.
func (b *InlineLayoutBox) initFromElementStyle(elem *dom.Element, style CalculatedStyleI) {
	initBoxWithElement(&b.Box, elem, style, false)
	b.SetSelf(b)
}

// NewInlineLayoutBox creates an inline layout box. elem and style may be nil.
func NewInlineLayoutBox(c *LayoutContext, elem *dom.Element, style CalculatedStyleI, cbWidth int) *InlineLayoutBox {
	b := &InlineLayoutBox{}
	b.initFromElementStyle(elem, style)
	b.containingBlockWidth = cbWidth
	b.SetMarginTop(c, 0)
	b.SetMarginBottom(c, 0)
	b.SetPending(true)
	b.CalculateHeight(c)
	return b
}

func (b *InlineLayoutBox) CopyOf() *InlineLayoutBox {
	result := newInlineLayoutBoxFromElementStyle(b.GetElement(), b.GetStyle())
	result.SetHeight(b.GetHeight())
	result.pending = b.pending
	result.SetContainingLayer(b.GetContainingLayer())
	return result
}

func (b *InlineLayoutBox) CalculateHeight(c *LayoutContext) {
	border := b.GetBorder(c)
	padding := b.GetPadding(c)

	metrics := b.GetStyle().GetFSFontMetrics(c)

	b.SetHeight(calculatedStyleFloatToInt(float32(math.Ceil(float64(border.Top() + padding.Top() + metrics.GetAscent() +
		metrics.GetDescent() + padding.Bottom() + border.Bottom())))))
}

func (b *InlineLayoutBox) GetBaseline() int {
	return b.baseline
}

func (b *InlineLayoutBox) SetBaseline(baseline int) {
	b.baseline = baseline
}

func (b *InlineLayoutBox) GetInlineChildCount() int {
	return len(b.inlineChildren)
}

func (b *InlineLayoutBox) AddInlineChild(c *LayoutContext, child InlineChild) {
	b.AddInlineChildWithCallUnmarkPending(c, child, true)
}

func (b *InlineLayoutBox) AddInlineChildWithCallUnmarkPending(c *LayoutContext, child InlineChild, callUnmarkPending bool) {
	b.inlineChildren = append(b.inlineChildren, child)

	if callUnmarkPending && b.IsPending() {
		b.UnmarkPending(c)
	}

	if box, ok := child.(BoxI); ok {
		box.SetParent(b)
		box.InitContainingLayer(c)
	} else if inlineText, ok := child.(*InlineText); ok {
		inlineText.SetParent(b)
	} else {
		panic(NewXRRuntimeException(fmt.Sprintf("Inline child of type %T not supported", child)))
	}
}

func (b *InlineLayoutBox) GetInlineChildren() []InlineChild {
	return b.inlineChildren
}

func (b *InlineLayoutBox) GetInlineChild(i int) InlineChild {
	return b.inlineChildren[i]
}

// GetInlineWidthWithCssContext is getInlineWidth(CssContext).
func (b *InlineLayoutBox) GetInlineWidthWithCssContext(cssContext CssContext) int {
	return b.inlineWidth
}

func (b *InlineLayoutBox) PrunePending() {
	if b.GetInlineChildCount() > 0 {
		for i := b.GetInlineChildCount() - 1; i >= 0; i-- {
			child := b.GetInlineChild(i)
			iB, ok := child.(*InlineLayoutBox)
			if !ok {
				break
			}

			iB.PrunePending()

			if iB.IsPending() {
				b.RemoveChildInt(i)
			} else {
				break
			}
		}
	}
}

func (b *InlineLayoutBox) IsEndsHere() bool {
	return b.endsHere
}

func (b *InlineLayoutBox) SetEndsHere(endsHere bool) {
	b.endsHere = endsHere
}

func (b *InlineLayoutBox) IsStartsHere() bool {
	return b.startsHere
}

func (b *InlineLayoutBox) SetStartsHere(startsHere bool) {
	b.startsHere = startsHere
}

func (b *InlineLayoutBox) IsPending() bool {
	return b.pending
}

func (b *InlineLayoutBox) SetPending(pending bool) {
	b.pending = pending
}

func (b *InlineLayoutBox) UnmarkPending(c *LayoutContext) {
	b.pending = false

	if iB, ok := b.GetParent().(*InlineLayoutBox); ok {
		if iB.IsPending() {
			iB.UnmarkPending(c)
		}
	}

	b.SetStartsHere(true)

	if b.GetStyle().RequiresLayer() {
		c.PushLayerBox(b)
		b.GetLayer().SetInline(true)
		b.ConnectChildrenToCurrentLayer(c)
	}
}

func (b *InlineLayoutBox) ConnectChildrenToCurrentLayer(c *LayoutContext) {
	if b.GetInlineChildCount() > 0 {
		for i := 0; i < b.GetInlineChildCount(); i++ {
			obj := b.GetInlineChild(i)
			if box, ok := obj.(BoxI); ok {
				box.SetContainingLayer(c.GetLayer())
				box.ConnectChildrenToCurrentLayer(c)
			}
		}
	}
}

func (b *InlineLayoutBox) PaintSelection(c *RenderingContext) {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		child := b.GetInlineChild(i)
		if inlineText, ok := child.(*InlineText); ok {
			inlineText.PaintSelection(c)
		}
	}
}

func (b *InlineLayoutBox) PaintInline(c *RenderingContext) {
	if !b.GetStyle().IsVisible() {
		return
	}

	b.PaintBackground(c)
	b.PaintBorder(c)

	if c.DebugDrawInlineBoxes() {
		b.PaintDebugOutline(c)
	}

	textDecorations := b.GetTextDecorations()
	if textDecorations != nil {
		for _, tD := range textDecorations {
			ident := tD.GetIdentValue()
			if ident == IdentValueUnderline || ident == IdentValueOverline {
				c.GetOutputDevice().DrawTextDecorationWithIBDecoration(c, b, tD)
			}
		}
	}

	for i := 0; i < b.GetInlineChildCount(); i++ {
		child := b.GetInlineChild(i)
		if inlineText, ok := child.(*InlineText); ok {
			inlineText.Paint(c)
		}
	}

	if textDecorations != nil {
		for _, tD := range textDecorations {
			ident := tD.GetIdentValue()
			if ident == IdentValueLineThrough {
				c.GetOutputDevice().DrawTextDecorationWithIBDecoration(c, b, tD)
			}
		}
	}
}

func (b *InlineLayoutBox) GetBorderSides() int {
	result := BorderPainterTop + BorderPainterBottom

	if b.startsHere {
		result += BorderPainterLeft
	}
	if b.endsHere {
		result += BorderPainterRight
	}

	return result
}

func (b *InlineLayoutBox) GetBorderEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	// x, y pins the content area of the box so subtract off top border and padding
	// too

	margin := b.getMarginLeftRight(cssCtx)

	border := b.GetBorder(cssCtx)
	padding := b.GetPadding(cssCtx)

	return geom.NewRectangle(
		calculatedStyleFloatToInt(float32(left)+margin.left),
		calculatedStyleFloatToInt(float32(top)-border.Top()-padding.Top()),
		calculatedStyleFloatToInt(float32(b.GetInlineWidthWithCssContext(cssCtx))-margin.left-margin.right),
		b.GetHeight())
}

// inlineLayoutBoxMarginLeftRight is the private record MarginLeftRight.
type inlineLayoutBoxMarginLeftRight struct {
	left  float32
	right float32
}

func (b *InlineLayoutBox) getMarginLeftRight(cssCtx CssContext) inlineLayoutBoxMarginLeftRight {
	var marginLeft float32 = 0
	var marginRight float32 = 0
	if b.startsHere || b.endsHere {
		margin := b.GetMargin(cssCtx)
		if b.startsHere {
			marginLeft = margin.Left()
		}
		if b.endsHere {
			marginRight = margin.Right()
		}
	}
	return inlineLayoutBoxMarginLeftRight{marginLeft, marginRight}
}

// GetMarginEdgeWithLeftTop is getMarginEdge(int left, int top, CssContext
// cssCtx, int tx, int ty).
func (b *InlineLayoutBox) GetMarginEdgeWithLeftTop(left int, top int, cssCtx CssContext, tx int, ty int) *geom.Rectangle {
	result := b.GetBorderEdge(left, top, cssCtx)
	margin := b.getMarginLeftRight(cssCtx)
	if margin.right > 0 {
		result.Width += calculatedStyleFloatToInt(margin.right)
	}
	if margin.left > 0 {
		result.X -= calculatedStyleFloatToInt(margin.left)
		result.Width += calculatedStyleFloatToInt(margin.left)
	}
	result.Translate(tx, ty)
	return result
}

func (b *InlineLayoutBox) GetContentAreaEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	border := b.GetBorder(cssCtx)
	padding := b.GetPadding(cssCtx)

	var marginLeft float32 = 0
	var marginRight float32 = 0

	var borderLeft float32 = 0
	var borderRight float32 = 0

	var paddingLeft float32 = 0
	var paddingRight float32 = 0

	if b.startsHere || b.endsHere {
		margin := b.GetMargin(cssCtx)
		if b.startsHere {
			marginLeft = margin.Left()
			borderLeft = border.Left()
			paddingLeft = padding.Left()
		}
		if b.endsHere {
			marginRight = margin.Right()
			borderRight = border.Right()
			paddingRight = padding.Right()
		}
	}

	return geom.NewRectangle(
		calculatedStyleFloatToInt(float32(left)+marginLeft+borderLeft+paddingLeft),
		calculatedStyleFloatToInt(float32(top)-border.Top()-padding.Top()),
		calculatedStyleFloatToInt(float32(b.GetInlineWidthWithCssContext(cssCtx))-marginLeft-borderLeft-paddingLeft-
			paddingRight-borderRight-marginRight),
		b.GetHeight())
}

func (b *InlineLayoutBox) GetLeftMarginBorderPadding(cssCtx CssContext) int {
	if b.startsHere {
		return b.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeLeft)
	} else {
		return 0
	}
}

func (b *InlineLayoutBox) GetRightMarginPaddingBorder(cssCtx CssContext) int {
	if b.endsHere {
		return b.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeRight)
	} else {
		return 0
	}
}

func (b *InlineLayoutBox) GetInlineWidth() int {
	return b.inlineWidth
}

func (b *InlineLayoutBox) SetInlineWidth(inlineWidth int) {
	b.inlineWidth = inlineWidth
}

func (b *InlineLayoutBox) IsContainsVisibleContent() bool {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		child := b.GetInlineChild(i)
		switch child := child.(type) {
		case *InlineText:
			if !child.IsEmpty() {
				return true
			}
		case *InlineLayoutBox:
			if child.IsContainsVisibleContent() {
				return true
			}
		case BoxI:
			if child.GetWidth() > 0 || child.GetHeight() > 0 {
				return true
			}
		default:
			panic(NewXRRuntimeException(fmt.Sprintf("Unexpected value: %v", child)))
		}
	}
	return false
}

func (b *InlineLayoutBox) IntersectsInlineBlocks(cssCtx CssContext, clip geom.Shape) bool {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		obj := b.GetInlineChild(i)

		switch obj := obj.(type) {
		case *InlineLayoutBox:
			possibleResult := obj.IntersectsInlineBlocks(cssCtx, clip)
			if possibleResult {
				return true
			}
		case BoxI:
			collector := NewBoxCollector()
			if collector.IntersectsAny(cssCtx, clip, obj) {
				return true
			}
		default:
		}
	}

	return false
}

// GetTextDecorations may return nil.
func (b *InlineLayoutBox) GetTextDecorations() []*TextDecoration {
	return b.textDecorations
}

// SetTextDecorations stores the list. Java rejects null here with a
// NullPointerException; an empty Go slice may be nil, so nil is accepted and
// reads as "no decorations".
func (b *InlineLayoutBox) SetTextDecorations(textDecoration []*TextDecoration) {
	b.textDecorations = textDecoration
}

func (b *InlineLayoutBox) addToContentList(list *[]BoxI) {
	*list = append(*list, b)

	for i := 0; i < b.GetInlineChildCount(); i++ {
		child := b.GetInlineChild(i)
		switch child := child.(type) {
		case *InlineLayoutBox:
			child.addToContentList(list)
		case BoxI:
			*list = append(*list, child)
		default:
		}
	}
}

func (b *InlineLayoutBox) GetLineBox() *LineBox {
	box := b.GetParent()
	for {
		if lineBox, ok := box.(*LineBox); ok {
			return lineBox
		}
		box = box.GetParent()
	}
}

func (b *InlineLayoutBox) GetElementWithContent() []BoxI {
	// inefficient, but the lists in question shouldn't be very long

	result := []BoxI{}

	container := b.GetLineBox().GetParent().(BlockBoxI)
	for {
		elementBoxes := container.GetElementBoxes(b.GetElement())
		for _, elementBox := range elementBoxes {
			elementBox.(*InlineLayoutBox).addToContentList(&result)
		}

		if _, ok := container.(*AnonymousBlockBox); !ok ||
			b.containsEnd(result) {
			break
		}

		following := b.addFollowingBlockBoxes(container, &result)

		if following == nil {
			break
		}
		container = following
	}

	return result
}

// addFollowingBlockBoxes may return nil.
func (b *InlineLayoutBox) addFollowingBlockBoxes(container BlockBoxI, result *[]BoxI) *AnonymousBlockBox {
	parent := container.GetParent()
	current := 0
	for ; current < parent.GetChildCount(); current++ {
		if parent.GetChild(current).AsBox() == container.AsBox() {
			current++
			break
		}
	}

	for ; current < parent.GetChildCount(); current++ {
		if _, ok := parent.GetChild(current).(*AnonymousBlockBox); ok {
			break
		} else {
			*result = append(*result, parent.GetChild(current))
		}
	}

	if current == parent.GetChildCount() {
		return nil
	}
	return parent.GetChild(current).(*AnonymousBlockBox)
}

func (b *InlineLayoutBox) containsEnd(result []BoxI) bool {
	for _, box := range result {
		if iB, ok := box.(*InlineLayoutBox); ok {
			if b.GetElement() == iB.GetElement() && iB.IsEndsHere() {
				return true
			}
		}
	}
	return false
}

func (b *InlineLayoutBox) GetElementBoxes(elem *dom.Element) []BoxI {
	result := []BoxI{}
	for i := 0; i < b.GetInlineChildCount(); i++ {
		child := b.GetInlineChild(i)
		if box, ok := child.(BoxI); ok {
			if box.GetElement() == elem {
				result = append(result, box)
			}
			result = append(result, box.GetElementBoxes(elem)...)
		}
	}
	return result
}

func (b *InlineLayoutBox) PositionRelative(cssCtx CssContext) *geom.Dimension {
	delta := b.Box.PositionRelative(cssCtx)

	b.SetX(b.GetX() - delta.Width)
	b.SetY(b.GetY() - delta.Height)

	toTranslate := b.GetElementWithContent()

	for _, box := range toTranslate {
		box.SetX(box.GetX() + delta.Width)
		box.SetY(box.GetY() + delta.Height)

		box.CalcCanvasLocation()
		box.CalcChildLocations()
	}

	return delta
}

// AddAllChildrenWithLayer is addAllChildren(List<Box> list, Layer layer); it
// appends to *list.
func (b *InlineLayoutBox) AddAllChildrenWithLayer(list *[]BoxI, layer *Layer) {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		child := b.GetInlineChild(i)
		if box, ok := child.(BoxI); ok {
			if box.GetContainingLayer() == layer {
				*list = append(*list, box)
				if inlineLayoutBox, ok := child.(*InlineLayoutBox); ok {
					inlineLayoutBox.AddAllChildrenWithLayer(list, layer)
				}
			}
		}
	}
}

func (b *InlineLayoutBox) PaintDebugOutline(c *RenderingContext) {
	c.GetOutputDevice().DrawDebugOutline(c, b, FSRGBColorBlue)
}

func (b *InlineLayoutBox) ResetChildren(c *LayoutContext) {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		object := b.GetInlineChild(i)
		if box, ok := object.(BoxI); ok {
			box.Reset(c)
		}
	}
}

// inlineLayoutBoxIsSameChild reports whether the inline child obj is the box
// child (Java compares the references).
func inlineLayoutBoxIsSameChild(obj InlineChild, child BoxI) bool {
	if child == nil {
		return obj == nil
	}
	box, ok := obj.(BoxI)
	return ok && box.AsBox() == child.AsBox()
}

// RemoveChildBox is removeChild(Box child): it removes the first occurrence
// of child from the inline children.
func (b *InlineLayoutBox) RemoveChildBox(child BoxI) {
	for i, obj := range b.inlineChildren {
		if inlineLayoutBoxIsSameChild(obj, child) {
			b.RemoveChildInt(i)
			return
		}
	}
}

// RemoveChildInt is removeChild(int i).
func (b *InlineLayoutBox) RemoveChildInt(i int) {
	if i < 0 || i >= len(b.inlineChildren) {
		panic(NewXRRuntimeException(fmt.Sprintf("Index %d out of bounds for length %d", i, len(b.inlineChildren))))
	}
	b.inlineChildren = append(b.inlineChildren[:i], b.inlineChildren[i+1:]...)
}

// GetPrevious may return nil. As in Java, the last inline child is never
// examined.
func (b *InlineLayoutBox) GetPrevious(child BoxI) BoxI {
	for i := 0; i < len(b.inlineChildren)-1; i++ {
		obj := b.inlineChildren[i]
		if inlineLayoutBoxIsSameChild(obj, child) {
			if i == 0 {
				return nil
			} else {
				previous := b.inlineChildren[i-1]
				if box, ok := previous.(BoxI); ok {
					return box
				}
				return nil
			}
		}
	}

	return nil
}

// GetNext may return nil.
func (b *InlineLayoutBox) GetNext(child BoxI) BoxI {
	for i := 0; i < len(b.inlineChildren)-1; i++ {
		obj := b.inlineChildren[i]
		if inlineLayoutBoxIsSameChild(obj, child) {
			next := b.inlineChildren[i+1]
			if box, ok := next.(BoxI); ok {
				return box
			}
			return nil
		}
	}

	return nil
}

func (b *InlineLayoutBox) CalcCanvasLocation() {
	lineBox := b.GetLineBox()
	b.SetAbsX(lineBox.GetAbsX() + b.GetX())
	b.SetAbsY(lineBox.GetAbsY() + b.GetY())
}

func (b *InlineLayoutBox) CalcChildLocations() {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		obj := b.GetInlineChild(i)
		if child, ok := obj.(BoxI); ok {
			child.CalcCanvasLocation()
			child.CalcChildLocations()
		}
	}
}

// ClearSelection appends the boxes whose selection changed to *modified.
func (b *InlineLayoutBox) ClearSelection(modified *[]BoxI) {
	changed := false
	for i := 0; i < b.GetInlineChildCount(); i++ {
		obj := b.GetInlineChild(i)
		switch obj := obj.(type) {
		case BoxI:
			obj.ClearSelection(modified)
		default:
			// As in Java, the text's selection is only cleared while no
			// earlier text of this box reported a change.
			changed = changed || obj.(*InlineText).ClearSelection()
		}
	}

	if changed {
		*modified = append(*modified, b)
	}
}

func (b *InlineLayoutBox) SelectAll() {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		obj := b.GetInlineChild(i)
		switch obj := obj.(type) {
		case BoxI:
			obj.SelectAll()
		case *InlineText:
			obj.SelectAll()
		default:
			panic(NewXRRuntimeException(fmt.Sprintf("Unexpected inline child type: %T", obj)))
		}
	}
}

func (b *InlineLayoutBox) CalcChildPaintingInfo(
	c CssContext, result *PaintingInfo, useCache bool) {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		obj := b.GetInlineChild(i)
		if box, ok := obj.(BoxI); ok {
			info := box.CalcPaintingInfo(c, useCache)
			b.MoveIfGreater(result.GetOuterMarginCorner(), info.GetOuterMarginCorner())
			result.GetAggregateBounds().Add(info.GetAggregateBounds())
		}
	}
}

func (b *InlineLayoutBox) LookForDynamicFunctions(c *RenderingContext) {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		obj := b.GetInlineChild(i)
		switch obj := obj.(type) {
		case *InlineText:
			if obj.IsDynamicFunction() {
				obj.UpdateDynamicValue(c)
			}
		case *InlineLayoutBox:
			obj.LookForDynamicFunctions(c)
		default:
		}
	}
}

// FindTrailingText may return nil.
func (b *InlineLayoutBox) FindTrailingText() *InlineText {
	if b.GetInlineChildCount() == 0 {
		return nil
	}

	var result *InlineText

	for offset := b.GetInlineChildCount() - 1; offset >= 0; offset-- {
		child := b.GetInlineChild(offset)
		switch child := child.(type) {
		case *InlineText:
			result = child
			if result.IsEmpty() {
				continue
			}
			return result
		case *InlineLayoutBox:
			result = child.FindTrailingText()
			if result != nil && result.IsEmpty() {
				continue
			}
			return result
		default:
			return nil
		}
	}

	return result
}

func (b *InlineLayoutBox) CalculateTextDecoration(c *LayoutContext) {
	decorations :=
		InlineBoxingCalculateTextDecorations(c, b, b.GetBaseline(), b.GetStyle().GetFSFontMetrics(c))
	b.SetTextDecorations(decorations)
}

// Find may return nil.
func (b *InlineLayoutBox) Find(cssCtx CssContext, absX int, absY int, findAnonymous bool) BoxI {
	pI := b.GetPaintingInfo()
	if pI != nil && !pI.GetAggregateBounds().Contains(absX, absY) {
		return nil
	}

	for i := 0; i < b.GetInlineChildCount(); i++ {
		child := b.GetInlineChild(i)
		if childBox, ok := child.(BoxI); ok {
			result := childBox.Find(cssCtx, absX, absY, findAnonymous)
			if result != nil {
				return result
			}
		}
	}

	edge := b.GetContentAreaEdge(b.GetAbsX(), b.GetAbsY(), cssCtx)
	var result BoxI
	if edge.Contains(absX, absY) && b.GetStyle().IsVisible() {
		result = b
	}

	if !findAnonymous && result != nil && b.GetElement() == nil {
		return b.GetParent().GetParent()
	} else {
		return result
	}
}

func (b *InlineLayoutBox) GetContainingBlockWidth() int {
	return b.containingBlockWidth
}

func (b *InlineLayoutBox) ToString() string {
	var result strings.Builder
	result.WriteString("InlineLayoutBox")
	result.WriteString(": ")
	if b.GetElement() != nil {
		result.WriteString("<")
		result.WriteString(b.GetElement().GetNodeName())
		result.WriteString("> ")
	} else {
		result.WriteString("(anonymous) ")
	}
	if b.IsStartsHere() || b.IsEndsHere() {
		result.WriteString("(")
		if b.IsStartsHere() {
			result.WriteString("S")
		}
		if b.IsEndsHere() {
			result.WriteString("E")
		}
		result.WriteString(") ")
	}
	result.WriteString("(baseline=")
	result.WriteString(fmt.Sprint(b.baseline))
	result.WriteString(") ")
	b.AppendPosition(&result)
	b.AppendSize(&result)
	return result.String()
}

func (b *InlineLayoutBox) String() string {
	return b.ToString()
}

func (b *InlineLayoutBox) Dump(c *LayoutContext, indent string, which BoxDump) string {
	if which != BoxDumpRender {
		panic(NewXRRuntimeException(fmt.Sprintf("Which %v not supported", which)))
	}

	result := indent
	result += b.ToString()

	for _, obj := range b.GetInlineChildren() {
		result += "\n"
		if box, ok := obj.(BoxI); ok {
			result += box.Dump(c, indent+"  ", which)
			if result[len(result)-1] == '\n' {
				result = result[:len(result)-1]
			}
		} else {
			result += indent + "  " + fmt.Sprint(obj)
		}
	}

	return result
}

func (b *InlineLayoutBox) Restyle(c *LayoutContext) {
	b.Box.Restyle(c)
	b.CalculateTextDecoration(c)
}

func (b *InlineLayoutBox) RestyleChildren(c *LayoutContext) {
	for i := 0; i < b.GetInlineChildCount(); i++ {
		obj := b.GetInlineChild(i)
		if box, ok := obj.(BoxI); ok {
			box.Restyle(c)
		}
	}
}

func (b *InlineLayoutBox) GetRestyleTarget() BoxI {
	// Inline boxes may be broken across lines so back out
	// to the nearest block box
	result := b.GetParent()
	for {
		if _, ok := result.(*InlineLayoutBox); !ok {
			break
		}
		result = result.GetParent()
	}
	return result.GetParent()
}

func (b *InlineLayoutBox) CollectText(c *RenderingContext, buffer *strings.Builder) error {
	for _, obj := range b.GetInlineChildren() {
		if inlineText, ok := obj.(*InlineText); ok {
			buffer.WriteString(inlineText.GetTextExportText())
		} else if box, ok := obj.(BoxI); ok {
			if err := box.CollectText(c, buffer); err != nil {
				return err
			}
		} else {
			panic(NewXRRuntimeException(fmt.Sprintf("Unexpected inline child type: %T", obj)))
		}
	}
	return nil
}

func (b *InlineLayoutBox) CountJustifiableChars(counts *CharCounts) {
	justifyThis := b.GetStyle().IsTextJustify()
	for _, o := range b.GetInlineChildren() {
		if inlineLayoutBox, ok := o.(*InlineLayoutBox); ok {
			inlineLayoutBox.CountJustifiableChars(counts)
		} else if inlineText, ok := o.(*InlineText); ok && justifyThis {
			inlineText.CountJustifiableChars(counts)
		}
	}
}

func (b *InlineLayoutBox) AdjustHorizontalPosition(info *JustificationInfo, adjust float32) float32 {
	runningTotal := adjust

	var result float32 = 0.0

	for _, o := range b.GetInlineChildren() {
		switch o := o.(type) {
		case *InlineText:
			o.SetX(o.GetX() + propertyBuilderRound(result))

			adj := o.CalcTotalAdjustment(info)
			result += adj
			runningTotal += adj
		case BoxI:
			o.SetX(o.GetX() + propertyBuilderRound(runningTotal))

			if iB, ok := o.(*InlineLayoutBox); ok {
				adj := iB.AdjustHorizontalPosition(info, runningTotal)
				result += adj
				runningTotal += adj
			}
		default:
		}
	}

	return result
}

func (b *InlineLayoutBox) GetEffectiveWidth() int {
	return b.GetInlineWidth()
}
