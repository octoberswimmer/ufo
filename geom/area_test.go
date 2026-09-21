package geom

import (
	"strings"
	"testing"
)

func TestAreaOfRectanglesMatchesJDK(t *testing.T) {
	a := NewArea(NewRectangle(0, 0, 10, 10))
	a.Intersect(NewArea(NewRectangle(5, 5, 10, 10)))
	if len(a.Shapes()) != 2 {
		t.Fatalf("Shapes = %v", a.Shapes())
	}
	if got := a.GetBounds(); !got.Equals(NewRectangle(5, 5, 5, 5)) {
		t.Errorf("GetBounds = %v", got)
	}
	if !a.IsRectangular() || a.IsEmpty() {
		t.Errorf("IsRectangular = %v, IsEmpty = %v", a.IsRectangular(), a.IsEmpty())
	}
	if !a.Intersects(NewRectangle(9, 9, 5, 5)) || a.Intersects(NewRectangle(0, 0, 5, 5)) {
		t.Errorf("Intersects is wrong")
	}
	// As the JDK's Area traces it (checked with OpenJDK 25): down the left
	// side, along the bottom, up the right side, closed without a line back.
	pi := a.GetPathIterator(nil)
	if pi.GetWindingRule() != PathIteratorWindNonZero {
		t.Errorf("winding rule = %d", pi.GetWindingRule())
	}
	assertSegments(t, "intersection", collectDouble(pi), []seg{
		{0, 5, 5}, {1, 5, 10}, {1, 10, 10}, {1, 10, 5}, {4},
	}, 0)
	assertSegments(t, "intersection, transformed", collectDouble(a.GetPathIterator(AffineTransformGetScaleInstance(2, 2))), []seg{
		{0, 10, 10}, {1, 10, 20}, {1, 20, 20}, {1, 20, 10}, {4},
	}, 0)

	a.Intersect(NewArea(NewRectangle(50, 50, 10, 10)))
	if !a.IsEmpty() {
		t.Errorf("the intersection of disjoint rectangles is not empty")
	}
	if got := a.GetBounds(); !got.Equals(NewRectangleEmpty()) {
		t.Errorf("GetBounds of an empty area = %v", got)
	}
	if got := a.GetBounds2D(); !got.Equals(NewRectangle2D()) {
		t.Errorf("GetBounds2D of an empty area = %v", got)
	}
	assertSegments(t, "empty", collectDouble(a.GetPathIterator(nil)), nil, 0)
	if a.Intersects(NewRectangle(0, 0, 100, 100)) {
		t.Errorf("an empty area intersects a rectangle")
	}
}

func TestAreaOfSingleShapeUsesItsIterator(t *testing.T) {
	e := NewEllipse2DDouble(0, 0, 10, 20)
	a := NewArea(e)
	assertSegments(t, "single", collectDouble(a.GetPathIterator(nil)), collectDouble(e.GetPathIterator(nil)), 0)
	if got := a.GetBounds(); !got.Equals(NewRectangle(0, 0, 10, 20)) {
		t.Errorf("GetBounds = %v", got)
	}
	if a.IsRectangular() {
		t.Errorf("an ellipse is reported as rectangular")
	}
	if !a.Intersects(NewRectangle(4, 9, 2, 2)) || a.Intersects(NewRectangle(0, 0, 1, 1)) {
		t.Errorf("Intersects is wrong")
	}
}

// ITextOutputDevice keeps its clip as an Area of shapes that went through
// AffineTransform.CreateTransformedShape, which turns rectangles into paths.
func TestAreaOfTransformedRectanglesIsRectangular(t *testing.T) {
	at := AffineTransformGetScaleInstance(0.5, 0.5)
	at.Translate(10, 10)
	a := NewArea(at.CreateTransformedShape(NewRectangle(0, 0, 100, 100)))
	a.Intersect(NewArea(at.CreateTransformedShape(NewRectangle(50, 60, 100, 100))))
	if !a.IsRectangular() {
		t.Fatalf("an area of transformed rectangles is not rectangular")
	}
	if got := a.GetBounds2D(); !got.Equals(NewRectangle2DDouble(30, 35, 25, 20)) {
		t.Errorf("GetBounds2D = %v", got)
	}
	inv, err := at.CreateInverse()
	if err != nil {
		t.Fatal(err)
	}
	// This is ITextOutputDevice.getClip.
	back, ok := inv.CreateTransformedShape(a).(*Area)
	if !ok {
		t.Fatalf("CreateTransformedShape of an *Area did not return an *Area")
	}
	if got := back.GetBounds(); !got.Equals(NewRectangle(50, 60, 50, 40)) {
		t.Errorf("GetBounds after the inverse transform = %v", got)
	}
	if len(a.Shapes()) != 2 || len(back.Shapes()) != 2 {
		t.Errorf("member counts = %d, %d", len(a.Shapes()), len(back.Shapes()))
	}
	if !back.Intersects(NewRectangle(90, 90, 20, 20)) || back.Intersects(NewRectangle(0, 0, 50, 60)) {
		t.Errorf("Intersects after the inverse transform is wrong")
	}

	rotated := AffineTransformGetRotateInstance(0.5).CreateTransformedShape(NewRectangle(0, 0, 10, 10))
	if NewArea(rotated).IsRectangular() {
		t.Errorf("a rotated rectangle is reported as rectangular")
	}
	quadrant := AffineTransformGetRotateInstance(-1.5707963267948966).CreateTransformedShape(NewRectangle(0, 0, 10, 20))
	if !NewArea(quadrant).IsRectangular() {
		t.Errorf("a rectangle rotated by 90 degrees is not reported as rectangular")
	}
}

