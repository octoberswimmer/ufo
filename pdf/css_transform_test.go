// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/CssTransformTest.java

package pdf

import (
	"math"
	"strconv"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/geom"
)

const cssTransformTestTolerance = 0.00001

// cssTransformTestContext stands for the Mockito mock of CssContext: only
// GetDotsPerPixel is stubbed (it returns 1); any other method panics on the
// nil embedded interface.
type cssTransformTestContext struct {
	ufo.CssContext
}

func (c *cssTransformTestContext) GetDotsPerPixel() int {
	return 1
}

// cssTransformTestStyle stands for the mock of CalculatedStyle.
type cssTransformTestStyle struct {
	ufo.CalculatedStyleI
	transforms      []*ufo.PropertyValue
	transformOrigin *ufo.BackgroundPosition
}

func (s *cssTransformTestStyle) GetTransforms() []*ufo.PropertyValue {
	return s.transforms
}

func (s *cssTransformTestStyle) GetTransformOrigin() *ufo.BackgroundPosition {
	return s.transformOrigin
}

// cssTransformTestBox stands for the mock of Box.
type cssTransformTestBox struct {
	ufo.BoxI
	style              ufo.CalculatedStyleI
	paintingBorderEdge *geom.Rectangle
}

func (b *cssTransformTestBox) GetStyle() ufo.CalculatedStyleI {
	return b.style
}

func (b *cssTransformTestBox) GetPaintingBorderEdge(cssCtx ufo.CssContext) *geom.Rectangle {
	return b.paintingBorderEdge
}

var cssTransformTestCtx ufo.CssContext = &cssTransformTestContext{}

func cssTransformTestBounds() *geom.Rectangle {
	return geom.NewRectangle(0, 0, 100, 50)
}

func cssTransformTestFloatString(value float32) string {
	return strconv.FormatFloat(float64(value), 'f', -1, 32)
}

func cssTransformTestPx(value float32) *ufo.PropertyValue {
	return ufo.NewPropertyValueFloat(ufo.CSSPrimitiveValueCssPx, value, cssTransformTestFloatString(value)+"px")
}

func cssTransformTestPercent(value float32) *ufo.PropertyValue {
	return ufo.NewPropertyValueFloat(ufo.CSSPrimitiveValueCssPercentage, value, cssTransformTestFloatString(value)+"%")
}

func cssTransformTestNumber(value float32) *ufo.PropertyValue {
	return ufo.NewPropertyValueFloat(ufo.CSSPrimitiveValueCssNumber, value, cssTransformTestFloatString(value))
}

func cssTransformTestDeg(value float32) *ufo.PropertyValue {
	return ufo.NewPropertyValueFloat(ufo.CSSPrimitiveValueCssDeg, value, cssTransformTestFloatString(value)+"deg")
}

func cssTransformTestRad(value float32) *ufo.PropertyValue {
	return ufo.NewPropertyValueFloat(ufo.CSSPrimitiveValueCssRad, value, cssTransformTestFloatString(value)+"rad")
}

func cssTransformTestGrad(value float32) *ufo.PropertyValue {
	return ufo.NewPropertyValueFloat(ufo.CSSPrimitiveValueCssGrad, value, cssTransformTestFloatString(value)+"grad")
}

func cssTransformTestFunction(name string, args ...*ufo.PropertyValue) *ufo.FSFunction {
	return ufo.NewFSFunction(name, args)
}

func cssTransformTestAssertCloseTo(t *testing.T, what string, actual float64, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > cssTransformTestTolerance {
		t.Errorf("%s = %v, want %v", what, actual, expected)
	}
}

func cssTransformTestAssertMatrixEquals(t *testing.T, actual *geom.AffineTransform, m00, m10, m01, m11, m02, m12 float64) {
	t.Helper()
	matrix := make([]float64, 6)
	actual.GetMatrix(matrix)
	cssTransformTestAssertCloseTo(t, "m00", matrix[0], m00)
	cssTransformTestAssertCloseTo(t, "m10", matrix[1], m10)
	cssTransformTestAssertCloseTo(t, "m01", matrix[2], m01)
	cssTransformTestAssertCloseTo(t, "m11", matrix[3], m11)
	cssTransformTestAssertCloseTo(t, "m02", matrix[4], m02)
	cssTransformTestAssertCloseTo(t, "m12", matrix[5], m12)
}

func cssTransformTestToMatrix(function *ufo.FSFunction) *geom.AffineTransform {
	return CssTransformToMatrix(cssTransformTestCtx, &cssTransformTestStyle{}, function, cssTransformTestBounds())
}

func TestCssTransform_matrix(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("matrix",
		cssTransformTestNumber(1), cssTransformTestNumber(2), cssTransformTestNumber(3),
		cssTransformTestNumber(4), cssTransformTestNumber(5), cssTransformTestNumber(6)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 2, 3, 4, 5, 6)
}

