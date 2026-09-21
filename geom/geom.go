// Package geom replaces the java.awt geometry classes that Flying Saucer uses:
// Rectangle, Point, Dimension, Insets, Shape, PathIterator, Path2D, Line2D,
// Ellipse2D, Arc2D, Rectangle2D, Point2D, AffineTransform, Area, Stroke and
// BasicStroke. The types behave as the JDK classes do and carry the JDK method
// names, capitalized, so ported code maps onto them one to one.
//
// Differences from the JDK:
//
//   - The JDK has a Float and a Double variant of each java.awt.geom class.
//     Here each class is one type that stores float64. The Float constructors
//     (NewLine2DFloat, NewArc2DFloat, ...) take float32 arguments, so the stored
//     values are the ones the JDK would store. Path2D keeps a flag for the Float
//     variant and rounds every stored coordinate to float32, as Path2D.Float
//     does.
//   - Area is not a general constructive area geometry implementation. It is
//     the intersection of a list of shapes; see Area.
//   - Path2D.GetBounds2D returns the bounds of the curves themselves, as the JDK
//     has done since version 19, not the bounds of the control points.
//   - BasicStroke.CreateStrokedShape is not implemented and panics.
//   - JDK exceptions are panics with the typed values declared in this file,
//     except NoninvertibleTransformException, which is a returned error.
package geom

import (
	"math"
	"strconv"
	"strings"
)

// IllegalPathStateException is the panic value for a path operation that is
// invalid in the path's current state, such as LineTo before MoveTo.
type IllegalPathStateException struct {
	Message string
}

func (e *IllegalPathStateException) Error() string {
	return e.Message
}

// IllegalArgumentException is the panic value for an argument the JDK rejects
// with java.lang.IllegalArgumentException.
type IllegalArgumentException struct {
	Message string
}

func (e *IllegalArgumentException) Error() string {
	return e.Message
}

// NoSuchElementException is the panic value for reading a segment from a
// PathIterator that is done.
type NoSuchElementException struct {
	Message string
}

func (e *NoSuchElementException) Error() string {
	return e.Message
}

// UnsupportedOperationException is the panic value for an operation this
// package does not implement.
type UnsupportedOperationException struct {
	Message string
}

func (e *UnsupportedOperationException) Error() string {
	return e.Message
}

// NoninvertibleTransformException is the error AffineTransform.CreateInverse
// and AffineTransform.InverseTransform return for a matrix without an inverse.
type NoninvertibleTransformException struct {
	Message string
}

func (e *NoninvertibleTransformException) Error() string {
	return e.Message
}

// javaInt converts as a Java (int) cast of a double does: truncation toward
// zero, NaN to 0, and values outside the 32-bit range to the nearest bound.
func javaInt(v float64) int {
	switch {
	case math.IsNaN(v):
		return 0
	case v >= math.MaxInt32:
		return math.MaxInt32
	case v <= math.MinInt32:
		return math.MinInt32
	}
	return int(v)
}

// clipInt32 limits v to the range of a Java int.
func clipInt32(v int64) int64 {
	if v < math.MinInt32 {
		return math.MinInt32
	}
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	return v
}

// toRadians is java.lang.Math.toRadians.
func toRadians(angdeg float64) float64 {
	return angdeg * 0.017453292519943295
}

// javaDoubleString formats as java.lang.Double.toString does: decimal
// notation with at least one fractional digit for magnitudes in [1e-3, 1e7),
// and computerized scientific notation ("1.0E10") otherwise.
func javaDoubleString(d float64) string {
	switch {
	case math.IsNaN(d):
		return "NaN"
	case math.IsInf(d, 1):
		return "Infinity"
	case math.IsInf(d, -1):
		return "-Infinity"
	case d == 0:
		if math.Signbit(d) {
			return "-0.0"
		}
		return "0.0"
	}
	abs := math.Abs(d)
	if abs >= 1e-3 && abs < 1e7 {
		s := strconv.FormatFloat(d, 'f', -1, 64)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	}
	s := strconv.FormatFloat(d, 'e', -1, 64)
	mant, exp, _ := strings.Cut(s, "e")
	if !strings.Contains(mant, ".") {
		mant += ".0"
	}
	n, _ := strconv.Atoi(exp)
	return mant + "E" + strconv.Itoa(n)
}
