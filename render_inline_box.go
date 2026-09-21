// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/InlineBox.java

package ufo

import (
	"strings"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo/dom"
)

// InlineBox represents a portion of an inline element. If an inline element
// does not contain any nested elements, then a single InlineBox
// object will contain the content for the entire element. Otherwise, multiple
// InlineBox objects will be created corresponding to each
// discrete chunk of text appearing in the element. It is not rendered directly
// (and hence does not extend from Box), but does play an important
// role in layout (for example, when calculating min/max widths). Note that it
// does not contain children. Inline content is stored as a flat list in the
// layout tree. However, InlineBox does contain enough
// information to reconstruct the original element nesting and this is, in fact,
// done during inline layout.
//
// See InlineLayoutBox.
type InlineBox struct {
	// element may be nil.
	element *dom.Element

	originalText        string
	text                string
	removableWhitespace bool
	startsHere          bool
	endsHere            bool

	// style may be nil.
	style CalculatedStyleI

	// contentFunction may be nil.
	contentFunction ContentFunction
	// function may be nil.
	function *FSFunction

	minMaxCalculated bool
	maxWidth         int
	minWidth         int

	firstLineWidth int

	// pseudoElementOrClass is "" where Java has null.
	pseudoElementOrClass string

	// textNode may be nil.
	textNode *dom.Text
}

func NewInlineBox(text string, textNode *dom.Text) *InlineBox {
	return NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClass(text, textNode, nil, nil, nil, "")
}

func NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClass(text string, textNode *dom.Text,
	contentFunction ContentFunction, function *FSFunction,
	element *dom.Element, pseudoElementOrClass string) *InlineBox {
	return NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClassStyle(text, textNode, contentFunction, function, element, pseudoElementOrClass, nil)
}

func NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClassStyle(text string, textNode *dom.Text,
	contentFunction ContentFunction, function *FSFunction,
	element *dom.Element, pseudoElementOrClass string,
	style CalculatedStyleI) *InlineBox {
	return &InlineBox{
		text:                 text,
		originalText:         text,
		textNode:             textNode,
		contentFunction:      contentFunction,
		function:             function,
		element:              element,
		pseudoElementOrClass: pseudoElementOrClass,
		style:                style,
	}
}

func (b *InlineBox) GetText() string {
	return b.text
}

func (b *InlineBox) SetText(text string) {
	b.text = text
	b.originalText = text
}

func (b *InlineBox) ApplyTextTransform() {
	b.text = b.originalText
	b.text = LayoutTextUtilTransformText(b.text, b.GetStyle())
}

func (b *InlineBox) IsRemovableWhitespace() bool {
	return b.removableWhitespace
}

func (b *InlineBox) SetRemovableWhitespace(removableWhitespace bool) {
	b.removableWhitespace = removableWhitespace
}

func (b *InlineBox) IsEndsHere() bool {
	return b.endsHere
}

func (b *InlineBox) SetEndsHere(endsHere bool) {
	b.endsHere = endsHere
}

func (b *InlineBox) IsStartsHere() bool {
	return b.startsHere
}

func (b *InlineBox) SetStartsHere(startsHere bool) {
	b.startsHere = startsHere
}

// GetStyle may return nil.
func (b *InlineBox) GetStyle() CalculatedStyleI {
	return b.style
}

func (b *InlineBox) SetStyle(style CalculatedStyleI) {
	b.style = style
}

// GetElement may return nil.
func (b *InlineBox) GetElement() *dom.Element {
	return b.element
}

func (b *InlineBox) SetElement(element *dom.Element) {
	b.element = element
}

// GetContentFunction may return nil.
func (b *InlineBox) GetContentFunction() ContentFunction {
	return b.contentFunction
}

func (b *InlineBox) IsDynamicFunction() bool {
	return b.contentFunction != nil
}

func (b *InlineBox) getTextWidth(c *LayoutContext, s string) int {
	return LayoutTextUtilTextWidth(c, b.GetStyle(), c.GetFont(b.GetStyle().GetFont(c)), s)
}

func (b *InlineBox) getMaxCharWidth(c *LayoutContext, s string) int {
	result := 0
	for i := 0; i < len(s); {
		_, size := utf8.DecodeRuneInString(s[i:])
		width := b.getTextWidth(c, s[i:i+size])
		if width > result {
			result = width
		}
		i += size
	}
	return result
}