func TestAreaOfSeveralCurvedShapes(t *testing.T) {
	a := NewArea(NewEllipse2DDouble(0, 0, 20, 20))
	a.Intersect(NewArea(NewRectangle(10, 0, 30, 30)))
	if got := a.GetBounds(); !got.Equals(NewRectangle(10, 0, 10, 20)) {
		t.Errorf("GetBounds = %v", got)
	}
	if !a.Intersects(NewRectangle(12, 8, 2, 2)) {
		t.Errorf("a rectangle inside both members does not intersect")
	}
	if a.Intersects(NewRectangle(2, 8, 2, 2)) {
		t.Errorf("a rectangle outside the rectangle member intersects")
	}
	if a.Intersects(NewRectangle(19, 0, 1, 1)) {
		t.Errorf("a rectangle outside the ellipse member intersects")
	}
	defer func() {
		e, ok := recover().(*UnsupportedOperationException)
		if !ok || !strings.Contains(e.Message, "Shapes()") {
			t.Errorf("panic = %v", e)
		}
	}()
	a.GetPathIterator(nil)
}

func TestNewAreaCopiesAnArea(t *testing.T) {
	a := NewArea(NewRectangle(0, 0, 10, 10))
	b := NewArea(a)
	b.Intersect(NewArea(NewRectangle(5, 5, 10, 10)))
	if len(a.Shapes()) != 1 || len(b.Shapes()) != 2 {
		t.Errorf("member counts = %d, %d", len(a.Shapes()), len(b.Shapes()))
	}
	defer func() {
		if _, ok := recover().(*IllegalArgumentException); !ok {
			t.Errorf("NewArea(nil) did not panic with *IllegalArgumentException")
		}
	}()
	NewArea(nil)
}

// The results below are what OpenJDK 25's Area produces for the same
// operations.
func TestAreaReducesAClippedRoundedRectangleLikeTheJDK(t *testing.T) {
	// A rounded rectangle far taller than the clip: its corners lie outside,
	// so the intersection is a rectangle.
	tall := NewPath2DFloat()
	tall.Append(NewRoundRectangle2DForTest(10, -1000, 300, 3000, 8, 8), false)
	a := NewArea(tall)
	a.Intersect(NewArea(NewRectangle(0, 0, 400, 500)))
	if !a.IsRectangular() {
		t.Fatalf("a clipped rounded rectangle whose corners are outside is not rectangular")
	}
	assertSegments(t, "clipped", collectDouble(a.GetPathIterator(nil)), []seg{
		{0, 10, 0}, {1, 10, 500}, {1, 310, 500}, {1, 310, 0}, {4},
	}, 0)

	// A float outline that ends a little away from its start, as
	// BorderPainter's does, is still convex.
	jitter := NewPath2DFloat()
	jitter.MoveTo(206.08984375, 527.384765625)
	jitter.CurveTo(218.47216796875, 497.4921875, 247.64306640625, 478, 280, 478)
	jitter.LineTo(14120, 478)
	jitter.CurveTo(14152.357421875, 478, 14193.91015625, 497.4921875, 14200, 558)
	jitter.LineTo(14200, 49194)
	jitter.CurveTo(14200, 49215.21875, 14141.216796875, 49274, 14120, 49274)
	jitter.LineTo(280, 49274)
	jitter.CurveTo(258.7822265625, 49274, 200, 49215.21875, 200, 49194)
	jitter.LineTo(200, 558)
	jitter.CurveTo(200, 547.49609375, 202.0693359375, 537.08984375, 206.08984375, 527.38671875)
	j := NewArea(jitter)
	j.Intersect(NewArea(NewRectangle(0, 19200, 14400, 19199)))
	if !j.IsRectangular() {
		t.Errorf("a rounded border whose outline does not close exactly is not reduced to a rectangle")
	}

	// With the corners inside the clip, the curves are part of the outline.
	inside := NewPath2DFloat()
	inside.Append(NewRoundRectangle2DForTest(10, 10, 100, 100, 20, 20), false)
	b := NewArea(inside)
	b.Intersect(NewArea(NewRectangle(0, 0, 400, 500)))
	if b.IsRectangular() {
		t.Errorf("a rounded rectangle inside the clip is reported as rectangular")
	}

	// A rectangle inside a rounded corner's cut-out stays conservative.
	corner := NewPath2DFloat()
	corner.Append(NewRoundRectangle2DForTest(0, 0, 100, 100, 40, 40), false)
	c := NewArea(corner)
	c.Intersect(NewArea(NewRectangle(0, 0, 3, 3)))
	if c.IsRectangular() {
		t.Errorf("a rectangle in the corner cut-out is reported as rectangular")
	}
}

// NewRoundRectangle2DForTest builds a rounded rectangle outline from lines and
// cubic quarter circles, the shape BorderPainter produces for a border with a
// radius.
func NewRoundRectangle2DForTest(x, y, w, h, arcW, arcH float64) Shape {
	rx, ry := arcW/2, arcH/2
	const k = 0.5522847498307933
	p := NewPath2DDouble()
	p.MoveTo(x+rx, y)
	p.LineTo(x+w-rx, y)
	p.CurveTo(x+w-rx+k*rx, y, x+w, y+ry-k*ry, x+w, y+ry)
	p.LineTo(x+w, y+h-ry)
	p.CurveTo(x+w, y+h-ry+k*ry, x+w-rx+k*rx, y+h, x+w-rx, y+h)
	p.LineTo(x+rx, y+h)
	p.CurveTo(x+rx-k*rx, y+h, x, y+h-ry+k*ry, x, y+h-ry)
	p.LineTo(x, y+ry)
	p.CurveTo(x, y+ry-k*ry, x+rx-k*rx, y, x+rx, y)
	p.ClosePath()
	return p
}
