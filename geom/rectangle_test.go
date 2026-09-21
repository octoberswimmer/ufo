package geom

import (
	"math"
	"testing"
)

func (j jdkRect) rectangle() *Rectangle {
	return NewRectangle(j.x, j.y, j.w, j.h)
}

func assertRect(t *testing.T, what string, got *Rectangle, want jdkRect) {
	t.Helper()
	if got.X != want.x || got.Y != want.y || got.Width != want.w || got.Height != want.h {
		t.Errorf("%s = %v, want %+v", what, got, want)
	}
}

func TestRectanglePairOperationsMatchJDK(t *testing.T) {
	for _, tc := range jdkRectanglePairs {
		a := tc.a.rectangle()
		b := tc.b.rectangle()
		name := a.String() + " / " + b.String()
		assertRect(t, name+" Intersection", a.Intersection(b), tc.intersection)
		assertRect(t, name+" Union", a.Union(b), tc.union)
		if got := a.Intersects(b); got != tc.intersects {
			t.Errorf("%s Intersects = %v, want %v", name, got, tc.intersects)
		}
		if got := a.ContainsRectangle(b); got != tc.contains {
			t.Errorf("%s ContainsRectangle = %v, want %v", name, got, tc.contains)
		}
		assertRect(t, name+" operands unchanged (a)", a, tc.a)
		assertRect(t, name+" operands unchanged (b)", b, tc.b)
		a.Add(b)
		assertRect(t, name+" Add", a, tc.add)
	}
}

func TestRectanglePointOperationsMatchJDK(t *testing.T) {
	for _, tc := range jdkRectanglePoints {
		r := tc.r.rectangle()
		name := r.String()
		if got := r.Contains(tc.x, tc.y); got != tc.contains {
			t.Errorf("%s Contains(%d, %d) = %v, want %v", name, tc.x, tc.y, got, tc.contains)
		}
		if got := r.ContainsPoint(NewPoint(tc.x, tc.y)); got != tc.contains {
			t.Errorf("%s ContainsPoint(%d, %d) = %v, want %v", name, tc.x, tc.y, got, tc.contains)
		}
		r.AddXY(tc.x, tc.y)
		assertRect(t, name+" AddXY", r, tc.add)
		r = tc.r.rectangle()
		r.AddPoint(NewPoint(tc.x, tc.y))
		assertRect(t, name+" AddPoint", r, tc.add)
	}
}

func TestRectangleTranslateAndGrowMatchJDK(t *testing.T) {
	for _, tc := range jdkRectangleMoves {
		r := tc.r.rectangle()
		name := r.String()
		r.Translate(tc.dx, tc.dy)
		assertRect(t, name+" Translate", r, tc.translated)
		r = tc.r.rectangle()
		r.Grow(tc.dx, tc.dy)
		assertRect(t, name+" Grow", r, tc.grown)
	}
}

func TestRectangleIsEmptyAndStringMatchJDK(t *testing.T) {
	for _, tc := range jdkRectangleStrings {
		r := tc.r.rectangle()
		if got := r.IsEmpty(); got != tc.empty {
			t.Errorf("%v IsEmpty = %v, want %v", r, got, tc.empty)
		}
		if got := r.String(); got != tc.str {
			t.Errorf("String = %q, want %q", got, tc.str)
		}
	}
}

func TestRectangleIntersectionOfDisjointRectanglesHasNegativeSize(t *testing.T) {
	got := NewRectangle(0, 0, 10, 10).Intersection(NewRectangle(20, 30, 5, 5))
	assertRect(t, "Intersection", got, jdkRect{20, 30, -10, -20})
	if !got.IsEmpty() {
		t.Errorf("IsEmpty = false for %v", got)
	}
}

