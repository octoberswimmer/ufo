// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/LineBox.java

package ufo

import (
	"fmt"
	"io"
	"strings"

	"github.com/octoberswimmer/ufo/geom"
)

const lineBoxJustifyNonSpaceShare float32 = 0.20

// lineBoxJustifySpaceShare is computed in float32 arithmetic, as Java does.
var lineBoxJustifySpaceShare = float32(1) - lineBoxJustifyNonSpaceShare

// LineBox contains a single line of text (or other inline content). It
// is created during layout. It also tracks floated and absolute content
// added while laying out the line.
type LineBox struct {
	Box

	endsOnNL                  bool
	containsContent           bool
	containsBlockLevelContent bool

	// floatDistances may be nil.
	floatDistances FloatDistances

	// textDecorations is nil where Java has null.
	textDecorations []*TextDecoration

	paintingTop    int
	paintingHeight int

	// nonFlowContent is nil until AddNonFlowContent is called.
	nonFlowContent []BoxI

	// markerData may be nil.
	markerData *MarkerData

	containsDynamicFunction bool

	contentStart int

	baseline int

	// justificationInfo may be nil.
	justificationInfo *JustificationInfo
}

// NewLineBox creates a line box. parent and style may be nil.
func NewLineBox(parent BoxI, style CalculatedStyleI) *LineBox {
	l := &LineBox{}
	initBox(&l.Box, parent, style)
	l.SetSelf(l)
	return l
}

func (l *LineBox) Dump(c *LayoutContext, indent string, which BoxDump) string {
	if which != BoxDumpRender {
		panic(NewXRRuntimeException(fmt.Sprintf("Unsupported which: %v (expected: %v)", which, BoxDumpRender)))
	}

	var result strings.Builder
	result.WriteString(indent)
	result.WriteString(l.ToString())
	result.WriteByte('\n')

	l.DumpBoxes(c, indent, l.GetNonFlowContent(), BoxDumpRender, &result)
	if len(l.GetNonFlowContent()) != 0 {
		result.WriteByte('\n')
	}
	l.DumpBoxes(c, indent, l.GetChildren(), BoxDumpRender, &result)

	return result.String()
}

func (l *LineBox) GetMarginEdge(cssCtx CssContext, tx int, ty int) *geom.Rectangle {
	result := geom.NewRectangle(l.GetX(), l.GetY(), l.GetContentWidth(), l.GetHeight())
	result.Translate(tx, ty)
	return result
}

func (l *LineBox) PaintInline(c *RenderingContext) {
	if !l.GetParent().GetStyle().IsVisible() {
		return
	}

	if l.IsContainsDynamicFunction() {
		l.lookForDynamicFunctions(c)
		totalLineWidth := InlineBoxingPositionHorizontally(c, l, 0)
		l.SetContentWidth(totalLineWidth)
		l.CalcChildLocations()
		l.Align(true)
		l.CalcPaintingInfo(c, false)
	}

	if l.textDecorations != nil {
		c.GetOutputDevice().DrawTextDecoration(c, l)
	}

	if c.DebugDrawLineBoxes() {
		c.GetOutputDevice().DrawDebugOutline(c, l, FSRGBColorGreen)
	}
}

func (l *LineBox) lookForDynamicFunctions(c *RenderingContext) {
	if l.GetChildCount() > 0 {
		for i := 0; i < l.GetChildCount(); i++ {
			b := l.GetChild(i)
			if inlineLayoutBox, ok := b.(*InlineLayoutBox); ok {
				inlineLayoutBox.LookForDynamicFunctions(c)
			}
		}
	}
}

func (l *LineBox) IsFirstLine() bool {
	parent := l.GetParent()
	return parent != nil && parent.GetChildCount() > 0 && parent.GetChild(0).AsBox() == &l.Box
}

func (l *LineBox) PrunePendingInlineBoxes() {
	if l.GetChildCount() > 0 {
		for i := l.GetChildCount() - 1; i >= 0; i-- {
			b := l.GetChild(i)
			iB, ok := b.(*InlineLayoutBox)
			if !ok {
				break
			}
			iB.PrunePending()
			if iB.IsPending() {
				l.RemoveChildInt(i)
			}
		}
	}
}

