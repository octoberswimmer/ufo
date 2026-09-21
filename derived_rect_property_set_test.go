// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/style/derived/RectPropertySetTest.java

package ufo

import "testing"

func TestRectPropertySet_isAllZeros(t *testing.T) {
	cases := []struct {
		rect *RectPropertySet
		want bool
	}{
		{RectPropertySetAllZeros, true},
		{NewRectPropertySet(0, 0, 0, 0), true},
		{NewRectPropertySet(0, 0, 0, 1), false},
		{NewRectPropertySet(0, 0, 1, 0), false},
		{NewRectPropertySet(0, 1, 0, 0), false},
		{NewRectPropertySet(1, 0, 0, 0), false},
	}
	for _, c := range cases {
		if got := c.rect.IsAllZeros(); got != c.want {
			t.Errorf("%s.IsAllZeros() = %v, want %v", c.rect, got, c.want)
		}
	}
}

func TestRectPropertySet_hasNegativeValues(t *testing.T) {
	cases := []struct {
		rect *RectPropertySet
		want bool
	}{
		{RectPropertySetAllZeros, false},
		{NewRectPropertySet(0, 0, 0, 0), false},
		{NewRectPropertySet(1, 1, 1, 1), false},
		{NewRectPropertySet(0, 0, 0, -1), true},
		{NewRectPropertySet(0, 0, -1, 0), true},
		{NewRectPropertySet(0, -1, 0, 0), true},
		{NewRectPropertySet(-1, 0, 0, 0), true},
	}
	for _, c := range cases {
		if got := c.rect.HasNegativeValues(); got != c.want {
			t.Errorf("%s.HasNegativeValues() = %v, want %v", c.rect, got, c.want)
		}
	}
}
