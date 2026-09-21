// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/ListBreakPointsProvider.java

package ufo

// ListBreakPointsProvider returns the break points of a list in order.
//
// Author of the Java class: Lukas Zaruba, lukas.zaruba@gmail.com
type ListBreakPointsProvider struct {
	breakPoints []*BreakPoint
	// index of the element that Next returns; stands for the Java Iterator
	index int
}

func NewListBreakPointsProvider(breakPoints []*BreakPoint) *ListBreakPointsProvider {
	return &ListBreakPointsProvider{breakPoints: breakPoints}
}

func (p *ListBreakPointsProvider) Next() *BreakPoint {
	if p.index >= len(p.breakPoints) {
		return BreakPointGetDonePoint()
	}
	result := p.breakPoints[p.index]
	p.index++
	return result
}
