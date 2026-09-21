// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/FSLinearGradient.java

package ufo

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

type FSLinearGradient struct {
	x1         int
	y1         int
	x2         int
	y2         int
	stopPoints []*FSLinearGradientStopValue
}

type FSLinearGradientStopValue struct {
	color      FSColor
	lengthType int16
	// length is nil when the stop has no length.
	length *float32
	// dotsValue is nil until the stop has been normalized.
	dotsValue *float32
}

func newFSLinearGradientStopValueWithValueLengthType(color FSColor, value float32, lengthType int16) *FSLinearGradientStopValue {
	return &FSLinearGradientStopValue{
		color:      color,
		length:     &value,
		lengthType: lengthType,
	}
}

func newFSLinearGradientStopValue(color FSColor) *FSLinearGradientStopValue {
	return &FSLinearGradientStopValue{
		color:      color,
		length:     nil,
		lengthType: 0,
	}
}

func (s *FSLinearGradientStopValue) GetColor() FSColor {
	return s.color
}

// GetLength may return nil.
func (s *FSLinearGradientStopValue) GetLength() *float32 {
	return s.dotsValue
}

func (s *FSLinearGradientStopValue) ToString() string {
	dotsValue := "null"
	if s.dotsValue != nil {
		dotsValue = lengthValueFloatToString(*s.dotsValue)
	}
	return "[" + listValueToString(s.color) + "](" + dotsValue + ")"
}

func (s *FSLinearGradientStopValue) String() string {
	return s.ToString()
}

func (s *FSLinearGradientStopValue) CompareTo(arg0 *FSLinearGradientStopValue) int {
	if fsLinearGradientFloatEquals(s.dotsValue, arg0.dotsValue) {
		return 0
	}
	if *s.dotsValue < *arg0.dotsValue {
		return -1
	}

	return 1
}

// fsLinearGradientFloatEquals is Objects.equals on two java.lang.Float
// values: Float.equals compares the floatToIntBits representations, so NaN
// equals NaN and 0.0 does not equal -0.0.
func fsLinearGradientFloatEquals(a *float32, b *float32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return fsLinearGradientFloatToIntBits(*a) == fsLinearGradientFloatToIntBits(*b)
}

func fsLinearGradientFloatToIntBits(f float32) uint32 {
	if f != f {
		return 0x7fc00000
	}
	return math.Float32bits(f)
}

func (g *FSLinearGradient) GetStopPoints() []*FSLinearGradientStopValue {
	return g.stopPoints
}

func (g *FSLinearGradient) deg2rad(deg float32) float32 {
	// Math.toRadians
	return float32(float64(deg) * 0.017453292519943295)
}

func (g *FSLinearGradient) rad2deg(rad float32) float32 {
	// Math.toDegrees
	return float32(float64(rad) * 57.29577951308232)
}

// Compute the endpoints so that a gradient of the given angle
// covers a box of the given size.
// From: https://github.com/WebKit/webkit/blob/master/Source/WebCore/css/CSSGradientValue.cpp
func (g *FSLinearGradient) endPointsFromAngle(angleDeg float32, w int, h int) {
	angleDeg = float32(math.Mod(float64(angleDeg), 360))
	if angleDeg < 0 {
		angleDeg += 360
	}

	if angleDeg == 0 {
		g.x1 = 0
		g.y1 = h

		g.x2 = 0
		g.y2 = 0
		return
	}

	if angleDeg == 90 {
		g.x1 = 0
		g.y1 = 0

		g.x2 = w
		g.y2 = 0
		return
	}

	if angleDeg == 180 {
		g.x1 = 0
		g.y1 = 0

		g.x2 = 0
		g.y2 = h
		return
	}

	if angleDeg == 270 {
		g.x1 = w
		g.y1 = 0

		g.x2 = 0
		g.y2 = 0
		return
	}

	// angleDeg is a "bearing angle" (0deg = N, 90deg = E),
	// but tan expects 0deg = E, 90deg = N.
	slope := float32(math.Tan(float64(g.deg2rad(90 - angleDeg))))

	// We find the endpoint by computing the intersection of the line formed by the slope,
	// and a line perpendicular to it that intersects the corner.
	perpendicularSlope := -1 / slope

	// Compute start corner relative to center, in Cartesian space (+y = up).
	halfHeight := float32(h) / 2.0
	halfWidth := float32(w) / 2.0
	var xEnd, yEnd float32

	if angleDeg < 90 {
		xEnd = halfWidth
		yEnd = halfHeight
	} else if angleDeg < 180 {
		xEnd = halfWidth
		yEnd = -halfHeight
	} else if angleDeg < 270 {
		xEnd = -halfWidth
		yEnd = -halfHeight
	} else {
		xEnd = -halfWidth
		yEnd = halfHeight
	}

	// Compute c (of y = mx + c) using the corner point.
	c := yEnd - float32(perpendicularSlope*xEnd)
	endX := c / (slope - perpendicularSlope)
	endY := float32(perpendicularSlope*endX) + c

	// We computed the end point, so set the second point,
	// taking into account the moved origin and the fact that we're in drawing space (+y = down).
	g.x2 = calculatedStyleFloatToInt(halfWidth + endX)
	g.y2 = calculatedStyleFloatToInt(halfHeight - endY)

	// Reflect around the center for the start point.
	g.x1 = calculatedStyleFloatToInt(halfWidth - endX)
	g.y1 = calculatedStyleFloatToInt(halfHeight + endY)
}

