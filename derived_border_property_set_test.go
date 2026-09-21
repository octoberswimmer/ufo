// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/style/derived/BorderPropertySetTest.java

package ufo

import "testing"

func borderPropertySetTestAssertSides(t *testing.T, borders RectPropertySetI, top float32, right float32, bottom float32, left float32) {
	t.Helper()
	if borders.Top() != top {
		t.Errorf("Top() = %v, want %v", borders.Top(), top)
	}
	if borders.Right() != right {
		t.Errorf("Right() = %v, want %v", borders.Right(), right)
	}
	if borders.Bottom() != bottom {
		t.Errorf("Bottom() = %v, want %v", borders.Bottom(), bottom)
	}
	if borders.Left() != left {
		t.Errorf("Left() = %v, want %v", borders.Left(), left)
	}
}

func TestBorderPropertySet_constructor(t *testing.T) {
	borders := NewBorderPropertySetWithTopRightBottomLeft(
		NewCollapsedBorderValue(nil, 1, FSRGBColorBlue, 0),
		NewCollapsedBorderValue(nil, 2, FSRGBColorBlue, 0),
		NewCollapsedBorderValue(nil, 3, FSRGBColorBlue, 0),
		NewCollapsedBorderValue(nil, 4, FSRGBColorBlue, 0),
	)

	borderPropertySetTestAssertSides(t, borders, 1.0, 2.0, 3.0, 4.0)
}

func TestBorderPropertySet_resetNegativeValues(t *testing.T) {
	negativeBorders := NewBorderPropertySetWithTopRightBottomLeft(
		NewCollapsedBorderValue(nil, -1, FSRGBColorBlue, 0),
		NewCollapsedBorderValue(nil, -2, FSRGBColorBlue, 0),
		NewCollapsedBorderValue(nil, 3, FSRGBColorBlue, 0),
		NewCollapsedBorderValue(nil, 4, FSRGBColorBlue, 0),
	)

	borders, ok := negativeBorders.ResetNegativeValues().(*BorderPropertySet)
	if !ok {
		t.Fatalf("ResetNegativeValues() did not return a *BorderPropertySet")
	}

	borderPropertySetTestAssertSides(t, borders, 0.0, 0.0, 3.0, 4.0)
}
