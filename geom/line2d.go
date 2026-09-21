package geom

// Line2D is java.awt.geom.Line2D: a line segment. It covers Line2D.Float and
// Line2D.Double.
type Line2D struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
}

// NewLine2DDouble is new Line2D.Double(x1, y1, x2, y2).
func NewLine2DDouble(x1, y1, x2, y2 float64) *Line2D {
	return &Line2D{X1: x1, Y1: y1, X2: x2, Y2: y2}
}

// NewLine2DFloat is new Line2D.Float(x1, y1, x2, y2).
func NewLine2DFloat(x1, y1, x2, y2 float32) *Line2D {
	return &Line2D{X1: float64(x1), Y1: float64(y1), X2: float64(x2), Y2: float64(y2)}
}

func (l *Line2D) GetX1() float64 {
	return l.X1
}

func (l *Line2D) GetY1() float64 {
	return l.Y1
}

func (l *Line2D) GetX2() float64 {
	return l.X2
}

func (l *Line2D) GetY2() float64 {
	return l.Y2
}

func (l *Line2D) GetP1() *Point2D {
	return &Point2D{X: l.X1, Y: l.Y1}
}

func (l *Line2D) GetP2() *Point2D {
	return &Point2D{X: l.X2, Y: l.Y2}
}

func (l *Line2D) SetLine(x1, y1, x2, y2 float64) {
	l.X1 = x1
	l.Y1 = y1
	l.X2 = x2
	l.Y2 = y2
}

func (l *Line2D) GetBounds2D() *Rectangle2D {
	var x, y, w, h float64
	if l.X1 < l.X2 {
		x = l.X1
		w = l.X2 - l.X1
	} else {
		x = l.X2
		w = l.X1 - l.X2
	}
	if l.Y1 < l.Y2 {
		y = l.Y1
		h = l.Y2 - l.Y1
	} else {
		y = l.Y2
		h = l.Y1 - l.Y2
	}
	return &Rectangle2D{X: x, Y: y, Width: w, Height: h}
}

func (l *Line2D) GetBounds() *Rectangle {
	return l.GetBounds2D().GetBounds()
}

// Intersects reports whether the segment touches the rectangle.
func (l *Line2D) Intersects(r *Rectangle) bool {
	return r.GetBounds2D().IntersectsLine(l.X1, l.Y1, l.X2, l.Y2)
}

func (l *Line2D) IntersectsWithXYWH(x, y, w, h float64) bool {
	return (&Rectangle2D{X: x, Y: y, Width: w, Height: h}).IntersectsLine(l.X1, l.Y1, l.X2, l.Y2)
}

// Contains is always false: a line has no area.
func (l *Line2D) Contains(x, y float64) bool {
	return false
}

// GetPathIterator returns a move to the first point and a line to the second.
func (l *Line2D) GetPathIterator(at *AffineTransform) PathIterator {
	return &shapeIterator{
		windingRule: PathIteratorWindNonZero,
		count:       2,
		affine:      at,
		outOfBounds: "line iterator out of bounds",
		segment: func(index int, coords []float64) int {
			if index == 0 {
				coords[0] = l.X1
				coords[1] = l.Y1
				return PathIteratorSegMoveto
			}
			coords[0] = l.X2
			coords[1] = l.Y2
			return PathIteratorSegLineto
		},
	}
}

func (l *Line2D) Clone() *Line2D {
	c := *l
	return &c
}
