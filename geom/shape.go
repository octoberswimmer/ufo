package geom

import "math"

// Shape is java.awt.Shape, limited to the methods Flying Saucer calls.
//
// Intersects takes a *Rectangle because every caller in Flying Saucer passes
// one; the JDK method takes a Rectangle2D, which a Rectangle is.
// IntersectsWithXYWH is the JDK intersects(double x, double y, double w,
// double h).
type Shape interface {
	GetBounds() *Rectangle
	GetBounds2D() *Rectangle2D
	Intersects(r *Rectangle) bool
	IntersectsWithXYWH(x, y, w, h float64) bool
	GetPathIterator(at *AffineTransform) PathIterator
}

// PathIterator is java.awt.geom.PathIterator. CurrentSegmentFloat and
// CurrentSegmentDouble are the two currentSegment overloads; each writes up
// to six coordinates and returns the segment type.
type PathIterator interface {
	GetWindingRule() int
	IsDone() bool
	Next()
	CurrentSegmentFloat(coords []float32) int
	CurrentSegmentDouble(coords []float64) int
}

// Winding rules and segment types, with the JDK values. The segment types are
// declared under both the spelling the porting rules derive from the JDK name
// (SEG_MOVETO -> SegMoveto) and the camel case spelling (SegMoveTo).
const (
	PathIteratorWindEvenOdd = 0
	PathIteratorWindNonZero = 1

	PathIteratorSegMoveto  = 0
	PathIteratorSegLineto  = 1
	PathIteratorSegQuadto  = 2
	PathIteratorSegCubicto = 3
	PathIteratorSegClose   = 4

	PathIteratorSegMoveTo  = PathIteratorSegMoveto
	PathIteratorSegLineTo  = PathIteratorSegLineto
	PathIteratorSegQuadTo  = PathIteratorSegQuadto
	PathIteratorSegCubicTo = PathIteratorSegCubicto
)

// segmentPointCounts is the number of points each segment type carries.
var segmentPointCounts = [5]int{1, 1, 2, 3, 0}

// shapeIterator implements PathIterator for every shape in the package. The
// shape supplies the number of segments and a function that writes the
// untransformed coordinates of one segment.
type shapeIterator struct {
	windingRule int
	index       int
	count       int
	affine      *AffineTransform
	segment     func(index int, coords []float64) int
	outOfBounds string
}

func (it *shapeIterator) GetWindingRule() int {
	return it.windingRule
}

func (it *shapeIterator) IsDone() bool {
	return it.index >= it.count
}

func (it *shapeIterator) Next() {
	it.index++
}

func (it *shapeIterator) CurrentSegmentDouble(coords []float64) int {
	if it.IsDone() {
		panic(&NoSuchElementException{Message: it.outOfBounds})
	}
	segType := it.segment(it.index, coords)
	if it.affine != nil {
		it.affine.TransformDouble(coords, 0, coords, 0, segmentPointCounts[segType])
	}
	return segType
}

// CurrentSegmentFloat rounds the untransformed coordinates to float32 and then
// transforms them, rounding the results to float32, which is the order the JDK
// iterators use.
func (it *shapeIterator) CurrentSegmentFloat(coords []float32) int {
	if it.IsDone() {
		panic(&NoSuchElementException{Message: it.outOfBounds})
	}
	var tmp [6]float64
	segType := it.segment(it.index, tmp[:])
	n := segmentPointCounts[segType] * 2
	for i := 0; i < n; i++ {
		coords[i] = float32(tmp[i])
	}
	if it.affine != nil {
		it.affine.TransformFloat(coords, 0, coords, 0, n/2)
	}
	return segType
}

// emptyPathIterator returns an iterator with no segments.
func emptyPathIterator(windingRule int) PathIterator {
	return &shapeIterator{windingRule: windingRule, outOfBounds: "path iterator out of bounds"}
}

// rectanglePathIterator is the JDK RectIterator: a move, four lines (the last
// one back to the start) and a close. A rectangle with a negative width or
// height has no segments.
func rectanglePathIterator(x, y, w, h float64, at *AffineTransform) PathIterator {
	it := &shapeIterator{
		windingRule: PathIteratorWindNonZero,
		count:       6,
		affine:      at,
		outOfBounds: "rect iterator out of bounds",
	}
	if w < 0 || h < 0 {
		it.count = 0
	}
	it.segment = func(index int, coords []float64) int {
		if index == 5 {
			return PathIteratorSegClose
		}
		coords[0] = x
		coords[1] = y
		if index == 1 || index == 2 {
			coords[0] += w
		}
		if index == 2 || index == 3 {
			coords[1] += h
		}
		if index == 0 {
			return PathIteratorSegMoveto
		}
		return PathIteratorSegLineto
	}
	return it
}

// boundsOfDoubles is RectangularShape.getBounds and Rectangle2D.getBounds: the
// smallest integer rectangle that encloses the given one, or an empty
// rectangle at the origin when a dimension is negative.
func boundsOfDoubles(x, y, width, height float64) *Rectangle {
	if width < 0 || height < 0 {
		return &Rectangle{}
	}
	x1 := math.Floor(x)
	y1 := math.Floor(y)
	x2 := math.Ceil(x + width)
	y2 := math.Ceil(y + height)
	return &Rectangle{X: javaInt(x1), Y: javaInt(y1), Width: javaInt(x2 - x1), Height: javaInt(y2 - y1)}
}

var (
	_ Shape = (*Rectangle)(nil)
	_ Shape = (*Rectangle2D)(nil)
	_ Shape = (*Path2D)(nil)
	_ Shape = (*Line2D)(nil)
	_ Shape = (*Ellipse2D)(nil)
	_ Shape = (*Arc2D)(nil)
	_ Shape = (*Area)(nil)
)
