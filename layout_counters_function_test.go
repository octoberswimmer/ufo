// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/layout/CountersFunctionTest.java

package ufo

import "testing"

func TestCountersFunctionUpperRoman(t *testing.T) {
	function := NewCountersFunction([]int{2, 4, 6, 8}, " | ", IdentValueUpperRoman)
	if got := function.Evaluate(); got != "II | IV | VI | VIII" {
		t.Errorf("got %q", got)
	}
}

func TestCountersFunctionLowerLatin(t *testing.T) {
	function := NewCountersFunction([]int{3, 6, 9, 12}, " / ", IdentValueLowerLatin)
	if got := function.Evaluate(); got != "c / f / i / l" {
		t.Errorf("got %q", got)
	}
}
