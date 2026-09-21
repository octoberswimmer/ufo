// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/property/AbstractPropertyBuilderTest.java

package ufo

import (
	"fmt"
	"testing"
)

// propertyBuilderTestCatch runs f and returns the *CSSParseException it
// panics with. It fails the test when f does not panic with one. The ported
// tests of the property builders share it.
func propertyBuilderTestCatch(t *testing.T, f func()) (result *CSSParseException) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected a CSSParseException, but nothing was thrown")
		}
		e, ok := r.(*CSSParseException)
		if !ok {
			t.Fatalf("expected a CSSParseException, but got %T: %v", r, r)
		}
		result = e
	}()
	f()
	return nil
}

func TestAbstractPropertyBuilder_lengthCssTypes(t *testing.T) {
	builder := &AbstractPropertyBuilder{}
	for _, typ := range []int16{CSSPrimitiveValueCssEms, CSSPrimitiveValueCssExs, CSSPrimitiveValueCssPx, CSSPrimitiveValueCssIn,
		CSSPrimitiveValueCssCm, CSSPrimitiveValueCssMm, CSSPrimitiveValueCssPt, CSSPrimitiveValueCssPc} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			if !builder.IsLength(NewPropertyValueFloat(typ, 123.45, "?")) {
				t.Errorf("IsLength is false for type %d", typ)
			}
		})
	}
}

func TestAbstractPropertyBuilder_cssNumber(t *testing.T) {
	builder := &AbstractPropertyBuilder{}
	if !builder.IsLength(NewPropertyValueFloat(CSSPrimitiveValueCssNumber, 0, "?")) {
		t.Errorf("IsLength is false for the number 0")
	}
	if !builder.IsLength(NewPropertyValueFloat(CSSPrimitiveValueCssNumber, 0.0, "?")) {
		t.Errorf("IsLength is false for the number 0.0")
	}
	if builder.IsLength(NewPropertyValueFloat(CSSPrimitiveValueCssNumber, 123.45, "?")) {
		t.Errorf("IsLength is true for the number 123.45")
	}
}

func TestAbstractPropertyBuilder_otherCssTypes(t *testing.T) {
	builder := &AbstractPropertyBuilder{}
	for _, typ := range []int16{CSSPrimitiveValueCssUnknown, CSSPrimitiveValueCssPercentage, CSSPrimitiveValueCssDeg,
		CSSPrimitiveValueCssString, CSSPrimitiveValueCssUri} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			if builder.IsLength(NewPropertyValueFloat(typ, 123.45, "?")) {
				t.Errorf("IsLength is true for type %d", typ)
			}
		})
	}
}
