// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/constants/ValueConstantsTest.java

package ufo

import "testing"

func TestValueConstants_stringForSACPrimitiveType(t *testing.T) {
	cases := []struct {
		primitiveType int16
		want          string
	}{
		{CSSPrimitiveValueCssEms, "em"},
		{CSSPrimitiveValueCssPx, "px"},
		{CSSPrimitiveValueCssPercentage, "%"},
		{CSSPrimitiveValueCssPt, "pt"},
		{CSSPrimitiveValueCssMm, "mm"},
	}
	for _, c := range cases {
		if got := ValueConstantsStringForSACPrimitiveType(c.primitiveType); got != c.want {
			t.Errorf("ValueConstantsStringForSACPrimitiveType(%d) = %q, want %q", c.primitiveType, got, c.want)
		}
	}
}
