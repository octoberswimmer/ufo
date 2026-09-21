package geom

import (
	"math"
	"strconv"
)

// Rectangle is java.awt.Rectangle: an area with integer coordinates. A
// rectangle whose width or height is zero or negative is empty. A rectangle
// whose width or height is negative is also non-existent: Union and Add treat
// it as absent, while a zero-size rectangle still contributes its location.
//
// The JDK computes with 32-bit ints and limits results to that range. The
// methods here limit results to the same range, so they return what the JDK
// returns for every input that fits a Java int.
type Rectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewRectangle is new Rectangle(x, y, width, height).
func NewRectangle(x, y, width, height int) *Rectangle {
	return &Rectangle{X: x, Y: y, Width: width, Height: height}
}

// NewRectangleEmpty is new Rectangle(): all four fields zero.
func NewRectangleEmpty() *Rectangle {
	return &Rectangle{}
}

// NewRectangleFromRectangle is new Rectangle(Rectangle).
func NewRectangleFromRectangle(r *Rectangle) *Rectangle {
	return &Rectangle{X: r.X, Y: r.Y, Width: r.Width, Height: r.Height}
}

// NewRectangleWidthHeight is new Rectangle(width, height), located at the
// origin.
func NewRectangleWidthHeight(width, height int) *Rectangle {
	return &Rectangle{Width: width, Height: height}
}

// NewRectangleFromPointDimension is new Rectangle(Point, Dimension).
func NewRectangleFromPointDimension(p *Point, d *Dimension) *Rectangle {
	return &Rectangle{X: p.X, Y: p.Y, Width: d.Width, Height: d.Height}
}

// NewRectangleFromPoint is new Rectangle(Point): zero size at the point.
func NewRectangleFromPoint(p *Point) *Rectangle {
	return &Rectangle{X: p.X, Y: p.Y}
}

// NewRectangleFromDimension is new Rectangle(Dimension), located at the
// origin.
func NewRectangleFromDimension(d *Dimension) *Rectangle {
	return &Rectangle{Width: d.Width, Height: d.Height}
}

func (r *Rectangle) GetX() float64 {
	return float64(r.X)
}

func (r *Rectangle) GetY() float64 {
	return float64(r.Y)
}

func (r *Rectangle) GetWidth() float64 {
	return float64(r.Width)
}

func (r *Rectangle) GetHeight() float64 {
	return float64(r.Height)
}

func (r *Rectangle) GetMinX() float64 {
	return r.GetX()
}

func (r *Rectangle) GetMinY() float64 {
	return r.GetY()
}

func (r *Rectangle) GetMaxX() float64 {
	return r.GetX() + r.GetWidth()
}

func (r *Rectangle) GetMaxY() float64 {
	return r.GetY() + r.GetHeight()
}

func (r *Rectangle) GetCenterX() float64 {
	return r.GetX() + r.GetWidth()/2.0
}

func (r *Rectangle) GetCenterY() float64 {
	return r.GetY() + r.GetHeight()/2.0
}

// GetBounds returns a copy of the rectangle.
func (r *Rectangle) GetBounds() *Rectangle {
	return &Rectangle{X: r.X, Y: r.Y, Width: r.Width, Height: r.Height}
}

func (r *Rectangle) GetBounds2D() *Rectangle2D {
	return &Rectangle2D{X: float64(r.X), Y: float64(r.Y), Width: float64(r.Width), Height: float64(r.Height)}
}

func (r *Rectangle) SetBounds(x, y, width, height int) {
	r.X = x
	r.Y = y
	r.Width = width
	r.Height = height
}

// SetBoundsRectangle is setBounds(Rectangle).
func (r *Rectangle) SetBoundsRectangle(o *Rectangle) {
	r.SetBounds(o.X, o.Y, o.Width, o.Height)
}

// SetRect is setRect(double, double, double, double): the smallest integer
// rectangle that encloses the given one, limited to the range of a Java int.
func (r *Rectangle) SetRect(x, y, width, height float64) {
	var newx, newy, neww, newh int
	if x > 2.0*math.MaxInt32 {
		// Too far in positive X direction to represent.
		newx = math.MaxInt32
		neww = -1
	} else {
		newx = rectangleClip(x, false)
		if width >= 0 {
			width += x - float64(newx)
		}
		neww = rectangleClip(width, width >= 0)
	}
	if y > 2.0*math.MaxInt32 {
		// Too far in positive Y direction to represent.
		newy = math.MaxInt32
		newh = -1
	} else {
		newy = rectangleClip(y, false)
		if height >= 0 {
			height += y - float64(newy)
		}
		newh = rectangleClip(height, height >= 0)
	}
	r.SetBounds(newx, newy, neww, newh)
}

