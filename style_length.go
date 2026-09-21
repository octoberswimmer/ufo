// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/Length.java

package ufo

import (
	"fmt"
	"math"
)

// LengthLengthType ports the enum Length.LengthType.
type LengthLengthType int

const (
	LengthLengthTypeVariable LengthLengthType = iota
	LengthLengthTypeFixed
	LengthLengthTypePercent
)

// Should use something more reasonable here (e.g. a few feet based on the current
// DPI)
const LengthMaxWidth = math.MaxInt32 / 2

var LengthZero = newLengthZero()

// Length is a simplified version of KHTML's Length type.  It's very convenient to be able
// to treat length values (including auto) in a uniform matter when calculating
// table column widths.  Our own LengthValue is too heavyweight for this purpose and
// doesn't encompass variable (auto) widths.
type Length struct {
	typ   LengthLengthType
	value int64
}

func newLengthZero() *Length {
	return NewLength(0, LengthLengthTypeVariable)
}

func NewLength(value int64, typ LengthLengthType) *Length {
	return &Length{value: value, typ: typ}
}

func (l *Length) Value() int64 {
	return l.value
}

func (l *Length) Type() LengthLengthType {
	return l.typ
}

func (l *Length) IsVariable() bool {
	return l.typ == LengthLengthTypeVariable
}

func (l *Length) IsFixed() bool {
	return l.typ == LengthLengthTypeFixed
}

func (l *Length) IsPercent() bool {
	return l.typ == LengthLengthTypePercent
}

func (l *Length) Width(maxWidth int) int64 {
	switch l.typ {
	case LengthLengthTypeFixed:
		return l.value
	case LengthLengthTypePercent:
		return int64(maxWidth) * l.value / 100
	default:
		return int64(maxWidth)
	}
}

func (l *Length) MinWidth(maxWidth int) int64 {
	switch l.typ {
	case LengthLengthTypeFixed:
		return l.value
	case LengthLengthTypePercent:
		return int64(maxWidth) * l.value / 100
	default:
		return 0
	}
}

func (l *Length) ToString() string {
	var typ string
	switch l.typ {
	case LengthLengthTypeFixed:
		typ = "fixed"
	case LengthLengthTypePercent:
		typ = "percent"
	case LengthLengthTypeVariable:
		typ = "variable"
	}
	return fmt.Sprintf("Length (type=%s, value=%d)", typ, l.value)
}

func (l *Length) String() string {
	return l.ToString()
}
