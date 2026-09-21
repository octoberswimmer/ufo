package geom

import "math"

// Area stands in for java.awt.geom.Area. Flying Saucer uses Area only to
// intersect clip shapes, so an Area here is the intersection of a list of
// member shapes, kept as that list. It does not compute the outline of the
// intersection of curved shapes. An output device clips to the area by
// clipping to each of Shapes() in turn, since successive clips intersect.
//
// The JDK operations add, subtract and exclusiveOr are not provided; Flying
// Saucer does not call them.
//
// Where every member is an axis-aligned rectangle (a Rectangle, a Rectangle2D,
// or a Path2D that traces one, which is what
// AffineTransform.CreateTransformedShape makes of a Rectangle), the
// intersection is computed exactly and the area behaves as the JDK's does.
// Otherwise GetBounds2D, IsEmpty and Intersects work from the members'
// bounds and individual intersection tests. They never report a smaller area
// than the true intersection, but they can report a larger one.
type Area struct {
	shapes []Shape
}

// NewArea is new Area(Shape s). An *Area argument is copied. A nil shape
// panics, as it throws in the JDK.
func NewArea(s Shape) *Area {
	if s == nil {
		panic(&IllegalArgumentException{Message: "geom: NewArea of a nil Shape"})
	}
	if other, ok := s.(*Area); ok {
		return other.Clone()
	}
	return &Area{shapes: []Shape{s}}
}

// Intersect makes this area the intersection of itself and rhs by appending
// the member shapes of rhs.
func (a *Area) Intersect(rhs *Area) {
	a.shapes = append(a.shapes, rhs.shapes...)
}

// Shapes returns the member shapes whose intersection is this area.
func (a *Area) Shapes() []Shape {
	return a.shapes
}

// Clone returns an area with a copy of the member list. The member shapes
// themselves are shared.
func (a *Area) Clone() *Area {
	return &Area{shapes: append([]Shape(nil), a.shapes...)}
}

// rectangularShape returns the rectangle s traces when s is a Rectangle, a
// Rectangle2D, or a Path2D made of a move, three or four lines and an optional
// close whose points are the corners of an axis-aligned rectangle.
func rectangularShape(s Shape) (*Rectangle2D, bool) {
	switch v := s.(type) {
	case *Rectangle:
		return v.GetBounds2D(), true
	case *Rectangle2D:
		return v.GetBounds2D(), true
	case *Path2D:
		types := v.pointTypes
		n := len(types)
		if n > 0 && types[n-1] == PathIteratorSegClose {
			n--
		}
		if n != 4 && n != 5 {
			return nil, false
		}
		if types[0] != PathIteratorSegMoveto {
			return nil, false
		}
		for _, t := range types[1:n] {
			if t != PathIteratorSegLineto {
				return nil, false
			}
		}
		c := v.coords
		if n == 5 && (c[8] != c[0] || c[9] != c[1]) {
			return nil, false
		}
		horizontalFirst := c[1] == c[3] && c[2] == c[4] && c[5] == c[7] && c[6] == c[0]
		verticalFirst := c[0] == c[2] && c[3] == c[5] && c[4] == c[6] && c[7] == c[1]
		if !horizontalFirst && !verticalFirst {
			return nil, false
		}
		r := &Rectangle2D{}
		r.SetFrameFromDiagonal(c[0], c[1], c[4], c[5])
		return r, true
	}
	return nil, false
}

// rectangleIntersection returns the intersection of the members when every
// member is an axis-aligned rectangle. The result can have a zero or negative
// width or height, which means the area is empty.
func (a *Area) rectangleIntersection() (*Rectangle2D, bool) {
	var result *Rectangle2D
	var others []Shape
	for _, s := range a.shapes {
		r, ok := rectangularShape(s)
		if !ok {
			others = append(others, s)
			continue
		}
		if result == nil {
			result = r
		} else {
			Rectangle2DIntersect(result, r, result)
		}
	}
	if result == nil {
		return nil, false
	}
	// A convex member whose curved and slanted edges all lie outside the
	// rectangle meets the rectangle only along its straight axis-aligned
	// edges, which lie on its bounding box; within the rectangle it is its
	// bounding box. This is the case of a rounded border clipped to a page:
	// the JDK's Area reduces it to a rectangle, and so does this.
	for _, s := range others {
		if result.IsEmpty() || !convexWithCurvesOutside(s, result) {
			return nil, false
		}
		Rectangle2DIntersect(result, s.GetBounds2D(), result)
	}
	return result, true
}

