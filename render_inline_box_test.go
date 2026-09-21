package ufo

import "testing"

// Flying Saucer has no JUnit tests for InlineBox. The expected strings were
// produced by the Java class over the same inputs.
func TestInlineBoxToStringAgainstJava(t *testing.T) {
	forty := "0123456789012345678901234567890123456789"
	cases := []struct {
		text       string
		withPseudo string
		plain      string
	}{
		{"short", "InlineBox: (anonymous) :before (S) (short) ", "InlineBox: (anonymous) (E) (short) "},
		{"with\nnewline", "InlineBox: (anonymous) :before (S) (with newline) ", "InlineBox: (anonymous) (E) (with newline) "},
		{forty, "InlineBox: (anonymous) :before (S) (" + forty + "...) ", "InlineBox: (anonymous) (E) (" + forty + "...) "},
		{forty + "X", "InlineBox: (anonymous) :before (S) (" + forty + "...) ", "InlineBox: (anonymous) (E) (" + forty + "...) "},
		{forty[:39], "InlineBox: (anonymous) :before (S) (" + forty[:39] + ") ", "InlineBox: (anonymous) (E) (" + forty[:39] + ") "},
	}
	for _, tc := range cases {
		iB := NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClassStyle(tc.text, nil, nil, nil, nil, "before", NewEmptyStyle())
		iB.SetStartsHere(true)
		if got := iB.ToString(); got != tc.withPseudo {
			t.Errorf("%q: ToString() = %q, want %q", tc.text, got, tc.withPseudo)
		}

		iB2 := NewInlineBox(tc.text, nil)
		iB2.SetStyle(NewEmptyStyle())
		iB2.SetEndsHere(true)
		if got := iB2.String(); got != tc.plain {
			t.Errorf("%q: String() = %q, want %q", tc.text, got, tc.plain)
		}
	}
}

func TestInlineBoxTextAccessors(t *testing.T) {
	iB := NewInlineBox(" text ", nil)
	if iB.GetText() != " text " {
		t.Errorf("GetText() = %q", iB.GetText())
	}
	if got := iB.getTextWithTrimLeadingSpace(true); got != "text " {
		t.Errorf("getTextWithTrimLeadingSpace(true) = %q, want %q", got, "text ")
	}
	if got := iB.getTextWithTrimLeadingSpace(false); got != " text " {
		t.Errorf("getTextWithTrimLeadingSpace(false) = %q, want %q", got, " text ")
	}
	if iB.IsDynamicFunction() {
		t.Error("IsDynamicFunction() = true without a content function")
	}
	iB.SetText("other")
	iB.TruncateText()
	if iB.GetText() != "" {
		t.Errorf("GetText() after TruncateText() = %q", iB.GetText())
	}
	var _ Styleable = iB
}

func TestInlineBoxJavaTrimAndIndexFrom(t *testing.T) {
	if got := inlineBoxJavaTrim(" \t\n a b \r\n"); got != "a b" {
		t.Errorf("inlineBoxJavaTrim = %q, want %q", got, "a b")
	}
	if got := inlineBoxJavaTrim(" a "); got != " a " {
		t.Errorf("inlineBoxJavaTrim removed U+00A0: %q", got)
	}
	if got := inlineBoxIndexFrom("a\nb\nc", "\n", 2); got != 3 {
		t.Errorf("inlineBoxIndexFrom = %d, want 3", got)
	}
	if got := inlineBoxIndexFrom("a\nb", "\n", 2); got != -1 {
		t.Errorf("inlineBoxIndexFrom = %d, want -1", got)
	}
}
