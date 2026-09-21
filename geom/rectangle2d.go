package geom

// Outcodes for Rectangle2D.Outcode, with the JDK values.
const (
	Rectangle2DOutLeft   = 1
	Rectangle2DOutTop    = 2
	Rectangle2DOutRight  = 4
	Rectangle2DOutBottom = 8
)

// Rectangle2D is java.awt.geom.Rectangle2D: a rectangle with floating point
// coordinates. It covers both Rectangle2D.Float and Rectangle2D.Double.
type Rectangle2D struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewRectangle2D is new Rectangle2D.Double(): all four fields zero.
func NewRectangle2D() *Rectangle2D {
	return &Rectangle2D{}
}

// NewRectangle2DDouble is new Rectangle2D.Double(x, y, w, h).
func NewRectangle2DDouble(x, y, w, h float64) *Rectangle2D {
	return &Rectangle2D{X: x, Y: y, Width: w, Height: h}
}

// NewRectangle2DFloat is new Rectangle2D.Float(x, y, w, h).
func NewRectangle2DFloat(x, y, w, h float32) *Rectangle2D {
	return &Rectangle2D{X: float64(x), Y: float64(y), Width: float64(w), Height: float64(h)}
}

func (r *Rectangle2D) GetX() float64 {
	return r.X
}

func (r *Rectangle2D) GetY() float64 {
	return r.Y
}

func (r *Rectangle2D) GetWidth() float64 {
	return r.Width
}

func (r *Rectangle2D) GetHeight() float64 {
	return r.Height
}

func (r *Rectangle2D) GetMinX() float64 {
	return r.X
}

func (r *Rectangle2D) GetMinY() float64 {
	return r.Y
}

func (r *Rectangle2D) GetMaxX() float64 {
	return r.X + r.Width
}

func (r *Rectangle2D) GetMaxY() float64 {
	return r.Y + r.Height
}

func (r *Rectangle2D) GetCenterX() float64 {
	return r.X + r.Width/2.0
}

func (r *Rectangle2D) GetCenterY() float64 {
	return r.Y + r.Height/2.0
}

// IsEmpty reports whether the width or the height is zero or negative.
func (r *Rectangle2D) IsEmpty() bool {
	return r.Width <= 0.0 || r.Height <= 0.0
}

func (r *Rectangle2D) SetRect(x, y, w, h float64) {
	r.X = x
	r.Y = y
	r.Width = w
	r.Height = h
}

// SetRectRectangle2D is setRect(Rectangle2D).
func (r *Rectangle2D) SetRectRectangle2D(o *Rectangle2D) {
	r.SetRect(o.X, o.Y, o.Width, o.Height)
}

// SetFrameFromDiagonal sets the rectangle from two opposite corners given in
// any order.
func (r *Rectangle2D) SetFrameFromDiagonal(x1, y1, x2, y2 float64) {
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	r.SetRect(x1, y1, x2-x1, y2-y1)
}

// Outcode reports where the point lies in relation to the rectangle as a
// combination of the Rectangle2DOut constants; 0 means inside or on an edge.
func (r *Rectangle2D) Outcode(x, y float64) int {
	out := 0
	if r.Width <= 0 {
		out |= Rectangle2DOutLeft | Rectangle2DOutRight
	} else if x < r.X {
		out |= Rectangle2DOutLeft
	} else if x > r.X+r.Width {
		out |= Rectangle2DOutRight
	}
	if r.Height <= 0 {
		out |= Rectangle2DOutTop | Rectangle2DOutBottom
	} else if y < r.Y {
		out |= Rectangle2DOutTop
	} else if y > r.Y+r.Height {
		out |= Rectangle2DOutBottom
	}
	return out
}

// IntersectsLine reports whether the line segment from (x1, y1) to (x2, y2)
// touches the rectangle.
func (r *Rectangle2D) IntersectsLine(x1, y1, x2, y2 float64) bool {
	out2 := r.Outcode(x2, y2)
	if out2 == 0 {
		return true
	}
	for {
		out1 := r.Outcode(x1, y1)
		if out1 == 0 {
			return true
		}
		if (out1 & out2) != 0 {
			return false
		}
		if (out1 & (Rectangle2DOutLeft | Rectangle2DOutRight)) != 0 {
			x := r.X
			if (out1 & Rectangle2DOutRight) != 0 {
				x += r.Width
			}
			y1 = y1 + (x-x1)*(y2-y1)/(x2-x1)
			x1 = x
		} else {
			y := r.Y
			if (out1 & Rectangle2DOutBottom) != 0 {
				y += r.Height
			}
			x1 = x1 + (y-y1)*(x2-x1)/(y2-y1)
			y1 = y
		}
	}
}

// Contains reports whether the point is inside. The left and top edges are
// inside, the right and bottom edges are not.
func (r *Rectangle2D) Contains(x, y float64) bool {
	x0 := r.X
	y0 := r.Y
	return x >= x0 && y >= y0 && x < x0+r.Width && y < y0+r.Height
}

