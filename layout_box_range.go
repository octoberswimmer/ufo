// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BoxRange.java

package ufo

import "fmt"

type BoxRange struct {
	start int
	end   int
}

func NewBoxRange(start int, end int) *BoxRange {
	return &BoxRange{start: start, end: end}
}

func (r *BoxRange) GetStart() int {
	return r.start
}

func (r *BoxRange) GetEnd() int {
	return r.end
}

func (r *BoxRange) ToString() string {
	return fmt.Sprintf("[start=%d, end=%d]", r.start, r.end)
}

func (r *BoxRange) String() string {
	return r.ToString()
}
