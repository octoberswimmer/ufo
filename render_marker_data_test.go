package ufo

import "testing"

// Flying Saucer has no JUnit tests for MarkerData, TextDecoration or
// StrutMetrics; these check the behaviour the Java source states.
func TestMarkerDataLayoutWidth(t *testing.T) {
	strut := NewStrutMetrics(8, 10, 2)
	if got := NewMarkerData(strut, nil, nil, nil).GetLayoutWidth(); got != 0 {
		t.Errorf("no marker: GetLayoutWidth() = %d, want 0", got)
	}
	if got := NewMarkerData(strut, NewMarkerDataImageMarker(nil, 5), nil, nil).GetLayoutWidth(); got != 5 {
		t.Errorf("image marker: GetLayoutWidth() = %d, want 5", got)
	}
	if got := NewMarkerData(strut, nil, NewMarkerDataGlyphMarker(3, 6), nil).GetLayoutWidth(); got != 6 {
		t.Errorf("glyph marker: GetLayoutWidth() = %d, want 6", got)
	}
	text := NewMarkerDataTextMarker("1.", 7)
	data := NewMarkerData(strut, NewMarkerDataImageMarker(nil, 5), NewMarkerDataGlyphMarker(3, 6), text)
	if got := data.GetLayoutWidth(); got != 7 {
		t.Errorf("text marker: GetLayoutWidth() = %d, want 7", got)
	}
	if data.GetTextMarker().GetText() != "1." || data.GetGlyphMarker().GetDiameter() != 3 {
		t.Error("marker accessors returned other values than the constructors received")
	}
	if data.GetStructMetrics() != strut {
		t.Error("GetStructMetrics() returned another object")
	}
}

func TestMarkerDataReferenceLine(t *testing.T) {
	data := NewMarkerData(NewStrutMetrics(8, 10, 2), nil, nil, nil)
	first := &LineBox{}
	second := &LineBox{}
	data.SetReferenceLine(first)
	data.SetReferenceLine(second)

	data.RestorePreviousReferenceLine(first)
	if data.GetReferenceLine() != second {
		t.Error("RestorePreviousReferenceLine changed the line although another line was current")
	}
	data.RestorePreviousReferenceLine(second)
	if data.GetReferenceLine() != first {
		t.Error("RestorePreviousReferenceLine did not restore the previous line")
	}
}

func TestTextDecorationThickness(t *testing.T) {
	decoration := NewTextDecoration(IdentValueUnderline)
	decoration.SetThickness(0)
	if decoration.GetThickness() != 1 {
		t.Errorf("SetThickness(0): GetThickness() = %d, want 1", decoration.GetThickness())
	}
	decoration.SetThickness(4)
	decoration.SetOffset(-2)
	if decoration.GetThickness() != 4 || decoration.GetOffset() != -2 || decoration.GetIdentValue() != IdentValueUnderline {
		t.Error("accessors returned other values than were set")
	}
}

func TestStrutMetrics(t *testing.T) {
	strut := NewStrutMetrics(8.5, 10, 2.5)
	strut.SetBaseline(12)
	if strut.GetAscent() != 8.5 || strut.GetBaseline() != 12 || strut.GetDescent() != 2.5 {
		t.Errorf("got %v, %v, %v", strut.GetAscent(), strut.GetBaseline(), strut.GetDescent())
	}
}
