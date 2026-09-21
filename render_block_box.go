// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/BlockBox.java

package ufo

import (
	"fmt"
	"math"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

// BlockBoxPosition ports the enum BlockBox.Position.
type BlockBoxPosition int

const (
	BlockBoxPositionVertically BlockBoxPosition = iota
	BlockBoxPositionHorizontally
	BlockBoxPositionBoth
)

// BlockBoxContentType ports the enum BlockBox.ContentType.
type BlockBoxContentType int

const (
	BlockBoxContentTypeUnknown BlockBoxContentType = iota
	BlockBoxContentTypeInline
	BlockBoxContentTypeBlock
	BlockBoxContentTypeEmpty
)

// String returns the Java enum constant name.
func (t BlockBoxContentType) String() string {
	switch t {
	case BlockBoxContentTypeUnknown:
		return "UNKNOWN"
	case BlockBoxContentTypeInline:
		return "INLINE"
	case BlockBoxContentTypeBlock:
		return "BLOCK"
	case BlockBoxContentTypeEmpty:
		return "EMPTY"
	}
	return fmt.Sprintf("%d", int(t))
}

// String returns the Java enum constant name.
func (p BlockBoxPosition) String() string {
	switch p {
	case BlockBoxPositionVertically:
		return "VERTICALLY"
	case BlockBoxPositionHorizontally:
		return "HORIZONTALLY"
	case BlockBoxPositionBoth:
		return "BOTH"
	}
	return fmt.Sprintf("%d", int(p))
}

// BlockBoxNoBaseline is BlockBox.NO_BASELINE (Integer.MIN_VALUE).
const BlockBoxNoBaseline = math.MinInt32

// BlockBoxI lists every non-private method of BlockBox, including the ones
// inherited from Box. A Java value typed BlockBox is a BlockBoxI in Go.
type BlockBoxI interface {
	BoxI

	AsBlockBox() *BlockBox

	CopyOf() BlockBoxI
	GetExtraBoxDescription() string
	PaintListMarker(c *RenderingContext)
	PaintInline(c *RenderingContext)
	IsInline() bool
	GetLineBox() *LineBox
	PaintDebugOutline(c *RenderingContext)
	GetMarkerData() *MarkerData
	GetListCounter() int
	SetListCounter(listCounter int)
	GetPersistentBFC() *PersistentBFC
	SetPersistentBFC(persistentBFC *PersistentBFC)
	GetStaticEquivalent() BoxI
	SetStaticEquivalent(staticEquivalent BoxI)
	IsReplaced() bool
	CalcInitialFloatedCanvasLocation(c *LayoutContext)
	IsNeedPageClear() bool
	SetNeedPageClear(needPageClear bool)
	PositionAbsolute(cssCtx CssContext, direction BlockBoxPosition)
	PositionAbsoluteOnPage(c *LayoutContext)
	GetReplacedElement() ReplacedElement
	SetReplacedElement(replacedElement ReplacedElement)
	ResolveAutoMargins(c *LayoutContext, cssWidth int, padding RectPropertySetI, border *BorderPropertySet)
	CalcDimensions(c *LayoutContext)
	CalcDimensionsWithCssWidth(c *LayoutContext, cssWidth int)
	Layout(c *LayoutContext)
	LayoutWithContentStart(c *LayoutContext, contentStart int)
	IsAllowHeightToShrink() bool
	GetPageClearance() int
	CalcLayoutHeight(c *LayoutContext, border *BorderPropertySet, margin RectPropertySetI, padding RectPropertySetI)
	ApplyCSSMinMaxWidth(c CssContext)
	EnsureChildren(c *LayoutContext)
	BuildContainerAndAnalyzePageBreaks(c *LayoutContext, container *ContentLimitContainer) *ContentLimitContainer
	LayoutChildren(c *LayoutContext, contentStart int)
	LayoutInlineChildren(c *LayoutContext, contentStart int, breakAtLine int, tryAgain bool)
	GetChildrenContentType() BlockBoxContentType
	SetChildrenContentType(contentType BlockBoxContentType)
	GetInlineContent() []Styleable
	SetInlineContent(inlineContent []Styleable)
	IsSkipWhenCollapsingMargins() bool
	IsMayCollapseMarginsWithChildren() bool
	IsTopMarginCalculated() bool
	SetTopMarginCalculated(topMarginCalculated bool)
	IsBottomMarginCalculated() bool
	SetBottomMarginCalculated(bottomMarginCalculated bool)
	GetCSSWidth(c CssContext) int
	GetCSSWidthWithShrinkingToFit(c CssContext, shrinkingToFit bool) int
	GetCSSFitToWidth(c CssContext) int
	GetCSSHeight(c CssContext) int
	IsAutoHeight() bool
	GetAvailableWidth(c *LayoutContext) int
	IsFixedWidthAdvisoryOnly() bool
	CalcMinMaxWidth(c *LayoutContext)
	GetMaxWidth() int
	SetMaxWidth(maxWidth int)
	GetMinWidth() int
	SetMinWidth(minWidth int)
	StyleText(c *LayoutContext)
	StyleTextWithStyle(c *LayoutContext, style CalculatedStyleI)
	GetFirstLetterStyle() *CascadedStyle
	SetFirstLetterStyle(firstLetterStyle *CascadedStyle)
	GetFirstLineStyle() *CascadedStyle
	SetFirstLineStyle(firstLineStyle *CascadedStyle)
	IsMinMaxCalculated() bool
	SetMinMaxCalculated(minMaxCalculated bool)
	SetDimensionsCalculated(dimensionsCalculated bool)
	SetNeedShrinkToFitCalculation(needShrinkToFitCalculation bool)
	InitStaticPos(c *LayoutContext, parent BlockBoxI, childOffset int)
	CalcBaseline(c *LayoutContext) int
	CalcInitialBreakAtLine(c *LayoutContext) int
	IsCurrentBreakAtLineContext(c *LayoutContext) bool
	CalcBreakAtLineContext(c *LayoutContext) *BreakAtLineContext
	CalcInlineBaseline(c CssContext) int
	FindOffset(box BoxI) int
	FindLastNthLineBox(count int) *LineBox
	IsNeedsKeepWithInline(c *LayoutContext) bool
	IsFloated() bool
	GetFloatedBoxData() *FloatedBoxData
	SetFloatedBoxData(floatedBoxData *FloatedBoxData)
	GetChildrenHeight() int
	SetChildrenHeight(childrenHeight int)
	IsFromCaptionedTable() bool
	SetFromCaptionedTable(fromTable bool)
	IsInMainFlow() bool
	IsContainsInlineContent(c *LayoutContext) bool
	CheckPageContext(c *LayoutContext) bool
	IsNeedsClipOnPaint(c *RenderingContext) bool
	PropagateExtraSpace(c *LayoutContext, parentContainer *ContentLimitContainer, currentContainer *ContentLimitContainer, extraTop int, extraBottom int)
}

// BlockBox is a block box as defined in the CSS spec. It also provides a base
// class for other kinds of block content (for example table rows or cells).
type BlockBox struct {
	Box

	markerData *MarkerData

	listCounter int

	persistentBFC *PersistentBFC

	staticEquivalent BoxI

	needPageClear bool

	replacedElement ReplacedElement

	childrenContentType BlockBoxContentType

	inlineContent []Styleable

	topMarginCalculated    bool
	bottomMarginCalculated bool

	pendingCollapseCalculation *blockBoxMarginCollapseResult

	minWidth         int
	maxWidth         int
	minMaxCalculated bool

	dimensionsCalculated       bool
	needShrinkToFitCalculation bool

	firstLineStyle   *CascadedStyle
	firstLetterStyle *CascadedStyle

	floatedBoxData *FloatedBoxData

	childrenHeight int

	fromCaptionedTable bool
}

// initBlockBox is the body of the constructor
// BlockBox(Element element, CalculatedStyle style, boolean anonymous), together
// with BlockBox's field initializers, for a BlockBox embedded in another
// struct. It does not set self: the constructor of the outermost struct calls
// SetSelf next, before it calls any method. See initBox for setStyle.
func initBlockBox(b *BlockBox, element *dom.Element, style CalculatedStyleI, anonymous bool) {
	initBoxWithElement(&b.Box, element, style, anonymous)
	b.childrenContentType = BlockBoxContentTypeUnknown
}

// NewBlockBox is the protected constructor BlockBox().
func NewBlockBox() *BlockBox {
	return NewBlockBoxWithElementStyleAnonymous(nil, nil, false)
}

// NewBlockBoxWithElementStyleAnonymous is the constructor
// BlockBox(Element element, CalculatedStyle style, boolean anonymous).
// element and style may be nil.
func NewBlockBoxWithElementStyleAnonymous(element *dom.Element, style CalculatedStyleI, anonymous bool) *BlockBox {
	b := &BlockBox{}
	initBlockBox(b, element, style, anonymous)
	b.SetSelf(b)
	return b
}

func (b *BlockBox) AsBlockBox() *BlockBox {
	return b
}

// selfBlock is the outermost object, through which every call to a method
// that a subclass overrides is made.
func (b *BlockBox) selfBlock() BlockBoxI {
	return b.self.(BlockBoxI)
}

// blockBoxFromBox is the Java cast (BlockBox) box, which lets null through.
func blockBoxFromBox(box BoxI) BlockBoxI {
	if box == nil {
		return nil
	}
	return box.(BlockBoxI)
}

func (b *BlockBox) CopyOf() BlockBoxI {
	return NewBlockBoxWithElementStyleAnonymous(b.GetElement(), b.GetStyle(), b.IsAnonymous())
}

func (b *BlockBox) GetExtraBoxDescription() string {
	return ""
}

func (b *BlockBox) ToString() string {
	s := b.selfBlock()
	var result strings.Builder
	result.WriteString(boxSimpleClassName(s))
	result.WriteString(": ")
	if b.GetElement() != nil && !b.IsAnonymous() {
		result.WriteString("<")
		result.WriteString(b.GetElement().GetNodeName())
		id := b.GetElement().GetAttribute("id")
		if id != "" {
			result.WriteString(" id=\"")
			result.WriteString(id)
			result.WriteString("\"")
		}
		result.WriteString("> ")
	}
	if b.IsAnonymous() {
		result.WriteString("(anonymous) ")
	}
	if b.GetPseudoElementOrClass() != "" {
		result.WriteByte(':')
		result.WriteString(b.GetPseudoElementOrClass())
		result.WriteByte(' ')
	}
	result.WriteByte('(')
	result.WriteString(b.GetStyle().GetIdent(CSSNameDisplay).ToString())
	result.WriteString(") ")

	if b.GetStyle().IsRunning() {
		result.WriteString("(running) ")
	}

	switch b.GetChildrenContentType() {
	case BlockBoxContentTypeBlock:
		result.WriteString("(B) ")
	case BlockBoxContentTypeInline:
		result.WriteString("(I) ")
	case BlockBoxContentTypeEmpty:
		result.WriteString("(E) ")
	case BlockBoxContentTypeUnknown:
	}

	result.WriteString(s.GetExtraBoxDescription())

	UtilsAppendPositioningInfo(b.GetStyle(), &result)
	s.AppendPosition(&result)
	s.AppendSize(&result)
	return boxJavaTrim(result.String())
}

func (b *BlockBox) String() string {
	return b.self.ToString()
}

func (b *BlockBox) Dump(c *LayoutContext, indent string, which BoxDump) string {
	s := b.selfBlock()
	var result strings.Builder
	result.WriteString(indent)

	s.EnsureChildren(c)

	result.WriteString(s.ToString())

	margin := s.GetMargin(c)
	result.WriteString(" effMargin=[")
	result.WriteString(propertyBuilderFloatToString(margin.Top()))
	result.WriteString(", ")
	result.WriteString(propertyBuilderFloatToString(margin.Right()))
	result.WriteString(", ")
	result.WriteString(propertyBuilderFloatToString(margin.Bottom()))
	result.WriteString(", ")
	result.WriteString(propertyBuilderFloatToString(margin.Right()))
	result.WriteString("] ")
	styleMargin := s.GetStyleMargin(c)
	result.WriteString(" styleMargin=[")
	result.WriteString(propertyBuilderFloatToString(styleMargin.Top()))
	result.WriteString(", ")
	result.WriteString(propertyBuilderFloatToString(styleMargin.Right()))
	result.WriteString(", ")
	result.WriteString(propertyBuilderFloatToString(styleMargin.Bottom()))
	result.WriteString(", ")
	result.WriteString(propertyBuilderFloatToString(styleMargin.Right()))
	result.WriteString("] ")

	if b.GetChildrenContentType() != BlockBoxContentTypeEmpty {
		result.WriteByte('\n')
	}

	switch b.GetChildrenContentType() {
	case BlockBoxContentTypeBlock:
		s.DumpBoxes(c, indent, b.GetChildren(), which, &result)
	case BlockBoxContentTypeInline:
		if which == BoxDumpRender {
			s.DumpBoxes(c, indent, b.GetChildren(), which, &result)
		} else {
			inlineContent := b.GetInlineContent()
			for i, styleable := range inlineContent {
				if block, ok := styleable.(BlockBoxI); ok {
					dumped := block.Dump(c, indent+"  ", which)
					// Java appends the dump and then deletes a trailing
					// newline from the accumulated text.
					result.WriteString(dumped)
					if accumulated := result.String(); accumulated[len(accumulated)-1] == '\n' {
						result.Reset()
						result.WriteString(accumulated[:len(accumulated)-1])
					}
				} else {
					result.WriteString(indent)
					result.WriteString("  ")
					result.WriteString(fmt.Sprint(styleable))
				}
				if i < len(inlineContent)-1 {
					result.WriteByte('\n')
				}
			}
		}
	}

	return result.String()
}

func (b *BlockBox) PaintListMarker(c *RenderingContext) {
	if !b.GetStyle().IsVisible() {
		return
	}

	if b.GetStyle().IsListItem() {
		ListItemPainterPaint(c, b.selfBlock())
	}
}

func (b *BlockBox) GetPaintingClipEdge(cssCtx CssContext) *geom.Rectangle {
	result := b.Box.GetPaintingClipEdge(cssCtx)

	// HACK Don't know how wide the list marker is (or even where it is)
	// so extend the bounding box all the way over to the left edge of
	// the canvas
	if b.GetStyle().IsListItem() {
		delta := result.X
		result.X = 0
		result.Width += delta
	}

	return result
}

func (b *BlockBox) PaintInline(c *RenderingContext) {
	if !b.GetStyle().IsVisible() {
		return
	}

	b.GetContainingLayer().PaintAsLayer(c, b.selfBlock())
}

func (b *BlockBox) IsInline() bool {
	parent := b.GetParent()
	switch parent.(type) {
	case *LineBox, *InlineLayoutBox:
		return true
	}
	return false
}

// GetLineBox may return nil.
func (b *BlockBox) GetLineBox() *LineBox {
	if !b.IsInline() {
		return nil
	}
	box := b.GetParent()
	for {
		if lineBox, ok := box.(*LineBox); ok {
			return lineBox
		}
		box = box.GetParent()
	}
}

func (b *BlockBox) PaintDebugOutline(c *RenderingContext) {
	c.GetOutputDevice().DrawDebugOutline(c, b.self, FSRGBColorRed)
}

// GetMarkerData may return nil.
func (b *BlockBox) GetMarkerData() *MarkerData {
	return b.markerData
}

func (b *BlockBox) initMarkerData(c *LayoutContext) *MarkerData {
	if b.markerData == nil {
		b.markerData = b.createMarkerData(c)
	}
	return b.markerData
}

func (b *BlockBox) createMarkerData(c *LayoutContext) *MarkerData {
	strutMetrics := InlineBoxingCreateDefaultStrutMetrics(c, b.self)

	style := b.GetStyle()
	listStyle := style.GetIdent(CSSNameListStyleType)

	image := style.GetStringProperty(CSSNameListStyleImage)
	imageMarker := b.makeImageMarker(c, strutMetrics, image)
	if imageMarker != nil {
		return NewMarkerData(strutMetrics, imageMarker, nil, nil)
	} else if listStyle == IdentValueCircle || listStyle == IdentValueSquare || listStyle == IdentValueDisc {
		return NewMarkerData(strutMetrics, nil, b.makeGlyphMarker(strutMetrics), nil)
	} else if listStyle != IdentValueNone {
		return NewMarkerData(strutMetrics, nil, nil, b.makeTextMarker(c, listStyle))
	} else {
		return NewMarkerData(strutMetrics, nil, nil, nil)
	}
}

func (b *BlockBox) makeGlyphMarker(strutMetrics *StrutMetrics) *MarkerDataGlyphMarker {
	diameter := boxInt((strutMetrics.GetAscent() + strutMetrics.GetDescent()) / 3)

	return NewMarkerDataGlyphMarker(diameter, diameter*3)
}

func (b *BlockBox) makeImageMarker(c *LayoutContext, structMetrics *StrutMetrics, image string) *MarkerDataImageMarker {
	if image == "none" {
		return nil
	}
	img := c.GetUac().GetImageResource(image).GetImage()
	if img == nil {
		return nil
	}
	if float32(img.GetHeight()) > structMetrics.GetAscent() {
		img = img.Scale(-1, boxInt(structMetrics.GetAscent()))
	}
	return NewMarkerDataImageMarker(img, img.GetWidth()*2)
}

func (b *BlockBox) makeTextMarker(c *LayoutContext, listStyle *IdentValue) *MarkerDataTextMarker {
	listCounter := b.GetListCounter()
	text := CounterFunctionCreateCounterText(listStyle, listCounter) + ".  "

	w := c.GetTextRenderer().GetWidth(
		c.GetFontContext(),
		b.GetStyle().GetFSFont(c),
		text)

	return NewMarkerDataTextMarker(text, w)
}

func (b *BlockBox) GetListCounter() int {
	return b.listCounter
}

func (b *BlockBox) SetListCounter(listCounter int) {
	b.listCounter = listCounter
}

// GetPersistentBFC may return nil.
func (b *BlockBox) GetPersistentBFC() *PersistentBFC {
	return b.persistentBFC
}

func (b *BlockBox) SetPersistentBFC(persistentBFC *PersistentBFC) {
	b.persistentBFC = persistentBFC
}

// GetStaticEquivalent may return nil.
func (b *BlockBox) GetStaticEquivalent() BoxI {
	return b.staticEquivalent
}

func (b *BlockBox) SetStaticEquivalent(staticEquivalent BoxI) {
	b.staticEquivalent = staticEquivalent
}

func (b *BlockBox) IsReplaced() bool {
	return b.replacedElement != nil
}

func (b *BlockBox) CalcCanvasLocation() {
	if b.IsFloated() {
		manager := b.floatedBoxData.GetManager()
		if manager != nil {
			offset := manager.GetOffset(b.selfBlock())
			b.SetAbsX(manager.GetMaster().GetAbsX() + b.GetX() - offset.X)
			b.SetAbsY(manager.GetMaster().GetAbsY() + b.GetY() - offset.Y)
		}
	}

	lineBox := b.GetLineBox()
	if lineBox == nil {
		parent := b.GetParent()
		if parent != nil {
			b.SetAbsX(parent.GetAbsX() + parent.GetTx() + b.GetX())
			b.SetAbsY(parent.GetAbsY() + parent.GetTy() + b.GetY())
		} else if b.IsStyled() && b.GetStyle().IsAbsFixedOrInlineBlockEquiv() {
			cb := b.GetContainingBlock()
			if cb != nil {
				b.SetAbsX(cb.GetAbsX() + b.GetX())
				b.SetAbsY(cb.GetAbsY() + b.GetY())
			}
		}
	} else {
		b.SetAbsX(lineBox.GetAbsX() + b.GetX())
		b.SetAbsY(lineBox.GetAbsY() + b.GetY())
	}

	if b.IsReplaced() {
		location := b.GetReplacedElement().GetLocation()
		if location.X != b.GetAbsX() || location.Y != b.GetAbsY() {
			b.GetReplacedElement().SetLocation(b.GetAbsX(), b.GetAbsY())
		}
	}
}

func (b *BlockBox) CalcInitialFloatedCanvasLocation(c *LayoutContext) {
	offset := c.GetBlockFormattingContext().GetOffset()
	manager := c.GetBlockFormattingContext().GetFloatManager()
	b.SetAbsX(manager.GetMaster().GetAbsX() + b.GetX() - offset.X)
	b.SetAbsY(manager.GetMaster().GetAbsY() + b.GetY() - offset.Y)
}

func (b *BlockBox) CalcChildLocations() {
	b.Box.CalcChildLocations()

	if b.persistentBFC != nil {
		b.persistentBFC.GetFloatManager().CalcFloatLocations()
	}
}

func (b *BlockBox) IsNeedPageClear() bool {
	return b.needPageClear
}

func (b *BlockBox) SetNeedPageClear(needPageClear bool) {
	b.needPageClear = needPageClear
}

func (b *BlockBox) alignToStaticEquivalent() {
	if b.staticEquivalent.GetAbsY() != b.GetAbsY() {
		b.SetY(b.staticEquivalent.GetAbsY() - b.GetAbsY())
		b.SetAbsY(b.staticEquivalent.GetAbsY())
	}
}

func (b *BlockBox) PositionAbsolute(cssCtx CssContext, direction BlockBoxPosition) {
	s := b.selfBlock()
	style := b.GetStyle()
	cbContentHeight := b.GetContainingBlock().GetContentAreaEdge(0, 0, cssCtx).Height

	var boundingBox *geom.Rectangle
	if _, ok := b.GetContainingBlock().(BlockBoxI); ok {
		boundingBox = b.GetContainingBlock().GetPaddingEdge(0, 0, cssCtx)
	} else {
		boundingBox = b.GetContainingBlock().GetContentAreaEdge(0, 0, cssCtx)
	}

	if direction == BlockBoxPositionHorizontally || direction == BlockBoxPositionBoth {
		b.SetX(0)
		if !style.IsIdent(CSSNameLeft, IdentValueAuto) {
			b.SetX(boxInt(style.GetFloatPropertyProportionalWidth(CSSNameLeft, float32(b.GetContainingBlock().GetContentWidth()), cssCtx)))
		} else if !style.IsIdent(CSSNameRight, IdentValueAuto) {
			b.SetX(boundingBox.Width -
				boxInt(style.GetFloatPropertyProportionalWidth(CSSNameRight, float32(b.GetContainingBlock().GetContentWidth()), cssCtx)) - s.GetWidth())
		}
		b.SetX(b.GetX() + boundingBox.X)
	}

	if direction == BlockBoxPositionVertically || direction == BlockBoxPositionBoth {
		b.SetY(0)
		if !style.IsIdent(CSSNameTop, IdentValueAuto) {
			b.SetY(boxInt(style.GetFloatPropertyProportionalHeight(CSSNameTop, float32(cbContentHeight), cssCtx)))
		} else if !style.IsIdent(CSSNameBottom, IdentValueAuto) {
			b.SetY(boundingBox.Height -
				boxInt(style.GetFloatPropertyProportionalWidth(CSSNameBottom, float32(cbContentHeight), cssCtx)) - s.GetHeight())
		}

		// Can't do this before now because our containing block
		// must be completed laid out
		pinnedHeight := b.calcPinnedHeight(cssCtx)
		if pinnedHeight != -1 && s.GetCSSHeight(cssCtx) == -1 {
			b.SetHeight(pinnedHeight)
			b.applyCSSMinMaxHeight(cssCtx)
		}

		b.SetY(b.GetY() + boundingBox.Y)
	}

	s.CalcCanvasLocation()

	if (direction == BlockBoxPositionVertically || direction == BlockBoxPositionBoth) &&
		b.GetStyle().IsTopAuto() && b.GetStyle().IsBottomAuto() {
		b.alignToStaticEquivalent()
	}

	s.CalcChildLocations()
}

func (b *BlockBox) PositionAbsoluteOnPage(c *LayoutContext) {
	s := b.selfBlock()
	if c.IsPrint() &&
		(b.GetStyle().IsForcePageBreakBefore() || b.IsNeedPageClear()) {
		s.ForcePageBreakBefore(c, b.GetStyle().GetIdent(CSSNamePageBreakBefore), false)
		s.CalcCanvasLocation()
		s.CalcChildLocations()

		b.SetNeedPageClear(false)
	}
}

// GetReplacedElement may return nil.
func (b *BlockBox) GetReplacedElement() ReplacedElement {
	return b.replacedElement
}

func (b *BlockBox) SetReplacedElement(replacedElement ReplacedElement) {
	b.replacedElement = replacedElement
}

func (b *BlockBox) Reset(c *LayoutContext) {
	b.Box.Reset(c)
	b.SetTopMarginCalculated(false)
	b.SetBottomMarginCalculated(false)
	b.SetDimensionsCalculated(false)
	b.SetMinMaxCalculated(false)
	b.SetChildrenHeight(0)
	if b.IsReplaced() {
		b.GetReplacedElement().Detach(c)
		b.SetReplacedElement(nil)
	}
	if b.GetChildrenContentType() == BlockBoxContentTypeInline {
		b.RemoveAllChildren()
	}

	if b.IsFloated() {
		b.floatedBoxData.GetManager().RemoveFloat(b.selfBlock())
		b.floatedBoxData.GetDrawingLayer().RemoveFloat(b.selfBlock())
	}

	if b.GetStyle().IsRunning() {
		c.GetRootLayer().RemoveRunningBlock(b.selfBlock())
	}
}

func (b *BlockBox) calcPinnedContentWidth(c CssContext) int {
	if !b.GetStyle().IsIdent(CSSNameLeft, IdentValueAuto) &&
		!b.GetStyle().IsIdent(CSSNameRight, IdentValueAuto) {
		paddingEdge := b.GetContainingBlock().GetPaddingEdge(0, 0, c)

		left := boxInt(b.GetStyle().GetFloatPropertyProportionalTo(
			CSSNameLeft, float32(paddingEdge.Width), c))
		right := boxInt(b.GetStyle().GetFloatPropertyProportionalTo(
			CSSNameRight, float32(paddingEdge.Width), c))

		result := paddingEdge.Width - left - right - b.GetLeftMBP() - b.GetRightMBP()
		return max(result, 0)
	}

	return -1
}

func (b *BlockBox) calcPinnedHeight(c CssContext) int {
	if !b.GetStyle().IsIdent(CSSNameTop, IdentValueAuto) &&
		!b.GetStyle().IsIdent(CSSNameBottom, IdentValueAuto) {
		paddingEdge := b.GetContainingBlock().GetPaddingEdge(0, 0, c)

		top := boxInt(b.GetStyle().GetFloatPropertyProportionalTo(
			CSSNameTop, float32(paddingEdge.Height), c))
		bottom := boxInt(b.GetStyle().GetFloatPropertyProportionalTo(
			CSSNameBottom, float32(paddingEdge.Height), c))

		result := paddingEdge.Height - top - bottom
		return max(result, 0)
	}

	return -1
}

func (b *BlockBox) ResolveAutoMargins(c *LayoutContext, cssWidth int, padding RectPropertySetI, border *BorderPropertySet) {
	s := b.selfBlock()
	withoutMargins :=
		boxInt(border.Left()) + boxInt(padding.Left()) +
			cssWidth +
			boxInt(padding.Right()) + boxInt(border.Right())
	if withoutMargins < s.GetContainingBlockWidth() {
		available := s.GetContainingBlockWidth() - withoutMargins

		autoLeft := b.GetStyle().IsAutoLeftMargin()
		autoRight := b.GetStyle().IsAutoRightMargin()

		if autoLeft && autoRight {
			b.SetMarginLeft(c, available/2)
			b.SetMarginRight(c, available/2)
		} else if autoLeft {
			b.SetMarginLeft(c, available)
		} else if autoRight {
			b.SetMarginRight(c, available)
		}
	}
}

func (b *BlockBox) calcEffPageRelativeWidth(c *LayoutContext) int {
	totalLeftMBP := 0
	totalRightMBP := 0

	usePageRelativeWidth := true

	current := b.self
	for {
		style := current.GetStyle()
		if style.IsAutoWidth() && !style.IsCanBeShrunkToFit() {
			totalLeftMBP += current.GetLeftMBP()
			totalRightMBP += current.GetRightMBP()
		} else {
			usePageRelativeWidth = false
			break
		}

		if current.GetContainingBlock().IsInitialContainingBlock() {
			break
		} else {
			current = current.GetContainingBlock()
		}
	}

	if usePageRelativeWidth {
		currentPage := c.GetRootLayer().GetFirstPage(c, b.self)
		return currentPage.GetContentWidth(c) - totalLeftMBP - totalRightMBP
	} else {
		return b.self.GetContainingBlockWidth() - b.GetLeftMBP() - b.GetRightMBP()
	}
}

func (b *BlockBox) CalcDimensions(c *LayoutContext) {
	s := b.selfBlock()
	s.CalcDimensionsWithCssWidth(c, s.GetCSSWidth(c))
}

func (b *BlockBox) CalcDimensionsWithCssWidth(c *LayoutContext, cssWidth int) {
	s := b.selfBlock()
	if !b.isDimensionsCalculated() {
		style := b.GetStyle()

		padding := s.GetPadding(c)
		border := s.GetBorder(c)

		if cssWidth != -1 && !b.IsAnonymous() &&
			(b.GetStyle().IsIdent(CSSNameMarginLeft, IdentValueAuto) ||
				b.GetStyle().IsIdent(CSSNameMarginRight, IdentValueAuto)) &&
			b.GetStyle().IsNeedAutoMarginResolution() {
			s.ResolveAutoMargins(c, cssWidth, padding, border)
		}

		b.recalculateMargin(c)
		margin := s.GetMargin(c)

		// CLEAN: cast to int
		b.SetLeftMBP(boxInt(margin.Left()) + boxInt(border.Left()) + boxInt(padding.Left()))
		b.SetRightMBP(boxInt(padding.Right()) + boxInt(border.Right()) + boxInt(margin.Right()))
		if c.IsPrint() && b.GetStyle().IsDynamicAutoWidth() {
			b.SetContentWidth(b.calcEffPageRelativeWidth(c))
		} else {
			b.SetContentWidth(s.GetContainingBlockWidth() - b.GetLeftMBP() - b.GetRightMBP())
		}
		b.SetHeight(0)

		if !b.IsAnonymous() || b.IsFromCaptionedTable() && b.IsFloated() {
			pinnedContentWidth := -1

			borderBox := style.IsBorderBox()

			if cssWidth != -1 {
				if borderBox {
					b.SetContentWidth(cssWidth - boxInt(border.Width()) - boxInt(padding.Width()))
				} else {
					b.SetContentWidth(cssWidth)
				}
			} else if b.GetStyle().IsAbsolute() || b.GetStyle().IsFixed() {
				pinnedContentWidth = b.calcPinnedContentWidth(c)
				if pinnedContentWidth != -1 {
					b.SetContentWidth(pinnedContentWidth)
				}
			}

			cssHeight := s.GetCSSHeight(c)
			if cssHeight != -1 {
				if borderBox {
					b.SetHeight(cssHeight - boxInt(padding.Height()) - boxInt(border.Height()))
				} else {
					b.SetHeight(cssHeight)
				}

			}

			//check if replaced
			re := b.GetReplacedElement()
			if re == nil {
				re = c.GetReplacedElementFactory().CreateReplacedElement(
					c, s, c.GetUac(), cssWidth, cssHeight)
				if re != nil {
					re = b.fitReplacedElement(c, re)
				}
			}
			if re != nil {
				b.SetContentWidth(re.GetIntrinsicWidth())
				b.SetHeight(re.GetIntrinsicHeight())
				b.SetReplacedElement(re)
			} else if cssWidth == -1 && pinnedContentWidth == -1 &&
				style.IsCanBeShrunkToFit() {
				b.SetNeedShrinkToFitCalculation(true)
			}

			if !b.IsReplaced() {
				s.ApplyCSSMinMaxWidth(c)
			}
		}

		b.SetDimensionsCalculated(true)
	}
}

func (b *BlockBox) calcClearance(c *LayoutContext) {
	if b.GetStyle().IsCleared() && !b.GetStyle().IsFloated() {
		c.Translate(0, -b.GetY())
		c.GetBlockFormattingContext().Clear(c, b.self)
		c.Translate(0, b.GetY())
		b.self.CalcCanvasLocation()
	}
}

func (b *BlockBox) calcExtraPageClearance(c *LayoutContext) {
	if c.IsPageBreaksAllowed() &&
		c.GetExtraSpaceTop() > 0 && (b.GetStyle().IsSpecifiedAsBlock() || b.GetStyle().IsListItem()) {
		first := c.GetRootLayer().GetFirstPage(c, b.self)
		if first != nil && first.GetTop()+c.GetExtraSpaceTop() > b.GetAbsY() {
			diff := first.GetTop() + c.GetExtraSpaceTop() - b.GetAbsY()
			b.SetY(b.GetY() + diff)
			c.Translate(0, diff)
			b.self.CalcCanvasLocation()
		}
	}
}

func (b *BlockBox) addBoxID(c *LayoutContext) {
	if !b.IsAnonymous() {
		name := c.GetNamespaceHandler().GetAnchorName(b.GetElement())
		if name != "" {
			c.AddBoxId(name, b.self)
		}
		id := c.GetNamespaceHandler().GetID(b.GetElement())
		if id != "" {
			c.AddBoxId(id, b.self)
		}
	}
}

func (b *BlockBox) Layout(c *LayoutContext) {
	b.selfBlock().LayoutWithContentStart(c, 0)
}

func (b *BlockBox) LayoutWithContentStart(c *LayoutContext, contentStart int) {
	s := b.selfBlock()
	style := b.GetStyle()

	pushedLayer := false
	if b.IsRoot() || style.RequiresLayer() {
		pushedLayer = true
		if b.GetLayer() == nil {
			c.PushLayerBox(s)
		} else {
			c.PushLayerLayer(b.GetLayer())
		}
	}

	if style.IsFixedBackground() {
		c.GetRootLayer().SetFixedBackground(true)
	}

	b.calcClearance(c)

	if b.IsRoot() || b.GetStyle().EstablishesBFC() || s.IsMarginAreaRoot() {
		bfc := NewBlockFormattingContext(s, c)
		c.PushBFC(bfc)
	}

	b.addBoxID(c)

	if c.IsPrint() && b.GetStyle().IsIdent(CSSNameFsPageSequence, IdentValueStart) {
		c.GetRootLayer().AddPageSequence(s)
	}

	s.CalcDimensions(c)
	b.calcShrinkToFitWidthIfNeeded(c)
	b.collapseMargins(c)

	b.calcExtraPageClearance(c)

	if c.IsPrint() {
		if len(c.GetRootLayer().GetPages()) == 0 {
			c.GetRootLayer().AddPage(c)
		}

		firstPage := c.GetRootLayer().GetFirstPage(c, s)
		if firstPage != nil && firstPage.GetTop() == b.GetAbsY()-s.GetPageClearance() {
			b.ResetTopMargin(c)
		}
	}

	border := s.GetBorder(c)
	margin := s.GetMargin(c)
	padding := s.GetPadding(c)

	// save height in case fixed height
	originalHeight := s.GetHeight()

	if !b.IsReplaced() {
		b.SetHeight(0)
	}

	didSetMarkerData := false
	if b.GetStyle().IsListItem() {
		markerData := b.initMarkerData(c)
		c.SetCurrentMarkerData(markerData)
		didSetMarkerData = true
	}

	// do children's layout
	tx := boxInt(margin.Left()) + boxInt(border.Left()) + boxInt(padding.Left())
	ty := boxInt(margin.Top()) + boxInt(border.Top()) + boxInt(padding.Top())
	b.SetTx(tx)
	b.SetTy(ty)
	c.Translate(b.GetTx(), b.GetTy())
	if !b.IsReplaced() {
		s.LayoutChildren(c, contentStart)
	} else {
		b.SetState(BoxStateDone)
	}
	c.Translate(-b.GetTx(), -b.GetTy())

	b.SetChildrenHeight(s.GetHeight())

	if !b.IsReplaced() {
		if !s.IsAutoHeight() {
			delta := originalHeight - s.GetHeight()
			if delta > 0 || s.IsAllowHeightToShrink() {
				b.SetHeight(originalHeight)
			}
		}

		b.applyCSSMinMaxHeight(c)
	}

	if b.IsRoot() || b.GetStyle().EstablishesBFC() {
		if b.GetStyle().IsAutoHeight() {
			delta :=
				c.GetBlockFormattingContext().GetFloatManager().GetClearDelta(
					c, b.GetTy()+s.GetHeight())
			if delta > 0 {
				b.SetHeight(s.GetHeight() + delta)
				b.SetChildrenHeight(b.GetChildrenHeight() + delta)
			}
		}
	}

	if didSetMarkerData {
		c.SetCurrentMarkerData(nil)
	}

	s.CalcLayoutHeight(c, border, margin, padding)

	if b.IsRoot() || b.GetStyle().EstablishesBFC() {
		c.PopBFC()
	}

	if pushedLayer {
		c.PopLayer()
	}
}

func (b *BlockBox) IsAllowHeightToShrink() bool {
	return true
}

func (b *BlockBox) GetPageClearance() int {
	return 0
}

func (b *BlockBox) CalcLayoutHeight(c *LayoutContext, border *BorderPropertySet, margin RectPropertySetI, padding RectPropertySetI) {
	b.SetHeight(b.self.GetHeight() + boxInt(margin.Top()) + boxInt(border.Top()) + boxInt(padding.Top()) +
		boxInt(padding.Bottom()) + boxInt(border.Bottom()) + boxInt(margin.Bottom()))
	b.SetChildrenHeight(b.GetChildrenHeight() + boxInt(margin.Top()) + boxInt(border.Top()) + boxInt(padding.Top()) +
		boxInt(padding.Bottom()) + boxInt(border.Bottom()) + boxInt(margin.Bottom()))
}

func (b *BlockBox) calcShrinkToFitWidthIfNeeded(c *LayoutContext) {
	if b.isNeedShrinkToFitCalculation() {
		b.SetContentWidth(b.calcShrinkToFitWidth(c) - b.GetLeftMBP() - b.GetRightMBP())
		b.selfBlock().ApplyCSSMinMaxWidth(c)
		b.SetNeedShrinkToFitCalculation(false)
	}
}

func (b *BlockBox) ApplyCSSMinMaxWidth(c CssContext) {
	if !b.GetStyle().IsMaxWidthNone() {
		cssMaxWidth := b.getCSSMaxWidth(c)
		if b.self.GetContentWidth() > cssMaxWidth {
			b.SetContentWidth(cssMaxWidth)
		}
	}
	cssMinWidth := b.getCSSMinWidth(c)
	if cssMinWidth > 0 && b.self.GetContentWidth() < cssMinWidth {
		b.SetContentWidth(cssMinWidth)
	}
}

func (b *BlockBox) applyCSSMinMaxHeight(c CssContext) {
	if !b.GetStyle().IsMaxHeightNone() {
		cssMaxHeight := b.getCSSMaxHeight(c)
		if b.self.GetHeight() > cssMaxHeight {
			b.SetHeight(cssMaxHeight)
		}
	}
	cssMinHeight := b.getCSSMinHeight(c)
	if cssMinHeight > 0 && b.self.GetHeight() < cssMinHeight {
		b.SetHeight(cssMinHeight)
	}
}

func (b *BlockBox) EnsureChildren(c *LayoutContext) {
	if b.GetChildrenContentType() == BlockBoxContentTypeUnknown {
		BoxBuilderCreateChildren(c, b.selfBlock())
	}
}

// BuildContainerAndAnalyzePageBreaks accepts a nil container.
func (b *BlockBox) BuildContainerAndAnalyzePageBreaks(c *LayoutContext, container *ContentLimitContainer) *ContentLimitContainer {
	contentLimitContainer := NewContentLimitContainer(container, c, b.GetAbsY())

	if container != nil {
		container.UpdateTop(c, b.GetAbsY())
		container.UpdateBottom(c, b.GetAbsY()+b.self.GetHeight())
	}

	for _, child := range b.GetChildren() {
		child.AnalyzePageBreaks(c, contentLimitContainer)
	}
	return contentLimitContainer
}

func (b *BlockBox) LayoutChildren(c *LayoutContext, contentStart int) {
	s := b.selfBlock()
	b.SetState(BoxStateChildrenFlux)
	s.EnsureChildren(c)

	if b.GetFirstLetterStyle() != nil {
		c.GetFirstLettersTracker().AddStyle(b.GetFirstLetterStyle())
	}
	if b.GetFirstLineStyle() != nil {
		c.GetFirstLinesTracker().AddStyle(b.GetFirstLineStyle())
	}

	switch b.GetChildrenContentType() {
	case BlockBoxContentTypeInline:
		s.LayoutInlineChildren(c, contentStart, s.CalcInitialBreakAtLine(c), true)
	case BlockBoxContentTypeBlock:
		BlockBoxingLayoutContent(c, s, contentStart)
	}

	if b.GetFirstLetterStyle() != nil {
		c.GetFirstLettersTracker().RemoveLast()
	}
	if b.GetFirstLineStyle() != nil {
		c.GetFirstLinesTracker().RemoveLast()
	}

	b.SetState(BoxStateDone)
}

func (b *BlockBox) LayoutInlineChildren(c *LayoutContext, contentStart int, breakAtLine int, tryAgain bool) {
	InlineBoxingLayoutContent(c, b.selfBlock(), contentStart, breakAtLine)

	if c.IsPrint() && c.IsPageBreaksAllowed() && b.GetChildCount() > 1 {
		b.satisfyWidowsAndOrphans(c, contentStart, tryAgain)
	}

	if tryAgain && b.GetStyle().IsTextJustify() {
		b.justifyText()
	}
}

func (b *BlockBox) justifyText() {
	for _, box := range b.GetChildren() {
		line := box.(*LineBox)
		line.Justify()
	}
}

func (b *BlockBox) satisfyWidowsAndOrphans(c *LayoutContext, contentStart int, tryAgain bool) {
	firstLineBox := b.GetChild(0).(*LineBox)
	firstPage := c.GetRootLayer().GetFirstPage(c, firstLineBox)

	if firstPage == nil {
		return
	}

	noContentLBs := 0
	i := 0
	cCount := b.GetChildCount()
	for i < cCount {
		lB := b.GetChild(i).(*LineBox)
		if lB.GetAbsY() >= firstPage.GetBottom() {
			break
		}
		if !lB.IsContainsContent() {
			noContentLBs++
		}
		i++
	}

	if i != cCount {
		orphans := boxInt(b.GetStyle().AsFloat(CSSNameOrphans))
		if i-noContentLBs < orphans {
			b.SetNeedPageClear(true)
		} else {
			lastLineBox := b.GetChild(cCount - 1).(*LineBox)
			pages := c.GetRootLayer().GetPages()
			lastPage := pages[firstPage.GetPageNo()+1]
			for lastPage.GetPageNo() != len(pages)-1 &&
				lastPage.GetBottom() < lastLineBox.GetAbsY() {
				lastPage = pages[lastPage.GetPageNo()+1]
			}

			noContentLBs = 0
			i = cCount - 1
			for i >= 0 && b.GetChild(i).GetAbsY() >= lastPage.GetTop() {
				lB := b.GetChild(i).(*LineBox)
				if lB.GetAbsY() < lastPage.GetTop() {
					break
				}
				if !lB.IsContainsContent() {
					noContentLBs++
				}
				i--
			}

			widows := boxInt(b.GetStyle().AsFloat(CSSNameWidows))
			if cCount-1-i-noContentLBs < widows {
				if cCount-1-widows < orphans {
					b.SetNeedPageClear(true)
				} else if tryAgain {
					breakAtLine := cCount - 1 - widows

					b.self.ResetChildren(c)
					b.RemoveAllChildren()

					b.selfBlock().LayoutInlineChildren(c, contentStart, breakAtLine, false)
				}
			}
		}
	}
}

func (b *BlockBox) GetChildrenContentType() BlockBoxContentType {
	return b.childrenContentType
}

func (b *BlockBox) SetChildrenContentType(contentType BlockBoxContentType) {
	b.childrenContentType = contentType
}

func (b *BlockBox) GetInlineContent() []Styleable {
	return b.inlineContent
}

func (b *BlockBox) SetInlineContent(inlineContent []Styleable) {
	b.inlineContent = inlineContent

	for _, child := range inlineContent {
		if childBox, ok := child.(BoxI); ok {
			childBox.SetContainingBlock(b.self)
		}
	}
}

func (b *BlockBox) IsSkipWhenCollapsingMargins() bool {
	return false
}

func (b *BlockBox) IsMayCollapseMarginsWithChildren() bool {
	return !b.IsRoot() && b.GetStyle().IsMayCollapseMarginsWithChildren()
}

// This will require a rethink if we ever truly layout incrementally
// Should only ever collapse top margin and pick up collapsable
// bottom margins by looking back up the tree.
func (b *BlockBox) collapseMargins(c *LayoutContext) {
	if !b.IsTopMarginCalculated() || !b.IsBottomMarginCalculated() {
		b.recalculateMargin(c)
		margin := b.GetMargin(c)

		if !b.IsTopMarginCalculated() && !b.IsBottomMarginCalculated() && b.isVerticalMarginsAdjoin(c) {
			collapsedMargin := b.pendingCollapseCalculation
			if collapsedMargin == nil {
				collapsedMargin = &blockBoxMarginCollapseResult{}
			}
			b.collapseEmptySubtreeMargins(c, collapsedMargin)
			b.setCollapsedBottomMargin(c, margin, collapsedMargin)
		} else {
			if !b.IsTopMarginCalculated() {
				collapsedMargin := b.pendingCollapseCalculation
				if collapsedMargin == nil {
					collapsedMargin = &blockBoxMarginCollapseResult{}
				}

				b.collapseTopMargin(c, true, collapsedMargin)
				if boxInt(margin.Top()) != collapsedMargin.getMargin() {
					b.SetMarginTop(c, collapsedMargin.getMargin())
				}
			}

			if !b.IsBottomMarginCalculated() {
				collapsedMargin := &blockBoxMarginCollapseResult{}
				b.collapseBottomMargin(c, true, collapsedMargin)

				b.setCollapsedBottomMargin(c, margin, collapsedMargin)
			}
		}
	}
}

func (b *BlockBox) setCollapsedBottomMargin(c *LayoutContext, margin RectPropertySetI, collapsedMargin *blockBoxMarginCollapseResult) {
	var next BlockBoxI
	if !b.IsInline() {
		next = b.getNextCollapsableSibling(collapsedMargin)
	}
	nextIsAnonymous := false
	if next != nil {
		_, nextIsAnonymous = next.(*AnonymousBlockBox)
	}
	if !(next == nil || nextIsAnonymous) &&
		collapsedMargin.hasMargin() {
		next.AsBlockBox().pendingCollapseCalculation = collapsedMargin
		b.SetMarginBottom(c, 0)
	} else if boxInt(margin.Bottom()) != collapsedMargin.getMargin() {
		b.SetMarginBottom(c, collapsedMargin.getMargin())
	}
}

func (b *BlockBox) getNextCollapsableSibling(collapsedMargin *blockBoxMarginCollapseResult) BlockBoxI {
	next := blockBoxFromBox(b.GetNextSibling())
	for next != nil {
		if anonymousBlockBox, ok := next.(*AnonymousBlockBox); ok {
			anonymousBlockBox.ProvideSiblingMarginToFloats(collapsedMargin.getMargin())
		}
		if !next.IsSkipWhenCollapsingMargins() {
			break
		} else {
			next = blockBoxFromBox(next.GetNextSibling())
		}
	}
	return next
}

func (b *BlockBox) collapseTopMargin(c *LayoutContext, calculationRoot bool, result *blockBoxMarginCollapseResult) {
	s := b.selfBlock()
	if !b.IsTopMarginCalculated() {
		if !s.IsSkipWhenCollapsingMargins() {
			s.CalcDimensions(c)
			if c.IsPrint() && b.GetStyle().IsDynamicAutoWidthApplicable() {
				// Force recalculation once box is positioned
				b.SetDimensionsCalculated(false)
			}
			margin := b.GetMargin(c)
			result.update(boxInt(margin.Top()))

			if !calculationRoot && boxInt(margin.Top()) != 0 {
				b.SetMarginTop(c, 0)
			}

			if s.IsMayCollapseMarginsWithChildren() && b.isNoTopPaddingOrBorder(c) {
				s.EnsureChildren(c)
				if b.GetChildrenContentType() == BlockBoxContentTypeBlock {
					for _, box := range b.GetChildren() {
						child := box.(BlockBoxI)
						child.AsBlockBox().collapseTopMargin(c, false, result)

						if child.IsSkipWhenCollapsingMargins() {
							continue
						}

						break
					}
				}
			}
		}

		b.SetTopMarginCalculated(true)
	}
}

func (b *BlockBox) collapseBottomMargin(c *LayoutContext, calculationRoot bool, result *blockBoxMarginCollapseResult) {
	s := b.selfBlock()
	if !b.IsBottomMarginCalculated() {
		if !s.IsSkipWhenCollapsingMargins() {
			s.CalcDimensions(c)
			if c.IsPrint() && b.GetStyle().IsDynamicAutoWidthApplicable() {
				// Force recalculation once box is positioned
				b.SetDimensionsCalculated(false)
			}
			margin := b.GetMargin(c)
			result.update(boxInt(margin.Bottom()))

			if !calculationRoot && boxInt(margin.Bottom()) != 0 {
				b.SetMarginBottom(c, 0)
			}

			if s.IsMayCollapseMarginsWithChildren() &&
				!b.GetStyle().IsTable() && b.isNoBottomPaddingOrBorder(c) {
				s.EnsureChildren(c)
				if b.GetChildrenContentType() == BlockBoxContentTypeBlock {
					for i := b.GetChildCount() - 1; i >= 0; i-- {
						child := b.GetChild(i).(BlockBoxI)

						if child.IsSkipWhenCollapsingMargins() {
							continue
						}

						child.AsBlockBox().collapseBottomMargin(c, false, result)

						break
					}
				}
			}
		}

		b.SetBottomMarginCalculated(true)
	}
}

func (b *BlockBox) isNoTopPaddingOrBorder(c *LayoutContext) bool {
	padding := b.self.GetPadding(c)
	border := b.self.GetBorder(c)

	return boxInt(padding.Top()) == 0 && boxInt(border.Top()) == 0
}

func (b *BlockBox) isNoBottomPaddingOrBorder(c *LayoutContext) bool {
	padding := b.self.GetPadding(c)
	border := b.self.GetBorder(c)

	return boxInt(padding.Bottom()) == 0 && boxInt(border.Bottom()) == 0
}

func (b *BlockBox) collapseEmptySubtreeMargins(c *LayoutContext, result *blockBoxMarginCollapseResult) {
	margin := b.GetMargin(c)
	result.update(boxInt(margin.Top()))
	result.update(boxInt(margin.Bottom()))

	b.SetMarginTop(c, 0)
	b.SetTopMarginCalculated(true)
	b.SetMarginBottom(c, 0)
	b.SetBottomMarginCalculated(true)

	b.selfBlock().EnsureChildren(c)
	if b.GetChildrenContentType() == BlockBoxContentTypeBlock {
		for _, box := range b.GetChildren() {
			child := box.(BlockBoxI)
			child.AsBlockBox().collapseEmptySubtreeMargins(c, result)
		}
	}
}

func (b *BlockBox) isVerticalMarginsAdjoin(c *LayoutContext) bool {
	s := b.selfBlock()
	style := b.GetStyle()

	borderWidth := style.GetBorder(c)
	padding := s.GetPadding(c)

	bordersOrPadding :=
		boxInt(borderWidth.Top()) != 0 || boxInt(borderWidth.Bottom()) != 0 ||
			boxInt(padding.Top()) != 0 || boxInt(padding.Bottom()) != 0

	if bordersOrPadding {
		return false
	}

	s.EnsureChildren(c)
	if b.GetChildrenContentType() == BlockBoxContentTypeInline {
		return false
	} else if b.GetChildrenContentType() == BlockBoxContentTypeBlock {
		for _, box := range b.GetChildren() {
			child := box.(BlockBoxI)
			if child.IsSkipWhenCollapsingMargins() || !child.AsBlockBox().isVerticalMarginsAdjoin(c) {
				return false
			}
		}
	}

	return style.AsFloat(CSSNameMinHeight) == 0 &&
		(s.IsAutoHeight() || style.AsFloat(CSSNameHeight) == 0)
}

func (b *BlockBox) IsTopMarginCalculated() bool {
	return b.topMarginCalculated
}

func (b *BlockBox) SetTopMarginCalculated(topMarginCalculated bool) {
	b.topMarginCalculated = topMarginCalculated
}

func (b *BlockBox) IsBottomMarginCalculated() bool {
	return b.bottomMarginCalculated
}

func (b *BlockBox) SetBottomMarginCalculated(bottomMarginCalculated bool) {
	b.bottomMarginCalculated = bottomMarginCalculated
}

func (b *BlockBox) GetCSSWidth(c CssContext) int {
	return b.selfBlock().GetCSSWidthWithShrinkingToFit(c, false)
}

func (b *BlockBox) GetCSSWidthWithShrinkingToFit(c CssContext, shrinkingToFit bool) int {
	if !b.IsAnonymous() {
		if !b.GetStyle().IsAutoWidth() {
			if shrinkingToFit && !b.GetStyle().IsAbsoluteWidth() {
				return -1
			} else {
				result := boxInt(b.GetStyle().GetFloatPropertyProportionalWidth(
					CSSNameWidth, float32(b.GetContainingBlock().GetContentWidth()), c))
				if result >= 0 {
					return result
				}
				return -1
			}
		}
	}

	return -1
}

func (b *BlockBox) GetCSSFitToWidth(c CssContext) int {
	if !b.IsAnonymous() {
		if !b.GetStyle().IsIdent(CSSNameFsFitImagesToWidth, IdentValueAuto) {
			result := boxInt(b.GetStyle().GetFloatPropertyProportionalWidth(
				CSSNameFsFitImagesToWidth, float32(b.GetContainingBlock().GetContentWidth()), c))
			if result >= 0 {
				return result
			}
			return -1
		}
	}

	return -1
}

func (b *BlockBox) GetCSSHeight(c CssContext) int {
	if !b.IsAnonymous() {
		if !b.selfBlock().IsAutoHeight() {
			if b.GetStyle().HasAbsoluteUnit(CSSNameHeight) {
				return boxInt(b.GetStyle().GetFloatPropertyProportionalHeight(CSSNameHeight, 0, c))
			} else {
				return boxInt(b.GetStyle().GetFloatPropertyProportionalHeight(
					CSSNameHeight,
					float32(b.GetContainingBlock().(BlockBoxI).GetCSSHeight(c)),
					c))
			}
		}
	}

	return -1
}

func (b *BlockBox) IsAutoHeight() bool {
	if b.GetStyle().IsAutoHeight() {
		return true
	} else if b.GetStyle().HasAbsoluteUnit(CSSNameHeight) {
		return false
	} else {
		// We have a percentage height, defer to our block parent (if applicable)
		cb := b.GetContainingBlock()
		blockBox, isBlockBox := cb.(BlockBoxI)
		if cb.IsStyled() && isBlockBox {
			return blockBox.IsAutoHeight()
		} else {
			return !isBlockBox || !cb.IsInitialContainingBlock()
		}
	}
}

func (b *BlockBox) getCSSMinWidth(c CssContext) int {
	result := b.GetStyle().GetMinWidth(c, b.self.GetContainingBlockWidth())
	if result > 0 && b.GetStyle().IsBorderBox() {
		// min-width in border-box mode refers to the total outer width.
		// Subtract paddingBorderWidth so calcMinMaxWidth stores the correct
		// content-width floor — preventing AutoTableLayout from over-allocating
		// the column before applyCSSMinMaxWidth runs.
		padding := b.self.GetPadding(c)
		border := b.self.GetBorder(c)
		paddingBorderWidth := boxInt(padding.Width()) + boxInt(border.Width())
		result = max(0, result-paddingBorderWidth)
	}
	return result
}

func (b *BlockBox) getCSSMaxWidth(c CssContext) int {
	result := b.GetStyle().GetMaxWidth(c, b.self.GetContainingBlockWidth())
	if b.GetStyle().IsBorderBox() {
		// max-width in border-box mode refers to the total outer width.
		// Subtract paddingBorderWidth to convert to a content-width ceiling
		// so calcMinMaxWidth and applyCSSMinMaxWidth both operate on the
		// same axis as contentWidth.
		padding := b.self.GetPadding(c)
		border := b.self.GetBorder(c)
		paddingBorderWidth := boxInt(padding.Width()) + boxInt(border.Width())
		result = max(0, result-paddingBorderWidth)
	}
	return result
}

func (b *BlockBox) getCSSMinHeight(c CssContext) int {
	return b.GetStyle().GetMinHeight(c, b.getContainingBlockCSSHeight(c))
}

func (b *BlockBox) getCSSMaxHeight(c CssContext) int {
	return b.GetStyle().GetMaxHeight(c, b.getContainingBlockCSSHeight(c))
}

// Use only when the height of the containing block is required for
// resolving percentage values.  Does not represent the actual (resolved) height
// of the containing block.
func (b *BlockBox) getContainingBlockCSSHeight(c CssContext) int {
	if !b.GetContainingBlock().IsStyled() ||
		b.GetContainingBlock().GetStyle().IsAutoHeight() {
		return 0
	} else {
		if b.GetContainingBlock().GetStyle().HasAbsoluteUnit(CSSNameHeight) {
			return boxInt(b.GetContainingBlock().GetStyle().GetFloatPropertyProportionalTo(
				CSSNameHeight, 0, c))
		} else {
			return 0
		}
	}
}

func (b *BlockBox) calcShrinkToFitWidth(c *LayoutContext) int {
	s := b.selfBlock()
	s.CalcMinMaxWidth(c)

	return min(max(b.GetMinWidth(), s.GetAvailableWidth(c)), b.GetMaxWidth())
}

func (b *BlockBox) GetAvailableWidth(c *LayoutContext) int {
	if !b.GetStyle().IsAbsolute() {
		return b.self.GetContainingBlockWidth()
	} else {
		left := 0
		right := 0
		if !b.GetStyle().IsIdent(CSSNameLeft, IdentValueAuto) {
			left =
				boxInt(b.GetStyle().GetFloatPropertyProportionalTo(CSSNameLeft,
					float32(b.GetContainingBlock().GetContentWidth()), c))
		}

		if !b.GetStyle().IsIdent(CSSNameRight, IdentValueAuto) {
			right =
				boxInt(b.GetStyle().GetFloatPropertyProportionalTo(CSSNameRight,
					float32(b.GetContainingBlock().GetContentWidth()), c))
		}

		return b.GetContainingBlock().GetPaddingWidth(c) - left - right
	}
}

func (b *BlockBox) IsFixedWidthAdvisoryOnly() bool {
	return false
}

func (b *BlockBox) recalculateMargin(c *LayoutContext) {
	if b.IsTopMarginCalculated() && b.IsBottomMarginCalculated() {
		return
	}

	// Check if we're a potential candidate upfront to avoid expensive
	// getStyleMargin(c, false) call
	topMargin := b.GetStyle().ValueByName(CSSNameMarginTop)
	_, topIsLength := topMargin.(*LengthValue)
	resetTop := topIsLength && !topMargin.HasAbsoluteUnit()

	bottomMargin := b.GetStyle().ValueByName(CSSNameMarginBottom)
	_, bottomIsLength := bottomMargin.(*LengthValue)
	resetBottom := bottomIsLength && !bottomMargin.HasAbsoluteUnit()

	if !resetTop && !resetBottom {
		return
	}

	styleMargin := b.GetStyleMarginNoCache(c)
	workingMargin := b.GetMargin(c)

	// A shrink-to-fit calculation may have set incorrect values for
	// percentage margins (as the containing block width
	// hasn't been calculated yet).  Reset top and bottom margins
	// in this case.
	if !b.IsTopMarginCalculated() &&
		styleMargin.Top() != workingMargin.Top() {
		b.SetMarginTop(c, boxInt(styleMargin.Top()))
	}

	if !b.IsBottomMarginCalculated() &&
		styleMargin.Bottom() != workingMargin.Bottom() {
		b.SetMarginBottom(c, boxInt(styleMargin.Bottom()))
	}
}

func (b *BlockBox) CalcMinMaxWidth(c *LayoutContext) {
	s := b.selfBlock()
	if !b.IsMinMaxCalculated() {
		margin := s.GetMargin(c)
		border := s.GetBorder(c)
		padding := s.GetPadding(c)

		width := s.GetCSSWidthWithShrinkingToFit(c, true)

		if width == -1 {
			if b.IsReplaced() {
				width = b.GetReplacedElement().GetIntrinsicWidth()
			} else {
				height := s.GetCSSHeight(c)
				re := c.GetReplacedElementFactory().CreateReplacedElement(
					c, s, c.GetUac(), width, height)
				if re != nil {
					re = b.fitReplacedElement(c, re)
					b.SetReplacedElement(re)
					width = b.GetReplacedElement().GetIntrinsicWidth()
				}
			}
		}

		if b.IsReplaced() || width != -1 && !s.IsFixedWidthAdvisoryOnly() {
			b.minWidth =
				boxInt(margin.Left()) + boxInt(border.Left()) + boxInt(padding.Left()) +
					width +
					boxInt(margin.Right()) + boxInt(border.Right()) + boxInt(padding.Right())
			b.maxWidth = b.minWidth
		} else {
			cw := -1
			if width != -1 {
				// Set a provisional content width on table cells so
				// percentage values resolve correctly (but save and reset
				// the existing value)
				cw = s.GetContentWidth()
				b.SetContentWidth(width)
			}

			b.minWidth =
				boxInt(margin.Left()) + boxInt(border.Left()) + boxInt(padding.Left()) +
					boxInt(margin.Right()) + boxInt(border.Right()) + boxInt(padding.Right())
			b.maxWidth = b.minWidth

			minimumMaxWidth := b.maxWidth
			if width != -1 {
				minimumMaxWidth += width
			}

			s.EnsureChildren(c)

			switch b.GetChildrenContentType() {
			case BlockBoxContentTypeBlock:
				b.calcMinMaxWidthBlockChildren(c)
			case BlockBoxContentTypeInline:
				b.calcMinMaxWidthInlineChildren(c)
			}

			if minimumMaxWidth > b.maxWidth {
				b.maxWidth = minimumMaxWidth
			}

			if cw != -1 {
				b.SetContentWidth(cw)
			}
		}

		if !b.IsReplaced() {
			b.calcMinMaxCSSMinMaxWidth(c, margin, border, padding)
		}

		b.SetMinMaxCalculated(true)
	}
}

func (b *BlockBox) fitReplacedElement(c *LayoutContext, re ReplacedElement) ReplacedElement {
	maxImageWidth := b.GetCSSFitToWidth(c)
	if maxImageWidth > -1 && re.GetIntrinsicWidth() > maxImageWidth {
		oldWidth := float64(re.GetIntrinsicWidth())
		scale := float64(maxImageWidth) / oldWidth
		re = c.GetReplacedElementFactory().CreateReplacedElement(
			c, b.selfBlock(), c.GetUac(), maxImageWidth, blockBoxDoubleToInt(math.RoundToEven(scale*float64(re.GetIntrinsicHeight()))))
	}
	return re
}

// blockBoxDoubleToInt is Java's (int) cast of a double: truncation toward
// zero, NaN to 0, and saturation at the int range.
func blockBoxDoubleToInt(d float64) int {
	if d != d {
		return 0
	}
	if d >= math.MaxInt32 {
		return math.MaxInt32
	}
	if d <= math.MinInt32 {
		return math.MinInt32
	}
	return int(d)
}

func (b *BlockBox) calcMinMaxCSSMinMaxWidth(c *LayoutContext, margin RectPropertySetI, border *BorderPropertySet, padding RectPropertySetI) {
	cssMinWidth := b.getCSSMinWidth(c)
	if cssMinWidth > 0 {
		cssMinWidth +=
			boxInt(margin.Left()) + boxInt(border.Left()) + boxInt(padding.Left()) +
				boxInt(margin.Right()) + boxInt(border.Right()) + boxInt(padding.Right())
		if b.minWidth < cssMinWidth {
			b.minWidth = cssMinWidth
		}
	}
	if !b.GetStyle().IsMaxWidthNone() {
		cssMaxWidth := b.getCSSMaxWidth(c)
		cssMaxWidth +=
			boxInt(margin.Left()) + boxInt(border.Left()) + boxInt(padding.Left()) +
				boxInt(margin.Right()) + boxInt(border.Right()) + boxInt(padding.Right())
		if b.maxWidth > cssMaxWidth {
			b.maxWidth = max(cssMaxWidth, b.minWidth)
		}
	}
}

func (b *BlockBox) calcMinMaxWidthBlockChildren(c *LayoutContext) {
	childMinWidth := 0
	childMaxWidth := 0

	for _, box := range b.GetChildren() {
		child := box.(BlockBoxI)
		child.CalcMinMaxWidth(c)
		if child.GetMinWidth() > childMinWidth {
			childMinWidth = child.GetMinWidth()
		}
		if child.GetMaxWidth() > childMaxWidth {
			childMaxWidth = child.GetMaxWidth()
		}
	}

	b.minWidth += childMinWidth
	b.maxWidth += childMaxWidth
}

func (b *BlockBox) calcMinMaxWidthInlineChildren(c *LayoutContext) {
	textIndent := boxInt(b.GetStyle().GetFloatPropertyProportionalWidth(
		CSSNameTextIndent, float32(b.self.GetContentWidth()), c))

	if b.GetStyle().IsListItem() && b.GetStyle().IsListMarkerInside() {
		markerData := b.initMarkerData(c)
		textIndent += markerData.GetLayoutWidth()
	}

	childMinWidth := 0
	childMaxWidth := 0
	lineWidth := 0

	var trimmableIB *InlineBox

	for _, child := range b.inlineContent {
		if child.GetStyle().IsAbsolute() || child.GetStyle().IsFixed() || child.GetStyle().IsRunning() {
			continue
		}

		if child.GetStyle().IsFloated() || child.GetStyle().IsInlineBlock() ||
			child.GetStyle().IsInlineTable() {
			if child.GetStyle().IsFloated() && child.GetStyle().IsCleared() {
				if trimmableIB != nil {
					lineWidth -= trimmableIB.GetTrailingSpaceWidth(c)
				}
				if lineWidth > childMaxWidth {
					childMaxWidth = lineWidth
				}
				lineWidth = 0
			}
			trimmableIB = nil
			block := child.(BlockBoxI)
			block.CalcMinMaxWidth(c)
			lineWidth += block.GetMaxWidth()
			if block.GetMinWidth() > childMinWidth {
				childMinWidth = block.GetMinWidth()
			}
		} else { /* child.getStyle().isInline() */
			iB := child.(*InlineBox)
			whitespace := iB.GetStyle().GetWhitespace()
			iB.CalcMinMaxWidth(c, b.self.GetContentWidth(), lineWidth == 0)

			if whitespace == IdentValueNowrap {
				lineWidth += textIndent + iB.GetMaxWidth()
				if iB.GetMinWidth() > childMinWidth {
					childMinWidth = iB.GetMinWidth()
				}
				trimmableIB = iB
			} else if whitespace == IdentValuePre {
				if trimmableIB != nil {
					lineWidth -= trimmableIB.GetTrailingSpaceWidth(c)
				}
				trimmableIB = nil
				if lineWidth > childMaxWidth {
					childMaxWidth = lineWidth
				}
				lineWidth = textIndent + iB.GetFirstLineWidth()
				if lineWidth > childMinWidth {
					childMinWidth = lineWidth
				}
				lineWidth = iB.GetMaxWidth()
				if lineWidth > childMinWidth {
					childMinWidth = lineWidth
				}
				if childMinWidth > childMaxWidth {
					childMaxWidth = childMinWidth
				}
				lineWidth = 0
			} else if whitespace == IdentValuePreWrap || whitespace == IdentValuePreLine {
				lineWidth += textIndent + iB.GetFirstLineWidth()
				if trimmableIB != nil {
					lineWidth -= trimmableIB.GetTrailingSpaceWidth(c)
				}
				if lineWidth > childMaxWidth {
					childMaxWidth = lineWidth
				}

				if iB.GetMaxWidth() > childMaxWidth {
					childMaxWidth = iB.GetMaxWidth()
				}
				if iB.GetMinWidth() > childMinWidth {
					childMinWidth = iB.GetMinWidth()
				}
				if whitespace == IdentValuePreLine {
					trimmableIB = iB
				} else {
					trimmableIB = nil
				}
				lineWidth = 0
			} else /* if (whitespace == IdentValue.NORMAL) */ {
				lineWidth += textIndent + iB.GetMaxWidth()
				if iB.GetMinWidth() > childMinWidth {
					childMinWidth = textIndent + iB.GetMinWidth()
				}
				trimmableIB = iB
			}

			if textIndent > 0 {
				textIndent = 0
			}
		}
	}

	if trimmableIB != nil {
		lineWidth -= trimmableIB.GetTrailingSpaceWidth(c)
	}
	if lineWidth > childMaxWidth {
		childMaxWidth = lineWidth
	}

	b.minWidth += childMinWidth
	b.maxWidth += childMaxWidth
}

func (b *BlockBox) GetMaxWidth() int {
	return b.maxWidth
}

func (b *BlockBox) SetMaxWidth(maxWidth int) {
	b.maxWidth = maxWidth
}

func (b *BlockBox) GetMinWidth() int {
	return b.minWidth
}

func (b *BlockBox) SetMinWidth(minWidth int) {
	b.minWidth = minWidth
}

func (b *BlockBox) StyleText(c *LayoutContext) {
	b.selfBlock().StyleTextWithStyle(c, b.GetStyle())
}

// FIXME Should be expanded into generic restyle facility
func (b *BlockBox) StyleTextWithStyle(c *LayoutContext, style CalculatedStyleI) {
	if b.GetChildrenContentType() == BlockBoxContentTypeInline {
		styles := []CalculatedStyleI{style}
		for _, child := range b.inlineContent {
			if iB, ok := child.(*InlineBox); ok {

				if iB.IsStartsHere() {
					var cs *CascadedStyle
					if iB.GetElement() != nil {
						if iB.GetPseudoElementOrClass() == "" {
							cs = c.GetCss().GetCascadedStyle(iB.GetElement(), false)
						} else {
							cs = c.GetCss().GetPseudoElementStyle(
								iB.GetElement(), iB.GetPseudoElementOrClass())
						}
						styles = append(styles, styles[len(styles)-1].DeriveStyle(cs))
					} else {
						styles = append(styles, style.CreateAnonymousStyle(IdentValueInline))
					}
				}

				iB.SetStyle(styles[len(styles)-1])
				iB.ApplyTextTransform()

				if iB.IsEndsHere() {
					styles = styles[:len(styles)-1]
				}
			}
		}
	}
}

func (b *BlockBox) CalcChildPaintingInfo(c CssContext, result *PaintingInfo, useCache bool) {
	if b.GetPersistentBFC() != nil {
		b.GetPersistentBFC().GetFloatManager().PerformFloatOperation(
			FloatManagerFloatOperationFunc(func(floater BoxI) {
				info := floater.CalcPaintingInfo(c, useCache)
				b.MoveIfGreater(
					result.GetOuterMarginCorner(),
					info.GetOuterMarginCorner())
			}))
	}
	b.Box.CalcChildPaintingInfo(c, result, useCache)
}

// GetFirstLetterStyle may return nil.
func (b *BlockBox) GetFirstLetterStyle() *CascadedStyle {
	return b.firstLetterStyle
}

func (b *BlockBox) SetFirstLetterStyle(firstLetterStyle *CascadedStyle) {
	b.firstLetterStyle = firstLetterStyle
}

// GetFirstLineStyle may return nil.
func (b *BlockBox) GetFirstLineStyle() *CascadedStyle {
	return b.firstLineStyle
}

func (b *BlockBox) SetFirstLineStyle(firstLineStyle *CascadedStyle) {
	b.firstLineStyle = firstLineStyle
}

func (b *BlockBox) IsMinMaxCalculated() bool {
	return b.minMaxCalculated
}

func (b *BlockBox) SetMinMaxCalculated(minMaxCalculated bool) {
	b.minMaxCalculated = minMaxCalculated
}

func (b *BlockBox) SetDimensionsCalculated(dimensionsCalculated bool) {
	b.dimensionsCalculated = dimensionsCalculated
}

func (b *BlockBox) isDimensionsCalculated() bool {
	return b.dimensionsCalculated
}

func (b *BlockBox) SetNeedShrinkToFitCalculation(needShrinkToFitCalculation bool) {
	b.needShrinkToFitCalculation = needShrinkToFitCalculation
}

func (b *BlockBox) isNeedShrinkToFitCalculation() bool {
	return b.needShrinkToFitCalculation
}

func (b *BlockBox) InitStaticPos(c *LayoutContext, parent BlockBoxI, childOffset int) {
	b.SetX(0)
	b.SetY(childOffset)
}

func (b *BlockBox) CalcBaseline(c *LayoutContext) int {
	for i := 0; i < b.GetChildCount(); i++ {
		child := b.GetChild(i)
		if lineBox, ok := child.(*LineBox); ok {
			return child.GetAbsY() + lineBox.GetBaseline()
		} else {
			if tableRow, ok := child.(*TableRowBox); ok {
				return child.GetAbsY() + tableRow.GetBaseline()
			} else {
				result := child.(BlockBoxI).CalcBaseline(c)
				if result != BlockBoxNoBaseline {
					return result
				}
			}
		}
	}

	return BlockBoxNoBaseline
}

func (b *BlockBox) CalcInitialBreakAtLine(c *LayoutContext) int {
	bContext := c.GetBreakAtLineContext()
	if bContext != nil && bContext.GetBlock() == b.selfBlock() {
		return bContext.GetLine()
	}
	return 0
}

func (b *BlockBox) IsCurrentBreakAtLineContext(c *LayoutContext) bool {
	bContext := c.GetBreakAtLineContext()
	return bContext != nil && bContext.GetBlock() == b.selfBlock()
}

// CalcBreakAtLineContext may return nil.
func (b *BlockBox) CalcBreakAtLineContext(c *LayoutContext) *BreakAtLineContext {
	if !c.IsPrint() || !b.GetStyle().IsKeepWithInline() {
		return nil
	}

	breakLine := b.FindLastNthLineBox(boxInt(b.GetStyle().AsFloat(CSSNameWidows)))
	if breakLine != nil {
		linePage := c.GetRootLayer().GetLastPageWithCBox(c, breakLine)
		ourPage := c.GetRootLayer().GetLastPageWithCBox(c, b.self)
		if linePage != nil && ourPage != nil && linePage.GetPageNo()+1 == ourPage.GetPageNo() {
			breakBox := breakLine.GetParent().(BlockBoxI)
			return NewBreakAtLineContext(breakBox, breakBox.FindOffset(breakLine))
		}
	}

	return nil
}

func (b *BlockBox) CalcInlineBaseline(c CssContext) int {
	if b.IsReplaced() && b.GetReplacedElement().HasBaseline() {
		bounds := b.self.GetContentAreaEdge(b.GetAbsX(), b.GetAbsY(), c)
		return bounds.Y + b.GetReplacedElement().GetBaseline() - b.GetAbsY()
	} else {
		lastLine := b.findLastLineBox()
		if lastLine == nil {
			return b.self.GetHeight()
		} else {
			return lastLine.GetAbsY() + lastLine.GetBaseline() - b.GetAbsY()
		}
	}
}

func (b *BlockBox) FindOffset(box BoxI) int {
	count := b.GetChildCount()
	for i := 0; i < count; i++ {
		if b.GetChild(i) == box {
			return i
		}
	}
	return -1
}

// FindLastNthLineBox may return nil.
func (b *BlockBox) FindLastNthLineBox(count int) *LineBox {
	context := &blockBoxLastLineBoxContext{current: count}
	b.findLastLineBoxWithContext(context)
	return context.line
}

type blockBoxLastLineBoxContext struct {
	current int
	line    *LineBox
}

func (b *BlockBox) findLastLineBoxWithContext(context *blockBoxLastLineBoxContext) {
	contentType := b.GetChildrenContentType()
	count := b.GetChildCount()
	if count > 0 {
		if contentType == BlockBoxContentTypeInline {
			for i := count - 1; i >= 0; i-- {
				child := b.GetChild(i).(*LineBox)
				if child.GetHeight() > 0 {
					context.line = child
					context.current--
					if context.current == 0 {
						return
					}
				}
			}
		} else if contentType == BlockBoxContentTypeBlock {
			for i := count - 1; i >= 0; i-- {
				b.GetChild(i).(BlockBoxI).AsBlockBox().findLastLineBoxWithContext(context)
				if context.current == 0 {
					break
				}
			}
		}
	}
}

func (b *BlockBox) findLastLineBox() *LineBox {
	contentType := b.GetChildrenContentType()
	count := b.GetChildCount()
	if count > 0 {
		if contentType == BlockBoxContentTypeInline {
			for i := count - 1; i >= 0; i-- {
				result := b.GetChild(i).(*LineBox)
				if result.GetHeight() > 0 {
					return result
				}
			}
		} else if contentType == BlockBoxContentTypeBlock {
			for i := count - 1; i >= 0; i-- {
				result := b.GetChild(i).(BlockBoxI).AsBlockBox().findLastLineBox()
				if result != nil {
					return result
				}
			}
		}
	}

	return nil
}

func (b *BlockBox) findFirstLineBox() *LineBox {
	contentType := b.GetChildrenContentType()
	count := b.GetChildCount()
	if count > 0 {
		if contentType == BlockBoxContentTypeInline {
			for i := 0; i < count; i++ {
				result := b.GetChild(i).(*LineBox)
				if result.GetHeight() > 0 {
					return result
				}
			}
		} else if contentType == BlockBoxContentTypeBlock {
			for i := 0; i < count; i++ {
				result := b.GetChild(i).(BlockBoxI).AsBlockBox().findFirstLineBox()
				if result != nil {
					return result
				}
			}
		}
	}

	return nil
}

func (b *BlockBox) IsNeedsKeepWithInline(c *LayoutContext) bool {
	if c.IsPrint() && b.GetStyle().IsKeepWithInline() {
		line := b.findFirstLineBox()
		if line != nil {
			linePage := c.GetRootLayer().GetFirstPage(c, line)
			ourPage := c.GetRootLayer().GetFirstPage(c, b.self)
			return linePage != nil && ourPage != nil && linePage.GetPageNo() == ourPage.GetPageNo()+1
		}
	}

	return false
}

func (b *BlockBox) IsFloated() bool {
	return b.floatedBoxData != nil
}

// GetFloatedBoxData may return nil.
func (b *BlockBox) GetFloatedBoxData() *FloatedBoxData {
	return b.floatedBoxData
}

func (b *BlockBox) SetFloatedBoxData(floatedBoxData *FloatedBoxData) {
	b.floatedBoxData = floatedBoxData
}

func (b *BlockBox) GetChildrenHeight() int {
	return b.childrenHeight
}

func (b *BlockBox) SetChildrenHeight(childrenHeight int) {
	b.childrenHeight = childrenHeight
}

func (b *BlockBox) IsFromCaptionedTable() bool {
	return b.fromCaptionedTable
}

func (b *BlockBox) SetFromCaptionedTable(fromTable bool) {
	b.fromCaptionedTable = fromTable
}

func (b *BlockBox) IsInlineBlock() bool {
	return b.IsInline()
}

func (b *BlockBox) IsInMainFlow() bool {
	flowRoot := b.self
	for flowRoot.GetParent() != nil {
		flowRoot = flowRoot.GetParent()
	}

	return flowRoot.IsRoot()
}

func (b *BlockBox) IsContainsInlineContent(c *LayoutContext) bool {
	b.selfBlock().EnsureChildren(c)
	switch b.GetChildrenContentType() {
	case BlockBoxContentTypeInline:
		return true
	case BlockBoxContentTypeEmpty:
		return false
	case BlockBoxContentTypeBlock:
		for _, value := range b.GetChildren() {
			box := value.(BlockBoxI)
			if box.IsContainsInlineContent(c) {
				return true
			}
		}
		return false
	default:
		panic(NewXRRuntimeException("internal error: no children"))
	}
}

func (b *BlockBox) CheckPageContext(c *LayoutContext) bool {
	if !b.GetStyle().IsIdent(CSSNamePage, IdentValueAuto) {
		pageName := b.GetStyle().GetStringProperty(CSSNamePage)
		if pageName != c.GetPageName() && b.IsInDocumentFlow() &&
			b.IsContainsInlineContent(c) {
			c.SetPendingPageName(pageName)
			return true
		}
	} else if c.GetPageName() != "" && b.IsInDocumentFlow() {
		c.SetPendingPageName("")
		return true
	}

	return false
}

func (b *BlockBox) IsNeedsClipOnPaint(c *RenderingContext) bool {
	return !b.IsReplaced() &&
		b.GetStyle().IsIdent(CSSNameOverflow, IdentValueHidden) &&
		b.GetStyle().IsOverflowApplies()
}

func (b *BlockBox) PropagateExtraSpace(
	c *LayoutContext,
	parentContainer *ContentLimitContainer, currentContainer *ContentLimitContainer,
	extraTop int, extraBottom int) {
	start := currentContainer.GetInitialPageNo()
	end := currentContainer.GetLastPageNo()
	current := start

	for current <= end {
		contentLimit :=
			currentContainer.GetContentLimit(current)

		if current != start {
			top := contentLimit.GetTop()
			if top != ContentLimitUndefined {
				parentContainer.UpdateTop(c, top-extraTop)
			}
		}

		if current != end {
			bottom := contentLimit.GetBottom()
			if bottom != ContentLimitUndefined {
				parentContainer.UpdateBottom(c, bottom+extraBottom)
			}
		}

		current++
	}
}

type blockBoxMarginCollapseResult struct {
	maxPositive int
	maxNegative int
}

func (r *blockBoxMarginCollapseResult) update(value int) {
	if value < 0 && value < r.maxNegative {
		r.maxNegative = value
	}

	if value > 0 && value > r.maxPositive {
		r.maxPositive = value
	}
}

func (r *blockBoxMarginCollapseResult) getMargin() int {
	return r.maxPositive + r.maxNegative
}

func (r *blockBoxMarginCollapseResult) hasMargin() bool {
	return r.maxPositive != 0 || r.maxNegative != 0
}
