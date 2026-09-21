package geom

import "math"

// Winding rules, with the JDK values.
const (
	Path2DWindEvenOdd = PathIteratorWindEvenOdd
	Path2DWindNonZero = PathIteratorWindNonZero

	GeneralPathWindEvenOdd = PathIteratorWindEvenOdd
	GeneralPathWindNonZero = PathIteratorWindNonZero
)

// Path2D is java.awt.geom.Path2D: a path made of move, line, quadratic curve,
// cubic curve and close segments. It covers Path2D.Float and Path2D.Double.
// The Float variant rounds every coordinate it stores to float32, including
// the results of Transform, as the JDK class does.
type Path2D struct {
	pointTypes  []byte
	coords      []float64
	windingRule int
	single      bool
}

// GeneralPath is java.awt.geom.GeneralPath, which is a Path2D.Float.
type GeneralPath = Path2D

// NewPath2DFloat is new Path2D.Float(): an empty path with the non-zero
// winding rule.
func NewPath2DFloat() *Path2D {
	return &Path2D{windingRule: Path2DWindNonZero, single: true}
}

// NewPath2DFloatWithRule is new Path2D.Float(int rule).
func NewPath2DFloatWithRule(rule int) *Path2D {
	p := NewPath2DFloat()
	p.SetWindingRule(rule)
	return p
}

// NewPath2DFloatFromShape is new Path2D.Float(Shape s).
func NewPath2DFloatFromShape(s Shape) *Path2D {
	return NewPath2DFloatFromShapeWithAt(s, nil)
}

// NewPath2DFloatFromShapeWithAt is new Path2D.Float(Shape s, AffineTransform
// at): the path of s transformed by at, which may be nil.
func NewPath2DFloatFromShapeWithAt(s Shape, at *AffineTransform) *Path2D {
	p := NewPath2DFloat()
	p.initFromShape(s, at)
	return p
}

// NewPath2DDouble is new Path2D.Double(): an empty path with the non-zero
// winding rule.
func NewPath2DDouble() *Path2D {
	return &Path2D{windingRule: Path2DWindNonZero}
}

// NewPath2DDoubleWithRule is new Path2D.Double(int rule).
func NewPath2DDoubleWithRule(rule int) *Path2D {
	p := NewPath2DDouble()
	p.SetWindingRule(rule)
	return p
}

// NewPath2DDoubleFromShape is new Path2D.Double(Shape s).
func NewPath2DDoubleFromShape(s Shape) *Path2D {
	return NewPath2DDoubleFromShapeWithAt(s, nil)
}

// NewPath2DDoubleFromShapeWithAt is new Path2D.Double(Shape s,
// AffineTransform at): the path of s transformed by at, which may be nil.
func NewPath2DDoubleFromShapeWithAt(s Shape, at *AffineTransform) *Path2D {
	p := NewPath2DDouble()
	p.initFromShape(s, at)
	return p
}

// NewGeneralPath is new GeneralPath().
func NewGeneralPath() *GeneralPath {
	return NewPath2DFloat()
}

// NewGeneralPathWithRule is new GeneralPath(int rule).
func NewGeneralPathWithRule(rule int) *GeneralPath {
	return NewPath2DFloatWithRule(rule)
}

// NewGeneralPathFromShape is new GeneralPath(Shape s).
func NewGeneralPathFromShape(s Shape) *GeneralPath {
	return NewPath2DFloatFromShape(s)
}

func (p *Path2D) initFromShape(s Shape, at *AffineTransform) {
	if src, ok := s.(*Path2D); ok {
		p.windingRule = src.windingRule
		p.pointTypes = append([]byte(nil), src.pointTypes...)
		p.coords = append([]float64(nil), src.coords...)
		if at != nil {
			at.TransformDouble(p.coords, 0, p.coords, 0, len(p.coords)/2)
		}
		p.roundCoords(0)
		return
	}
	pi := s.GetPathIterator(at)
	p.SetWindingRule(pi.GetWindingRule())
	p.AppendPathIterator(pi, false)
}

