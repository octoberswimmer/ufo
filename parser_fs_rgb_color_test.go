// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/FSRGBColorTest.java

package ufo

import "testing"

func TestFSRGBColor_toString_rgb(t *testing.T) {
	if got := NewFSRGBColor(255, 250, 180).ToString(); got != "#fffab4" {
		t.Errorf("got %q, want %q", got, "#fffab4")
	}
	if got := NewFSRGBColor(1, 2, 3).ToString(); got != "#010203" {
		t.Errorf("got %q, want %q", got, "#010203")
	}
}

func TestFSRGBColor_toString_rgba(t *testing.T) {
	if got := NewFSRGBColorWithAlpha(255, 250, 180, 0.42).ToString(); got != "rgba(255,250,180,0.42)" {
		t.Errorf("got %q, want %q", got, "rgba(255,250,180,0.42)")
	}
	if got := NewFSRGBColorWithAlpha(1, 2, 3, 0.01).ToString(); got != "rgba(1,2,3,0.01)" {
		t.Errorf("got %q, want %q", got, "rgba(1,2,3,0.01)")
	}
}

func TestFSRGBColor_toHSB(t *testing.T) {
	cases := []struct {
		color *FSRGBColor
		want  *HSBColor
	}{
		{NewFSRGBColor(255, 200, 100), NewHSBColor(0.107526876, 0.60784316, 1.0)},
		{NewFSRGBColor(1, 2, 3), NewHSBColor(0.5833333, 0.6666667, 0.011764706)},
		{NewFSRGBColorWithAlpha(1, 2, 3, 0.1), NewHSBColor(0.5833333, 0.6666667, 0.011764706)},
	}
	for _, c := range cases {
		if got := c.color.ToHSB(); !got.Equals(c.want) {
			t.Errorf("%s.ToHSB() = %v, want %v", c.color, *got, *c.want)
		}
	}
}
