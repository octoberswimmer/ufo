// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/InlineText.java

package ufo

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

// InlineText is a lightweight object which contains a chunk of text from an
// inline element. It will never extend across a line break nor will it extend
// across an element nested within its inline element.
//
// start and end are byte offsets into masterText. selectionStart and
// selectionEnd are glyph indexes (Java uses them as char indexes into the
// substring); they are treated as rune indexes into the substring here.
type InlineText struct {
	// parent may be nil.
	parent *InlineLayoutBox

	x int

	masterText string
	start      int
	end        int

	width int

	// functionData may be nil.
	functionData *FunctionData

	containedLF bool

	selectionStart int16
	selectionEnd   int16

	// glyphPositions is nil until ensureGlyphPositions has run.
	glyphPositions []float32

	trimmedLeadingSpace  bool
	trimmedTrailingSpace bool
	textNode             *dom.Text
}

func (t *InlineText) TrimTrailingSpace(c *LayoutContext) {
	if !t.IsEmpty() && t.masterText[t.end-1] == ' ' {
		t.end--
		t.SetWidth(LayoutTextUtilTextWidth(c, t.GetParent().GetStyle(),
			t.GetParent().GetStyle().GetFSFont(c),
			t.GetSubstring()))
		t.setTrimmedTrailingSpace()
	}
}

func (t *InlineText) IsEmpty() bool {
	return t.start == t.end && !t.containedLF
}

func (t *InlineText) GetSubstring() string {
	// Java also reports "No master text set for element" when the master text
	// is null. A Go string is never null, so that branch does not exist here.
	if t.start == -1 || t.end == -1 {
		panic(NewXRRuntimeException(fmt.Sprintf("negative index in InlineBox (start: %d, end: %d) for element %s", t.start, t.end, t.describeWithoutSubstring())))
	}
	if t.end < t.start {
		panic(NewXRRuntimeException(fmt.Sprintf("end is less than start (%d < %d) for element %s", t.end, t.start, t.describeWithoutSubstring())))
	}
	return t.masterText[t.start:t.end]
}

// describeWithoutSubstring is the text used for this object in the messages
// of GetSubstring. Java formats the object with toString() there, which calls
// getSubstring() again and ends in a StackOverflowError instead of the
// intended message; a Go stack overflow ends the process, so the messages use
// this description, which does not read the substring.
func (t *InlineText) describeWithoutSubstring() string {
	return "InlineText: " + t.flagsToString()
}

func NewInlineText(masterText string, textNode *dom.Text, start int, end int, width int) *InlineText {
	t := &InlineText{}
	t.masterText = masterText
	t.textNode = textNode
	t.width = width

	if end < start {
		UuPString(fmt.Sprintf("setting substring to: %d %d", start, end))
		panic(NewXRRuntimeException(fmt.Sprintf("end is less than start (%d < %d) for element %s", end, start, t.ToString())))
	} else if end < 0 || start < 0 {
		panic(NewXRRuntimeException(fmt.Sprintf("Trying to set negative index to inline box (start: %d, end: %d)", start, end)))
	}
	t.start = start
	t.end = end

	if t.end > 0 && rune(t.masterText[t.end-1]) == WhitespaceStripperEolc {
		t.containedLF = true
		t.end--
	}
	return t
}

func (t *InlineText) GetMasterText() string {
	return t.masterText
}

func (t *InlineText) GetX() int {
	return t.x
}

func (t *InlineText) SetX(x int) {
	t.x = x
}

func (t *InlineText) GetWidth() int {
	return t.width
}

func (t *InlineText) SetWidth(width int) {
	t.width = width
}

func (t *InlineText) Paint(c *RenderingContext) {
	c.GetOutputDevice().DrawText(c, t)
}

func (t *InlineText) PaintSelection(c *RenderingContext) {
	c.GetOutputDevice().DrawSelection(c, t)
}

// GetParent may return nil.
func (t *InlineText) GetParent() *InlineLayoutBox {
	return t.parent
}

func (t *InlineText) SetParent(parent *InlineLayoutBox) {
	t.parent = parent
}

func (t *InlineText) IsDynamicFunction() bool {
	return t.functionData != nil
}

func (t *InlineText) GetFunctionData() *FunctionData {
	return t.functionData
}

func (t *InlineText) SetFunctionData(functionData *FunctionData) {
	t.functionData = functionData
}

func (t *InlineText) UpdateDynamicValue(c *RenderingContext) {
	value := t.functionData.GetContentFunction().CalculateWithText(
		c, t.functionData.GetFunction(), t)
	t.start = 0
	t.end = len(value)
	t.masterText = value
	t.width = LayoutTextUtilTextWidth(c, t.GetParent().GetStyle(),
		t.GetParent().GetStyle().GetFSFont(c), value)
}

func (t *InlineText) flagsToString() string {
	var result strings.Builder
	if t.containedLF || t.IsDynamicFunction() {
		result.WriteString("(")
		if t.containedLF {
			result.WriteByte('L')
		}
		if t.IsDynamicFunction() {
			result.WriteByte('F')
		}
		result.WriteString(") ")
	}
	return result.String()
}

func (t *InlineText) ToString() string {
	var result strings.Builder
	result.WriteString("InlineText: ")
	result.WriteString(t.flagsToString())
	result.WriteByte('(')
	result.WriteString(t.GetSubstring())
	result.WriteByte(')')

	return result.String()
}

func (t *InlineText) String() string {
	return t.ToString()
}