func (g *FSLinearGradient) constructZero() {
	//BuilderUtil.cssNoThrowError(LangId.FUNCTION_GENERAL, "linear-gradient");

	// Just return a 1px wide (nearly) transparent gradient.
	g.x1 = 0
	g.y1 = 0

	g.x2 = 1
	g.y2 = 0

	g.stopPoints = g.stopPoints[:0]
	g.stopPoints = append(g.stopPoints, newFSLinearGradientStopValue(NewFSRGBColorWithAlpha(0, 0, 0, 0)))
	g.stopPoints = append(g.stopPoints, newFSLinearGradientStopValue(NewFSRGBColorWithAlpha(0, 0, 0, 0.0001)))

	var first float32 = 0.0
	var second float32 = 1.0
	g.stopPoints[0].dotsValue = &first
	g.stopPoints[1].dotsValue = &second
}

const fsLinearGradientRcssNumber = "(-)?((\\d){1,10}((\\.)(\\d){1,10})?)"
const fsLinearGradientRcssLength = "((0$)|((" + fsLinearGradientRcssNumber + ")+" + "((em)|(ex)|(px)|(cm)|(mm)|(in)|(pt)|(pc)|(%))))"

// Java compiles the pattern on every call and uses Matcher.matches(), which
// has to match the whole input; the anchors do that here.
var fsLinearGradientCssLengthPattern = regexp.MustCompile("^(?:" + fsLinearGradientRcssLength + ")$")

func FSLinearGradientLooksLikeALength(val string) bool {
	return fsLinearGradientCssLengthPattern.MatchString(val)
}

func FSLinearGradientLooksLikeABGPosition(val string) bool {
	backgroundPositionsIdents := []string{
		"top",
		"center",
		"bottom",
		"right",
		"left",
	}
	return fsLinearGradientContains(backgroundPositionsIdents, val) || FSLinearGradientLooksLikeALength(val)
}

func fsLinearGradientContains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func NewFSLinearGradient(function *FSFunction, style CalculatedStyleI, width int, height int, ctx CssContext) *FSLinearGradient {
	g := &FSLinearGradient{stopPoints: make([]*FSLinearGradientStopValue, 0, 2)}
	g.construct(function, style, width, height, ctx)
	return g
}

