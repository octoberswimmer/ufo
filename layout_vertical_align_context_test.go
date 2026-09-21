// Tests VerticalAlignContext (flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/VerticalAlignContext.java).
// Flying Saucer has no JUnit test for the class.

package ufo

import "testing"

func TestVerticalAlignContextTracksExtremesOfPushedMeasurements(t *testing.T) {
	initial := NewInlineBoxMeasurements(12, 2, 16, 0, 18, 0, 0)
	va := NewVerticalAlignContextInlineBoxMeasurements(initial)
	if va.GetParentMeasurements() != initial {
		t.Error("the initial measurements are not the parent measurements")
	}
	if va.GetParent() != nil {
		t.Error("a root context has a parent")
	}

	first := NewInlineBoxMeasurements(12, 2, 16, -3, 20, -1, 17)
	va.PushMeasurements(first)
	second := NewInlineBoxMeasurements(10, 4, 14, 1, 25, 3, 12)
	va.PushMeasurements(second)
	if va.GetParentMeasurements() != second {
		t.Error("GetParentMeasurements is not the last pushed measurements")
	}
	va.PopMeasurements()
	if va.GetParentMeasurements() != first {
		t.Error("PopMeasurements did not remove the last pushed measurements")
	}

	if va.GetInlineTop() != -3 || va.GetInlineBottom() != 25 {
		t.Errorf("inline top %d bottom %d, want -3 and 25", va.GetInlineTop(), va.GetInlineBottom())
	}
	if va.GetPaintingTop() != -1 || va.GetPaintingBottom() != 17 {
		t.Errorf("painting top %d bottom %d, want -1 and 17", va.GetPaintingTop(), va.GetPaintingBottom())
	}
	if va.GetLineBoxHeight() != 28 {
		t.Errorf("GetLineBoxHeight() = %d, want 28", va.GetLineBoxHeight())
	}
}

// The first update sets a tracked value even when it is on the far side of
// the zero value the field starts with.
func TestVerticalAlignContextFirstUpdateSetsValue(t *testing.T) {
	va := NewVerticalAlignContextVerticalAlignContext(nil)
	va.UpdateInlineTop(7)
	va.UpdateInlineBottom(-7)
	va.UpdatePaintingTop(5)
	va.UpdatePaintingBottom(-5)
	if va.GetInlineTop() != 7 || va.GetInlineBottom() != -7 || va.GetPaintingTop() != 5 || va.GetPaintingBottom() != -5 {
		t.Errorf("got %d %d %d %d", va.GetInlineTop(), va.GetInlineBottom(), va.GetPaintingTop(), va.GetPaintingBottom())
	}
	va.UpdateInlineTop(9)
	va.UpdateInlineBottom(-9)
	if va.GetInlineTop() != 7 || va.GetInlineBottom() != -7 {
		t.Error("a later update moved a tracked value in the wrong direction")
	}
}