// rectangleClip returns v rounded up or down and limited to the int range.
func rectangleClip(v float64, doceil bool) int {
	if v <= math.MinInt32 {
		return math.MinInt32
	}
	if v >= math.MaxInt32 {
		return math.MaxInt32
	}
	if doceil {
		return javaInt(math.Ceil(v))
	}
	return javaInt(math.Floor(v))
}

func (r *Rectangle) GetLocation() *Point {
	return &Point{X: r.X, Y: r.Y}
}

func (r *Rectangle) SetLocation(x, y int) {
	r.X = x
	r.Y = y
}

// SetLocationPoint is setLocation(Point).
func (r *Rectangle) SetLocationPoint(p *Point) {
	r.SetLocation(p.X, p.Y)
}

func (r *Rectangle) GetSize() *Dimension {
	return &Dimension{Width: r.Width, Height: r.Height}
}

func (r *Rectangle) SetSize(width, height int) {
	r.Width = width
	r.Height = height
}

// SetSizeDimension is setSize(Dimension).
func (r *Rectangle) SetSizeDimension(d *Dimension) {
	r.SetSize(d.Width, d.Height)
}

// Translate moves the rectangle by (dx, dy). When the new location does not
// fit a Java int, the location is set to the nearest bound and, for a
// non-negative size, the size is adjusted so that the far edge stays where the
// move put it.
func (r *Rectangle) Translate(dx, dy int) {
	r.X, r.Width = translateAxis(r.X, r.Width, dx)
	r.Y, r.Height = translateAxis(r.Y, r.Height, dy)
}

func translateAxis(pos, size, delta int) (int, int) {
	moved := int64(pos) + int64(delta)
	if moved >= math.MinInt32 && moved <= math.MaxInt32 {
		return int(moved), size
	}
	if int64(pos) < math.MinInt32 || int64(pos) > math.MaxInt32 {
		// The JDK cannot hold this location; move without range limiting.
		return pos + delta, size
	}
	// newv is the sum as Java's 32-bit addition produces it.
	newv := int32(moved)
	width := int32(size)
	if delta < 0 {
		// negative overflow
		// Only adjust width if it was valid (>= 0).
		if width >= 0 {
			// Conceptually the same as:
			// width += newv; newv = MIN_VALUE; width -= newv;
			width += newv - math.MinInt32
			// width may go negative if the right edge went past
			// MIN_VALUE, but it cannot overflow since it cannot
			// have moved more than MIN_VALUE and any non-negative
			// number + MIN_VALUE does not overflow.
		}
		return math.MinInt32, int(width)
	}
	// positive overflow
	if width >= 0 {
		// Conceptually the same as:
		// width += newv; newv = MAX_VALUE; width -= newv;
		width += newv - math.MaxInt32
		// With large widths and large displacements
		// we may overflow so we need to check it.
		if width < 0 {
			width = math.MaxInt32
		}
	}
	return math.MaxInt32, int(width)
}

// Contains reports whether the point (X, Y) is inside the rectangle. The
// left and top edges are inside, the right and bottom edges are not.
func (r *Rectangle) Contains(X, Y int) bool {
	w := r.Width
	h := r.Height
	if (w | h) < 0 {
		// At least one of the dimensions is negative...
		return false
	}
	// Note: if either dimension is zero, tests below must return false...
	x := r.X
	y := r.Y
	if X < x || Y < y {
		return false
	}
	w += x
	h += y
	//    overflow || intersect
	return (w < x || w > X) && (h < y || h > Y)
}

// ContainsPoint is contains(Point).
func (r *Rectangle) ContainsPoint(p *Point) bool {
	return r.Contains(p.X, p.Y)
}

// ContainsRectangle is contains(Rectangle).
func (r *Rectangle) ContainsRectangle(o *Rectangle) bool {
	return r.ContainsWithXYWH(o.X, o.Y, o.Width, o.Height)
}

// ContainsWithXYWH is contains(int X, int Y, int W, int H): whether this
// rectangle entirely contains the given one. An empty argument or an empty
// receiver gives false.
func (r *Rectangle) ContainsWithXYWH(X, Y, W, H int) bool {
	w := r.Width
	h := r.Height
	if (w | h | W | H) < 0 {
		// At least one of the dimensions is negative...
		return false
	}
	// Note: if any dimension is zero, tests below must return false...
	x := r.X
	y := r.Y
	if X < x || Y < y {
		return false
	}
	w += x
	W += X
	if W <= X {
		// X+W overflowed or W was zero, return false if...
		// either original w or W was zero or
		// x+w did not overflow or
		// the overflowed x+w is smaller than the overflowed X+W
		if w >= x || W > w {
			return false
		}
	} else {
		// X+W did not overflow and W was not zero, return false if...
		// original w was zero or
		// x+w did not overflow and x+w is smaller than X+W
		if w >= x && W > w {
			return false
		}
	}
	h += y
	H += Y
	if H <= Y {
		if h >= y || H > h {
			return false
		}
	} else {
		if h >= y && H > h {
			return false
		}
	}
	return true
}

