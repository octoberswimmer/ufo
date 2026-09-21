// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/FSLinearGradient.java.
// Flying Saucer has no JUnit test for the class.

package ufo

import "testing"

func TestFSLinearGradient_looksLikeALength(t *testing.T) {
	cases := map[string]bool{
		"0":      true,
		"10px":   true,
		"-1.5em": true,
		"50%":    true,
		"10":     false,
		"px":     false,
		"10pxx":  false,
		"top":    false,
	}
	for val, want := range cases {
		if got := FSLinearGradientLooksLikeALength(val); got != want {
			t.Errorf("FSLinearGradientLooksLikeALength(%q) = %v, want %v", val, got, want)
		}
	}
	if !FSLinearGradientLooksLikeABGPosition("center") || FSLinearGradientLooksLikeABGPosition("middle") {
		t.Errorf("FSLinearGradientLooksLikeABGPosition: center must match, middle must not")
	}
}

func fsLinearGradientTestStops(t *testing.T, gradient *FSLinearGradient, want ...float32) {
	t.Helper()
	stops := gradient.GetStopPoints()
	if len(stops) != len(want) {
		t.Fatalf("%s has %d stop points, want %d", gradient, len(stops), len(want))
	}
	for i, stop := range stops {
		if stop.GetLength() == nil || *stop.GetLength() != want[i] {
			t.Errorf("stop %d of %s, want %v", i, gradient, want[i])
		}
	}
}

func TestFSLinearGradient_toSide(t *testing.T) {
	ctx := &calculatedStyleTestContext{}
	style := NewEmptyStyle()
	function := NewFSFunction("linear-gradient", []*PropertyValue{
		NewPropertyValueString(CSSPrimitiveValueCssIdent, "to", "to"),
		NewPropertyValueString(CSSPrimitiveValueCssIdent, "right", "right"),
		NewPropertyValueFSColor(FSRGBColorRed),
		NewPropertyValueFSColor(FSRGBColorGreen),
		NewPropertyValueFSColor(FSRGBColorBlue),
	})

	gradient := NewFSLinearGradient(function, style, 200, 100, ctx)

	if gradient.GetStartX() != 0 || gradient.GetStartY() != 0 || gradient.GetEndX() != 200 || gradient.GetEndY() != 0 {
		t.Errorf("end points of %s, want [0, 0] to [200, 0]", gradient)
	}
	// the middle stop has no length and is placed half way
	fsLinearGradientTestStops(t, gradient, 0, 100, 200)
	if gradient.GetStopPoints()[1].GetColor() != FSColor(FSRGBColorGreen) {
		t.Errorf("middle stop color = %v, want green", gradient.GetStopPoints()[1].GetColor())
	}
}

func TestFSLinearGradient_angle(t *testing.T) {
	ctx := &calculatedStyleTestContext{}
	style := NewEmptyStyle()
	function := NewFSFunction("linear-gradient", []*PropertyValue{
		NewPropertyValueFloat(CSSPrimitiveValueCssDeg, 45, "45deg"),
		NewPropertyValueFSColor(FSRGBColorRed),
		NewPropertyValueFloat(CSSPrimitiveValueCssPercentage, 25, "25%"),
		NewPropertyValueFSColor(FSRGBColorBlue),
	})

	gradient := NewFSLinearGradient(function, style, 100, 100, ctx)

	if gradient.GetStartX() != 0 || gradient.GetStartY() != 100 || gradient.GetEndX() != 100 || gradient.GetEndY() != 0 {
		t.Errorf("end points of %s, want [0, 100] to [100, 0]", gradient)
	}
	fsLinearGradientTestStops(t, gradient, 25, 100)
}

func TestFSLinearGradient_malformedInputGivesTheTransparentGradient(t *testing.T) {
	ctx := &calculatedStyleTestContext{}
	style := NewEmptyStyle()
	functions := []*FSFunction{
		NewFSFunction("linear-gradient", []*PropertyValue{}),
		// no "to" keyword and no angle
		NewFSFunction("linear-gradient", []*PropertyValue{
			NewPropertyValueFSColor(FSRGBColorRed),
			NewPropertyValueFSColor(FSRGBColorBlue),
		}),
		// one color stop
		NewFSFunction("linear-gradient", []*PropertyValue{
			NewPropertyValueFloat(CSSPrimitiveValueCssDeg, 90, "90deg"),
			NewPropertyValueFSColor(FSRGBColorRed),
		}),
	}
	for _, function := range functions {
		gradient := NewFSLinearGradient(function, style, 100, 100, ctx)
		if gradient.GetStartX() != 0 || gradient.GetStartY() != 0 || gradient.GetEndX() != 1 || gradient.GetEndY() != 0 {
			t.Errorf("end points of %s, want [0, 0] to [1, 0]", gradient)
		}
		fsLinearGradientTestStops(t, gradient, 0, 1)
	}
}