func (t *InlineText) UpdateSelection(c *RenderingContext, selection *geom.Rectangle) bool {
	t.ensureGlyphPositions(c)
	positions := t.glyphPositions
	y := t.GetParent().GetAbsY()
	offset := t.GetParent().GetAbsX() + t.GetX()

	prevSelectionStart := t.selectionStart
	prevSelectionEnd := t.selectionEnd

	found := false
	t.selectionStart = 0
	t.selectionEnd = 0
	for i := 0; i < len(positions)-2; i += 2 {
		target := geom.NewRectangle(
			int(float32(offset)+(positions[i]+positions[i+2])/2),
			y,
			1,
			t.GetParent().GetHeight())
		if selection.Intersects(target) {
			if !found {
				found = true
				t.selectionStart = int16(i / 2)
				t.selectionEnd = int16(i/2 + 1)
			} else {
				t.selectionEnd++
			}
		}
	}

	return prevSelectionStart != t.selectionStart || prevSelectionEnd != t.selectionEnd
}

func (t *InlineText) ensureGlyphPositions(c *RenderingContext) {
	if t.glyphPositions == nil {
		glyphVector := c.GetTextRenderer().GetGlyphVector(
			c.GetOutputDevice(),
			t.GetParent().GetStyle().GetFSFont(c),
			t.GetSubstring())
		t.glyphPositions = c.GetTextRenderer().GetGlyphPositions(
			c.GetOutputDevice(),
			t.GetParent().GetStyle().GetFSFont(c),
			glyphVector)
	}
}

func (t *InlineText) ClearSelection() bool {
	result := t.selectionStart != 0 || t.selectionEnd != 0

	t.selectionStart = 0
	t.selectionEnd = 0

	return result
}

func (t *InlineText) IsSelected() bool {
	return t.selectionStart != t.selectionEnd
}

func (t *InlineText) GetSelectionEnd() int16 {
	return t.selectionEnd
}

func (t *InlineText) GetSelectionStart() int16 {
	return t.selectionStart
}

// GetSelection returns the selected part of the substring. The selection
// bounds are rune indexes; they are converted to byte offsets here. Bounds
// outside of the substring panic, as substring() throws in Java.
func (t *InlineText) GetSelection() string {
	s := t.GetSubstring()
	from := inlineTextByteOffset(s, int(t.selectionStart))
	to := inlineTextByteOffset(s, int(t.selectionEnd))
	return s[from:to]
}

// inlineTextByteOffset returns the byte offset in s of the rune with index
// runeIndex. runeIndex may equal the number of runes in s.
func inlineTextByteOffset(s string, runeIndex int) int {
	if runeIndex < 0 {
		panic(NewXRRuntimeException(fmt.Sprintf("String index out of range: %d", runeIndex)))
	}
	offset := 0
	for i := 0; i < runeIndex; i++ {
		if offset >= len(s) {
			panic(NewXRRuntimeException(fmt.Sprintf("String index out of range: %d", runeIndex)))
		}
		_, size := utf8.DecodeRuneInString(s[offset:])
		offset += size
	}
	return offset
}

func (t *InlineText) SelectAll() {
	t.selectionStart = 0
	t.selectionEnd = int16(utf8.RuneCountInString(t.GetSubstring()))
}

func (t *InlineText) GetTextExportText() string {
	var result strings.Builder
	if t.IsTrimmedLeadingSpace() {
		result.WriteByte(' ')
	}
	for _, c := range t.GetSubstring() {
		if c != '\n' {
			result.WriteRune(c)
		}
	}
	if t.isTrimmedTrailingSpace() {
		result.WriteByte(' ')
	}
	return result.String()
}

func (t *InlineText) IsTrimmedLeadingSpace() bool {
	return t.trimmedLeadingSpace
}

func (t *InlineText) SetTrimmedLeadingSpace(trimmedLeadingSpace bool) {
	t.trimmedLeadingSpace = trimmedLeadingSpace
}

func (t *InlineText) setTrimmedTrailingSpace() {
	t.trimmedTrailingSpace = true
}

func (t *InlineText) isTrimmedTrailingSpace() bool {
	return t.trimmedTrailingSpace
}

// CountJustifiableChars counts runes (see PORTING.md, text positions).
func (t *InlineText) CountJustifiableChars(counts *CharCounts) {
	s := t.GetSubstring()
	spaces := 0
	other := 0

	for _, c := range s {
		if c == ' ' || c == ' ' || c == '　' {
			spaces++
		} else {
			other++
		}
	}

	counts.SetSpaceCount(counts.GetSpaceCount() + spaces)
	counts.SetNonSpaceCount(counts.GetNonSpaceCount() + other)
}

func (t *InlineText) CalcTotalAdjustment(info *JustificationInfo) float32 {
	s := t.GetSubstring()

	var result float32 = 0.0
	for _, c := range s {
		if c == ' ' || c == ' ' || c == '　' {
			result += info.SpaceAdjust()
		} else {
			result += info.NonSpaceAdjust()
		}
	}

	return result
}

func (t *InlineText) GetStart() int {
	return t.start
}

func (t *InlineText) GetEnd() int {
	return t.end
}

func (t *InlineText) SetSelectionStart(s int16) {
	t.selectionStart = s
}

func (t *InlineText) SetSelectionEnd(s int16) {
	t.selectionEnd = s
}

func (t *InlineText) GetTextNode() *dom.Text {
	return t.textNode
}
