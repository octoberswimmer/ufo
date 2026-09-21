// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/LineBreakContext.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// LineBreakContext is a bean which serves as a way for the layout code to pass
// information to the line breaking code and for the line breaking code to pass
// instructions back to the layout code.
//
// start, end and savedEnd are byte offsets into master.
type LineBreakContext struct {
	master       string
	start        int
	end          int
	savedEnd     int
	unbreakable  bool
	needsNewLine bool
	width        int
	endsOnNL     bool
	textNode     *dom.Text
}

func NewLineBreakContext(master string, textNode *dom.Text) *LineBreakContext {
	return &LineBreakContext{master: master, textNode: textNode}
}

func (l *LineBreakContext) GetLast() int {
	return len(l.master)
}

func (l *LineBreakContext) Reset() {
	l.width = 0
	l.unbreakable = false
	l.needsNewLine = false
}

func (l *LineBreakContext) GetEnd() int {
	return l.end
}

func (l *LineBreakContext) SetEnd(end int) {
	l.end = end
}

func (l *LineBreakContext) GetMaster() string {
	return l.master
}

func (l *LineBreakContext) SetMaster(master string) {
	l.master = master
}

func (l *LineBreakContext) GetStart() int {
	return l.start
}

func (l *LineBreakContext) SetStart(start int) {
	l.start = start
}

func (l *LineBreakContext) GetStartSubstring() string {
	return l.master[l.start:]
}

func (l *LineBreakContext) GetCalculatedSubstring() string {
	// mimic the calculation in InlineText.setSubstring to strip newlines for our width calculations
	// the original text width calculation in InlineBox.calcMaxWidthFromLineLength() excludes the newline character
	// so if we include them here we get spurious newlines
	// apparently newlines do take up some width in most fonts
	if l.end > 0 && rune(l.master[l.end-1]) == WhitespaceStripperEolc {
		return l.master[l.start : l.end-1]
	}
	return l.master[l.start:l.end]
}

func (l *LineBreakContext) IsUnbreakable() bool {
	return l.unbreakable
}

func (l *LineBreakContext) SetUnbreakable(unbreakable bool) {
	l.unbreakable = unbreakable
}

func (l *LineBreakContext) IsNeedsNewLine() bool {
	return l.needsNewLine
}

func (l *LineBreakContext) SetNeedsNewLine(needsLineBreak bool) {
	l.needsNewLine = needsLineBreak
}

func (l *LineBreakContext) GetWidth() int {
	return l.width
}

func (l *LineBreakContext) SetWidth(width int) {
	l.width = width
}

func (l *LineBreakContext) IsFinished() bool {
	return l.end == len(l.GetMaster())
}

func (l *LineBreakContext) ResetEnd() {
	l.end = l.savedEnd
}

func (l *LineBreakContext) SaveEnd() {
	l.savedEnd = l.end
}

func (l *LineBreakContext) IsEndsOnNL() bool {
	return l.endsOnNL
}

func (l *LineBreakContext) SetEndsOnNL(b bool) {
	l.endsOnNL = b
}

func (l *LineBreakContext) GetTextNode() *dom.Text {
	return l.textNode
}