func (b *InlineBox) calcMaxWidthFromLineLength(c *LayoutContext, cbWidth int, trim bool) {
	last := 0

	for {
		current := inlineBoxIndexFrom(b.text, WhitespaceStripperEol, last)
		if current == -1 {
			break
		}
		target := b.text[last:current]
		if trim {
			target = inlineBoxJavaTrim(target)
		}
		length := b.getTextWidth(c, target)
		if last == 0 {
			length += b.GetStyle().GetMarginBorderPadding(c, cbWidth, CalculatedStyleEdgeLeft)
		}
		if length > b.maxWidth {
			b.maxWidth = length
		}
		if last == 0 {
			b.firstLineWidth = length
		}
		last = current + 1
	}

	target := b.text[last:]
	if trim {
		target = inlineBoxJavaTrim(target)
	}
	length := b.getTextWidth(c, target)
	length += b.GetStyle().GetMarginBorderPadding(c, cbWidth, CalculatedStyleEdgeRight)
	if length > b.maxWidth {
		b.maxWidth = length
	}
	if last == 0 {
		b.firstLineWidth = length
	}
}

// inlineBoxIndexFrom is String.indexOf(str, fromIndex).
func inlineBoxIndexFrom(s string, substr string, from int) int {
	if from > len(s) {
		return -1
	}
	i := strings.Index(s[from:], substr)
	if i == -1 {
		return -1
	}
	return from + i
}

// inlineBoxJavaTrim is String.trim(): it removes leading and trailing
// characters whose code is at most U+0020.
func inlineBoxJavaTrim(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] <= ' ' {
		start++
	}
	for start < end && s[end-1] <= ' ' {
		end--
	}
	return s[start:end]
}

func (b *InlineBox) GetSpaceWidth(c *LayoutContext) int {
	return LayoutTextUtilTextWidth(c, b.GetStyle(), b.GetStyle().GetFSFont(c), WhitespaceStripperSpace)
}

func (b *InlineBox) GetTrailingSpaceWidth(c *LayoutContext) int {
	if len(b.text) != 0 && b.text[len(b.text)-1] == ' ' {
		return b.GetSpaceWidth(c)
	} else {
		return 0
	}
}

func (b *InlineBox) calcMinWidthFromWordLength(
	c *LayoutContext, cbWidth int, trimLeadingSpace bool, includeWS bool) int {
	spaceWidth := b.GetSpaceWidth(c)

	last := 0
	maxWidth := 0
	spaceCount := 0

	// Java never sets haveFirstWord to true, so firstWord ends up holding the
	// last word's width as well.
	haveFirstWord := false
	firstWord := 0
	lastWord := 0

	text := b.getTextWithTrimLeadingSpace(trimLeadingSpace)

	breakIterator := BreakerGetBreakPointsProviderElement(text, c, b.GetElement(), b.GetStyle())

	// Breaker should be used
	for {
		current := breakIterator.Next().GetPosition()
		if current == BreakIteratorDone {
			break
		}
		currentWord := text[last:current]
		wordWidth := b.getTextWidth(c, currentWord)
		var minWordWidth int
		if b.GetStyle().GetWordWrap() == IdentValueBreakWord {
			minWordWidth = b.getMaxCharWidth(c, currentWord)
		} else {
			minWordWidth = wordWidth
		}

		if spaceCount > 0 {
			if includeWS {
				for i := 0; i < spaceCount; i++ {
					wordWidth += spaceWidth
					minWordWidth += spaceWidth
				}
			} else {
				maxWidth += spaceWidth
			}
			spaceCount = 0
		}
		if minWordWidth > 0 {
			if !haveFirstWord {
				firstWord = minWordWidth
			}
			lastWord = minWordWidth
		}

		b.adjustMinWidth(minWordWidth)
		maxWidth += wordWidth

		last = current
		for i := current; i < len(text); i++ {
			if text[i] == ' ' {
				spaceCount++
				last++
			} else {
				break
			}
		}
	}

	currentWord := text[last:]
	wordWidth := b.getTextWidth(c, currentWord)
	var minWordWidth int
	if b.GetStyle().GetWordWrap() == IdentValueBreakWord {
		minWordWidth = b.getMaxCharWidth(c, currentWord)
	} else {
		minWordWidth = wordWidth
	}
	if spaceCount > 0 {
		if includeWS {
			for i := 0; i < spaceCount; i++ {
				wordWidth += spaceWidth
				minWordWidth += spaceWidth
			}
		} else {
			maxWidth += spaceWidth
		}
	}
	if minWordWidth > 0 {
		if !haveFirstWord {
			firstWord = minWordWidth
		}
		lastWord = minWordWidth
	}
	b.adjustMinWidth(minWordWidth)
	maxWidth += wordWidth

	if b.IsStartsHere() {
		leftMBP := b.GetStyle().GetMarginBorderPadding(c, cbWidth, CalculatedStyleEdgeLeft)
		b.adjustMinWidth(firstWord + leftMBP)
		maxWidth += leftMBP
	}

	if b.IsEndsHere() {
		rightMBP := b.GetStyle().GetMarginBorderPadding(c, cbWidth, CalculatedStyleEdgeRight)
		b.adjustMinWidth(lastWord + rightMBP)
		maxWidth += rightMBP
	}

	return maxWidth
}