// roundCoords rounds the coordinates from index start on to float32 when the
// path is the Float variant.
func (p *Path2D) roundCoords(start int) {
	if !p.single {
		return
	}
	for i := start; i < len(p.coords); i++ {
		p.coords[i] = float64(float32(p.coords[i]))
	}
}

func (p *Path2D) needRoom(needMove bool) {
	if len(p.pointTypes) == 0 && needMove {
		panic(&IllegalPathStateException{Message: "missing initial moveto in path definition"})
	}
}

func (p *Path2D) addSegment(segType byte, values ...float64) {
	start := len(p.coords)
	p.pointTypes = append(p.pointTypes, segType)
	p.coords = append(p.coords, values...)
	p.roundCoords(start)
}

// MoveTo starts a new subpath at (x, y). A move directly after another move
// replaces it.
func (p *Path2D) MoveTo(x, y float64) {
	n := len(p.pointTypes)
	if n > 0 && p.pointTypes[n-1] == PathIteratorSegMoveto {
		start := len(p.coords) - 2
		p.coords[start] = x
		p.coords[start+1] = y
		p.roundCoords(start)
		return
	}
	p.addSegment(PathIteratorSegMoveto, x, y)
}

// LineTo adds a line to (x, y). It panics with *IllegalPathStateException
// when the path has no initial MoveTo; so do QuadTo, CurveTo and ClosePath.
func (p *Path2D) LineTo(x, y float64) {
	p.needRoom(true)
	p.addSegment(PathIteratorSegLineto, x, y)
}

// QuadTo adds a quadratic curve with control point (x1, y1) to (x2, y2).
func (p *Path2D) QuadTo(x1, y1, x2, y2 float64) {
	p.needRoom(true)
	p.addSegment(PathIteratorSegQuadto, x1, y1, x2, y2)
}

// CurveTo adds a cubic curve with control points (x1, y1) and (x2, y2) to
// (x3, y3).
func (p *Path2D) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	p.needRoom(true)
	p.addSegment(PathIteratorSegCubicto, x1, y1, x2, y2, x3, y3)
}

// ClosePath closes the current subpath with a line back to its MoveTo. It
// does nothing when the path is already closed.
func (p *Path2D) ClosePath() {
	n := len(p.pointTypes)
	if n == 0 || p.pointTypes[n-1] != PathIteratorSegClose {
		p.needRoom(true)
		p.addSegment(PathIteratorSegClose)
	}
}

// Append is append(Shape s, boolean connect).
func (p *Path2D) Append(s Shape, connect bool) {
	p.AppendPathIterator(s.GetPathIterator(nil), connect)
}

// AppendPathIterator is append(PathIterator pi, boolean connect). With
// connect, an initial move of the appended path becomes a line from the
// current point, or is dropped when it goes to the current point and the path
// is not closed. The Float variant reads the iterator's float32 coordinates.
func (p *Path2D) AppendPathIterator(pi PathIterator, connect bool) {
	var coords [6]float64
	var fcoords [6]float32
	for !pi.IsDone() {
		var segType int
		if p.single {
			segType = pi.CurrentSegmentFloat(fcoords[:])
			for i, v := range fcoords {
				coords[i] = float64(v)
			}
		} else {
			segType = pi.CurrentSegmentDouble(coords[:])
		}
		switch segType {
		case PathIteratorSegMoveto:
			n := len(p.pointTypes)
			nc := len(p.coords)
			if !connect || n < 1 || nc < 1 {
				p.MoveTo(coords[0], coords[1])
				break
			}
			if p.pointTypes[n-1] != PathIteratorSegClose &&
				p.coords[nc-2] == coords[0] &&
				p.coords[nc-1] == coords[1] {
				// Collapse out initial moveto/lineto
				break
			}
			p.LineTo(coords[0], coords[1])
		case PathIteratorSegLineto:
			p.LineTo(coords[0], coords[1])
		case PathIteratorSegQuadto:
			p.QuadTo(coords[0], coords[1], coords[2], coords[3])
		case PathIteratorSegCubicto:
			p.CurveTo(coords[0], coords[1], coords[2], coords[3], coords[4], coords[5])
		case PathIteratorSegClose:
			p.ClosePath()
		}
		pi.Next()
		connect = false
	}
}

