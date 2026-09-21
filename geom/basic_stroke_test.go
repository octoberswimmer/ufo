package geom

import "testing"

func TestBasicStrokeDefaults(t *testing.T) {
	s := NewBasicStroke()
	if s.GetLineWidth() != 1.0 || s.GetEndCap() != BasicStrokeCapSquare || s.GetLineJoin() != BasicStrokeJoinMiter ||
		s.GetMiterLimit() != 10.0 || s.GetDashArray() != nil || s.GetDashPhase() != 0.0 {
		t.Errorf("defaults = %+v", s)
	}
	w := NewBasicStrokeWithWidth(2.5)
	if w.GetLineWidth() != 2.5 || w.GetEndCap() != BasicStrokeCapSquare || w.GetMiterLimit() != 10.0 {
		t.Errorf("NewBasicStrokeWithWidth = %+v", w)
	}
	c := NewBasicStrokeWithWidthCapJoin(3, BasicStrokeCapRound, BasicStrokeJoinRound)
	if c.GetEndCap() != BasicStrokeCapRound || c.GetLineJoin() != BasicStrokeJoinRound || c.GetMiterLimit() != 10.0 {
		t.Errorf("NewBasicStrokeWithWidthCapJoin = %+v", c)
	}
	m := NewBasicStrokeWithWidthCapJoinMiterlimit(3, BasicStrokeCapButt, BasicStrokeJoinMiter, 4)
	if m.GetMiterLimit() != 4 {
		t.Errorf("NewBasicStrokeWithWidthCapJoinMiterlimit = %+v", m)
	}
	var _ Stroke = s
}

func TestBasicStrokeConstantsHaveTheJDKValues(t *testing.T) {
	if BasicStrokeCapButt != 0 || BasicStrokeCapRound != 1 || BasicStrokeCapSquare != 2 ||
		BasicStrokeJoinMiter != 0 || BasicStrokeJoinRound != 1 || BasicStrokeJoinBevel != 2 {
		t.Errorf("BasicStroke constants differ from the JDK")
	}
	if PathIteratorWindEvenOdd != 0 || PathIteratorWindNonZero != 1 || PathIteratorSegMoveTo != 0 ||
		PathIteratorSegLineTo != 1 || PathIteratorSegQuadTo != 2 || PathIteratorSegCubicTo != 3 || PathIteratorSegClose != 4 {
		t.Errorf("PathIterator constants differ from the JDK")
	}
	if Arc2DOpen != 0 || Arc2DChord != 1 || Arc2DPie != 2 {
		t.Errorf("Arc2D constants differ from the JDK")
	}
}

// BorderPainter builds its dashed strokes with a bevel join and a miter limit
// of 0, which the JDK accepts because the limit is checked for miter joins
// only.
func TestBasicStrokeDashedAsBorderPainterBuildsIt(t *testing.T) {
	pattern := []float32{6, 6}
	s := NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(4, BasicStrokeCapButt, BasicStrokeJoinBevel, 0, pattern, 3)
	pattern[0] = 99
	dash := s.GetDashArray()
	if len(dash) != 2 || dash[0] != 6 || dash[1] != 6 || s.GetDashPhase() != 3 {
		t.Errorf("dash = %v, phase = %v", dash, s.GetDashPhase())
	}
	// ITextOutputDevice.transformStroke scales the returned array in place.
	dash[0] *= 2
	if s.GetDashArray()[0] != 6 {
		t.Errorf("GetDashArray returned the stroke's own array")
	}
}

func TestBasicStrokeRejectsWhatTheJDKRejects(t *testing.T) {
	tests := []struct {
		width      float32
		cap, join  int
		miterlimit float32
		dash       []float32
		dashPhase  float32
		message    string
	}{
		{-1, 0, 0, 10, nil, 0, "negative width"},
		{1, 3, 0, 10, nil, 0, "illegal end cap value"},
		{1, 0, 0, 0.5, nil, 0, "miter limit < 1"},
		{1, 0, 3, 10, nil, 0, "illegal line join value"},
		{1, 0, 2, 0, []float32{1}, -1, "negative dash phase"},
		{1, 0, 2, 0, []float32{0, 0}, 0, "dash lengths all zero"},
		{1, 0, 2, 0, []float32{1, -1}, 0, "negative dash length"},
		{1, 0, 2, 0, []float32{}, 0, "dash lengths all zero"},
		{1, 0, 2, 0, []float32{2, 0}, 3, ""},
	}
	for _, tc := range tests {
		func() {
			defer func() {
				r := recover()
				if tc.message == "" {
					if r != nil {
						t.Errorf("%+v: unexpected panic %v", tc, r)
					}
					return
				}
				e, ok := r.(*IllegalArgumentException)
				if !ok || e.Message != tc.message {
					t.Errorf("%+v: panic = %v, want %q", tc, r, tc.message)
				}
			}()
			NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(tc.width, tc.cap, tc.join, tc.miterlimit, tc.dash, tc.dashPhase)
		}()
	}
}

func TestBasicStrokeEquals(t *testing.T) {
	a := NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(2, BasicStrokeCapButt, BasicStrokeJoinBevel, 0, []float32{1, 2}, 0)
	b := NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(2, BasicStrokeCapButt, BasicStrokeJoinBevel, 0, []float32{1, 2}, 0)
	c := NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(2, BasicStrokeCapButt, BasicStrokeJoinBevel, 0, []float32{1, 3}, 0)
	d := NewBasicStrokeWithWidthCapJoinMiterlimit(2, BasicStrokeCapButt, BasicStrokeJoinBevel, 0)
	if !a.Equals(b) || a.Equals(c) || a.Equals(d) || d.Equals(a) || !d.Equals(d) || a.Equals(nil) {
		t.Errorf("Equals is wrong")
	}
}

func TestBasicStrokeCreateStrokedShapeIsNotPorted(t *testing.T) {
	defer func() {
		if _, ok := recover().(*UnsupportedOperationException); !ok {
			t.Errorf("expected an *UnsupportedOperationException panic")
		}
	}()
	NewBasicStroke().CreateStrokedShape(NewRectangle(0, 0, 1, 1))
}