func (l *LineBox) IsContainsContent() bool {
	return l.containsContent
}

func (l *LineBox) SetContainsContent(containsContent bool) {
	l.containsContent = containsContent
}

func (l *LineBox) IsEndsOnNL() bool {
	return l.endsOnNL
}

func (l *LineBox) SetEndsOnNL(endsOnNL bool) {
	l.endsOnNL = endsOnNL
}

func (l *LineBox) Align(dynamic bool) {
	align := l.GetParent().GetStyle().GetIdent(CSSNameTextAlign)

	calcX := 0

	if align == IdentValueLeft || align == IdentValueJustify {
		floatDistance := l.GetFloatDistances().LeftFloatDistance()
		calcX = l.GetContentStart() + floatDistance
		if align == IdentValueJustify && dynamic {
			l.Justify()
		}
	} else if align == IdentValueCenter {
		leftFloatDistance := l.GetFloatDistances().LeftFloatDistance()
		rightFloatDistance := l.GetFloatDistances().RightFloatDistance()

		midpoint := leftFloatDistance +
			(l.GetParent().GetContentWidth()-leftFloatDistance-rightFloatDistance)/2

		calcX = midpoint - (l.GetContentWidth()+l.GetContentStart())/2
	} else if align == IdentValueRight {
		floatDistance := l.GetFloatDistances().RightFloatDistance()
		calcX = l.GetParent().GetContentWidth() - floatDistance - l.GetContentWidth()
	}

	if calcX != l.GetX() {
		l.SetX(calcX)
		l.CalcCanvasLocation()
		l.CalcChildLocations()
	}
}

func (l *LineBox) Justify() {
	if !l.isLastLineWithContent() {
		leftFloatDistance := l.GetFloatDistances().LeftFloatDistance()
		rightFloatDistance := l.GetFloatDistances().RightFloatDistance()

		available := l.GetParent().GetContentWidth() -
			leftFloatDistance - rightFloatDistance - l.GetContentStart()

		if available > l.GetContentWidth() {
			toAdd := available - l.GetContentWidth()

			counts := l.countJustifiableChars()

			var info *JustificationInfo
			if counts.GetSpaceCount() == 0 {
				info = lineBoxJustificationInfo(counts, toAdd, 1.0, 0.0)
			} else if !l.GetParent().GetStyle().IsIdent(CSSNameLetterSpacing, IdentValueNormal) {
				info = lineBoxJustificationInfo(counts, toAdd, 0.0, 1.0)
			} else {
				info = lineBoxJustificationInfo(counts, toAdd, lineBoxJustifyNonSpaceShare, lineBoxJustifySpaceShare)
			}

			l.adjustChildren(info)
			l.setJustificationInfo(info)
		}
	}
}

func lineBoxJustificationInfo(counts *CharCounts, toAdd int,
	nonSpaceShare float32, spaceShare float32) *JustificationInfo {
	var nonSpaceAdjust float32 = 0.0
	if counts.GetNonSpaceCount() > 1 {
		nonSpaceAdjust = float32(toAdd) * nonSpaceShare / float32(counts.GetNonSpaceCount()-1)
	}

	var spaceAdjust float32 = 0.0
	if counts.GetSpaceCount() > 0 {
		spaceAdjust = float32(toAdd) * spaceShare / float32(counts.GetSpaceCount())
	}

	return NewJustificationInfo(nonSpaceAdjust, spaceAdjust)
}

func (l *LineBox) adjustChildren(info *JustificationInfo) {
	var adjust float32 = 0.0
	for _, b := range l.GetChildren() {
		b.SetX(b.GetX() + propertyBuilderRound(adjust))

		if inlineLayoutBox, ok := b.(*InlineLayoutBox); ok {
			adjust += inlineLayoutBox.AdjustHorizontalPosition(info, adjust)
		}
	}

	l.CalcChildLocations()
}

// lineBoxNextSiblingLine is the Java cast (LineBox)box.getNextSibling(): nil
// when there is no next sibling, a panic when the sibling is not a line box.
func lineBoxNextSiblingLine(box *LineBox) *LineBox {
	next := box.GetNextSibling()
	if next == nil {
		return nil
	}
	return next.(*LineBox)
}