func TestRectangleUnionWithEmptyRectangles(t *testing.T) {
	base := NewRectangle(0, 0, 10, 10)
	// A zero-size rectangle takes part with its location.
	assertRect(t, "union with zero size", base.Union(NewRectangle(20, 20, 0, 0)), jdkRect{0, 0, 20, 20})
	// A rectangle with a negative dimension is ignored.
	assertRect(t, "union with negative size", base.Union(NewRectangle(50, 50, -1, 4)), jdkRect{0, 0, 10, 10})
	assertRect(t, "negative size union", NewRectangle(50, 50, -1, 4).Union(base), jdkRect{0, 0, 10, 10})
	// The result is a copy, not the operand.
	u := base.Union(NewRectangle(50, 50, -1, 4))
	u.X = 99
	if base.X != 0 {
		t.Errorf("Union returned the receiver instead of a copy")
	}
}

func TestRectangleAddPointOnFarEdgeIsNotContained(t *testing.T) {
	r := NewRectangle(0, 0, 10, 10)
	r.AddXY(15, 20)
	assertRect(t, "AddXY", r, jdkRect{0, 0, 15, 20})
	if r.Contains(15, 20) {
		t.Errorf("the added point is on the right/bottom edge and must not be contained")
	}
	neg := NewRectangle(3, 3, -1, -1)
	neg.AddXY(7, 8)
	assertRect(t, "AddXY on non-existent rectangle", neg, jdkRect{7, 8, 0, 0})
}

func TestRectangleSetRectAndRectangle2DGetBoundsMatchJDK(t *testing.T) {
	tests := []struct {
		in      [4]float64
		setRect jdkRect
		bounds  jdkRect
	}{
		{[4]float64{0.5, 0.5, 2.2, 3.7}, jdkRect{0, 0, 3, 5}, jdkRect{0, 0, 3, 5}},
		{[4]float64{-1.5, -2.5, 1.0, 1.0}, jdkRect{-2, -3, 2, 2}, jdkRect{-2, -3, 2, 2}},
		{[4]float64{1.0e12, 0.0, 5.0, 5.0}, jdkRect{2147483647, 0, -1, 5}, jdkRect{2147483647, 0, 5, 5}},
		{[4]float64{-1.0e12, -1.0e12, 3.0e12, 1.0e11}, jdkRect{-2147483648, -2147483648, 2147483647, -2147483648}, jdkRect{-2147483648, -2147483648, 2147483647, 2147483647}},
		{[4]float64{1.0, 1.0, -2.0, 3.0}, jdkRect{1, 1, -2, 3}, jdkRect{0, 0, 0, 0}},
		{[4]float64{math.NaN(), 0.0, 1.0, 1.0}, jdkRect{0, 0, 0, 1}, jdkRect{0, 0, 0, 1}},
	}
	for _, tc := range tests {
		r := NewRectangleEmpty()
		r.SetRect(tc.in[0], tc.in[1], tc.in[2], tc.in[3])
		assertRect(t, "SetRect", r, tc.setRect)
		r2d := NewRectangle2DDouble(tc.in[0], tc.in[1], tc.in[2], tc.in[3])
		assertRect(t, "Rectangle2D.GetBounds", r2d.GetBounds(), tc.bounds)
	}
}

