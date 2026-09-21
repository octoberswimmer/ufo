package geom

// Ellipse2D is java.awt.geom.Ellipse2D: the ellipse inscribed in a framing
// rectangle. It covers Ellipse2D.Float and Ellipse2D.Double.
type Ellipse2D struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewEllipse2DDouble is new Ellipse2D.Double(x, y, w, h).
func NewEllipse2DDouble(x, y, w, h float64) *Ellipse2D {
	return &Ellipse2D{X: x, Y: y, Width: w, Height: h}
}

// NewEllipse2DFloat is new Ellipse2D.Float(x, y, w, h).
func NewEllipse2DFloat(x, y, w, h float32) *Ellipse2D {
	return &Ellipse2D{X: float64(x), Y: float64(y), Width: float64(w), Height: float64(h)}
}

func (e *Ellipse2D) GetX() float64 {
	return e.X
}

func (e *Ellipse2D) GetY() float64 {
	return e.Y
}

func (e *Ellipse2D) GetWidth() float64 {
	return e.Width
}

func (e *Ellipse2D) GetHeight() float64 {
	return e.Height
}

func (e *Ellipse2D) IsEmpty() bool {
	return e.Width <= 0.0 || e.Height <= 0.0
}

func (e *Ellipse2D) SetFrame(x, y, w, h float64) {
	e.X = x
	e.Y = y
	e.Width = w
	e.Height = h
}

func (e *Ellipse2D) GetBounds2D() *Rectangle2D {
	return &Rectangle2D{X: e.X, Y: e.Y, Width: e.Width, Height: e.Height}
}

func (e *Ellipse2D) GetBounds() *Rectangle {
	return boundsOfDoubles(e.X, e.Y, e.Width, e.Height)
}

// Contains reports whether the point is inside the ellipse.
func (e *Ellipse2D) Contains(x, y float64) bool {
	// Normalize the coordinates compared to the ellipse
	// having a center at 0,0 and a radius of 0.5.
	ellw := e.Width
	if ellw <= 0.0 {
		return false
	}
	normx := (x-e.X)/ellw - 0.5
	ellh := e.Height
	if ellh <= 0.0 {
		return false
	}
	normy := (y-e.Y)/ellh - 0.5
	return (normx*normx + normy*normy) < 0.25
}

func (e *Ellipse2D) Intersects(r *Rectangle) bool {
	return e.IntersectsWithXYWH(float64(r.X), float64(r.Y), float64(r.Width), float64(r.Height))
}

// IntersectsWithXYWH reports whether the interior of the ellipse overlaps the
// interior of the rectangle.
func (e *Ellipse2D) IntersectsWithXYWH(x, y, w, h float64) bool {
	if w <= 0.0 || h <= 0.0 {
		return false
	}
	// Normalize the rectangular coordinates compared to the ellipse
	// having a center at 0,0 and a radius of 0.5.
	ellw := e.Width
	if ellw <= 0.0 {
		return false
	}
	normx0 := (x-e.X)/ellw - 0.5
	normx1 := normx0 + w/ellw
	ellh := e.Height
	if ellh <= 0.0 {
		return false
	}
	normy0 := (y-e.Y)/ellh - 0.5
	normy1 := normy0 + h/ellh
	// find nearest x (left edge, right edge, 0.0)
	// find nearest y (top edge, bottom edge, 0.0)
	// if nearest x,y is inside circle of radius 0.5, then intersects
	var nearx, neary float64
	if normx0 > 0.0 {
		// center to left of X extents
		nearx = normx0
	} else if normx1 < 0.0 {
		// center to right of X extents
		nearx = normx1
	}
	if normy0 > 0.0 {
		// center above Y extents
		neary = normy0
	} else if normy1 < 0.0 {
		// center below Y extents
		neary = normy1
	}
	return (nearx*nearx + neary*neary) < 0.25
}

// ellipseCtrlVal is the distance of a cubic control point from the end point
// of a 90 degree arc of a unit circle: 4/3 * tan(pi/8). It is a variable so
// that the two values derived from it are computed in float64 arithmetic, as
// the JDK computes them, and not as exact constants.
var ellipseCtrlVal = 0.5522847498307933

var (
	ellipsePcv = 0.5 + ellipseCtrlVal*0.5
	ellipseNcv = 0.5 - ellipseCtrlVal*0.5
)

// ellipseCtrlPts are the control points of the four cubic curves, as
// fractions of the framing rectangle, starting at the right-most point and
// proceeding through the bottom, left and top.
var ellipseCtrlPts = [4][6]float64{
	{1.0, ellipsePcv, ellipsePcv, 1.0, 0.5, 1.0},
	{ellipseNcv, 1.0, 0.0, ellipsePcv, 0.0, 0.5},
	{0.0, ellipseNcv, ellipseNcv, 0.0, 0.5, 0.0},
	{ellipsePcv, 0.0, 1.0, ellipseNcv, 1.0, 0.5},
}

// GetPathIterator returns a move to the right-most point, four cubic curves
// and a close. An ellipse with a negative width or height has no segments.
func (e *Ellipse2D) GetPathIterator(at *AffineTransform) PathIterator {
	x, y, w, h := e.X, e.Y, e.Width, e.Height
	it := &shapeIterator{
		windingRule: PathIteratorWindNonZero,
		count:       6,
		affine:      at,
		outOfBounds: "ellipse iterator out of bounds",
	}
	if w < 0 || h < 0 {
		it.count = 0
	}
	it.segment = func(index int, coords []float64) int {
		if index == 5 {
			return PathIteratorSegClose
		}
		if index == 0 {
			ctrls := ellipseCtrlPts[3]
			coords[0] = x + float64(ctrls[4]*w)
			coords[1] = y + float64(ctrls[5]*h)
			return PathIteratorSegMoveto
		}
		ctrls := ellipseCtrlPts[index-1]
		coords[0] = x + float64(ctrls[0]*w)
		coords[1] = y + float64(ctrls[1]*h)
		coords[2] = x + float64(ctrls[2]*w)
		coords[3] = y + float64(ctrls[3]*h)
		coords[4] = x + float64(ctrls[4]*w)
		coords[5] = y + float64(ctrls[5]*h)
		return PathIteratorSegCubicto
	}
	return it
}

func (e *Ellipse2D) Clone() *Ellipse2D {
	c := *e
	return &c
}
