package geom

import (
	"math"
	"strconv"
)

// Arc closure types, with the JDK values.
const (
	// Arc2DOpen leaves the arc without a closing segment.
	Arc2DOpen = 0
	// Arc2DChord closes the arc with a line from its end to its start.
	Arc2DChord = 1
	// Arc2DPie closes the arc with lines from its end to the center and from
	// the center to its start.
	Arc2DPie = 2
)

// Arc2D is java.awt.geom.Arc2D: a part of the ellipse inscribed in a framing
// rectangle. Angles are in degrees; 0 is at 3 o'clock and a positive extent
// goes counter-clockwise on screen (toward negative Y). It covers Arc2D.Float
// and Arc2D.Double.
type Arc2D struct {
	X       float64
	Y       float64
	Width   float64
	Height  float64
	Start   float64
	Extent  float64
	arcType int
}

// NewArc2DDouble is new Arc2D.Double(x, y, w, h, start, extent, type).
func NewArc2DDouble(x, y, w, h, start, extent float64, arcType int) *Arc2D {
	a := &Arc2D{X: x, Y: y, Width: w, Height: h, Start: start, Extent: extent}
	a.SetArcType(arcType)
	return a
}

// NewArc2DFloat is new Arc2D.Float(x, y, w, h, start, extent, type).
func NewArc2DFloat(x, y, w, h, start, extent float32, arcType int) *Arc2D {
	return NewArc2DDouble(float64(x), float64(y), float64(w), float64(h), float64(start), float64(extent), arcType)
}

func (a *Arc2D) GetX() float64 {
	return a.X
}

func (a *Arc2D) GetY() float64 {
	return a.Y
}

func (a *Arc2D) GetWidth() float64 {
	return a.Width
}

func (a *Arc2D) GetHeight() float64 {
	return a.Height
}

func (a *Arc2D) GetAngleStart() float64 {
	return a.Start
}

func (a *Arc2D) GetAngleExtent() float64 {
	return a.Extent
}

func (a *Arc2D) GetArcType() int {
	return a.arcType
}

func (a *Arc2D) SetArcType(arcType int) {
	if arcType < Arc2DOpen || arcType > Arc2DPie {
		panic(&IllegalArgumentException{Message: "invalid type for Arc: " + strconv.Itoa(arcType)})
	}
	a.arcType = arcType
}

func (a *Arc2D) SetArc(x, y, w, h, angSt, angExt float64, closure int) {
	a.SetArcType(closure)
	a.X = x
	a.Y = y
	a.Width = w
	a.Height = h
	a.Start = angSt
	a.Extent = angExt
}

func (a *Arc2D) IsEmpty() bool {
	return a.Width <= 0.0 || a.Height <= 0.0
}

// GetStartPoint returns the point where the arc starts.
func (a *Arc2D) GetStartPoint() *Point2D {
	angle := toRadians(-a.Start)
	x := a.X + (math.Cos(angle)*0.5+0.5)*a.Width
	y := a.Y + (math.Sin(angle)*0.5+0.5)*a.Height
	return &Point2D{X: x, Y: y}
}

// GetEndPoint returns the point where the arc ends.
func (a *Arc2D) GetEndPoint() *Point2D {
	angle := toRadians(-a.Start - a.Extent)
	x := a.X + (math.Cos(angle)*0.5+0.5)*a.Width
	y := a.Y + (math.Sin(angle)*0.5+0.5)*a.Height
	return &Point2D{X: x, Y: y}
}

// normalizeDegrees brings an angle into the range (-180, 180].
func normalizeDegrees(angle float64) float64 {
	if angle > 180.0 {
		if angle <= (180.0 + 360.0) {
			angle = angle - 360.0
		} else {
			angle = math.Remainder(angle, 360.0)
			// IEEEremainder can return -180 here for some input values...
			if angle == -180.0 {
				angle = 180.0
			}
		}
	} else if angle <= -180.0 {
		if angle > (-180.0 - 360.0) {
			angle = angle + 360.0
		} else {
			angle = math.Remainder(angle, 360.0)
			// IEEEremainder can return -180 here for some input values...
			if angle == -180.0 {
				angle = 180.0
			}
		}
	}
	return angle
}

// ContainsAngle reports whether the angle, in degrees, is within the extent
// of the arc.
func (a *Arc2D) ContainsAngle(angle float64) bool {
	angExt := a.Extent
	backwards := angExt < 0.0
	if backwards {
		angExt = -angExt
	}
	if angExt >= 360.0 {
		return true
	}
	angle = normalizeDegrees(angle) - normalizeDegrees(a.Start)
	if backwards {
		angle = -angle
	}
	if angle < 0.0 {
		angle += 360.0
	}
	return angle >= 0.0 && angle < angExt
}

