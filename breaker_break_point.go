// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/BreakPoint.java

package ufo

import "strconv"

// BreakPoint is a position at which a text may be broken, with the hyphen
// text to insert there, if any. The position is a byte offset.
//
// Author of the Java class: Lukas Zaruba, lukas.zaruba@gmail.com
type BreakPoint struct {
	position int
	// nil when no hyphen was set
	hyphen *string
}

func NewBreakPoint(position int) *BreakPoint {
	return &BreakPoint{position: position}
}

func (b *BreakPoint) GetPosition() int {
	return b.position
}

func (b *BreakPoint) SetPosition(position int) {
	b.position = position
}

func (b *BreakPoint) String() string {
	return "BreakPoint [position=" + strconv.Itoa(b.position) + "]"
}

func (b *BreakPoint) ToString() string {
	return b.String()
}

func (b *BreakPoint) CompareTo(o *BreakPoint) int {
	if b.position < o.position {
		return -1
	}
	if b.position == o.position {
		return 0
	}
	return 1
}

func (b *BreakPoint) SetHyphen(hyphen string) {
	b.hyphen = &hyphen
}

func (b *BreakPoint) GetHyphen() string {
	if b.hyphen == nil {
		return ""
	}
	return *b.hyphen
}

func BreakPointGetDonePoint() *BreakPoint {
	return NewBreakPoint(BreakIteratorDone)
}
