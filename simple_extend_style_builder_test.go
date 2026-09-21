// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/simple/extend/StyleBuilderTest.java

package ufo

import "testing"

func TestStyleBuilder_isInteger(t *testing.T) {
	handler := NewStyleBuilder()
	cases := []struct {
		value string
		want  bool
	}{
		{"", false},
		{"_", false},
		{"0", true},
		{"01234", true},
		{"123a", false},
		{"a234b", false},
	}
	for _, c := range cases {
		if got := handler.IsInteger(c.value); got != c.want {
			t.Errorf("IsInteger(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestStyleBuilder_convertToLength(t *testing.T) {
	handler := NewStyleBuilder()
	cases := []struct {
		value string
		want  string
	}{
		{"", ""},
		{"big", "big"},
		{"#eeffgg", "#eeffgg"},
		{"123a", "123a"},
		{"0", "0px"},
		{"123", "123px"},
	}
	for _, c := range cases {
		if got := handler.ConvertToLength(c.value); got != c.want {
			t.Errorf("ConvertToLength(%q) = %q, want %q", c.value, got, c.want)
		}
	}
}