// convexWithCurvesOutside reports whether s is convex (its outline, taken
// with the control points of its curves, turns one way only, and it has one
// subpath) and every segment of its outline other than a horizontal or
// vertical line lies outside the interior of r, judged by the bounding box of
// the segment's points, which contains the segment.
func convexWithCurvesOutside(s Shape, r *Rectangle2D) bool {
	it := s.GetPathIterator(nil)
	coords := make([]float64, 6)
	var points [][2]float64
	var curX, curY float64
	subpaths := 0
	outside := func(xs, ys []float64) bool {
		minX, maxX, minY, maxY := xs[0], xs[0], ys[0], ys[0]
		for i := range xs {
			minX, maxX = math.Min(minX, xs[i]), math.Max(maxX, xs[i])
			minY, maxY = math.Min(minY, ys[i]), math.Max(maxY, ys[i])
		}
		// Touching the rectangle's edge is not entering its interior.
		return maxX <= r.X || minX >= r.X+r.Width || maxY <= r.Y || minY >= r.Y+r.Height
	}
	for ; !it.IsDone(); it.Next() {
		switch it.CurrentSegmentDouble(coords) {
		case PathIteratorSegMoveto:
			subpaths++
			if subpaths > 1 {
				return false
			}
			curX, curY = coords[0], coords[1]
			points = append(points, [2]float64{curX, curY})
		case PathIteratorSegLineto:
			x, y := coords[0], coords[1]
			if x != curX && y != curY && !outside([]float64{curX, x}, []float64{curY, y}) {
				return false
			}
			curX, curY = x, y
			points = append(points, [2]float64{x, y})
		case PathIteratorSegQuadto:
			if !outside([]float64{curX, coords[0], coords[2]}, []float64{curY, coords[1], coords[3]}) {
				return false
			}
			points = append(points, [2]float64{coords[0], coords[1]}, [2]float64{coords[2], coords[3]})
			curX, curY = coords[2], coords[3]
		case PathIteratorSegCubicto:
			if !outside([]float64{curX, coords[0], coords[2], coords[4]}, []float64{curY, coords[1], coords[3], coords[5]}) {
				return false
			}
			points = append(points, [2]float64{coords[0], coords[1]}, [2]float64{coords[2], coords[3]}, [2]float64{coords[4], coords[5]})
			curX, curY = coords[4], coords[5]
		case PathIteratorSegClose:
		}
	}
	return convexPolygon(points)
}

// convexPolygon reports whether the closed polygon through points turns in one
// direction only. Repeated points and straight runs are allowed. Points closer
// together than a millionth of the polygon's size count as one: a path built
// from float arcs, as BorderPainter builds a rounded border, ends a few
// thousandths away from where it started, and the closing edge that leaves
// would otherwise turn the wrong way.
func convexPolygon(points [][2]float64) bool {
	if len(points) == 0 {
		return false
	}
	minX, maxX, minY, maxY := points[0][0], points[0][0], points[0][1], points[0][1]
	for _, p := range points {
		minX, maxX = math.Min(minX, p[0]), math.Max(maxX, p[0])
		minY, maxY = math.Min(minY, p[1]), math.Max(maxY, p[1])
	}
	tolerance := 1e-6 * math.Max(maxX-minX, maxY-minY)
	near := func(a, b [2]float64) bool {
		return math.Abs(a[0]-b[0]) <= tolerance && math.Abs(a[1]-b[1]) <= tolerance
	}
	var dedup [][2]float64
	for _, p := range points {
		if len(dedup) == 0 || !near(dedup[len(dedup)-1], p) {
			dedup = append(dedup, p)
		}
	}
	for len(dedup) > 1 && near(dedup[0], dedup[len(dedup)-1]) {
		dedup = dedup[:len(dedup)-1]
	}
	n := len(dedup)
	if n < 3 {
		return false
	}
	sign := 0
	for i := 0; i < n; i++ {
		p, q, r := dedup[i], dedup[(i+1)%n], dedup[(i+2)%n]
		ax, ay := q[0]-p[0], q[1]-p[1]
		bx, by := r[0]-q[0], r[1]-q[1]
		cross := ax*by - ay*bx
		// Where a curve continues tangent to the next segment the turn is
		// zero, and float control points make it a tiny angle either way;
		// turns with a sine below 1e-4 (about 0.006 degrees) count as none.
		epsilon := 1e-4 * math.Hypot(ax, ay) * math.Hypot(bx, by)
		switch {
		case cross > epsilon:
			if sign < 0 {
				return false
			}
			sign = 1
		case cross < -epsilon:
			if sign > 0 {
				return false
			}
			sign = -1
		}
	}
	return sign != 0
}

