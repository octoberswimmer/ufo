package geom

import (
	"math"
	"slices"
)

// End cap and line join styles, with the JDK values.
const (
	// BasicStrokeJoinMiter joins path segments by extending their outside
	// edges until they meet.
	BasicStrokeJoinMiter = 0
	// BasicStrokeJoinRound joins path segments by rounding off the corner at a
	// radius of half the line width.
	BasicStrokeJoinRound = 1
	// BasicStrokeJoinBevel joins path segments by connecting the outer corners
	// of their wide outlines with a straight segment.
	BasicStrokeJoinBevel = 2

	// BasicStrokeCapButt ends unclosed subpaths and dash segments with no
	// added decoration.
	BasicStrokeCapButt = 0
	// BasicStrokeCapRound ends unclosed subpaths and dash segments with a
	// round decoration that has a radius equal to half of the width of the pen.
	BasicStrokeCapRound = 1
	// BasicStrokeCapSquare ends unclosed subpaths and dash segments with a
	// square projection that extends beyond the end of the segment to a
	// distance equal to half of the line width.
	BasicStrokeCapSquare = 2
)

// Stroke is java.awt.Stroke.
type Stroke interface {
	CreateStrokedShape(p Shape) Shape
}

// BasicStroke is java.awt.BasicStroke: the width, end caps, line joins, miter
// limit and dash pattern of a pen. It is immutable.
type BasicStroke struct {
	width      float32
	join       int
	cap        int
	miterlimit float32
	dash       []float32
	dashPhase  float32
}

// NewBasicStroke is new BasicStroke(): width 1, square caps, miter joins and a
// miter limit of 10.
func NewBasicStroke() *BasicStroke {
	return NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(1.0, BasicStrokeCapSquare, BasicStrokeJoinMiter, 10.0, nil, 0.0)
}

// NewBasicStrokeWithWidth is new BasicStroke(float width).
func NewBasicStrokeWithWidth(width float32) *BasicStroke {
	return NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(width, BasicStrokeCapSquare, BasicStrokeJoinMiter, 10.0, nil, 0.0)
}

// NewBasicStrokeWithWidthCapJoin is new BasicStroke(float width, int cap, int
// join).
func NewBasicStrokeWithWidthCapJoin(width float32, cap, join int) *BasicStroke {
	return NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(width, cap, join, 10.0, nil, 0.0)
}

// NewBasicStrokeWithWidthCapJoinMiterlimit is new BasicStroke(float width, int
// cap, int join, float miterlimit).
func NewBasicStrokeWithWidthCapJoinMiterlimit(width float32, cap, join int, miterlimit float32) *BasicStroke {
	return NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(width, cap, join, miterlimit, nil, 0.0)
}

// NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase is new
// BasicStroke(float width, int cap, int join, float miterlimit, float[] dash,
// float dash_phase). It panics with *IllegalArgumentException for the
// arguments the JDK rejects: a negative width, an unknown cap or join, a miter
// limit below 1 with a miter join, a negative dash phase, a negative dash
// length, or dash lengths that are all zero. The miter limit is not checked
// for the other joins. The dash slice is copied.
func NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(width float32, cap, join int, miterlimit float32, dash []float32, dashPhase float32) *BasicStroke {
	if width < 0.0 {
		panic(&IllegalArgumentException{Message: "negative width"})
	}
	if cap != BasicStrokeCapButt && cap != BasicStrokeCapRound && cap != BasicStrokeCapSquare {
		panic(&IllegalArgumentException{Message: "illegal end cap value"})
	}
	if join == BasicStrokeJoinMiter {
		if miterlimit < 1.0 {
			panic(&IllegalArgumentException{Message: "miter limit < 1"})
		}
	} else if join != BasicStrokeJoinRound && join != BasicStrokeJoinBevel {
		panic(&IllegalArgumentException{Message: "illegal line join value"})
	}
	if dash != nil {
		if dashPhase < 0.0 {
			panic(&IllegalArgumentException{Message: "negative dash phase"})
		}
		allzero := true
		for _, d := range dash {
			if d > 0.0 {
				allzero = false
			} else if d < 0.0 {
				panic(&IllegalArgumentException{Message: "negative dash length"})
			}
		}
		if allzero {
			panic(&IllegalArgumentException{Message: "dash lengths all zero"})
		}
	}
	return &BasicStroke{
		width:      width,
		cap:        cap,
		join:       join,
		miterlimit: miterlimit,
		dash:       slices.Clone(dash),
		dashPhase:  dashPhase,
	}
}

// CreateStrokedShape is not implemented: the JDK computes the outline with
// its rasterizer's stroker, and the PDF output device strokes with the PDF
// line state instead of calling this for a BasicStroke.
func (s *BasicStroke) CreateStrokedShape(p Shape) Shape {
	panic(&UnsupportedOperationException{Message: "not ported: java.awt.BasicStroke.createStrokedShape"})
}

func (s *BasicStroke) GetLineWidth() float32 {
	return s.width
}

func (s *BasicStroke) GetEndCap() int {
	return s.cap
}

func (s *BasicStroke) GetLineJoin() int {
	return s.join
}

func (s *BasicStroke) GetMiterLimit() float32 {
	return s.miterlimit
}

// GetDashArray returns a copy of the dash lengths, or nil for a solid line.
func (s *BasicStroke) GetDashArray() []float32 {
	return slices.Clone(s.dash)
}

func (s *BasicStroke) GetDashPhase() float32 {
	return s.dashPhase
}

// Equals reports whether o is a BasicStroke with the same attributes.
func (s *BasicStroke) Equals(o Stroke) bool {
	bs, ok := o.(*BasicStroke)
	if !ok || bs == nil {
		return false
	}
	if s.width != bs.width || s.join != bs.join || s.cap != bs.cap ||
		s.miterlimit != bs.miterlimit || s.dashPhase != bs.dashPhase {
		return false
	}
	if (s.dash == nil) != (bs.dash == nil) {
		return false
	}
	// java.util.Arrays.equals(float[], float[]) compares bit patterns, so NaN
	// equals NaN and 0.0 differs from -0.0.
	return slices.EqualFunc(s.dash, bs.dash, func(a, b float32) bool {
		return math.Float32bits(a) == math.Float32bits(b)
	})
}