func (p *Path2D) GetWindingRule() int {
	return p.windingRule
}

func (p *Path2D) SetWindingRule(rule int) {
	if rule != Path2DWindEvenOdd && rule != Path2DWindNonZero {
		panic(&IllegalArgumentException{Message: "winding rule must be WIND_EVEN_ODD or WIND_NON_ZERO"})
	}
	p.windingRule = rule
}

// GetCurrentPoint returns the end point of the last segment, or nil for an
// empty path. After ClosePath it is the point of the subpath's MoveTo.
func (p *Path2D) GetCurrentPoint() *Point2D {
	index := len(p.coords)
	numTypes := len(p.pointTypes)
	if numTypes < 1 || index < 1 {
		return nil
	}
	if p.pointTypes[numTypes-1] == PathIteratorSegClose {
	loop:
		for i := numTypes - 2; i > 0; i-- {
			switch p.pointTypes[i] {
			case PathIteratorSegMoveto:
				break loop
			case PathIteratorSegLineto:
				index -= 2
			case PathIteratorSegQuadto:
				index -= 4
			case PathIteratorSegCubicto:
				index -= 6
			case PathIteratorSegClose:
			}
		}
	}
	return &Point2D{X: p.coords[index-2], Y: p.coords[index-1]}
}

// Reset empties the path. The winding rule is kept.
func (p *Path2D) Reset() {
	p.pointTypes = p.pointTypes[:0]
	p.coords = p.coords[:0]
}

// Transform transforms the path's coordinates in place.
func (p *Path2D) Transform(at *AffineTransform) {
	at.TransformDouble(p.coords, 0, p.coords, 0, len(p.coords)/2)
	p.roundCoords(0)
}

// CreateTransformedShape returns a transformed copy of the path, of the same
// variant.
func (p *Path2D) CreateTransformedShape(at *AffineTransform) Shape {
	c := p.Clone()
	if at != nil {
		c.Transform(at)
	}
	return c
}

func (p *Path2D) Clone() *Path2D {
	return &Path2D{
		pointTypes:  append([]byte(nil), p.pointTypes...),
		coords:      append([]float64(nil), p.coords...),
		windingRule: p.windingRule,
		single:      p.single,
	}
}

// GetBounds2D returns the bounds of the path: the end points of all segments,
// the points of all moves, and the extrema of the curves. An empty path gives
// an all-zero rectangle. This is the computation the JDK has used since
// version 19; earlier versions returned the bounds of the control points.
func (p *Path2D) GetBounds2D() *Rectangle2D {
	return pathBounds2D(p.GetPathIterator(nil))
}

// pathBounds2D is Path2D.getBounds2D(PathIterator).
func pathBounds2D(pi PathIterator) *Rectangle2D {
	var coeff [4]float64
	var derivCoeff [3]float64
	var coords [6]float64
	// bounds are stored as {leftX, rightX, topY, bottomY}
	var bounds [4]float64
	haveBounds := false
	lastX, lastY := 0.0, 0.0
	startX, startY := 0.0, 0.0
	for ; !pi.IsDone(); pi.Next() {
		var endX, endY float64
		segType := pi.CurrentSegmentDouble(coords[:])
		switch segType {
		case PathIteratorSegMoveto:
			if !haveBounds {
				bounds = [4]float64{coords[0], coords[0], coords[1], coords[1]}
				haveBounds = true
			}
			endX, endY = coords[0], coords[1]
			startX, startY = endX, endY
		case PathIteratorSegLineto:
			endX, endY = coords[0], coords[1]
		case PathIteratorSegQuadto:
			endX, endY = coords[2], coords[3]
		case PathIteratorSegCubicto:
			endX, endY = coords[4], coords[5]
		default:
			lastX, lastY = startX, startY
			continue
		}
		if endX < bounds[0] {
			bounds[0] = endX
		}
		if endX > bounds[1] {
			bounds[1] = endX
		}
		if endY < bounds[2] {
			bounds[2] = endY
		}
		if endY > bounds[3] {
			bounds[3] = endY
		}
		switch segType {
		case PathIteratorSegQuadto:
			accumulateExtremaBoundsForQuad(bounds[:], 0, lastX, coords[0], coords[2], coeff[:], derivCoeff[:])
			accumulateExtremaBoundsForQuad(bounds[:], 2, lastY, coords[1], coords[3], coeff[:], derivCoeff[:])
		case PathIteratorSegCubicto:
			accumulateExtremaBoundsForCubic(bounds[:], 0, lastX, coords[0], coords[2], coords[4], coeff[:], derivCoeff[:])
			accumulateExtremaBoundsForCubic(bounds[:], 2, lastY, coords[1], coords[3], coords[5], coeff[:], derivCoeff[:])
		}
		lastX, lastY = endX, endY
	}
	if haveBounds {
		return &Rectangle2D{X: bounds[0], Y: bounds[2], Width: bounds[1] - bounds[0], Height: bounds[3] - bounds[2]}
	}
	return &Rectangle2D{}
}