// Intersects reports whether the two rectangles share an area. An empty
// rectangle intersects nothing.
func (r *Rectangle) Intersects(o *Rectangle) bool {
	tw := r.Width
	th := r.Height
	rw := o.Width
	rh := o.Height
	if rw <= 0 || rh <= 0 || tw <= 0 || th <= 0 {
		return false
	}
	tx := r.X
	ty := r.Y
	rx := o.X
	ry := o.Y
	rw += rx
	rh += ry
	tw += tx
	th += ty
	//      overflow || intersect
	return (rw < rx || rw > tx) &&
		(rh < ry || rh > ty) &&
		(tw < tx || tw > rx) &&
		(th < ty || th > ry)
}

// IntersectsWithXYWH is the Rectangle2D intersects(double, double, double,
// double) that Rectangle inherits.
func (r *Rectangle) IntersectsWithXYWH(x, y, w, h float64) bool {
	return r.GetBounds2D().IntersectsWithXYWH(x, y, w, h)
}

// Intersection returns the rectangle the two have in common. When they do not
// overlap the result has a zero or negative width or height, and IsEmpty
// reports true for it.
func (r *Rectangle) Intersection(o *Rectangle) *Rectangle {
	tx1 := int64(r.X)
	ty1 := int64(r.Y)
	rx1 := int64(o.X)
	ry1 := int64(o.Y)
	tx2 := tx1 + int64(r.Width)
	ty2 := ty1 + int64(r.Height)
	rx2 := rx1 + int64(o.Width)
	ry2 := ry1 + int64(o.Height)
	if tx1 < rx1 {
		tx1 = rx1
	}
	if ty1 < ry1 {
		ty1 = ry1
	}
	if tx2 > rx2 {
		tx2 = rx2
	}
	if ty2 > ry2 {
		ty2 = ry2
	}
	tx2 -= tx1
	ty2 -= ty1
	// tx2,ty2 will never overflow (they will never be
	// larger than the smallest of the two source w,h)
	// they might underflow, though...
	if tx2 < math.MinInt32 {
		tx2 = math.MinInt32
	}
	if ty2 < math.MinInt32 {
		ty2 = math.MinInt32
	}
	return &Rectangle{X: int(tx1), Y: int(ty1), Width: int(tx2), Height: int(ty2)}
}

// Union returns the smallest rectangle that contains both. A rectangle with a
// negative dimension is non-existent and the result is a copy of the other
// one. A rectangle with a zero dimension still takes part with its location.
func (r *Rectangle) Union(o *Rectangle) *Rectangle {
	tx2 := int64(r.Width)
	ty2 := int64(r.Height)
	if (tx2 | ty2) < 0 {
		// This rectangle has negative dimensions...
		// If r has non-negative dimensions then it is the answer.
		// If r is non-existent (has a negative dimension), then both
		// are non-existent and we can return any non-existent rectangle
		// as an answer.  Thus, returning r meets that criterion.
		// Either way, r is our answer.
		return NewRectangleFromRectangle(o)
	}
	rx2 := int64(o.Width)
	ry2 := int64(o.Height)
	if (rx2 | ry2) < 0 {
		return NewRectangleFromRectangle(r)
	}
	tx1 := int64(r.X)
	ty1 := int64(r.Y)
	tx2 += tx1
	ty2 += ty1
	rx1 := int64(o.X)
	ry1 := int64(o.Y)
	rx2 += rx1
	ry2 += ry1
	if tx1 > rx1 {
		tx1 = rx1
	}
	if ty1 > ry1 {
		ty1 = ry1
	}
	if tx2 < rx2 {
		tx2 = rx2
	}
	if ty2 < ry2 {
		ty2 = ry2
	}
	tx2 -= tx1
	ty2 -= ty1
	// tx2,ty2 will never underflow since both original rectangles
	// were already proven to be non-empty
	// they might overflow, though...
	if tx2 > math.MaxInt32 {
		tx2 = math.MaxInt32
	}
	if ty2 > math.MaxInt32 {
		ty2 = math.MaxInt32
	}
	return &Rectangle{X: int(tx1), Y: int(ty1), Width: int(tx2), Height: int(ty2)}
}

