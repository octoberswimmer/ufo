// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/layout/breaker/BreakAnywhereLineBreakStrategyTest.java

package ufo

import (
	"reflect"
	"testing"
)

// Tests for BreakAnywhereLineBreakStrategy.
//
// Correct range: [1..length] — starts at 1 (minimum 1 char per line),
// ends at length (full string tested to detect overflow).

func breakAnywhereLineBreakStrategyTestCollectAllPositions(strategy *BreakAnywhereLineBreakStrategy) []int {
	positions := []int{}
	for bp := strategy.Next(); bp.GetPosition() != BreakIteratorDone; bp = strategy.Next() {
		positions = append(positions, bp.GetPosition())
	}
	return positions
}

// Core fix: full string length must be offered as a break point.
// Without position=length, doBreakText never detects the full string overflows avail.
func TestBreakAnywhereLineBreakStrategy_shouldOfferBreakPointAtFullStringLength(t *testing.T) {
	str := "ABC" // length = 3
	strategy := NewBreakAnywhereLineBreakStrategy(str)

	positions := breakAnywhereLineBreakStrategyTestCollectAllPositions(strategy)

	found := false
	for _, position := range positions {
		if position == len(str) {
			found = true
		}
	}
	if !found {
		t.Errorf("must include break point at position %d (= string length), got %v", len(str), positions)
	}
}

// Range must be [1..length], starting at 1 not 0.
// BreakPoint(0) = 0-char line = no progress = infinite loop risk
func TestBreakAnywhereLineBreakStrategy_shouldOfferBreakPointsFromOneToLengthInclusive(t *testing.T) {
	str := "HELLO" // length = 5
	strategy := NewBreakAnywhereLineBreakStrategy(str)

	positions := breakAnywhereLineBreakStrategyTestCollectAllPositions(strategy)

	if !reflect.DeepEqual(positions, []int{1, 2, 3, 4, 5}) {
		t.Errorf("expected [1 2 3 4 5] for %q — starts at 1 (not 0), ends at length; got %v", str, positions)
	}
}

// Single character string must offer exactly [1] — the only valid break
// (after the single char).
func TestBreakAnywhereLineBreakStrategy_shouldOfferSingleBreakPointForSingleCharString(t *testing.T) {
	strategy := NewBreakAnywhereLineBreakStrategy("A")

	positions := breakAnywhereLineBreakStrategyTestCollectAllPositions(strategy)

	if !reflect.DeepEqual(positions, []int{1}) {
		t.Errorf("single-char string must offer exactly [1], got %v", positions)
	}
}

// Empty string: no characters → nothing to break → DonePoint immediately →
// empty list.
func TestBreakAnywhereLineBreakStrategy_shouldOfferNoBreakPointsForEmptyString(t *testing.T) {
	strategy := NewBreakAnywhereLineBreakStrategy("")

	positions := breakAnywhereLineBreakStrategyTestCollectAllPositions(strategy)

	if len(positions) != 0 {
		t.Errorf("empty string has no characters to break → no positions offered, got %v", positions)
	}
}

// Supplementary characters must stay intact. A break point must land after a
// whole code point. Java expects the UTF-16 offsets 2 and 4; positions in the
// port are byte offsets, and each of the two emoji is 4 bytes long in UTF-8.
func TestBreakAnywhereLineBreakStrategy_shouldOfferBreakPointsOnCodePointBoundariesForSurrogatePairs(t *testing.T) {
	str := "😀😁" // two emoji, 8 bytes / 2 code points
	strategy := NewBreakAnywhereLineBreakStrategy(str)

	positions := breakAnywhereLineBreakStrategyTestCollectAllPositions(strategy)

	if !reflect.DeepEqual(positions, []int{4, 8}) {
		t.Errorf("break points must fall after each whole emoji (code point), got %v", positions)
	}
}