// ulp is java.lang.Math.ulp for a non-negative finite argument: the distance
// to the next larger float64.
func ulp(v float64) float64 {
	return math.Nextafter(v, math.Inf(1)) - v
}

// accumulateExtremaBoundsForQuad is
// sun.awt.geom.Curve.accumulateExtremaBoundsForQuad. It widens
// bounds[boundsOffset] and bounds[boundsOffset+1] to include the extremum of
// the quadratic curve in one dimension, plus a margin for the rounding error
// of the evaluation.
func accumulateExtremaBoundsForQuad(bounds []float64, boundsOffset int, x1, ctrlX, x2 float64, coeff, derivCoeff []float64) {
	if ctrlX < bounds[boundsOffset] || ctrlX > bounds[boundsOffset+1] {
		dx21 := ctrlX - x1
		coeff[2] = (x2 - ctrlX) - dx21 // A = P3 - P0 - 2 P2
		coeff[1] = 2.0 * dx21          // B = 2 (P2 - P1)
		coeff[0] = x1                  // C = P1

		derivCoeff[0] = coeff[1]
		derivCoeff[1] = 2.0 * coeff[2]

		t := -derivCoeff[0] / derivCoeff[1]
		if t > 0.0 && t < 1.0 {
			v := coeff[0] + float64(t*(coeff[1]+float64(t*coeff[2])))

			// error condition = sum ( abs (coeff) ):
			margin := ulp(math.Abs(coeff[0]) + math.Abs(coeff[1]) + math.Abs(coeff[2]))

			if v-margin < bounds[boundsOffset] {
				bounds[boundsOffset] = v - margin
			}
			if v+margin > bounds[boundsOffset+1] {
				bounds[boundsOffset+1] = v + margin
			}
		}
	}
}

// accumulateExtremaBoundsForCubic is
// sun.awt.geom.Curve.accumulateExtremaBoundsForCubic, the cubic counterpart of
// accumulateExtremaBoundsForQuad.
func accumulateExtremaBoundsForCubic(bounds []float64, boundsOffset int, x1, ctrlX1, ctrlX2, x2 float64, coeff, derivCoeff []float64) {
	if ctrlX1 < bounds[boundsOffset] || ctrlX1 > bounds[boundsOffset+1] ||
		ctrlX2 < bounds[boundsOffset] || ctrlX2 > bounds[boundsOffset+1] {
		dx32 := 3.0 * (ctrlX2 - ctrlX1)
		dx21 := 3.0 * (ctrlX1 - x1)
		coeff[3] = (x2 - x1) - dx32 // A = P3 - P0 - 3 (P2 - P1) = (P3 - P0) + 3 (P1 - P2)
		coeff[2] = dx32 - dx21      // B = 3 (P2 - P1) - 3(P1 - P0) = 3 (P2 + P0) - 6 P1
		coeff[1] = dx21             // C = 3 (P1 - P0)
		coeff[0] = x1               // D = P0

		derivCoeff[0] = coeff[1]
		derivCoeff[1] = 2.0 * coeff[2]
		derivCoeff[2] = 3.0 * coeff[3]

		// reuse this array, give it a new name for readability:
		tExtrema := derivCoeff

		tExtremaCount := solveQuadratic(derivCoeff, tExtrema)
		if tExtremaCount > 0 {
			// error condition = sum ( abs (coeff) ):
			margin := ulp(math.Abs(coeff[0]) + math.Abs(coeff[1]) + math.Abs(coeff[2]) + math.Abs(coeff[3]))

			for i := 0; i < tExtremaCount; i++ {
				t := tExtrema[i]
				if t > 0.0 && t < 1.0 {
					v := coeff[0] + float64(t*(coeff[1]+float64(t*(coeff[2]+float64(t*coeff[3])))))
					if v-margin < bounds[boundsOffset] {
						bounds[boundsOffset] = v - margin
					}
					if v+margin > bounds[boundsOffset+1] {
						bounds[boundsOffset+1] = v + margin
					}
				}
			}
		}
	}
}

