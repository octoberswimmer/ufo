// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/BreakAnywhereLineBreakStrategy.java

package ufo

import "unicode/utf8"

// BreakAnywhereLineBreakStrategy reports a break point after every character
// (code point) of a text.
//
// Author of the Java class: Lukas Zaruba, lukas.zaruba@gmail.com
type BreakAnywhereLineBreakStrategy struct {
	currentString string
	position      int
}

func NewBreakAnywhereLineBreakStrategy(currentString string) *BreakAnywhereLineBreakStrategy {
	return &BreakAnywhereLineBreakStrategy{currentString: currentString}
}

func (s *BreakAnywhereLineBreakStrategy) Next() *BreakPoint {
	if s.position >= len(s.currentString) {
		return BreakPointGetDonePoint()
	}
	// Java: currentString.offsetByCodePoints(position, 1)
	_, size := utf8.DecodeRuneInString(s.currentString[s.position:])
	s.position += size
	return NewBreakPoint(s.position)
}