// ContainsWithXYWH is contains(double x, double y, double w, double h).
func (r *Rectangle2D) ContainsWithXYWH(x, y, w, h float64) bool {
	if r.IsEmpty() || w <= 0 || h <= 0 {
		return false
	}
	x0 := r.X
	y0 := r.Y
	return x >= x0 && y >= y0 && (x+w) <= x0+r.Width && (y+h) <= y0+r.Height
}

func (r *Rectangle2D) Intersects(o *Rectangle) bool {
	return r.IntersectsWithXYWH(float64(o.X), float64(o.Y), float64(o.Width), float64(o.Height))
}

// IntersectsRectangle2D is intersects(Rectangle2D).
func (r *Rectangle2D) IntersectsRectangle2D(o *Rectangle2D) bool {
	return r.IntersectsWithXYWH(o.X, o.Y, o.Width, o.Height)
}

func (r *Rectangle2D) IntersectsWithXYWH(x, y, w, h float64) bool {
	if r.IsEmpty() || w <= 0 || h <= 0 {
		return false
	}
	x0 := r.X
	y0 := r.Y
	return x+w > x0 && y+h > y0 && x < x0+r.Width && y < y0+r.Height
}

// CreateIntersection returns the rectangle the two have in common. When they
// do not overlap the result has a negative width or height.
func (r *Rectangle2D) CreateIntersection(o *Rectangle2D) *Rectangle2D {
	dest := &Rectangle2D{}
	Rectangle2DIntersect(r, o, dest)
	return dest
}

// CreateUnion returns the smallest rectangle that contains both.
func (r *Rectangle2D) CreateUnion(o *Rectangle2D) *Rectangle2D {
	dest := &Rectangle2D{}
	Rectangle2DUnion(r, o, dest)
	return dest
}

// Rectangle2DIntersect is the static Rectangle2D.intersect(src1, src2, dest).
func Rectangle2DIntersect(src1, src2, dest *Rectangle2D) {
	x1 := max(src1.GetMinX(), src2.GetMinX())
	y1 := max(src1.GetMinY(), src2.GetMinY())
	x2 := min(src1.GetMaxX(), src2.GetMaxX())
	y2 := min(src1.GetMaxY(), src2.GetMaxY())
	dest.SetRect(x1, y1, x2-x1, y2-y1)
}

// Rectangle2DUnion is the static Rectangle2D.union(src1, src2, dest).
func Rectangle2DUnion(src1, src2, dest *Rectangle2D) {
	x1 := min(src1.GetMinX(), src2.GetMinX())
	y1 := min(src1.GetMinY(), src2.GetMinY())
	x2 := max(src1.GetMaxX(), src2.GetMaxX())
	y2 := max(src1.GetMaxY(), src2.GetMaxY())
	dest.SetFrameFromDiagonal(x1, y1, x2, y2)
}

// Add is add(double newx, double newy): the rectangle grows to contain the
// point.
func (r *Rectangle2D) Add(newx, newy float64) {
	x1 := min(r.GetMinX(), newx)
	x2 := max(r.GetMaxX(), newx)
	y1 := min(r.GetMinY(), newy)
	y2 := max(r.GetMaxY(), newy)
	r.SetRect(x1, y1, x2-x1, y2-y1)
}

// AddRectangle2D is add(Rectangle2D): the rectangle becomes the union of the
// two.
func (r *Rectangle2D) AddRectangle2D(o *Rectangle2D) {
	x1 := min(r.GetMinX(), o.GetMinX())
	x2 := max(r.GetMaxX(), o.GetMaxX())
	y1 := min(r.GetMinY(), o.GetMinY())
	y2 := max(r.GetMaxY(), o.GetMaxY())
	r.SetRect(x1, y1, x2-x1, y2-y1)
}

// GetBounds returns the smallest integer rectangle that encloses this one, or
// an empty rectangle at the origin when a dimension is negative.
func (r *Rectangle2D) GetBounds() *Rectangle {
	return boundsOfDoubles(r.X, r.Y, r.Width, r.Height)
}

// GetBounds2D returns a copy of the rectangle.
func (r *Rectangle2D) GetBounds2D() *Rectangle2D {
	return &Rectangle2D{X: r.X, Y: r.Y, Width: r.Width, Height: r.Height}
}

func (r *Rectangle2D) GetPathIterator(at *AffineTransform) PathIterator {
	return rectanglePathIterator(r.X, r.Y, r.Width, r.Height, at)
}

func (r *Rectangle2D) Clone() *Rectangle2D {
	return r.GetBounds2D()
}

func (r *Rectangle2D) Equals(o *Rectangle2D) bool {
	return o != nil && r.X == o.X && r.Y == o.Y && r.Width == o.Width && r.Height == o.Height
}

func (r *Rectangle2D) String() string {
	return "java.awt.geom.Rectangle2D$Double[x=" + javaDoubleString(r.X) +
		",y=" + javaDoubleString(r.Y) +
		",w=" + javaDoubleString(r.Width) +
		",h=" + javaDoubleString(r.Height) + "]"
}