// solveQuadratic is java.awt.geom.QuadCurve2D.solveQuadratic: it solves
// eqn[2]*x^2 + eqn[1]*x + eqn[0] = 0, stores the roots in res, which may be
// eqn, and returns their number, or -1 when the equation is a constant.
func solveQuadratic(eqn, res []float64) int {
	a := eqn[2]
	b := eqn[1]
	c := eqn[0]
	roots := 0
	if a == 0.0 {
		// The quadratic parabola has degenerated to a line.
		if b == 0.0 {
			// The line has degenerated to a constant.
			return -1
		}
		res[roots] = -c / b
		roots++
	} else {
		// From Numerical Recipes, 5.6, Quadratic and Cubic Equations
		d := float64(b*b) - float64(4.0*a*c)
		if d < 0.0 {
			// If d < 0.0, then there are no roots
			return 0
		}
		d = math.Sqrt(d)
		// For accuracy, calculate one root using:
		//     (-b +/- d) / 2a
		// and the other using:
		//     2c / (-b +/- d)
		// Choose the sign of the +/- so that b+d gets larger in magnitude
		if b < 0.0 {
			d = -d
		}
		q := (b + d) / -2.0
		// We already tested a for being 0 above
		res[roots] = q / a
		roots++
		if q != 0.0 {
			res[roots] = c / q
			roots++
		}
	}
	return roots
}

func (p *Path2D) GetBounds() *Rectangle {
	return p.GetBounds2D().GetBounds()
}

func (p *Path2D) Intersects(r *Rectangle) bool {
	return p.IntersectsWithXYWH(float64(r.X), float64(r.Y), float64(r.Width), float64(r.Height))
}

// IntersectsWithXYWH reports whether the interior of the path, under its
// winding rule, overlaps the interior of the rectangle. An unclosed subpath is
// treated as closed.
func (p *Path2D) IntersectsWithXYWH(x, y, w, h float64) bool {
	if math.IsNaN(x+w) || math.IsNaN(y+h) {
		return false
	}
	if w <= 0 || h <= 0 {
		return false
	}
	mask := 2
	if p.windingRule == Path2DWindNonZero {
		mask = -1
	}
	crossings := rectCrossingsForPath(p.GetPathIterator(nil), x, y, x+w, y+h)
	return crossings == rectIntersects || (crossings&mask) != 0
}

func (p *Path2D) GetPathIterator(at *AffineTransform) PathIterator {
	// offsets[i] is the index in coords of the first coordinate of segment i.
	offsets := make([]int, len(p.pointTypes))
	off := 0
	for i, t := range p.pointTypes {
		offsets[i] = off
		off += segmentPointCounts[t] * 2
	}
	return &shapeIterator{
		windingRule: p.windingRule,
		count:       len(p.pointTypes),
		affine:      at,
		outOfBounds: "path iterator out of bounds",
		segment: func(index int, coords []float64) int {
			t := int(p.pointTypes[index])
			n := segmentPointCounts[t] * 2
			copy(coords[:n], p.coords[offsets[index]:offsets[index]+n])
			return t
		},
	}
}