func TestRectangleConstructorsAndAccessors(t *testing.T) {
	r := NewRectangle(1, 2, 3, 4)
	c := NewRectangleFromRectangle(r)
	c.X = 50
	if r.X != 1 {
		t.Errorf("NewRectangleFromRectangle shares state with its argument")
	}
	assertRect(t, "NewRectangleWidthHeight", NewRectangleWidthHeight(7, 8), jdkRect{0, 0, 7, 8})
	assertRect(t, "NewRectangleEmpty", NewRectangleEmpty(), jdkRect{0, 0, 0, 0})
	assertRect(t, "NewRectangleFromPointDimension", NewRectangleFromPointDimension(NewPoint(1, 2), NewDimension(3, 4)), jdkRect{1, 2, 3, 4})
	assertRect(t, "NewRectangleFromPoint", NewRectangleFromPoint(NewPoint(1, 2)), jdkRect{1, 2, 0, 0})
	assertRect(t, "NewRectangleFromDimension", NewRectangleFromDimension(NewDimension(3, 4)), jdkRect{0, 0, 3, 4})

	if r.GetX() != 1.0 || r.GetY() != 2.0 || r.GetWidth() != 3.0 || r.GetHeight() != 4.0 {
		t.Errorf("getters = %v %v %v %v", r.GetX(), r.GetY(), r.GetWidth(), r.GetHeight())
	}
	if r.GetMaxX() != 4.0 || r.GetMaxY() != 6.0 || r.GetCenterX() != 2.5 || r.GetCenterY() != 4.0 {
		t.Errorf("max/center = %v %v %v %v", r.GetMaxX(), r.GetMaxY(), r.GetCenterX(), r.GetCenterY())
	}
	b := r.GetBounds()
	b.Width = 100
	if r.Width != 3 {
		t.Errorf("GetBounds returned the receiver instead of a copy")
	}
	r.SetBounds(5, 6, 7, 8)
	assertRect(t, "SetBounds", r, jdkRect{5, 6, 7, 8})
	r.SetBoundsRectangle(NewRectangle(1, 1, 2, 2))
	assertRect(t, "SetBoundsRectangle", r, jdkRect{1, 1, 2, 2})
	r.SetLocation(9, 10)
	r.SetSize(11, 12)
	assertRect(t, "SetLocation/SetSize", r, jdkRect{9, 10, 11, 12})
	if !r.GetLocation().Equals(NewPoint(9, 10)) || !r.GetSize().Equals(NewDimension(11, 12)) {
		t.Errorf("GetLocation/GetSize = %v %v", r.GetLocation(), r.GetSize())
	}
	if !r.Equals(NewRectangle(9, 10, 11, 12)) || r.Equals(NewRectangle(9, 10, 11, 13)) || r.Equals(nil) {
		t.Errorf("Equals is wrong")
	}
}

func TestRectangleIsAShape(t *testing.T) {
	var s Shape = NewRectangle(1, 2, 3, 4)
	assertRect(t, "GetBounds", s.GetBounds(), jdkRect{1, 2, 3, 4})
	if !s.GetBounds2D().Equals(NewRectangle2DDouble(1, 2, 3, 4)) {
		t.Errorf("GetBounds2D = %v", s.GetBounds2D())
	}
	if !s.IntersectsWithXYWH(3.5, 5.5, 1, 1) || s.IntersectsWithXYWH(4, 2, 1, 1) {
		t.Errorf("IntersectsWithXYWH is wrong")
	}
}

func TestPointDimensionInsets(t *testing.T) {
	p := NewPoint(1, 2)
	if p.String() != "java.awt.Point[x=1,y=2]" {
		t.Errorf("Point.String = %q", p.String())
	}
	p.Translate(3, -5)
	if !p.Equals(NewPoint(4, -3)) {
		t.Errorf("Translate = %v", p)
	}
	q := NewPointFromPoint(p)
	q.SetLocation(7, 8)
	if p.X != 4 || q.GetX() != 7.0 || q.GetY() != 8.0 {
		t.Errorf("copy or SetLocation is wrong: %v %v", p, q)
	}
	q.SetLocationDouble(2.5, -2.5)
	if !q.Equals(NewPoint(3, -2)) {
		t.Errorf("SetLocationDouble rounds with floor(v+0.5); got %v", q)
	}
	if d := NewPoint(0, 0).Distance(3, 4); d != 5 {
		t.Errorf("Distance = %v", d)
	}

	d := NewDimension(3, 4)
	if d.String() != "java.awt.Dimension[width=3,height=4]" {
		t.Errorf("Dimension.String = %q", d.String())
	}
	e := NewDimensionFromDimension(d)
	e.SetSize(5, 6)
	if d.Width != 3 || e.GetWidth() != 5.0 || e.GetHeight() != 6.0 {
		t.Errorf("copy or SetSize is wrong: %v %v", d, e)
	}
	e.SetSizeDouble(1.1, 2.0)
	if !e.Equals(NewDimension(2, 2)) {
		t.Errorf("SetSizeDouble rounds up; got %v", e)
	}
	if !e.GetSize().Equals(e) || e.Equals(nil) {
		t.Errorf("GetSize/Equals is wrong")
	}

	i := NewInsets(1, 2, 3, 4)
	if i.String() != "java.awt.Insets[top=1,left=2,bottom=3,right=4]" {
		t.Errorf("Insets.String = %q", i.String())
	}
	c := i.Clone()
	c.Set(5, 6, 7, 8)
	if !i.Equals(NewInsets(1, 2, 3, 4)) || !c.Equals(NewInsets(5, 6, 7, 8)) {
		t.Errorf("Insets Clone/Set is wrong: %v %v", i, c)
	}

	p2 := NewPoint2DDouble(1.5, -2.0e10)
	if p2.String() != "Point2D.Double[1.5, -2.0E10]" {
		t.Errorf("Point2D.String = %q", p2.String())
	}
	if f := NewPoint2DFloat(0.1, 0.2); f.X != float64(float32(0.1)) || f.Y != float64(float32(0.2)) {
		t.Errorf("NewPoint2DFloat = %v", f)
	}
}