// IsRectangular reports whether the area is an axis-aligned rectangle (or is
// empty): every member is one, or the members that are not are convex shapes
// whose curves lie outside the rectangle the others make.
func (a *Area) IsRectangular() bool {
	_, ok := a.rectangleIntersection()
	return ok
}

// boundsIntersection returns the intersection of the members' bounds.
func (a *Area) boundsIntersection() *Rectangle2D {
	if r, ok := a.rectangleIntersection(); ok {
		return r
	}
	var result *Rectangle2D
	for _, s := range a.shapes {
		b := s.GetBounds2D()
		if result == nil {
			result = b
		} else {
			Rectangle2DIntersect(result, b, result)
		}
	}
	if result == nil {
		return &Rectangle2D{}
	}
	return result
}

// IsEmpty reports whether the area encloses nothing. For members that are not
// all rectangles it reports whether the intersection of their bounds is empty.
func (a *Area) IsEmpty() bool {
	return a.boundsIntersection().IsEmpty()
}

// GetBounds2D returns the intersection of the members' bounds, or an all-zero
// rectangle when that intersection is empty, which is what the JDK returns for
// an empty area.
func (a *Area) GetBounds2D() *Rectangle2D {
	b := a.boundsIntersection()
	if b.IsEmpty() {
		return &Rectangle2D{}
	}
	return b
}

func (a *Area) GetBounds() *Rectangle {
	return a.GetBounds2D().GetBounds()
}

func (a *Area) Intersects(r *Rectangle) bool {
	return a.IntersectsWithXYWH(float64(r.X), float64(r.Y), float64(r.Width), float64(r.Height))
}

// IntersectsWithXYWH reports whether the rectangle overlaps the area. For
// members that are not all rectangles it reports true when the rectangle
// overlaps the intersection of the members' bounds and overlaps every member.
func (a *Area) IntersectsWithXYWH(x, y, w, h float64) bool {
	if !a.boundsIntersection().IntersectsWithXYWH(x, y, w, h) {
		return false
	}
	if a.IsRectangular() {
		return true
	}
	for _, s := range a.shapes {
		if !s.IntersectsWithXYWH(x, y, w, h) {
			return false
		}
	}
	return true
}

// GetPathIterator returns the member's iterator for an area with one member,
// and the iterator of the exact intersection rectangle (no segments when it is
// empty) for an area whose members are all rectangles. For any other area it
// panics: the outline of an intersection of curved shapes is not computed, and
// the caller must clip to each of Shapes() instead.
func (a *Area) GetPathIterator(at *AffineTransform) PathIterator {
	if r, ok := a.rectangleIntersection(); ok {
		if r.IsEmpty() {
			return emptyPathIterator(PathIteratorWindNonZero)
		}
		return areaRectanglePathIterator(r, at)
	}
	if len(a.shapes) == 1 {
		return a.shapes[0].GetPathIterator(at)
	}
	panic(&UnsupportedOperationException{Message: "geom: Area.GetPathIterator is not supported for an area of several shapes that are not all rectangles; use Shapes() and clip to each shape in turn"})
}

// areaRectanglePathIterator traces a rectangle the way the JDK's Area traces
// a rectangular area: from the top-left corner down the left side, along the
// bottom, up the right side, and closed, with no line back to the start.
func areaRectanglePathIterator(r *Rectangle2D, at *AffineTransform) PathIterator {
	path := NewPath2DDouble()
	path.MoveTo(r.X, r.Y)
	path.LineTo(r.X, r.Y+r.Height)
	path.LineTo(r.X+r.Width, r.Y+r.Height)
	path.LineTo(r.X+r.Width, r.Y)
	path.ClosePath()
	return path.GetPathIterator(at)
}

// Transform transforms every member shape.
func (a *Area) Transform(at *AffineTransform) {
	for i, s := range a.shapes {
		a.shapes[i] = at.CreateTransformedShape(s)
	}
}

// CreateTransformedArea returns a new area whose members are the transformed
// members of this one.
func (a *Area) CreateTransformedArea(at *AffineTransform) *Area {
	c := a.Clone()
	c.Transform(at)
	return c
}