// rectIntersects is sun.awt.geom.Curve.RECT_INTERSECTS: the crossing count
// value that says the path crosses the rectangle itself.
const rectIntersects = math.MinInt32

// rectCrossingsForPath is sun.awt.geom.Curve.rectCrossingsForPath. It
// accumulates the number of times the path crosses the shadow extending to
// the right of the rectangle, or returns rectIntersects as soon as a segment
// crosses the rectangle.
func rectCrossingsForPath(pi PathIterator, rxmin, rymin, rxmax, rymax float64) int {
	if rxmax <= rxmin || rymax <= rymin {
		return 0
	}
	if pi.IsDone() {
		return 0
	}
	var coords [6]float64
	if pi.CurrentSegmentDouble(coords[:]) != PathIteratorSegMoveto {
		panic(&IllegalPathStateException{Message: "missing initial moveto in path definition"})
	}
	pi.Next()
	curx, cury := coords[0], coords[1]
	movx, movy := curx, cury
	crossings := 0
	for crossings != rectIntersects && !pi.IsDone() {
		switch pi.CurrentSegmentDouble(coords[:]) {
		case PathIteratorSegMoveto:
			if curx != movx || cury != movy {
				crossings = rectCrossingsForLine(crossings, rxmin, rymin, rxmax, rymax, curx, cury, movx, movy)
			}
			// Count should always be a multiple of 2 here.
			movx, movy = coords[0], coords[1]
			curx, cury = movx, movy
		case PathIteratorSegLineto:
			endx, endy := coords[0], coords[1]
			crossings = rectCrossingsForLine(crossings, rxmin, rymin, rxmax, rymax, curx, cury, endx, endy)
			curx, cury = endx, endy
		case PathIteratorSegQuadto:
			endx, endy := coords[2], coords[3]
			crossings = rectCrossingsForQuad(crossings, rxmin, rymin, rxmax, rymax,
				curx, cury, coords[0], coords[1], endx, endy, 0)
			curx, cury = endx, endy
		case PathIteratorSegCubicto:
			endx, endy := coords[4], coords[5]
			crossings = rectCrossingsForCubic(crossings, rxmin, rymin, rxmax, rymax,
				curx, cury, coords[0], coords[1], coords[2], coords[3], endx, endy, 0)
			curx, cury = endx, endy
		case PathIteratorSegClose:
			if curx != movx || cury != movy {
				crossings = rectCrossingsForLine(crossings, rxmin, rymin, rxmax, rymax, curx, cury, movx, movy)
			}
			curx, cury = movx, movy
			// Count should always be a multiple of 2 here.
		}
		pi.Next()
	}
	if crossings != rectIntersects && (curx != movx || cury != movy) {
		crossings = rectCrossingsForLine(crossings, rxmin, rymin, rxmax, rymax, curx, cury, movx, movy)
	}
	// Count should always be a multiple of 2 here.
	return crossings
}

// shadowCrossings adjusts the count for a segment that lies entirely to the
// right of the rectangle and goes from y0 to y1.
func shadowCrossings(crossings int, rymin, rymax, y0, y1 float64) int {
	if y0 < y1 {
		// y-increasing line segment...
		// We know that y0 < rymax and y1 > rymin
		if y0 <= rymin {
			crossings++
		}
		if y1 >= rymax {
			crossings++
		}
	} else if y1 < y0 {
		// y-decreasing line segment...
		// We know that y1 < rymax and y0 > rymin
		if y1 <= rymin {
			crossings--
		}
		if y0 >= rymax {
			crossings--
		}
	}
	return crossings
}

