// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/StrutMetrics.java

package ufo

// StrutMetrics holds the ascent, baseline and descent of the strut of a block
// (the zero-width inline box with the block's font and line height which
// every line box starts with).
type StrutMetrics struct {
	baseline int
	ascent   float32
	descent  float32
}

func NewStrutMetrics(ascent float32, baseline int, descent float32) *StrutMetrics {
	return &StrutMetrics{ascent: ascent, baseline: baseline, descent: descent}
}

func (s *StrutMetrics) GetAscent() float32 {
	return s.ascent
}

func (s *StrutMetrics) GetBaseline() int {
	return s.baseline
}

func (s *StrutMetrics) SetBaseline(baseline int) {
	s.baseline = baseline
}

func (s *StrutMetrics) GetDescent() float32 {
	return s.descent
}