func TestCssTransform_translateWithBothArguments(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("translate", cssTransformTestPx(10), cssTransformTestPercent(20)))

	// percent(20) of bounds.height (50) == 10
	cssTransformTestAssertMatrixEquals(t, result, 1, 0, 0, 1, 10, 10)
}

func TestCssTransform_translateWithOnlyHorizontalArgumentDefaultsVerticalToZero(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("translate", cssTransformTestPx(10)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 0, 0, 1, 10, 0)
}

func TestCssTransform_translateX(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("translatex", cssTransformTestPx(15)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 0, 0, 1, 15, 0)
}

func TestCssTransform_translateY(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("translatey", cssTransformTestPx(15)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 0, 0, 1, 0, 15)
}

func TestCssTransform_scaleWithOneArgumentAppliesToBothAxes(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("scale", cssTransformTestNumber(2)))

	cssTransformTestAssertMatrixEquals(t, result, 2, 0, 0, 2, 0, 0)
}

func TestCssTransform_scaleWithTwoArguments(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("scale", cssTransformTestNumber(2), cssTransformTestNumber(3)))

	cssTransformTestAssertMatrixEquals(t, result, 2, 0, 0, 3, 0, 0)
}

func TestCssTransform_scaleX(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("scalex", cssTransformTestNumber(2)))

	cssTransformTestAssertMatrixEquals(t, result, 2, 0, 0, 1, 0, 0)
}

func TestCssTransform_scaleY(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("scaley", cssTransformTestNumber(3)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 0, 0, 3, 0, 0)
}

func TestCssTransform_rotate90Degrees(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("rotate", cssTransformTestDeg(90)))

	cssTransformTestAssertMatrixEquals(t, result, 0, 1, -1, 0, 0, 0)
}

func TestCssTransform_skewWithBothArguments(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("skew", cssTransformTestDeg(45), cssTransformTestDeg(0)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 0, 1, 1, 0, 0)
}

func TestCssTransform_skewX(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("skewx", cssTransformTestDeg(45)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 0, 1, 1, 0, 0)
}

func TestCssTransform_skewY(t *testing.T) {
	result := cssTransformTestToMatrix(cssTransformTestFunction("skewy", cssTransformTestDeg(45)))

	cssTransformTestAssertMatrixEquals(t, result, 1, 1, 0, 1, 0, 0)
}

func TestCssTransform_angleToRadiansConvertsDegrees(t *testing.T) {
	cssTransformTestAssertCloseTo(t, "radians", CssTransformAngleToRadians(cssTransformTestDeg(180)), math.Pi)
}

func TestCssTransform_angleToRadiansPassesRadiansThrough(t *testing.T) {
	cssTransformTestAssertCloseTo(t, "radians", CssTransformAngleToRadians(cssTransformTestRad(1.5)), 1.5)
}

func TestCssTransform_angleToRadiansConvertsGrads(t *testing.T) {
	// 200 grad == 180 deg == pi radians
	cssTransformTestAssertCloseTo(t, "radians", CssTransformAngleToRadians(cssTransformTestGrad(200)), math.Pi)
}

func TestCssTransform_toAffineTransformReturnsNullWhenThereIsNoTransform(t *testing.T) {
	box := &cssTransformTestBox{style: &cssTransformTestStyle{}}

	if result := CssTransformToAffineTransform(cssTransformTestCtx, box); result != nil {
		t.Errorf("CssTransformToAffineTransform = %v, want nil", result)
	}
}

func TestCssTransform_toAffineTransformAnchorsRotationAtTheTransformOrigin(t *testing.T) {
	style := &cssTransformTestStyle{
		transforms: []*ufo.PropertyValue{ufo.NewPropertyValueFSFunction(cssTransformTestFunction("rotate", cssTransformTestDeg(90)))},
		// transform-origin: left top -- the box's top-left corner, at absolute position (10, 20)
		transformOrigin: ufo.NewBackgroundPosition(cssTransformTestPercent(0), cssTransformTestPercent(0)),
	}
	box := &cssTransformTestBox{style: style, paintingBorderEdge: geom.NewRectangle(10, 20, 100, 50)}

	result := CssTransformToAffineTransform(cssTransformTestCtx, box)

	// The origin point itself must stay fixed under a rotation about itself.
	origin := result.Transform(geom.NewPoint2DFloat(10, 20), nil)
	cssTransformTestAssertCloseTo(t, "origin x", origin.GetX(), 10)
	cssTransformTestAssertCloseTo(t, "origin y", origin.GetY(), 20)

	// The box's top-right corner (110, 20) is 100 to the right of the origin;
	// rotating 90 degrees around the origin moves it 100 below the origin instead.
	corner := result.Transform(geom.NewPoint2DFloat(110, 20), nil)
	cssTransformTestAssertCloseTo(t, "corner x", corner.GetX(), 10)
	cssTransformTestAssertCloseTo(t, "corner y", corner.GetY(), 120)
}