func (l *LineBox) isLastLineWithContent() bool {
	current := lineBoxNextSiblingLine(l)
	if !l.endsOnNL {
		for current != nil {
			if current.IsContainsContent() {
				return false
			} else {
				current = lineBoxNextSiblingLine(current)
			}
		}
	}
	return true
}

func (l *LineBox) countJustifiableChars() *CharCounts {
	result := NewCharCounts()

	for _, b := range l.GetChildren() {
		if inlineLayoutBox, ok := b.(*InlineLayoutBox); ok {
			inlineLayoutBox.CountJustifiableChars(result)
		}
	}

	return result
}

func (l *LineBox) GetFloatDistances() FloatDistances {
	return l.floatDistances
}

// SetFloatDistances accepts nil.
func (l *LineBox) SetFloatDistances(floatDistances FloatDistances) {
	l.floatDistances = floatDistances
}

func (l *LineBox) IsContainsBlockLevelContent() bool {
	return l.containsBlockLevelContent
}

func (l *LineBox) SetContainsBlockLevelContent(containsBlockLevelContent bool) {
	l.containsBlockLevelContent = containsBlockLevelContent
}

func (l *LineBox) Intersects(cssCtx CssContext, clip geom.Shape) bool {
	return clip == nil || l.intersectsLine(cssCtx, clip) ||
		l.IsContainsBlockLevelContent() && l.intersectsInlineBlocks(cssCtx, clip)
}

func (l *LineBox) intersectsLine(cssCtx CssContext, clip geom.Shape) bool {
	result := l.GetPaintingClipEdge(cssCtx)
	return clip.Intersects(result)
}

func (l *LineBox) GetPaintingClipEdge(cssCtx CssContext) *geom.Rectangle {
	parent := l.GetParent()
	if parent.GetStyle().IsIdent(
		CSSNameFsTextDecorationExtent, IdentValueBlock) ||
		l.GetJustificationInfo() != nil {
		return geom.NewRectangle(
			l.GetAbsX(), l.GetAbsY()+l.paintingTop,
			parent.GetAbsX()+parent.GetTx()+parent.GetContentWidth()-l.GetAbsX(),
			l.paintingHeight)
	} else {
		return geom.NewRectangle(
			l.GetAbsX(), l.GetAbsY()+l.paintingTop, l.GetContentWidth(), l.paintingHeight)
	}
}

func (l *LineBox) intersectsInlineBlocks(cssCtx CssContext, clip geom.Shape) bool {
	for i := 0; i < l.GetChildCount(); i++ {
		child := l.GetChild(i)
		if inlineLayoutBox, ok := child.(*InlineLayoutBox); ok {
			possibleResult := inlineLayoutBox.IntersectsInlineBlocks(cssCtx, clip)
			if possibleResult {
				return true
			}
		} else {
			collector := NewBoxCollector()
			if collector.IntersectsAny(cssCtx, clip, child) {
				return true
			}
		}
	}

	return false
}

// GetTextDecorations may return nil.
func (l *LineBox) GetTextDecorations() []*TextDecoration {
	return l.textDecorations
}

func (l *LineBox) SetTextDecorations(textDecorations []*TextDecoration) {
	l.textDecorations = textDecorations
}

func (l *LineBox) GetPaintingHeight() int {
	return l.paintingHeight
}

func (l *LineBox) SetPaintingHeight(paintingHeight int) {
	l.paintingHeight = paintingHeight
}

func (l *LineBox) GetPaintingTop() int {
	return l.paintingTop
}

func (l *LineBox) SetPaintingTop(paintingTop int) {
	l.paintingTop = paintingTop
}

// AddAllChildrenWithLayer is addAllChildren(List<Box> list, Layer layer); it
// appends to *list.
func (l *LineBox) AddAllChildrenWithLayer(list *[]BoxI, layer *Layer) {
	for i := 0; i < l.GetChildCount(); i++ {
		child := l.GetChild(i)
		if l.GetContainingLayer() == layer {
			*list = append(*list, child)
			if inlineLayoutBox, ok := child.(*InlineLayoutBox); ok {
				inlineLayoutBox.AddAllChildrenWithLayer(list, layer)
			}
		}
	}
}