func (b *InlineBox) adjustMinWidth(minWordWidth int) {
	if minWordWidth > b.minWidth {
		b.minWidth = minWordWidth
	}
}

// getTextWithTrimLeadingSpace is the private getText(boolean trimLeadingSpace).
func (b *InlineBox) getTextWithTrimLeadingSpace(trimLeadingSpace bool) string {
	if !trimLeadingSpace {
		return b.GetText()
	} else {
		if len(b.text) != 0 && b.text[0] == ' ' {
			return b.text[1:]
		} else {
			return b.text
		}
	}
}

func (b *InlineBox) getInlineMBP(c *LayoutContext, cbWidth int) int {
	return b.GetStyle().GetMarginBorderPadding(c, cbWidth, CalculatedStyleEdgeLeft) +
		b.GetStyle().GetMarginBorderPadding(c, cbWidth, CalculatedStyleEdgeRight)
}

func (b *InlineBox) CalcMinMaxWidth(c *LayoutContext, cbWidth int, trimLeadingSpace bool) {
	if !b.minMaxCalculated {
		whitespace := b.GetStyle().GetWhitespace()
		if whitespace == IdentValueNowrap {
			b.maxWidth = b.getInlineMBP(c, cbWidth) + b.getTextWidth(c, b.getTextWithTrimLeadingSpace(trimLeadingSpace))
			b.minWidth = b.maxWidth
		} else if whitespace == IdentValuePre {
			b.calcMaxWidthFromLineLength(c, cbWidth, false)
			b.minWidth = b.maxWidth
		} else if whitespace == IdentValuePreWrap {
			b.calcMinWidthFromWordLength(c, cbWidth, false, true)
			b.calcMaxWidthFromLineLength(c, cbWidth, false)
		} else if whitespace == IdentValuePreLine {
			b.calcMinWidthFromWordLength(c, cbWidth, trimLeadingSpace, false)
			b.calcMaxWidthFromLineLength(c, cbWidth, true)
		} else /* if (whitespace == IdentValue.NORMAL) */ {
			b.maxWidth = b.calcMinWidthFromWordLength(c, cbWidth, trimLeadingSpace, false)
		}
		b.minWidth = min(b.maxWidth, b.minWidth)
		b.minMaxCalculated = true
	}
}

func (b *InlineBox) GetMaxWidth() int {
	return b.maxWidth
}

func (b *InlineBox) GetMinWidth() int {
	return b.minWidth
}

func (b *InlineBox) GetFirstLineWidth() int {
	return b.firstLineWidth
}

// GetPseudoElementOrClass returns "" where Java returns null.
func (b *InlineBox) GetPseudoElementOrClass() string {
	return b.pseudoElementOrClass
}

func (b *InlineBox) ToString() string {
	var result strings.Builder
	result.WriteString("InlineBox: ")
	if b.GetElement() != nil {
		result.WriteString("<")
		result.WriteString(b.GetElement().GetNodeName())
		result.WriteString("> ")
	} else {
		result.WriteString("(anonymous) ")
	}
	if b.GetPseudoElementOrClass() != "" {
		result.WriteByte(':')
		result.WriteString(b.GetPseudoElementOrClass())
		result.WriteByte(' ')
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

	UtilsAppendPositioningInfo(b.GetStyle(), &result)

	result.WriteString("(")
	result.WriteString(b.shortText())
	result.WriteString(") ")
	return result.String()
}

func (b *InlineBox) String() string {
	return b.ToString()
}

// shortText returns the first 40 characters of the text with line feeds
// replaced by spaces, followed by "..." when 40 characters were copied. Java
// counts UTF-16 chars; this counts runes, which is the same for BMP text.
func (b *InlineBox) shortText() string {
	var result strings.Builder
	count := 0
	for _, c := range b.text {
		if count >= 40 {
			break
		}
		if c == '\n' {
			result.WriteByte(' ')
		} else {
			result.WriteRune(c)
		}
		count++
	}
	if count == 40 {
		result.WriteString("...")
	}
	return result.String()
}

// GetFunction may return nil.
func (b *InlineBox) GetFunction() *FSFunction {
	return b.function
}

func (b *InlineBox) TruncateText() {
	b.text = ""
	b.originalText = ""
}

// GetTextNode may return nil.
func (b *InlineBox) GetTextNode() *dom.Text {
	return b.textNode
}
