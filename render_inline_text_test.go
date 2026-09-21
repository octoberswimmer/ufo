package ufo

import (
	"math"
	"strings"
	"testing"
)

// Flying Saucer has no JUnit tests for InlineText. The expected values below
// were produced by running the Java class over the same inputs.
func TestInlineTextAgainstJava(t *testing.T) {
	cases := []struct {
		text          string
		toString      string
		empty         bool
		end           int
		spaces        int
		nonSpaces     int
		adjustBits    uint32
		exportText    string
		trimmedLeader bool
	}{
		{"hello world", "InlineText: (hello world)", false, 11, 1, 10, 1083598439, "hello world", false},
		{"a b c　d", "InlineText: (a b c　d)", false, len("a b c　d"), 3, 4, 1086953882, "a b c　d", false},
		{"line\n", "InlineText: (L) (line)", false, 4, 0, 4, 1067030938, "line", false},
		{" lead and trail ", "InlineText: ( lead and trail )", false, 16, 4, 12, 1093035624, "  lead and trail ", true},
		{"", "InlineText: ()", true, 0, 0, 0, 0, "", false},
		{"x", "InlineText: (x)", false, 1, 0, 1, 1050253722, "x", false},
		{"café au lait\n", "InlineText: (L) (café au lait)", false, len("café au lait"), 2, 10, 1087163598, "café au lait", false},
		{"tab\there\nmid", "InlineText: (tab\there\nmid)", false, 12, 0, 12, 1080452709, "tab\theremid", false},
	}
	for _, tc := range cases {
		it := NewInlineText(tc.text, nil, 0, len(tc.text), 7)
		if got := it.ToString(); got != tc.toString {
			t.Errorf("%q: ToString() = %q, want %q", tc.text, got, tc.toString)
		}
		if got := it.String(); got != tc.toString {
			t.Errorf("%q: String() = %q, want %q", tc.text, got, tc.toString)
		}
		if got := it.IsEmpty(); got != tc.empty {
			t.Errorf("%q: IsEmpty() = %v, want %v", tc.text, got, tc.empty)
		}
		if it.GetStart() != 0 || it.GetEnd() != tc.end {
			t.Errorf("%q: start, end = %d, %d, want 0, %d", tc.text, it.GetStart(), it.GetEnd(), tc.end)
		}
		counts := NewCharCounts()
		it.CountJustifiableChars(counts)
		if counts.GetSpaceCount() != tc.spaces || counts.GetNonSpaceCount() != tc.nonSpaces {
			t.Errorf("%q: counts = %d spaces, %d non-spaces, want %d, %d", tc.text,
				counts.GetSpaceCount(), counts.GetNonSpaceCount(), tc.spaces, tc.nonSpaces)
		}
		adjust := it.CalcTotalAdjustment(NewJustificationInfo(0.3, 1.7))
		if got := math.Float32bits(adjust); got != tc.adjustBits {
			t.Errorf("%q: CalcTotalAdjustment bits = %d, want %d", tc.text, got, tc.adjustBits)
		}
		it.SetTrimmedLeadingSpace(tc.trimmedLeader)
		if got := it.GetTextExportText(); got != tc.exportText {
			t.Errorf("%q: GetTextExportText() = %q, want %q", tc.text, got, tc.exportText)
		}
		if it.GetWidth() != 7 {
			t.Errorf("%q: GetWidth() = %d, want 7", tc.text, it.GetWidth())
		}
	}
}

func TestInlineTextCountJustifiableCharsAccumulates(t *testing.T) {
	counts := NewCharCounts()
	NewInlineText("a b", nil, 0, 3, 0).CountJustifiableChars(counts)
	NewInlineText("c d e", nil, 0, 5, 0).CountJustifiableChars(counts)
	if counts.GetSpaceCount() != 3 || counts.GetNonSpaceCount() != 5 {
		t.Errorf("counts = %d, %d, want 3, 5", counts.GetSpaceCount(), counts.GetNonSpaceCount())
	}
}

func TestInlineTextSubstringUsesByteOffsets(t *testing.T) {
	master := "héllo wörld"
	start := strings.Index(master, "w")
	it := NewInlineText(master, nil, start, len(master), 0)
	if got := it.GetSubstring(); got != "wörld" {
		t.Errorf("GetSubstring() = %q, want %q", got, "wörld")
	}
	if it.GetMasterText() != master {
		t.Errorf("GetMasterText() = %q", it.GetMasterText())
	}
}

func TestInlineTextSelection(t *testing.T) {
	it := NewInlineText("héllo", nil, 0, len("héllo"), 0)
	if it.IsSelected() {
		t.Error("a new InlineText is selected")
	}
	it.SelectAll()
	if it.GetSelectionStart() != 0 || it.GetSelectionEnd() != 5 {
		t.Errorf("selection = %d..%d, want 0..5", it.GetSelectionStart(), it.GetSelectionEnd())
	}
	if got := it.GetSelection(); got != "héllo" {
		t.Errorf("GetSelection() = %q", got)
	}
	it.SetSelectionStart(1)
	it.SetSelectionEnd(3)
	if got := it.GetSelection(); got != "él" {
		t.Errorf("GetSelection() = %q, want %q", got, "él")
	}
	if !it.ClearSelection() {
		t.Error("ClearSelection() = false with a selection set")
	}
	if it.ClearSelection() {
		t.Error("ClearSelection() = true without a selection")
	}
}

func TestInlineTextConstructorRejectsBadOffsets(t *testing.T) {
	expectPanic := func(name string, want string, f func()) {
		t.Helper()
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("%s: no panic", name)
				return
			}
			e, ok := r.(*XRRuntimeException)
			if !ok {
				t.Errorf("%s: panic value %T, want *XRRuntimeException", name, r)
				return
			}
			if !strings.Contains(e.Error(), want) {
				t.Errorf("%s: message %q does not contain %q", name, e.Error(), want)
			}
		}()
		f()
	}
	expectPanic("end before start", "end is less than start (2 < 3) for element InlineText: ()", func() {
		NewInlineText("abcdef", nil, 3, 2, 0)
	})
	expectPanic("negative", "Trying to set negative index to inline box (start: -1, end: 2)", func() {
		NewInlineText("abcdef", nil, -1, 2, 0)
	})
}