// GetNonFlowContent returns an empty list when no content was added.
func (l *LineBox) GetNonFlowContent() []BoxI {
	return l.nonFlowContent
}

func (l *LineBox) AddNonFlowContent(box BlockBoxI) {
	l.nonFlowContent = append(l.nonFlowContent, box)
}

func (l *LineBox) Reset(c *LayoutContext) {
	for i := 0; i < len(l.GetNonFlowContent()); i++ {
		content := l.GetNonFlowContent()[i]
		content.Reset(c)
	}
	if l.markerData != nil {
		l.markerData.RestorePreviousReferenceLine(l)
	}
	l.Box.Reset(c)
}

func (l *LineBox) CalcCanvasLocation() {
	parent := l.GetParent()
	if parent == nil {
		panic(NewXRRuntimeException("calcCanvasLocation() called with no parent"))
	}
	l.SetAbsX(parent.GetAbsX() + parent.GetTx() + l.GetX())
	l.SetAbsY(parent.GetAbsY() + parent.GetTy() + l.GetY())
}

func (l *LineBox) CalcChildLocations() {
	l.Box.CalcChildLocations()

	// Update absolute boxes too.  Not necessary most of the time, but
	// it doesn't hurt (revisit this)
	for i := 0; i < len(l.GetNonFlowContent()); i++ {
		content := l.GetNonFlowContent()[i]
		if content.GetStyle().IsAbsolute() {
			content.CalcCanvasLocation()
			content.CalcChildLocations()
		}
	}
}

// GetMarkerData may return nil.
func (l *LineBox) GetMarkerData() *MarkerData {
	return l.markerData
}

func (l *LineBox) SetMarkerData(markerData *MarkerData) {
	l.markerData = markerData
}

func (l *LineBox) IsContainsDynamicFunction() bool {
	return l.containsDynamicFunction
}

func (l *LineBox) SetContainsDynamicFunction(containsPageCounter bool) {
	l.containsDynamicFunction = l.containsDynamicFunction || containsPageCounter
}

func (l *LineBox) GetContentStart() int {
	return l.contentStart
}

func (l *LineBox) SetContentStart(contentOffset int) {
	l.contentStart = contentOffset
}

// FindTrailingText may return nil.
func (l *LineBox) FindTrailingText() *InlineText {
	if l.GetChildCount() == 0 {
		return nil
	}

	for offset := l.GetChildCount() - 1; offset >= 0; offset-- {
		child := l.GetChild(offset)
		if inlineLayoutBox, ok := child.(*InlineLayoutBox); ok {
			result := inlineLayoutBox.FindTrailingText()
			if result != nil && result.IsEmpty() {
				continue
			}
			return result
		} else {
			return nil
		}
	}

	return nil
}

func (l *LineBox) TrimTrailingSpace(c *LayoutContext) {
	text := l.FindTrailingText()

	if text != nil {
		iB := text.GetParent()
		whitespace := iB.GetStyle().GetWhitespace()
		if whitespace == IdentValueNormal || whitespace == IdentValueNowrap {
			text.TrimTrailingSpace(c)
		}
	}
}

// Find may return nil.
func (l *LineBox) Find(cssCtx CssContext, absX int, absY int, findAnonymous bool) BoxI {
	pI := l.GetPaintingInfo()
	if pI != nil && !pI.GetAggregateBounds().Contains(absX, absY) {
		return nil
	}

	for i := 0; i < l.GetChildCount(); i++ {
		child := l.GetChild(i)
		result := child.Find(cssCtx, absX, absY, findAnonymous)
		if result != nil {
			return result
		}
	}

	return nil
}

func (l *LineBox) GetBaseline() int {
	return l.baseline
}

func (l *LineBox) SetBaseline(baseline int) {
	l.baseline = baseline
}

func (l *LineBox) IsContainsOnlyBlockLevelContent() bool {
	if !l.IsContainsBlockLevelContent() {
		return false
	}

	for i := 0; i < l.GetChildCount(); i++ {
		b := l.GetChild(i)
		if _, ok := b.(BlockBoxI); !ok {
			return false
		}
	}

	return true
}