// Add is add(Rectangle): the rectangle becomes the union of the two.
func (r *Rectangle) Add(o *Rectangle) {
	tx2 := int64(r.Width)
	ty2 := int64(r.Height)
	if (tx2 | ty2) < 0 {
		r.SetBounds(o.X, o.Y, o.Width, o.Height)
	}
	rx2 := int64(o.Width)
	ry2 := int64(o.Height)
	if (rx2 | ry2) < 0 {
		return
	}
	tx1 := int64(r.X)
	ty1 := int64(r.Y)
	tx2 += tx1
	ty2 += ty1
	rx1 := int64(o.X)
	ry1 := int64(o.Y)
	rx2 += rx1
	ry2 += ry1
	if tx1 > rx1 {
		tx1 = rx1
	}
	if ty1 > ry1 {
		ty1 = ry1
	}
	if tx2 < rx2 {
		tx2 = rx2
	}
	if ty2 < ry2 {
		ty2 = ry2
	}
	tx2 -= tx1
	ty2 -= ty1
	// tx2,ty2 will never underflow since both original
	// rectangles were non-empty
	// they might overflow, though...
	if tx2 > math.MaxInt32 {
		tx2 = math.MaxInt32
	}
	if ty2 > math.MaxInt32 {
		ty2 = math.MaxInt32
	}
	r.SetBounds(int(tx1), int(ty1), int(tx2), int(ty2))
}

// AddXY is add(int newx, int newy): the rectangle grows to the smallest one
// that contains both the original rectangle and the point. A rectangle with a
// negative dimension becomes the zero-size rectangle at the point. A point
// added on the right or bottom edge is not inside the result, because Contains
// excludes those edges.
func (r *Rectangle) AddXY(newx, newy int) {
	if (r.Width | r.Height) < 0 {
		r.X = newx
		r.Y = newy
		r.Width = 0
		r.Height = 0
		return
	}
	x1 := int64(r.X)
	y1 := int64(r.Y)
	x2 := int64(r.Width)
	y2 := int64(r.Height)
	x2 += x1
	y2 += y1
	if x1 > int64(newx) {
		x1 = int64(newx)
	}
	if y1 > int64(newy) {
		y1 = int64(newy)
	}
	if x2 < int64(newx) {
		x2 = int64(newx)
	}
	if y2 < int64(newy) {
		y2 = int64(newy)
	}
	x2 -= x1
	y2 -= y1
	if x2 > math.MaxInt32 {
		x2 = math.MaxInt32
	}
	if y2 > math.MaxInt32 {
		y2 = math.MaxInt32
	}
	r.SetBounds(int(x1), int(y1), int(x2), int(y2))
}

// AddPoint is add(Point).
func (r *Rectangle) AddPoint(p *Point) {
	r.AddXY(p.X, p.Y)
}

// Grow moves the left and right edges outward by h and the top and bottom
// edges outward by v. Negative values shrink the rectangle, and the result can
// have negative dimensions.
func (r *Rectangle) Grow(h, v int) {
	x0 := int64(r.X)
	y0 := int64(r.Y)
	x1 := int64(r.Width)
	y1 := int64(r.Height)
	x1 += x0
	y1 += y0

	x0 -= int64(h)
	y0 -= int64(v)
	x1 += int64(h)
	y1 += int64(v)

	x0, x1 = growAxis(x0, x1)
	y0, y1 = growAxis(y0, y1)
	r.SetBounds(int(x0), int(y0), int(x1), int(y1))
}

// growAxis turns the two edges of one axis into a location and a size, each
// limited to the int range.
func growAxis(x0, x1 int64) (int64, int64) {
	if x1 < x0 {
		// Non-existent in X direction
		// Final width must remain negative so subtract x0 before
		// it is clipped so that we avoid the risk that the clipping
		// of x0 will reverse the ordering of x0 and x1.
		x1 -= x0
		if x1 < math.MinInt32 {
			x1 = math.MinInt32
		}
		return clipInt32(x0), x1
	}
	// Clip x0 before we subtract it from x1 in case the clipping
	// affects the representable area of the rectangle.
	x0 = clipInt32(x0)
	x1 -= x0
	// The only way x1 can be negative now is if we clipped
	// x0 against MIN and x1 is less than MIN - in which case
	// we want to leave the width negative since the result
	// did not intersect the representable area.
	return x0, clipInt32(x1)
}

// IsEmpty reports whether the width or the height is zero or negative.
func (r *Rectangle) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

func (r *Rectangle) GetPathIterator(at *AffineTransform) PathIterator {
	return rectanglePathIterator(float64(r.X), float64(r.Y), float64(r.Width), float64(r.Height), at)
}

func (r *Rectangle) Clone() *Rectangle {
	return NewRectangleFromRectangle(r)
}

func (r *Rectangle) Equals(o *Rectangle) bool {
	return o != nil && r.X == o.X && r.Y == o.Y && r.Width == o.Width && r.Height == o.Height
}

func (r *Rectangle) String() string {
	return "java.awt.Rectangle[x=" + strconv.Itoa(r.X) + ",y=" + strconv.Itoa(r.Y) +
		",width=" + strconv.Itoa(r.Width) + ",height=" + strconv.Itoa(r.Height) + "]"
}