// rectCrossingsForLine is sun.awt.geom.Curve.rectCrossingsForLine.
func rectCrossingsForLine(crossings int, rxmin, rymin, rxmax, rymax, x0, y0, x1, y1 float64) int {
	if y0 >= rymax && y1 >= rymax {
		return crossings
	}
	if y0 <= rymin && y1 <= rymin {
		return crossings
	}
	if x0 <= rxmin && x1 <= rxmin {
		return crossings
	}
	if x0 >= rxmax && x1 >= rxmax {
		// Line is entirely to the right of the rect
		// and the vertical ranges of the two overlap by a non-empty amount
		// Thus, this line segment is partially in the "right-shadow"
		// Path may have done a complete crossing
		// Or path may have entered or exited the right-shadow
		return shadowCrossings(crossings, rymin, rymax, y0, y1)
	}
	// Remaining case:
	// Both x and y ranges overlap by a non-empty amount
	// First do trivial INTERSECTS rejection of the cases
	// where one of the endpoints is inside the rectangle.
	if (x0 > rxmin && x0 < rxmax && y0 > rymin && y0 < rymax) ||
		(x1 > rxmin && x1 < rxmax && y1 > rymin && y1 < rymax) {
		return rectIntersects
	}
	// Otherwise calculate the y intercepts and see where
	// they fall with respect to the rectangle
	xi0 := x0
	if y0 < rymin {
		xi0 += (rymin - y0) * (x1 - x0) / (y1 - y0)
	} else if y0 > rymax {
		xi0 += (rymax - y0) * (x1 - x0) / (y1 - y0)
	}
	xi1 := x1
	if y1 < rymin {
		xi1 += (rymin - y1) * (x0 - x1) / (y0 - y1)
	} else if y1 > rymax {
		xi1 += (rymax - y1) * (x0 - x1) / (y0 - y1)
	}
	if xi0 <= rxmin && xi1 <= rxmin {
		return crossings
	}
	if xi0 >= rxmax && xi1 >= rxmax {
		return shadowCrossings(crossings, rymin, rymax, y0, y1)
	}
	return rectIntersects
}

// curveShadowCrossings adjusts the count for a curve that lies entirely to the
// right of the rectangle and goes from y0 to y1.
func curveShadowCrossings(crossings int, rymin, rymax, y0, y1 float64) int {
	if y0 < y1 {
		// y-increasing line segment...
		if y0 <= rymin && y1 > rymin {
			crossings++
		}
		if y0 < rymax && y1 >= rymax {
			crossings++
		}
	} else if y1 < y0 {
		// y-decreasing line segment...
		if y1 <= rymin && y0 > rymin {
			crossings--
		}
		if y1 < rymax && y0 >= rymax {
			crossings--
		}
	}
	return crossings
}

// rectCrossingsForQuad is sun.awt.geom.Curve.rectCrossingsForQuad.
func rectCrossingsForQuad(crossings int, rxmin, rymin, rxmax, rymax, x0, y0, xc, yc, x1, y1 float64, level int) int {
	if y0 >= rymax && yc >= rymax && y1 >= rymax {
		return crossings
	}
	if y0 <= rymin && yc <= rymin && y1 <= rymin {
		return crossings
	}
	if x0 <= rxmin && xc <= rxmin && x1 <= rxmin {
		return crossings
	}
	if x0 >= rxmax && xc >= rxmax && x1 >= rxmax {
		// Quad is entirely to the right of the rect
		// and the vertical range of the 3 Y coordinates of the quad
		// overlaps the vertical range of the rect by a non-empty amount
		// We now judge the crossings solely based on the line segment
		// connecting the endpoints of the quad.
		// Note that we may have 0, 1, or 2 crossings as the control
		// point may be causing the Y range intersection while the
		// two endpoints are entirely above or below.
		return curveShadowCrossings(crossings, rymin, rymax, y0, y1)
	}
	// The intersection of ranges is more complicated
	// First do trivial INTERSECTS rejection of the cases
	// where one of the endpoints is inside the rectangle.
	if (x0 < rxmax && x0 > rxmin && y0 < rymax && y0 > rymin) ||
		(x1 < rxmax && x1 > rxmin && y1 < rymax && y1 > rymin) {
		return rectIntersects
	}
	// Otherwise, subdivide and look for one of the cases above.
	// double precision only has 52 bits of mantissa
	if level > 52 {
		return rectCrossingsForLine(crossings, rxmin, rymin, rxmax, rymax, x0, y0, x1, y1)
	}
	x0c := (x0 + xc) / 2
	y0c := (y0 + yc) / 2
	xc1 := (xc + x1) / 2
	yc1 := (yc + y1) / 2
	xc = (x0c + xc1) / 2
	yc = (y0c + yc1) / 2
	if math.IsNaN(xc) || math.IsNaN(yc) {
		// [xy]c are NaN if any of [xy]0c or [xy]c1 are NaN
		// [xy]0c or [xy]c1 are NaN if any of [xy][0c1] are NaN
		// These values are also NaN if opposing infinities are added
		return 0
	}
	crossings = rectCrossingsForQuad(crossings, rxmin, rymin, rxmax, rymax, x0, y0, x0c, y0c, xc, yc, level+1)
	if crossings != rectIntersects {
		crossings = rectCrossingsForQuad(crossings, rxmin, rymin, rxmax, rymax, xc, yc, xc1, yc1, x1, y1, level+1)
	}
	return crossings
}