func (l *LineBox) GetRestyleTarget() BoxI {
	return l.GetParent()
}

func (l *LineBox) Restyle(c *LayoutContext) {
	parent := l.GetParent()
	e := parent.GetElement()
	if e != nil {
		style := c.GetSharedContext().GetStyleWithRestyle(e, true)
		l.SetStyle(style.CreateAnonymousStyle(IdentValueBlock))
	}

	l.RestyleChildren(c)
}

func (l *LineBox) IsContainsVisibleContent() bool {
	for _, b := range l.GetChildren() {
		switch child := b.(type) {
		case BlockBoxI:
			if b.GetWidth() > 0 || b.GetHeight() > 0 {
				return true
			}
		case *InlineLayoutBox:
			maybeResult := child.IsContainsVisibleContent()
			if maybeResult {
				return true
			}
		default:
			panic(NewXRRuntimeException(fmt.Sprintf("Unexpected child type: %T", b)))
		}
	}

	return false
}

// ClearSelection appends the boxes whose selection changed to *modified.
func (l *LineBox) ClearSelection(modified *[]BoxI) {
	for _, b := range l.GetNonFlowContent() {
		b.ClearSelection(modified)
	}

	l.Box.ClearSelection(modified)
}

func (l *LineBox) SelectAll() {
	for _, value := range l.GetNonFlowContent() {
		box := value.(BlockBoxI)
		box.SelectAll()
	}

	l.Box.SelectAll()
}

func (l *LineBox) CollectText(c *RenderingContext, buffer *strings.Builder) error {
	for _, b := range l.GetNonFlowContent() {
		if err := b.CollectText(c, buffer); err != nil {
			return err
		}
	}
	if l.IsContainsDynamicFunction() {
		l.lookForDynamicFunctions(c)
	}
	return l.Box.CollectText(c, buffer)
}

func (l *LineBox) ExportText(c *RenderingContext, writer io.Writer) error {
	baselinePos := l.GetAbsY() + l.GetBaseline()
	if baselinePos >= c.GetPage().GetBottom() && l.IsInDocumentFlow() {
		if err := l.ExportPageBoxTextWithYPos(c, writer, baselinePos); err != nil {
			return err
		}
	}

	for _, b := range l.GetNonFlowContent() {
		if err := b.ExportText(c, writer); err != nil {
			return err
		}
	}

	if l.IsContainsContent() {
		var result strings.Builder
		if err := l.CollectText(c, &result); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, inlineBoxJavaTrim(result.String())); err != nil {
			return err
		}
		// Java writes System.lineSeparator().
		if _, err := io.WriteString(writer, "\n"); err != nil {
			return err
		}
	}
	return nil
}

func (l *LineBox) AnalyzePageBreaks(c *LayoutContext, container *ContentLimitContainer) {
	container.UpdateTop(c, l.GetAbsY())
	container.UpdateBottom(c, l.GetAbsY()+l.GetHeight())
}

func (l *LineBox) CheckPagePosition(c *LayoutContext, alwaysBreak bool) {
	if !c.IsPageBreaksAllowed() {
		return
	}

	pageBox := c.GetRootLayer().GetFirstPage(c, l)
	if pageBox != nil {
		needsPageBreak :=
			alwaysBreak || l.GetAbsY()+l.GetHeight() >= pageBox.GetBottom()-c.GetExtraSpaceBottom()

		if needsPageBreak {
			l.ForcePageBreakBefore(c, IdentValueAlways, false)
			l.CalcCanvasLocation()
		} else if pageBox.GetTop()+c.GetExtraSpaceTop() > l.GetAbsY() {
			diff := pageBox.GetTop() + c.GetExtraSpaceTop() - l.GetAbsY()

			l.SetY(l.GetY() + diff)
			l.CalcCanvasLocation()
		}
	}
}

func (l *LineBox) GetJustificationInfo() *JustificationInfo {
	return l.justificationInfo
}

func (l *LineBox) setJustificationInfo(justificationInfo *JustificationInfo) {
	l.justificationInfo = justificationInfo
}