func TestRectangle2D(t *testing.T) {
	r := NewRectangle2DDouble(1, 2.5, 1.0e-5, 12345678.9)
	if got, want := r.String(), "java.awt.geom.Rectangle2D$Double[x=1.0,y=2.5,w=1.0E-5,h=1.23456789E7]"; got != want {
		t.Errorf("String = %q, want %q", got, want)
	}
	a := NewRectangle2DDouble(0, 0, 10, 10)
	b := NewRectangle2DDouble(20.5, 30, 5, 5)
	if got := a.CreateIntersection(b); !got.Equals(NewRectangle2DDouble(20.5, 30, -10.5, -20)) {
		t.Errorf("CreateIntersection of disjoint rectangles = %v", got)
	}
	if got := a.CreateUnion(b); !got.Equals(NewRectangle2DDouble(0, 0, 25.5, 35)) {
		t.Errorf("CreateUnion = %v", got)
	}
	c := a.Clone()
	c.Add(-2, 12)
	if !c.Equals(NewRectangle2DDouble(-2, 0, 12, 12)) {
		t.Errorf("Add = %v", c)
	}
	c.AddRectangle2D(b)
	if !c.Equals(NewRectangle2DDouble(-2, 0, 27.5, 35)) {
		t.Errorf("AddRectangle2D = %v", c)
	}
	if !a.Contains(0, 0) || a.Contains(10, 5) || a.Contains(5, 10) {
		t.Errorf("Contains must include the left/top edges and exclude the right/bottom edges")
	}
	if !a.ContainsWithXYWH(0, 0, 10, 10) || a.ContainsWithXYWH(0, 0, 10.5, 10) || a.ContainsWithXYWH(1, 1, 0, 1) {
		t.Errorf("ContainsWithXYWH is wrong")
	}
	if a.IntersectsWithXYWH(10, 0, 5, 5) || !a.IntersectsWithXYWH(9.5, 0, 5, 5) || a.IntersectsWithXYWH(1, 1, 0, 5) {
		t.Errorf("IntersectsWithXYWH is wrong")
	}
	if !a.Intersects(NewRectangle(9, 9, 5, 5)) || a.Intersects(NewRectangle(10, 10, 5, 5)) {
		t.Errorf("Intersects is wrong")
	}
	if got := a.Outcode(-1, 11); got != Rectangle2DOutLeft|Rectangle2DOutBottom {
		t.Errorf("Outcode = %d", got)
	}
	if !a.IntersectsLine(-5, 5, 15, 5) || a.IntersectsLine(-5, -1, 15, -1) || !a.IntersectsLine(-5, 5, 5, -5) || a.IntersectsLine(-5, 4, 4, -5.5) {
		t.Errorf("IntersectsLine is wrong")
	}
	if got := NewRectangle2DFloat(0.1, 0.2, 0.3, 0.4); got.X != float64(float32(0.1)) || got.Height != float64(float32(0.4)) {
		t.Errorf("NewRectangle2DFloat = %v", got)
	}
}