// rectCrossingsForCubic is sun.awt.geom.Curve.rectCrossingsForCubic.
func rectCrossingsForCubic(crossings int, rxmin, rymin, rxmax, rymax, x0, y0, xc0, yc0, xc1, yc1, x1, y1 float64, level int) int {
	if y0 >= rymax && yc0 >= rymax && yc1 >= rymax && y1 >= rymax {
		return crossings
	}
	if y0 <= rymin && yc0 <= rymin && yc1 <= rymin && y1 <= rymin {
		return crossings
	}
	if x0 <= rxmin && xc0 <= rxmin && xc1 <= rxmin && x1 <= rxmin {
		return crossings
	}
	if x0 >= rxmax && xc0 >= rxmax && xc1 >= rxmax && x1 >= rxmax {
		// Cubic is entirely to the right of the rect
		// and the vertical range of the 4 Y coordinates of the cubic
		// overlaps the vertical range of the rect by a non-empty amount
		// We now judge the crossings solely based on the line segment
		// connecting the endpoints of the cubic.
		return curveShadowCrossings(crossings, rymin, rymax, y0, y1)
	}
	// The intersection of ranges is more complicated
	// First do trivial INTERSECTS rejection of the cases
	// where one of the endpoints is inside the rectangle.
	if (x0 > rxmin && x0 < rxmax && y0 > rymin && y0 < rymax) ||
		(x1 > rxmin && x1 < rxmax && y1 > rymin && y1 < rymax) {
		return rectIntersects
	}
	// Otherwise, subdivide and look for one of the cases above.
	// double precision only has 52 bits of mantissa
	if level > 52 {
		return rectCrossingsForLine(crossings, rxmin, rymin, rxmax, rymax, x0, y0, x1, y1)
	}
	xmid := (xc0 + xc1) / 2
	ymid := (yc0 + yc1) / 2
	xc0 = (x0 + xc0) / 2
	yc0 = (y0 + yc0) / 2
	xc1 = (xc1 + x1) / 2
	yc1 = (yc1 + y1) / 2
	xc0m := (xc0 + xmid) / 2
	yc0m := (yc0 + ymid) / 2
	xmc1 := (xmid + xc1) / 2
	ymc1 := (ymid + yc1) / 2
	xmid = (xc0m + xmc1) / 2
	ymid = (yc0m + ymc1) / 2
	if math.IsNaN(xmid) || math.IsNaN(ymid) {
		// [xy]mid are NaN if any of [xy]c0m or [xy]mc1 are NaN
		// [xy]c0m or [xy]mc1 are NaN if any of [xy][c][01] are NaN
		// These values are also NaN if opposing infinities are added
		return 0
	}
	crossings = rectCrossingsForCubic(crossings, rxmin, rymin, rxmax, rymax,
		x0, y0, xc0, yc0, xc0m, yc0m, xmid, ymid, level+1)
	if crossings != rectIntersects {
		crossings = rectCrossingsForCubic(crossings, rxmin, rymin, rxmax, rymax,
			xmid, ymid, xmc1, ymc1, xc1, yc1, x1, y1, level+1)
	}
	return crossings
}