// construct is the body of the Java constructor, which returns early on
// malformed input.
func (g *FSLinearGradient) construct(function *FSFunction, style CalculatedStyleI, width int, height int, ctx CssContext) {
	params := function.GetParameters()
	i := 1

	if len(params) == 0 {
		g.constructZero()
		return
	}

	if GeneralUtilCiEquals(params[0].GetStringValue(), "to") {
		// The "to" keyword is followed by one or two position
		// idents (in any order).
		// linear-gradient( to left top, blue, red);
		// linear-gradient( to top right, blue, red);
		for ; i < len(params); i++ {
			// PropertyValue.GetStringValue returns "" where Java returns null.
			if params[i].GetStringValue() == "" || !FSLinearGradientLooksLikeABGPosition(params[i].GetStringValue()) {
				break
			}
		}

		positions := []string{}

		if i == 2 {
			positions = []string{strings.ToLower(params[1].GetStringValue())}
		} else if i == 3 {
			positions = []string{
				strings.ToLower(params[1].GetStringValue()),
				strings.ToLower(params[2].GetStringValue())}
		}

		if fsLinearGradientContains(positions, "top") && fsLinearGradientContains(positions, "left") {
			g.x1 = width
			g.y1 = height

			g.x2 = 0
			g.y2 = 0
		} else if fsLinearGradientContains(positions, "top") && fsLinearGradientContains(positions, "right") {
			g.x1 = 0
			g.y1 = height

			g.x2 = width
			g.y2 = 0
		} else if fsLinearGradientContains(positions, "bottom") && fsLinearGradientContains(positions, "left") {
			g.x1 = width
			g.y1 = 0

			g.x2 = 0
			g.y2 = height
		} else if fsLinearGradientContains(positions, "bottom") && fsLinearGradientContains(positions, "right") {
			g.x1 = 0
			g.y1 = 0

			g.x2 = width
			g.y2 = height
		} else if fsLinearGradientContains(positions, "bottom") {
			g.x1 = 0
			g.y1 = 0

			g.x2 = 0
			g.y2 = height
		} else if fsLinearGradientContains(positions, "top") {
			g.x1 = 0
			g.y1 = height

			g.x2 = 0
			g.y2 = 0
		} else if fsLinearGradientContains(positions, "left") {
			g.x1 = width
			g.y1 = 0

			g.x2 = 0
			g.y2 = 0
		} else if fsLinearGradientContains(positions, "right") {
			g.x1 = 0
			g.y1 = 0

			g.x2 = width
			g.y2 = 0
		} else {
			g.constructZero()
			return
		}
	} else if params[0].GetPrimitiveType() == CSSPrimitiveValueCssDeg {
		// linear-gradient(45deg, ...)
		g.endPointsFromAngle(params[0].GetFloatValue(), width, height)
	} else if params[0].GetPrimitiveType() == CSSPrimitiveValueCssRad {
		// linear-gradient(2rad, ...)
		g.endPointsFromAngle(g.rad2deg(params[0].GetFloatValue()), width, height)
	} else {
		// linear-gradient function must begin with the word 'to' or an angle.
		g.constructZero()
		return
	}

	if len(params)-i < 2 {
		// Less than two color stops provided.
		g.constructZero()
		return
	}

	for ; i < len(params); i++ {
		// Each stop point can have a color and optionally a length.
		value := params[i]
		var color *FSRGBColor
		switch value.GetPrimitiveType() {
		case CSSPrimitiveValueCssIdent:
			color = ConversionsGetColor(value.GetStringValue())
		default:
			if fsColor := value.GetFSColor(); fsColor != nil {
				color = fsColor.(*FSRGBColor)
			}
		}

		if color == nil {
			// Invalid color.
			g.constructZero()
			return
		}

		if i+1 < len(params) &&
			(BuilderUtilIsLength(params[i+1]) ||
				params[i+1].GetPrimitiveType() == CSSPrimitiveValueCssPercentage) {
			val2 := params[i+1]
			g.stopPoints = append(g.stopPoints, newFSLinearGradientStopValueWithValueLengthType(color, val2.GetFloatValue(), val2.GetPrimitiveType()))
			i++
		} else {
			g.stopPoints = append(g.stopPoints, newFSLinearGradientStopValue(color))
		}
	}

	// Normalize lengths into dots values.
	for m := 0; m < len(g.stopPoints); m++ {
		pt := g.stopPoints[m]
		if pt.length != nil {
			dots := LengthValueCalcFloatProportionalValue(style, CSSNameBackgroundImage, "", *pt.length, pt.lengthType, float32(width), ctx)
			pt.dotsValue = &dots
		} else if m == 0 {
			// First value is zero.
			var dots float32 = 0.0
			pt.dotsValue = &dots
		} else if m == len(g.stopPoints)-1 {
			// Last value is 100%.
			dots := LengthValueCalcFloatProportionalValue(style, CSSNameBackgroundImage, "100%", 100.0, CSSPrimitiveValueCssPercentage, float32(width), ctx)
			pt.dotsValue = &dots
		}
	}

	var lastValue float32 = 0.0
	var nextValue float32 = 100.0
	var increment float32 = 0.0

	// TODO: Confirm below is correct, no divide by zero and
	// no endless loop.

	// Now normalize those stop points without a length.
	for j := 1; j < len(g.stopPoints); j++ {
		if j+1 < len(g.stopPoints) &&
			g.stopPoints[j].dotsValue == nil &&
			increment == 0.0 {
			k := j + 1

			for ; k < len(g.stopPoints); k++ {
				if g.stopPoints[k].dotsValue != nil {
					nextValue = *g.stopPoints[k].dotsValue
					break
				}
			}

			// k now contains the number of values that we had to skip to find a provided
			// value. We use this to get the increment for unprovided values.
			increment = (nextValue - lastValue) / float32(k)
		}

		if g.stopPoints[j].dotsValue != nil {
			increment = 0
		} else {
			dots := lastValue + increment
			g.stopPoints[j].dotsValue = &dots
		}
		lastValue = *g.stopPoints[j].dotsValue
	}

	// Collections.sort is a stable sort.
	sort.SliceStable(g.stopPoints, func(a, b int) bool {
		return g.stopPoints[a].CompareTo(g.stopPoints[b]) < 0
	})

	for b := 0; b < len(g.stopPoints)-1; b++ {
		if fsLinearGradientFloatEquals(g.stopPoints[b].dotsValue,
			g.stopPoints[b+1].dotsValue) {
			// Duplicate lengths.
			g.constructZero()
			return
		}
	}
}

// These function get the x, y of the starting and ending points of the gradient.
// They assume a start at zero, so should be offset when used.

func (g *FSLinearGradient) GetStartX() int {
	return g.x1
}

func (g *FSLinearGradient) GetEndX() int {
	return g.x2
}

func (g *FSLinearGradient) GetStartY() int {
	return g.y1
}

func (g *FSLinearGradient) GetEndY() int {
	return g.y2
}

func (g *FSLinearGradient) ToString() string {
	stops := make([]string, len(g.stopPoints))
	for i, stop := range g.stopPoints {
		stops[i] = stop.ToString()
	}
	return fmt.Sprintf("[%d, %d] to [%d, %d](%s)", g.x1, g.y1, g.x2, g.y2, "["+strings.Join(stops, ", ")+"]")
}

func (g *FSLinearGradient) String() string {
	return g.ToString()
}
