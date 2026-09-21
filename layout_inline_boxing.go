// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/InlineBoxing.java

package ufo

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// InlineBoxing is responsible for flowing inline content into lines.  Block
// content which participates in an inline formatting context is also handled
// here as well as floating and absolutely positioned content.

const inlineBoxingMaxIterationCount = 100000

// inlineBoxingRound is Math.round(float): the closest int, ties rounding up,
// limited to the range of a Java int, 0 for NaN.
func inlineBoxingRound(f float32) int {
	if f != f {
		return 0
	}
	r := math.Floor(float64(f) + 0.5)
	if r >= math.MaxInt32 {
		return math.MaxInt32
	}
	if r <= math.MinInt32 {
		return math.MinInt32
	}
	return int(r)
}

// inlineBoxingDoubleToInt is Java's (int) cast of a double.
func inlineBoxingDoubleToInt(d float64) int {
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

func InlineBoxingLayoutContent(c *LayoutContext, box BlockBoxI, initialY int, breakAtLine int) {
	maxAvailableWidth := box.GetContentWidth()
	remainingWidth := maxAvailableWidth

	currentLine := inlineBoxingNewLineWithY(c, initialY, box)

	var currentIB *InlineLayoutBox
	var previousIB *InlineLayoutBox

	contentStart := 0

	var openInlineBoxes []*InlineBox

	iBMap := map[*InlineBox]*InlineLayoutBox{}

	if anonymousBlockBox, ok := box.(*AnonymousBlockBox); ok {
		openInlineBoxes = anonymousBlockBox.GetOpenInlineBoxes()
		if openInlineBoxes != nil {
			openInlineBoxes = append([]*InlineBox{}, openInlineBoxes...)
			currentIB = inlineBoxingAddOpenInlineBoxes(
				c, currentLine, openInlineBoxes, maxAvailableWidth, iBMap)
		}
	}

	if openInlineBoxes == nil {
		openInlineBoxes = []*InlineBox{}
	}

	remainingWidth -= c.GetBlockFormattingContext().GetFloatDistance(c, currentLine, remainingWidth)

	parentStyle := box.GetStyle()
	minimumLineHeight := calculatedStyleFloatToInt(parentStyle.GetLineHeight(c))
	indent := calculatedStyleFloatToInt(parentStyle.GetFloatPropertyProportionalWidth(CSSNameTextIndent, float32(maxAvailableWidth), c))
	remainingWidth -= indent
	contentStart += indent

	markerData := c.GetCurrentMarkerData()
	if markerData != nil && box.GetStyle().IsListMarkerInside() {
		remainingWidth -= markerData.GetLayoutWidth()
		contentStart += markerData.GetLayoutWidth()
	}
	c.SetCurrentMarkerData(nil)

	// Never nil: LayoutUtilLayoutFloated takes a nil slice for Java's null.
	pendingFloats := []*FloatLayoutResult{}
	pendingLeftMBP := 0
	pendingRightMBP := 0

	hasFirstLinePEs := false
	pendingInlineLayers := []*Layer{}

	if c.GetFirstLinesTracker().HasStyles() {
		box.StyleTextWithStyle(c, c.GetFirstLinesTracker().DeriveAll(box.GetStyle()))
		hasFirstLinePEs = true
	}

	needFirstLetter := c.GetFirstLettersTracker().HasStyles()
	zeroWidthInlineBlock := false
	lineOffset := 0

	for _, node := range box.GetInlineContent() {
		if node.GetStyle().IsInline() {
			iB := node.(*InlineBox)

			style := iB.GetStyle()
			if iB.IsStartsHere() {
				previousIB = currentIB
				currentIB = NewInlineLayoutBox(c, iB.GetElement(), style, maxAvailableWidth)

				openInlineBoxes = append(openInlineBoxes, iB)
				iBMap[iB] = currentIB

				if previousIB == nil {
					currentLine.AddChildForLayout(c, currentIB)
				} else {
					previousIB.AddInlineChild(c, currentIB)
				}

				if currentIB.GetElement() != nil {
					// NamespaceHandler returns "" where Java returns null.
					name := c.GetNamespaceHandler().GetAnchorName(currentIB.GetElement())
					if name != "" {
						c.AddBoxId(name, currentIB)
					}
					id := c.GetNamespaceHandler().GetID(currentIB.GetElement())
					if id != "" {
						c.AddBoxId(id, currentIB)
					}
				}

				//To break the line well, assume we don't just want to paint padding on next line
				pendingLeftMBP += style.GetMarginBorderPadding(
					c, maxAvailableWidth, CalculatedStyleEdgeLeft)
				pendingRightMBP += style.GetMarginBorderPadding(
					c, maxAvailableWidth, CalculatedStyleEdgeRight)
			}

			var master string
			if iB.IsDynamicFunction() {
				master = iB.GetContentFunction().GetLayoutReplacementText()
			} else {
				master = iB.GetText()
			}
			lbContext := NewLineBreakContext(master, iB.GetTextNode())

			q := 0
			// Java: do { ... } while (!lbContext.isFinished()); a continue in
			// the body goes to the test of the condition.
			for firstIteration := true; firstIteration || !lbContext.IsFinished(); firstIteration = false {
				q++
				if q-1 > inlineBoxingMaxIterationCount {
					panic(NewXRRuntimeException("Too many iterations (" + strconv.Itoa(q) + ") in InlineBoxing, giving up."))
				}

				lbContext.Reset()

				fit := 0
				if lbContext.GetStart() == 0 {
					fit += pendingLeftMBP + pendingRightMBP
				}

				trimmedLeadingSpace := false
				if inlineBoxingHasTrimmableLeadingSpace(
					currentLine, style, lbContext, zeroWidthInlineBlock) {
					trimmedLeadingSpace = true
					inlineBoxingTrimLeadingSpace(lbContext)
				}

				lbContext.SetEndsOnNL(false)

				zeroWidthInlineBlock = false

				if lbContext.GetStartSubstring() == "" {
					break
				}

				if needFirstLetter && !lbContext.IsFinished() {
					firstLetter :=
						inlineBoxingAddFirstLetterBox(c, currentLine, currentIB, lbContext,
							maxAvailableWidth, remainingWidth)
					remainingWidth -= firstLetter.GetInlineWidth()

					if currentIB.IsStartsHere() {
						pendingLeftMBP -= currentIB.GetStyle().GetMarginBorderPadding(
							c, maxAvailableWidth, CalculatedStyleEdgeLeft)
					}

					needFirstLetter = false
				} else {
					lbContext.SaveEnd()
					inlineText := inlineBoxingLayoutText(
						c, iB.GetStyle(), remainingWidth-fit, maxAvailableWidth, lbContext, false)
					if lbContext.IsUnbreakable() && !currentLine.IsContainsContent() {
						delta := c.GetBlockFormattingContext().GetNextLineBoxDelta(c, currentLine, maxAvailableWidth)
						if delta > 0 {
							currentLine.SetY(currentLine.GetY() + delta)
							currentLine.CalcCanvasLocation()
							remainingWidth = maxAvailableWidth
							remainingWidth -= c.GetBlockFormattingContext().GetFloatDistance(c, currentLine, maxAvailableWidth)
							lbContext.ResetEnd()
							continue
						}
					}

					if !lbContext.IsUnbreakable() ||
						lbContext.IsUnbreakable() && !currentLine.IsContainsContent() {
						if iB.IsDynamicFunction() {
							inlineText.SetFunctionData(NewFunctionData(
								iB.GetContentFunction(), iB.GetFunction()))
						}
						inlineText.SetTrimmedLeadingSpace(trimmedLeadingSpace)
						currentLine.SetContainsDynamicFunction(inlineText.IsDynamicFunction())
						currentIB.AddInlineChild(c, inlineText)
						currentLine.SetContainsContent(true)
						lbContext.SetStart(lbContext.GetEnd())
						remainingWidth -= inlineText.GetWidth()

						if currentIB.IsStartsHere() {
							marginBorderPadding :=
								currentIB.GetStyle().GetMarginBorderPadding(
									c, maxAvailableWidth, CalculatedStyleEdgeLeft)
							pendingLeftMBP -= marginBorderPadding
							remainingWidth -= marginBorderPadding
						}
					} else {
						lbContext.ResetEnd()
					}
				}

				if lbContext.IsNeedsNewLine() {
					if iB.GetStyle().IsTextJustify() {
						currentLine.TrimTrailingSpace(c)
					}
					inlineBoxingSaveLine(currentLine, c, box, minimumLineHeight,
						maxAvailableWidth, &pendingFloats,
						hasFirstLinePEs, &pendingInlineLayers, markerData,
						contentStart, inlineBoxingIsAlwaysBreak(c, box, breakAtLine, lineOffset))
					lineOffset++
					markerData = nil
					contentStart = 0
					if currentLine.IsFirstLine() && hasFirstLinePEs {
						lbContext.SetMaster(LayoutTextUtilTransformText(iB.GetText(), iB.GetStyle()))
					}
					if lbContext.IsEndsOnNL() {
						currentLine.SetEndsOnNL(true)
					}
					previousLine := currentLine
					currentLine = inlineBoxingNewLine(c, previousLine, box)
					currentIB = inlineBoxingAddOpenInlineBoxes(
						c, currentLine, openInlineBoxes, maxAvailableWidth, iBMap)
					remainingWidth = maxAvailableWidth
					remainingWidth -= c.GetBlockFormattingContext().GetFloatDistance(c, currentLine, remainingWidth)
				}
			}

			if iB.IsEndsHere() {
				rightMBP := style.GetMarginBorderPadding(
					c, maxAvailableWidth, CalculatedStyleEdgeRight)

				pendingRightMBP -= rightMBP
				remainingWidth -= rightMBP

				openInlineBoxes = openInlineBoxes[:len(openInlineBoxes)-1]

				if currentIB.IsPending() {
					currentIB.UnmarkPending(c)

					// Reset to correct value
					currentIB.SetStartsHere(iB.IsStartsHere())
				}

				currentIB.SetEndsHere(true)

				if currentIB.GetStyle().RequiresLayer() {
					if !currentIB.IsPending() && (currentIB.GetElement() == nil ||
						currentIB.GetElement() != c.GetLayer().GetMaster().GetElement()) {
						panic(NewXRRuntimeException("internal error"))
					}
					if !currentIB.IsPending() {
						c.GetLayer().SetEnd(currentIB)
						c.PopLayer()
						pendingInlineLayers = append(pendingInlineLayers, currentIB.GetContainingLayer())
					}
				}

				if _, ok := currentIB.GetParent().(*LineBox); ok {
					currentIB = nil
				} else {
					currentIB = currentIB.GetParent().(*InlineLayoutBox)
				}
			}
		} else {
			child := node.(BlockBoxI)

			if child.GetStyle().IsNonFlowContent() {
				remainingWidth -= inlineBoxingProcessOutOfFlowContent(
					c, currentLine, child, remainingWidth, &pendingFloats)
			} else if child.GetStyle().IsInlineBlock() || child.GetStyle().IsInlineTable() {
				inlineBoxingLayoutInlineBlockContent(c, box, child, initialY)

				if child.GetWidth() > remainingWidth && currentLine.IsContainsContent() {
					inlineBoxingSaveLine(currentLine, c, box, minimumLineHeight,
						maxAvailableWidth, &pendingFloats, hasFirstLinePEs,
						&pendingInlineLayers, markerData, contentStart,
						inlineBoxingIsAlwaysBreak(c, box, breakAtLine, lineOffset))
					lineOffset++
					markerData = nil
					contentStart = 0
					previousLine := currentLine
					currentLine = inlineBoxingNewLine(c, previousLine, box)
					currentIB = inlineBoxingAddOpenInlineBoxes(
						c, currentLine, openInlineBoxes, maxAvailableWidth, iBMap)
					remainingWidth = maxAvailableWidth
					remainingWidth -= c.GetBlockFormattingContext().GetFloatDistance(c, currentLine, remainingWidth)

					child.Reset(c)
					inlineBoxingLayoutInlineBlockContent(c, box, child, initialY)
				}

				if currentIB == nil {
					currentLine.AddChildForLayout(c, child)
				} else {
					currentIB.AddInlineChild(c, child)
				}

				currentLine.SetContainsContent(true)
				currentLine.SetContainsBlockLevelContent(true)

				remainingWidth -= child.GetWidth()

				if currentIB != nil && currentIB.IsStartsHere() {
					pendingLeftMBP -= currentIB.GetStyle().GetMarginBorderPadding(
						c, maxAvailableWidth, CalculatedStyleEdgeLeft)
				}

				needFirstLetter = false

				if child.GetWidth() == 0 {
					zeroWidthInlineBlock = true
				}
			}
		}
	}

	currentLine.TrimTrailingSpace(c)
	inlineBoxingSaveLine(currentLine, c, box, minimumLineHeight,
		maxAvailableWidth, &pendingFloats, hasFirstLinePEs,
		&pendingInlineLayers, markerData, contentStart,
		inlineBoxingIsAlwaysBreak(c, box, breakAtLine, lineOffset))
	if currentLine.IsFirstLine() && currentLine.GetHeight() == 0 && markerData != nil {
		c.SetCurrentMarkerData(markerData)
	}

	box.SetContentWidth(maxAvailableWidth)
	box.SetHeight(currentLine.GetY() + currentLine.GetHeight())
}

func inlineBoxingIsAlwaysBreak(c *LayoutContext, parent BlockBoxI, breakAtLine int, lineOffset int) bool {
	if parent.IsCurrentBreakAtLineContext(c) {
		return lineOffset == breakAtLine
	} else {
		return breakAtLine > 0 && lineOffset == breakAtLine
	}
}

func inlineBoxingAddFirstLetterBox(c *LayoutContext, current *LineBox,
	currentIB *InlineLayoutBox, lbContext *LineBreakContext, maxAvailableWidth int,
	remainingWidth int) *InlineLayoutBox {
	previous := currentIB.GetStyle()

	currentIB.SetStyle(c.GetFirstLettersTracker().DeriveAll(currentIB.GetStyle()))

	iB := NewInlineLayoutBox(c, nil, currentIB.GetStyle(), maxAvailableWidth)
	iB.SetStartsHere(true)
	iB.SetEndsHere(true)

	currentIB.AddInlineChild(c, iB)
	current.SetContainsContent(true)

	text := inlineBoxingLayoutText(c, iB.GetStyle(), remainingWidth, maxAvailableWidth, lbContext, true)
	iB.AddInlineChild(c, text)
	iB.SetInlineWidth(text.GetWidth())

	lbContext.SetStart(lbContext.GetEnd())

	c.GetFirstLettersTracker().ClearStyles()
	currentIB.SetStyle(previous)

	return iB
}

func inlineBoxingLayoutInlineBlockContent(
	c *LayoutContext, containingBlock BlockBoxI, inlineBlock BlockBoxI, initialY int) {
	inlineBlock.SetContainingBlock(containingBlock)
	inlineBlock.SetContainingLayer(c.GetLayer())
	inlineBlock.InitStaticPos(c, containingBlock, initialY)
	inlineBlock.CalcCanvasLocation()
	inlineBlock.Layout(c)
}

// InlineBoxingPositionHorizontally is the public overload
// positionHorizontally(CssContext, Box, int).
func InlineBoxingPositionHorizontally(c CssContext, current BoxI, start int) int {
	x := start

	var currentIB *InlineLayoutBox

	if inlineLayoutBox, ok := current.(*InlineLayoutBox); ok {
		currentIB = inlineLayoutBox
		x += currentIB.GetLeftMarginBorderPadding(c)
	}

	for i := 0; i < current.GetChildCount(); i++ {
		b := current.GetChild(i)
		if iB, ok := b.(*InlineLayoutBox); ok {
			iB.SetX(x)
			x += inlineBoxingPositionHorizontallyInlineLayoutBox(c, iB, x)
		} else {
			b.SetX(x)
			x += b.GetWidth()
		}
	}

	if currentIB != nil {
		x += currentIB.GetRightMarginPaddingBorder(c)
		currentIB.SetInlineWidth(x - start)
	}

	return x - start
}

// inlineBoxingPositionHorizontallyInlineLayoutBox is the private overload
// positionHorizontally(CssContext, InlineLayoutBox, int).
func inlineBoxingPositionHorizontallyInlineLayoutBox(c CssContext, current *InlineLayoutBox, start int) int {
	x := start

	x += current.GetLeftMarginBorderPadding(c)

	for i := 0; i < current.GetInlineChildCount(); i++ {
		child := current.GetInlineChild(i)
		switch child := child.(type) {
		case *InlineLayoutBox:
			iB := child
			iB.SetX(x)
			x += inlineBoxingPositionHorizontallyInlineLayoutBox(c, iB, x)
		case *InlineText:
			iT := child
			iT.SetX(x - start)
			x += iT.GetWidth()
		case BlockBoxI:
			b := child
			b.SetX(x)
			x += b.GetWidth()
		default:
			panic(NewXRRuntimeException("Unexpected inline child: " + fmt.Sprint(child)))
		}
	}

	x += current.GetRightMarginPaddingBorder(c)

	current.SetInlineWidth(x - start)

	return x - start
}

func InlineBoxingCreateDefaultStrutMetrics(c *LayoutContext, container BoxI) *StrutMetrics {
	strutM := container.GetStyle().GetFSFontMetrics(c)
	measurements := inlineBoxingGetInitialMeasurements(c, container, strutM)

	return NewStrutMetrics(
		strutM.GetAscent(), measurements.GetBaseline(), strutM.GetDescent())
}

// markerData may be nil.
func inlineBoxingPositionVertically(
	c *LayoutContext, container BoxI, current *LineBox, markerData *MarkerData) {
	if current.GetChildCount() == 0 || !current.IsContainsVisibleContent() {
		current.SetHeight(0)
	} else {
		strutM := container.GetStyle().GetFSFontMetrics(c)
		measurements := inlineBoxingGetInitialMeasurements(c, container, strutM)
		vaContext := NewVerticalAlignContextInlineBoxMeasurements(measurements)

		lBDecorations := InlineBoxingCalculateTextDecorations(c, container, measurements.GetBaseline(), strutM)
		current.SetTextDecorations(lBDecorations)

		for i := 0; i < current.GetChildCount(); i++ {
			child := current.GetChild(i)
			inlineBoxingPositionInlineContentVertically(c, vaContext, child)
		}

		vaContext.AlignChildren()

		current.SetHeight(vaContext.GetLineBoxHeight())

		paintingTop := vaContext.GetPaintingTop()
		paintingBottom := vaContext.GetPaintingBottom()

		if vaContext.GetInlineTop() < 0 {
			inlineBoxingMoveLineContents(current, -vaContext.GetInlineTop())
			for _, lBDecoration := range lBDecorations {
				lBDecoration.SetOffset(lBDecoration.GetOffset() - vaContext.GetInlineTop())
			}
			paintingTop -= vaContext.GetInlineTop()
			paintingBottom -= vaContext.GetInlineTop()
		}

		if markerData != nil {
			strutMetrics := markerData.GetStructMetrics()
			strutMetrics.SetBaseline(measurements.GetBaseline() - vaContext.GetInlineTop())
			markerData.SetReferenceLine(current)
			current.SetMarkerData(markerData)
		}

		current.SetBaseline(measurements.GetBaseline() - vaContext.GetInlineTop())

		current.SetPaintingTop(paintingTop)
		current.SetPaintingHeight(paintingBottom - paintingTop)
	}
}

func inlineBoxingPositionInlineVertically(c *LayoutContext,
	vaContext *VerticalAlignContext, iB *InlineLayoutBox) {
	iBMeasurements := inlineBoxingCalculateInlineMeasurements(c, iB, vaContext)
	vaContext.PushMeasurements(iBMeasurements)
	inlineBoxingPositionInlineChildrenVertically(c, iB, vaContext)
	vaContext.PopMeasurements()
}

func inlineBoxingPositionInlineBlockVertically(
	c *LayoutContext, vaContext *VerticalAlignContext, inlineBlock BlockBoxI) {
	baseline := inlineBlock.CalcInlineBaseline(c)
	descent := inlineBlock.GetHeight() - baseline
	inlineBoxingAlignInlineContent(c, inlineBlock, float32(baseline), float32(descent), vaContext)

	vaContext.UpdateInlineTop(inlineBlock.GetY())
	vaContext.UpdatePaintingTop(inlineBlock.GetY())

	vaContext.UpdateInlineBottom(inlineBlock.GetY() + inlineBlock.GetHeight())
	vaContext.UpdatePaintingBottom(inlineBlock.GetY() + inlineBlock.GetHeight())
}

func inlineBoxingMoveLineContents(current *LineBox, ty int) {
	for i := 0; i < current.GetChildCount(); i++ {
		child := current.GetChild(i)
		child.SetY(child.GetY() + ty)
		if inlineLayoutBox, ok := child.(*InlineLayoutBox); ok {
			inlineBoxingMoveInlineContents(inlineLayoutBox, ty)
		}
	}
}

func inlineBoxingMoveInlineContents(box *InlineLayoutBox, ty int) {
	for i := 0; i < box.GetInlineChildCount(); i++ {
		obj := box.GetInlineChild(i)
		if childBox, ok := obj.(BoxI); ok {
			childBox.SetY(childBox.GetY() + ty)

			if inlineLayoutBox, ok := obj.(*InlineLayoutBox); ok {
				inlineBoxingMoveInlineContents(inlineLayoutBox, ty)
			}
		}
	}
}

func inlineBoxingCalculateInlineMeasurements(c *LayoutContext, iB *InlineLayoutBox,
	vaContext *VerticalAlignContext) *InlineBoxMeasurements {
	fm := iB.GetStyle().GetFSFontMetrics(c)

	style := iB.GetStyle()
	lineHeight := style.GetLineHeight(c)

	halfLeading := inlineBoxingRound((lineHeight - iB.GetStyle().GetFont(c).Size()) / 2)
	if halfLeading > 0 {
		halfLeading = inlineBoxingRound((lineHeight -
			(fm.GetDescent() + fm.GetAscent())) / 2)
	}

	iB.SetBaseline(inlineBoxingRound(fm.GetAscent()))

	inlineBoxingAlignInlineContent(c, iB, fm.GetAscent(), fm.GetDescent(), vaContext)
	decorations := InlineBoxingCalculateTextDecorations(c, iB, iB.GetBaseline(), fm)
	iB.SetTextDecorations(decorations)

	padding := iB.GetPadding(c)
	border := iB.GetBorder(c)

	baseline := iB.GetY() + iB.GetBaseline()
	inlineTop := iB.GetY() - halfLeading

	return NewInlineBoxMeasurements(
		baseline,
		iB.GetY(), calculatedStyleFloatToInt(float32(baseline)+fm.GetDescent()),
		inlineTop, inlineBoxingRound(float32(inlineTop)+lineHeight),
		inlineBoxingDoubleToInt(math.Floor(float64(float32(iB.GetY())-border.Top()-padding.Top()))),
		inlineBoxingDoubleToInt(math.Ceil(float64(float32(iB.GetY())+
			fm.GetAscent()+fm.GetDescent()+
			border.Bottom()+padding.Bottom()))),
	)
}

func inlineBoxingContainsIdent(idents []FSDerivedValue, ident *IdentValue) bool {
	for _, value := range idents {
		if value == FSDerivedValue(ident) {
			return true
		}
	}
	return false
}

// InlineBoxingCalculateTextDecorations returns an empty list when the style
// of box has no text decorations.
func InlineBoxingCalculateTextDecorations(c *LayoutContext, box BoxI, baseline int, fm FSFontMetrics) []*TextDecoration {
	idents := box.GetStyle().GetTextDecorations()
	if idents == nil {
		return []*TextDecoration{}
	}

	result := make([]*TextDecoration, 0, len(idents))
	if inlineBoxingContainsIdent(idents, IdentValueUnderline) {
		decoration := inlineBoxingCalculateTextDecoration(c, box, baseline, fm)
		result = append(result, decoration)
	}

	if inlineBoxingContainsIdent(idents, IdentValueLineThrough) {
		decoration := NewTextDecoration(IdentValueLineThrough)
		decoration.SetOffset(inlineBoxingRound(float32(baseline) + fm.GetStrikethroughOffset()))
		decoration.SetThickness(inlineBoxingRound(fm.GetStrikethroughThickness()))
		result = append(result, decoration)
	}

	if inlineBoxingContainsIdent(idents, IdentValueOverline) {
		decoration := NewTextDecoration(IdentValueOverline)
		decoration.SetOffset(0)
		decoration.SetThickness(inlineBoxingRound(fm.GetUnderlineThickness()))
		result = append(result, decoration)
	}
	return result
}

func inlineBoxingCalculateTextDecoration(c *LayoutContext, box BoxI, baseline int, fm FSFontMetrics) *TextDecoration {
	decoration := NewTextDecoration(IdentValueUnderline)
	style := box.GetStyle()

	// Get CSS properties
	position := style.GetTextUnderlinePosition()
	offsetValue := style.GetTextUnderlineOffset()

	// Step 1: Determine base position (text-underline-position)
	var basePosition int
	if position == IdentValueUnder {
		// Place underline at the bottom of the em box, consistent with
		// Blink's BottomOfEmHeight. getDescent() is based on the font's
		// bounding box and can be excessively large for CJK fonts (e.g.
		// NotoSansJP). getUnderlineOffset() is derived from the font
		// descriptor's typographic descent, which accurately reflects
		// the em box bottom edge.
		effectiveDescent := inlineBoxingMinFloat(fm.GetDescent(), fm.GetUnderlineOffset())
		basePosition = inlineBoxingRound(float32(baseline) + effectiveDescent + 1)
	} else {
		basePosition = baseline
	}

	// Calculate offset based on text-underline-offset
	var offset int
	if offsetValue.IsIdent() && offsetValue.AsIdentValue() == IdentValueAuto {
		offset = basePosition
	} else {
		// User-specified offset: move from base position
		offsetFloat := style.GetFloatPropertyProportionalTo(CSSNameTextUnderlineOffset, 0, c)
		offset = basePosition + inlineBoxingRound(offsetFloat)
	}
	if position == IdentValueAuto && offsetValue.IsIdent() && offsetValue.AsIdentValue() == IdentValueAuto {
		// JDK on Linux returns some goofy values for
		// LineMetrics.getUnderlineOffset(). Compensate by always
		// making sure underline fits inside the descender
		if fm.GetUnderlineOffset() == 0 { // HACK, are we running under the JDK
			maxOffset :=
				baseline + calculatedStyleFloatToInt(fm.GetDescent()) - decoration.GetThickness()
			if offset > maxOffset {
				offset = maxOffset
			}
		}
	}

	decoration.SetOffset(offset)
	decoration.SetThickness(inlineBoxingRound(fm.GetUnderlineThickness()))

	return decoration
}

// inlineBoxingMinFloat is Math.min(float, float): NaN if either argument is
// NaN, and -0.0 is smaller than 0.0.
func inlineBoxingMinFloat(a float32, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}

// XXX vertical-align: super/middle/sub could be improved (in particular,
// super and sub should be sized by the measurements of our inline parent,
// not us)
func inlineBoxingAlignInlineContent(c *LayoutContext, box BoxI,
	ascent float32, descent float32, vaContext *VerticalAlignContext) {
	measurements := vaContext.GetParentMeasurements()

	style := box.GetStyle()

	if style.IsLength(CSSNameVerticalAlign) {
		box.SetY(calculatedStyleFloatToInt(float32(measurements.GetBaseline()) - ascent -
			style.GetFloatPropertyProportionalTo(CSSNameVerticalAlign, style.GetLineHeight(c), c)))
	} else {
		vAlign := style.GetIdent(CSSNameVerticalAlign)

		if vAlign == IdentValueBaseline {
			box.SetY(inlineBoxingRound(float32(measurements.GetBaseline()) - ascent))
		} else if vAlign == IdentValueTextTop {
			box.SetY(measurements.GetTextTop())
		} else if vAlign == IdentValueTextBottom {
			box.SetY(inlineBoxingRound(float32(measurements.GetTextBottom()) - descent - ascent))
		} else if vAlign == IdentValueMiddle {
			// FIXME: findbugs, loss of precision, try / (float)2
			box.SetY(inlineBoxingRound(float32(measurements.GetBaseline()-measurements.GetTextTop())/2 -
				(ascent+descent)/2))
		} else if vAlign == IdentValueSuper {
			box.SetY(inlineBoxingRound(float32(measurements.GetBaseline()) - 3*ascent/2))
		} else if vAlign == IdentValueSub {
			box.SetY(inlineBoxingRound(float32(measurements.GetBaseline()) - ascent/2))
		} else {
			box.SetY(inlineBoxingRound(float32(measurements.GetBaseline()) - ascent))
		}
	}
}

func inlineBoxingGetInitialMeasurements(
	c *LayoutContext, container BoxI, strutM FSFontMetrics) *InlineBoxMeasurements {
	lineHeight := container.GetStyle().GetLineHeight(c)

	halfLeading := inlineBoxingRound((lineHeight -
		container.GetStyle().GetFont(c).Size()) / 2)
	if halfLeading > 0 {
		halfLeading = inlineBoxingRound((lineHeight -
			(strutM.GetDescent() + strutM.GetAscent())) / 2)
	}

	baseline := calculatedStyleFloatToInt(float32(halfLeading) + strutM.GetAscent())
	return NewInlineBoxMeasurements(baseline,
		halfLeading, calculatedStyleFloatToInt(float32(baseline)+strutM.GetDescent()),
		halfLeading, calculatedStyleFloatToInt(float32(halfLeading)+lineHeight),
		0, 0,
	)
}

func inlineBoxingPositionInlineChildrenVertically(c *LayoutContext, current *InlineLayoutBox,
	vaContext *VerticalAlignContext) {
	for i := 0; i < current.GetInlineChildCount(); i++ {
		child := current.GetInlineChild(i)
		if childBox, ok := child.(BoxI); ok {
			inlineBoxingPositionInlineContentVertically(c, vaContext, childBox)
		}
	}
}

func inlineBoxingPositionInlineContentVertically(c *LayoutContext,
	vaContext *VerticalAlignContext, child BoxI) {
	vaTarget := vaContext
	if !child.GetStyle().IsLength(CSSNameVerticalAlign) {
		vAlign := child.GetStyle().GetIdent(
			CSSNameVerticalAlign)
		if vAlign == IdentValueTop || vAlign == IdentValueBottom {
			vaTarget = vaContext.CreateChild(child)
		}
	}
	if iB, ok := child.(*InlineLayoutBox); ok {
		inlineBoxingPositionInlineVertically(c, vaTarget, iB)
	} else { // any other Box class
		inlineBoxingPositionInlineBlockVertically(c, vaTarget, child.(BlockBoxI))
	}
}

// pendingFloats and pendingInlineLayers point to the lists of the caller,
// which this function empties. markerData may be nil.
func inlineBoxingSaveLine(current *LineBox, c *LayoutContext,
	block BlockBoxI, minHeight int,
	maxAvailableWidth int, pendingFloats *[]*FloatLayoutResult,
	hasFirstLinePCs bool, pendingInlineLayers *[]*Layer,
	markerData *MarkerData, contentStart int, alwaysBreak bool) {
	current.SetContentStart(contentStart)
	current.PrunePendingInlineBoxes()

	totalLineWidth := InlineBoxingPositionHorizontally(c, current, 0)
	current.SetContentWidth(totalLineWidth)

	inlineBoxingPositionVertically(c, block, current, markerData)

	// XXX Revisit this.  Do we need this when dealing with unbreakable
	// text?  Is a line required to always have a minimum height?
	if current.GetHeight() != 0 &&
		current.GetHeight() < minHeight &&
		!current.IsContainsOnlyBlockLevelContent() {
		current.SetHeight(minHeight)
	}

	if c.IsPrint() {
		current.CheckPagePosition(c, alwaysBreak)
	}

	inlineBoxingAlignLine(c, current, maxAvailableWidth)

	current.CalcChildLocations()

	block.AddChildForLayout(c, current)

	if len(*pendingInlineLayers) != 0 {
		inlineBoxingFinishPendingInlineLayers(c, *pendingInlineLayers)
		*pendingInlineLayers = (*pendingInlineLayers)[:0]
	}

	if hasFirstLinePCs && current.IsFirstLine() {
		c.GetFirstLinesTracker().ClearStyles()
		block.StyleText(c)
	}

	if len(*pendingFloats) != 0 {
		for _, layoutResult := range *pendingFloats {
			LayoutUtilLayoutFloated(c, current, layoutResult.GetBlock(), maxAvailableWidth, nil)
			current.AddNonFlowContent(layoutResult.GetBlock())
		}
		// stays non-nil, see InlineBoxingLayoutContent
		*pendingFloats = (*pendingFloats)[:0]
	}
}

func inlineBoxingAlignLine(c *LayoutContext, current *LineBox, maxAvailableWidth int) {
	var distances FloatDistances
	if !current.IsContainsDynamicFunction() && !current.GetParent().GetStyle().IsTextJustify() {
		distances = &inlineBoxingDynamicFloatDistances{c: c, current: current, maxAvailableWidth: maxAvailableWidth}
	} else {
		distances = newInlineBoxingStaticFloatDistances(c, current, maxAvailableWidth)
	}
	current.SetFloatDistances(distances)
	current.Align(false)
	if !current.IsContainsDynamicFunction() && !current.GetParent().GetStyle().IsTextJustify() {
		current.SetFloatDistances(nil)
	}
}

// inlineBoxingStaticFloatDistances is the private record
// InlineBoxing.StaticFloatDistances.
type inlineBoxingStaticFloatDistances struct {
	leftFloatDistance  int
	rightFloatDistance int
}

func newInlineBoxingStaticFloatDistances(c *LayoutContext, current *LineBox, maxAvailableWidth int) *inlineBoxingStaticFloatDistances {
	return &inlineBoxingStaticFloatDistances{
		leftFloatDistance:  c.GetBlockFormattingContext().GetLeftFloatDistance(c, current, maxAvailableWidth),
		rightFloatDistance: c.GetBlockFormattingContext().GetRightFloatDistance(c, current, maxAvailableWidth),
	}
}

func (d *inlineBoxingStaticFloatDistances) LeftFloatDistance() int {
	return d.leftFloatDistance
}

func (d *inlineBoxingStaticFloatDistances) RightFloatDistance() int {
	return d.rightFloatDistance
}

// inlineBoxingDynamicFloatDistances is the private record
// InlineBoxing.DynamicFloatDistances.
type inlineBoxingDynamicFloatDistances struct {
	c                 *LayoutContext
	current           *LineBox
	maxAvailableWidth int
}

func (d *inlineBoxingDynamicFloatDistances) LeftFloatDistance() int {
	return d.c.GetBlockFormattingContext().GetLeftFloatDistance(d.c, d.current, d.maxAvailableWidth)
}

func (d *inlineBoxingDynamicFloatDistances) RightFloatDistance() int {
	return d.c.GetBlockFormattingContext().GetRightFloatDistance(d.c, d.current, d.maxAvailableWidth)
}

func inlineBoxingFinishPendingInlineLayers(c *LayoutContext, layers []*Layer) {
	for _, l := range layers {
		l.PositionChildren(c)
	}
}

func inlineBoxingLayoutText(c *LayoutContext, style CalculatedStyleI,
	remainingWidth int, maxAvailableWidth int,
	lbContext *LineBreakContext, needFirstLetter bool) *InlineText {
	masterText := lbContext.GetMaster()
	if needFirstLetter {
		masterText = LayoutTextUtilTransformFirstLetterText(masterText, style)
		lbContext.SetMaster(masterText)
		BreakerBreakFirstLetter(c, lbContext, remainingWidth, style)
	} else {
		BreakerBreakText(c, lbContext, remainingWidth, maxAvailableWidth, style)
	}

	return NewInlineText(lbContext.GetMaster(), lbContext.GetTextNode(),
		lbContext.GetStart(), lbContext.GetEnd(),
		lbContext.GetWidth())
}

// pendingFloats points to the list of the caller, to which a pending float is
// added.
func inlineBoxingProcessOutOfFlowContent(
	c *LayoutContext, current *LineBox, block BlockBoxI,
	available int, pendingFloats *[]*FloatLayoutResult) int {
	result := 0
	style := block.GetStyle()
	if style.IsAbsolute() || style.IsFixed() {
		LayoutUtilLayoutAbsolute(c, current, block)
		current.AddNonFlowContent(block)
	} else if style.IsFloated() {
		layoutResult := LayoutUtilLayoutFloated(
			c, current, block, available, *pendingFloats)
		if layoutResult.IsPending() {
			*pendingFloats = append(*pendingFloats, layoutResult)
		} else {
			result = layoutResult.GetBlock().GetWidth()
			current.AddNonFlowContent(layoutResult.GetBlock())
		}
	} else if style.IsRunning() {
		block.SetStaticEquivalent(current)
		c.GetRootLayer().AddRunningBlock(block)
	}

	return result
}

func inlineBoxingHasTrimmableLeadingSpace(
	line *LineBox, style CalculatedStyleI, lbContext *LineBreakContext,
	zeroWidthInlineBlock bool) bool {
	if (!line.IsContainsContent() || zeroWidthInlineBlock) &&
		strings.HasPrefix(lbContext.GetStartSubstring(), WhitespaceStripperSpace) {
		whitespace := style.GetWhitespace()
		return whitespace == IdentValueNormal ||
			whitespace == IdentValueNowrap ||
			whitespace == IdentValuePreLine ||
			whitespace == IdentValuePreWrap &&
				lbContext.GetStart() > 0 &&
				len(lbContext.GetMaster()) > lbContext.GetStart()-1 &&
				rune(lbContext.GetMaster()[lbContext.GetStart()-1]) != WhitespaceStripperEolc
	}
	return false
}

func inlineBoxingTrimLeadingSpace(lbContext *LineBreakContext) {
	s := lbContext.GetStartSubstring()
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	lbContext.SetStart(lbContext.GetStart() + i)
}

// inlineBoxingNewLine is newLine(LayoutContext, LineBox, Box); previousLine
// may be nil.
func inlineBoxingNewLine(c *LayoutContext, previousLine *LineBox, box BoxI) *LineBox {
	y := 0

	if previousLine != nil {
		y = previousLine.GetY() + previousLine.GetHeight()
	}

	return inlineBoxingNewLineWithY(c, y, box)
}

// inlineBoxingNewLineWithY is newLine(LayoutContext, int, Box).
func inlineBoxingNewLineWithY(c *LayoutContext, y int, box BoxI) *LineBox {
	result := NewLineBox(box, box.GetStyle().CreateAnonymousStyle(IdentValueBlock))
	result.InitContainingLayer(c)
	result.SetY(y)
	result.CalcCanvasLocation()
	return result
}

// inlineBoxingAddOpenInlineBoxes returns nil when openParents is empty.
func inlineBoxingAddOpenInlineBoxes(
	c *LayoutContext, line *LineBox, openParents []*InlineBox, cbWidth int,
	iBMap map[*InlineBox]*InlineLayoutBox) *InlineLayoutBox {
	var currentIB *InlineLayoutBox
	var previousIB *InlineLayoutBox

	first := true
	for _, iB := range openParents {
		currentIB = NewInlineLayoutBox(
			c, iB.GetElement(), iB.GetStyle(), cbWidth)

		prev := iBMap[iB]
		if prev != nil {
			currentIB.SetPending(prev.IsPending())
		}

		iBMap[iB] = currentIB

		if first {
			line.AddChildForLayout(c, currentIB)
			first = false
		} else {
			previousIB.AddInlineChildWithCallUnmarkPending(c, currentIB, false)
		}
		previousIB = currentIB
	}

	return currentIB
}
