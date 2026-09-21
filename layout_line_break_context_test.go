// Tests LineBreakContext (flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/LineBreakContext.java).
// Flying Saucer has no JUnit test for the class.

package ufo

import "testing"

func TestLineBreakContextGetCalculatedSubstringStripsTrailingNewline(t *testing.T) {
	ctx := NewLineBreakContext("été one\ntwo", nil)
	ctx.SetEnd(len("été one\n"))
	if got := ctx.GetCalculatedSubstring(); got != "été one" {
		t.Errorf("got %q", got)
	}
	ctx.SetStart(ctx.GetEnd())
	ctx.SetEnd(ctx.GetLast())
	if got := ctx.GetCalculatedSubstring(); got != "two" {
		t.Errorf("got %q", got)
	}
	if got := ctx.GetStartSubstring(); got != "two" {
		t.Errorf("got %q", got)
	}
	if !ctx.IsFinished() {
		t.Error("IsFinished() = false at the end of the master text")
	}
}

func TestLineBreakContextSaveEndResetEndAndReset(t *testing.T) {
	ctx := NewLineBreakContext("abcdef", nil)
	ctx.SetEnd(2)
	ctx.SaveEnd()
	ctx.SetEnd(5)
	ctx.ResetEnd()
	if ctx.GetEnd() != 2 {
		t.Errorf("GetEnd() = %d after ResetEnd, want 2", ctx.GetEnd())
	}

	ctx.SetWidth(10)
	ctx.SetUnbreakable(true)
	ctx.SetNeedsNewLine(true)
	ctx.SetEndsOnNL(true)
	ctx.Reset()
	if ctx.GetWidth() != 0 || ctx.IsUnbreakable() || ctx.IsNeedsNewLine() {
		t.Error("Reset did not clear width, unbreakable and needsNewLine")
	}
	if !ctx.IsEndsOnNL() || ctx.GetEnd() != 2 {
		t.Error("Reset changed endsOnNL or end")
	}
}