// GetBounds2D returns the bounds of the arc itself, not of its framing
// rectangle: the start and end points, the points at multiples of 90 degrees
// that the arc passes through and, for a pie, the center.
func (a *Arc2D) GetBounds2D() *Rectangle2D {
	if a.IsEmpty() {
		return &Rectangle2D{X: a.X, Y: a.Y, Width: a.Width, Height: a.Height}
	}
	var x1, y1, x2, y2 float64
	if a.arcType == Arc2DPie {
		x1, y1, x2, y2 = 0.0, 0.0, 0.0, 0.0
	} else {
		x1, y1 = 1.0, 1.0
		x2, y2 = -1.0, -1.0
	}
	angle := 0.0
	for i := 0; i < 6; i++ {
		if i < 4 {
			// 0-3 are the four quadrants
			angle += 90.0
			if !a.ContainsAngle(angle) {
				continue
			}
		} else if i == 4 {
			// 4 is start angle
			angle = a.Start
		} else {
			// 5 is end angle
			angle += a.Extent
		}
		rads := toRadians(-angle)
		xe := math.Cos(rads)
		ye := math.Sin(rads)
		x1 = min(x1, xe)
		y1 = min(y1, ye)
		x2 = max(x2, xe)
		y2 = max(y2, ye)
	}
	w := a.Width
	h := a.Height
	x2 = (x2 - x1) * 0.5 * w
	y2 = (y2 - y1) * 0.5 * h
	x1 = a.X + (x1*0.5+0.5)*w
	y1 = a.Y + (y1*0.5+0.5)*h
	return &Rectangle2D{X: x1, Y: y1, Width: x2, Height: y2}
}

// GetBounds returns the integer bounds of the framing rectangle, which is what
// the JDK's RectangularShape.getBounds returns for an arc.
func (a *Arc2D) GetBounds() *Rectangle {
	return boundsOfDoubles(a.X, a.Y, a.Width, a.Height)
}

func (a *Arc2D) Intersects(r *Rectangle) bool {
	return a.IntersectsWithXYWH(float64(r.X), float64(r.Y), float64(r.Width), float64(r.Height))
}

// IntersectsWithXYWH tests the rectangle against the arc's path, with an open
// arc treated as closed by its chord. The JDK makes this test with closed-form
// computations on the ellipse instead of on the path's cubic curves, so
// results can differ for a rectangle within the curve approximation error of
// the arc's edge.
func (a *Arc2D) IntersectsWithXYWH(x, y, w, h float64) bool {
	return NewPath2DDoubleFromShape(a).IntersectsWithXYWH(x, y, w, h)
}

// arcBtan is the JDK's ArcIterator.btan: the distance of the cubic control
// points from the end points, along the tangents, for an arc of a unit circle
// spanning increment radians: 4/3 * tan(increment / 4).
func arcBtan(increment float64) float64 {
	increment /= 2.0
	return 4.0 / 3.0 * math.Sin(increment) / (1.0 + math.Cos(increment))
}

// GetPathIterator is the JDK's ArcIterator: a move to the start point, one
// cubic curve for each 90 degrees of extent or part of it (all curves span
// the same angle), and then nothing (open), a close (chord), or a line to the
// center and a close (pie). An arc with a negative width or height has no
// segments.
func (a *Arc2D) GetPathIterator(at *AffineTransform) PathIterator {
	w := a.Width / 2
	h := a.Height / 2
	x := a.X + w
	y := a.Y + h
	angStRad := -toRadians(a.Start)
	ext := -a.Extent
	var arcSegs int
	var increment, cv float64
	if ext >= 360.0 || ext <= -360 {
		arcSegs = 4
		increment = math.Pi / 2
		// btan(Math.PI / 2);
		cv = 0.5522847498307933
		if ext < 0 {
			increment = -increment
			cv = -cv
		}
	} else {
		arcSegs = javaInt(math.Ceil(math.Abs(ext) / 90.0))
		increment = toRadians(ext / float64(arcSegs))
		cv = arcBtan(increment)
		if cv == 0 {
			arcSegs = 0
		}
	}
	lineSegs := 0
	switch a.arcType {
	case Arc2DChord:
		lineSegs = 1
	case Arc2DPie:
		lineSegs = 2
	}
	it := &shapeIterator{
		windingRule: PathIteratorWindNonZero,
		count:       arcSegs + lineSegs + 1,
		affine:      at,
		outOfBounds: "arc iterator out of bounds",
	}
	if w < 0 || h < 0 {
		it.count = 0
	}
	it.segment = func(index int, coords []float64) int {
		angle := angStRad
		if index == 0 {
			coords[0] = x + float64(math.Cos(angle)*w)
			coords[1] = y + float64(math.Sin(angle)*h)
			return PathIteratorSegMoveto
		}
		if index > arcSegs {
			if index == arcSegs+lineSegs {
				return PathIteratorSegClose
			}
			coords[0] = x
			coords[1] = y
			return PathIteratorSegLineto
		}
		angle += float64(increment * float64(index-1))
		relx := math.Cos(angle)
		rely := math.Sin(angle)
		coords[0] = x + float64((relx-float64(cv*rely))*w)
		coords[1] = y + float64((rely+float64(cv*relx))*h)
		angle += increment
		relx = math.Cos(angle)
		rely = math.Sin(angle)
		coords[2] = x + float64((relx+float64(cv*rely))*w)
		coords[3] = y + float64((rely-float64(cv*relx))*h)
		coords[4] = x + float64(relx*w)
		coords[5] = y + float64(rely*h)
		return PathIteratorSegCubicto
	}
	return it
}

func (a *Arc2D) Clone() *Arc2D {
	c := *a
	return &c
}
