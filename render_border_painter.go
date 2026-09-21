// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/BorderPainter.java

package ufo

import (
	"math"

	"github.com/octoberswimmer/ufo/geom"
)

const (
	BorderPainterTop    = 1
	BorderPainterLeft   = 2
	BorderPainterBottom = 4
	BorderPainterRight  = 8
	BorderPainterAll    = BorderPainterTop | BorderPainterLeft | BorderPainterBottom | BorderPainterRight
)

// BorderPainterGenerateBorderBounds generates a full round rectangle that is
// made of bounds and border.
//
// bounds: dimensions of the rect. border: the border specs. inside: set true
// if you want the inner bounds of borders. Returns a path that is all sides of
// the round rectangle.
func BorderPainterGenerateBorderBounds(bounds *geom.Rectangle, border *BorderPropertySet, inside bool) geom.Shape {
	var scaledOffset float32
	if inside {
		scaledOffset = 1
	}
	path := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterTop, border, false, scaledOffset, 1, false)
	path.Append(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterRight, border, false, scaledOffset, 1, false), true)
	path.Append(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterBottom, border, false, scaledOffset, 1, false), true)
	path.Append(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterLeft, border, false, scaledOffset, 1, false), true)
	return path
}

// BorderPainterGenerateBorderShape generates one side of a border.
//
// bounds: bounds of the container. side: what side you want. border: border
// props. drawInterior: if you want it to be 2d or not, if false it will be
// just a line. Returns a path for the side chosen.
func BorderPainterGenerateBorderShape(bounds *geom.Rectangle, side int, border *BorderPropertySet, drawInterior bool) *geom.Path2D {
	return BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, side, border, drawInterior, 0, 1, true)
}

// BorderPainterGenerateBorderShapeWithScaledOffset generates one side of a
// border.
//
// scaledOffset insets the border by multiplying border widths by this
// variable, best use would be 1 or .5, cant see it for much other than that.
func BorderPainterGenerateBorderShapeWithScaledOffset(bounds *geom.Rectangle, side int, border *BorderPropertySet, drawInterior bool, scaledOffset float32) *geom.Path2D {
	return BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, side, border, drawInterior, scaledOffset, 1, true)
}

// BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap generates
// one side of a border.
//
// scaledOffset insets the border by multiplying border widths by this
// variable, best use would be 1 or .5, cant see it for much other than that.
// widthScale scales the border widths by this factor, useful for drawing half
// borders for border types like groove or double.
//
// The explicit float32 conversions around products keep each operation
// rounded to float as Java does; without them Go may fuse a multiplication
// and an addition into one operation on some architectures.
func BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds *geom.Rectangle, side int, border *BorderPropertySet, drawInterior bool, scaledOffset float32, widthScale float32, overlap bool) *geom.Path2D {
	/*
	 * Function overview: Prior to creating the path we check what side were building this on. All the coordinates in this function assume its building a top border
	 * the border is then rotated and translated to its appropriate side. Uses of "left" and "right" are assuming a perspective of inside the shape looking out.
	 */
	// we do not want any overlap for shape used to render background area
	// we want overlap for borders slightly, reducing the integer roundoff error in painting. This removes the tiny white line
	// between the 2 different borders. The better way would be to calculate the end location of the other border side and use that instead.
	var overlapAngle float32
	if overlap {
		overlapAngle = 1
	}
	border = border.NormalizedInstance(geom.NewRectangleWidthHeight(bounds.Width, bounds.Height))

	props := newBorderPainterRelativeBorderProperties(border, side, widthScale)
	var sideWidth float32
	if props.isDimensionsSwapped() {
		sideWidth = float32(bounds.Height) - float32(float32((1+scaledOffset)*widthScale)*(border.Top()+border.Bottom()))
	} else {
		sideWidth = float32(bounds.Width) - float32(float32((1+scaledOffset)*widthScale)*(border.Left()+border.Right()))
	}
	path := geom.NewPath2DFloat()

	var fullAngle float32 = 90
	defaultAngle := fullAngle / 2
	angle := defaultAngle
	widthSum := props.getTop() + props.getLeft()
	if widthSum != 0.0 { // Avoid NaN
		angle = float32(fullAngle*props.getTop()) / widthSum
	}
	borderPainterAppendPath(path, 0-props.getLeft(), 0-props.getTop(), props.getLeftCorner().Left(), props.getLeftCorner().Right(), 90+angle+overlapAngle, -angle-overlapAngle, props.getTop(), props.getLeft(), scaledOffset, true)

	angle = defaultAngle
	widthSum = props.getTop() + props.getRight()
	if widthSum != 0.0 { // Avoid NaN
		angle = float32(fullAngle*props.getTop()) / widthSum
	}
	borderPainterAppendPath(path, sideWidth+props.getRight(), 0-props.getTop(), props.getRightCorner().Right(), props.getRightCorner().Left(), 90, -angle-overlapAngle, props.getTop(), props.getRight(), scaledOffset, false)

	if drawInterior {
		//border = border.normalizeBorderRadius(new Rectangle((int)(bounds.width), (int)(bounds.height)));
		//props = new RelativeBorderProperties(bounds, border, 0f, side, 1+scaledOffset, 1);

		borderPainterAppendPath(path, sideWidth, 0, props.getRightCorner().Right(), props.getRightCorner().Left(), 90-angle-overlapAngle, angle+overlapAngle, props.getTop(), props.getRight(), scaledOffset+1, false)

		angle = defaultAngle
		widthSum = props.getTop() + props.getLeft()
		if widthSum != 0.0 { // Avoid NaN
			angle = float32(fullAngle*props.getTop()) / widthSum
		}
		borderPainterAppendPath(path, 0, 0, props.getLeftCorner().Left(), props.getLeftCorner().Right(), 90, angle+overlapAngle, props.getTop(), props.getLeft(), scaledOffset+1, true)
		path.ClosePath()
	}

	var tx, ty float32
	if !props.isDimensionsSwapped() {
		tx = float32(-bounds.Width) / 2.0
	} else {
		tx = float32(-bounds.Height) / 2.0
	}
	if props.isDimensionsSwapped() {
		ty = float32(-bounds.Width) / 2.0
	} else {
		ty = float32(-bounds.Height) / 2.0
	}
	path.Transform(geom.AffineTransformGetTranslateInstance(
		float64(tx+float32((scaledOffset+1)*props.getLeft())),
		float64(ty+float32((scaledOffset+1)*props.getTop()))))
	path.Transform(geom.AffineTransformGetRotateInstance(
		props.getRotation()))
	// empirical: add 0.5 to better play with rasterization rules
	path.Transform(geom.AffineTransformGetTranslateInstance(
		float64(float32(bounds.Width)/2.0+float32(bounds.X)), float64(float32(bounds.Height)/2.0+float32(bounds.Y))))

	return path
}

func borderPainterAppendPath(path *geom.Path2D, xOffset float32, yOffset float32, radiusVert float32, radiusHoriz float32, startAngle float32, distance float32, topWidth float32, sideWidth float32, scaleOffset float32, left bool) {
	innerWidth := float32(2*radiusHoriz) - float32(scaleOffset*sideWidth) - float32(scaleOffset*sideWidth)
	innerHeight := float32(2*radiusVert) - float32(scaleOffset*topWidth) - float32(scaleOffset*topWidth)

	if innerWidth > 0 && innerHeight > 0 {
		// do arc
		x := xOffset
		if !left {
			x = xOffset - innerWidth
		}
		arc := geom.NewArc2DFloat(
			x,
			yOffset,
			innerWidth,
			innerHeight, startAngle, distance, geom.Arc2DOpen)
		path.Append(arc, true)
	} else {
		// do line
		if path.GetCurrentPoint() == nil {
			path.MoveTo(float64(xOffset), float64(yOffset))
		} else {
			path.LineTo(float64(xOffset), float64(yOffset))
		}
	}
}

// borderPainterRelativeBorderProperties ports the nested class
// BorderPainter.RelativeBorderProperties.
type borderPainterRelativeBorderProperties struct {
	top         float32
	left        float32
	right       float32
	leftCorner  *BorderRadiusCorner
	rightCorner *BorderRadiusCorner

	rotation          float64
	dimensionsSwapped bool
}

func newBorderPainterRelativeBorderProperties(props *BorderPropertySet, side int, widthScale float32) *borderPainterRelativeBorderProperties {
	r := &borderPainterRelativeBorderProperties{}
	if (side & BorderPainterTop) == BorderPainterTop {
		r.top = props.Top() * widthScale
		r.left = props.Left() * widthScale
		r.right = props.Right() * widthScale
		r.leftCorner = props.GetTopLeft()
		r.rightCorner = props.GetTopRight()
		r.rotation = 0
		r.dimensionsSwapped = false
	} else if (side & BorderPainterRight) == BorderPainterRight {
		r.top = props.Right() * widthScale
		r.left = props.Top() * widthScale
		r.right = props.Bottom() * widthScale
		r.leftCorner = props.GetTopRight()
		r.rightCorner = props.GetBottomRight()
		r.rotation = math.Pi / 2
		r.dimensionsSwapped = true
	} else if (side & BorderPainterBottom) == BorderPainterBottom {
		r.top = props.Bottom() * widthScale
		r.left = props.Right() * widthScale
		r.right = props.Left() * widthScale
		r.leftCorner = props.GetBottomRight()
		r.rightCorner = props.GetBottomLeft()
		r.rotation = math.Pi
		r.dimensionsSwapped = false
	} else if (side & BorderPainterLeft) == BorderPainterLeft {
		r.top = props.Left() * widthScale
		r.left = props.Bottom() * widthScale
		r.right = props.Top() * widthScale
		r.leftCorner = props.GetBottomLeft()
		r.rightCorner = props.GetTopLeft()
		r.rotation = 3 * math.Pi / 2
		r.dimensionsSwapped = true
	} else {
		panic(NewXRRuntimeException("No side found"))
	}
	return r
}

func (r *borderPainterRelativeBorderProperties) getRightCorner() *BorderRadiusCorner {
	return r.rightCorner
}

func (r *borderPainterRelativeBorderProperties) getLeftCorner() *BorderRadiusCorner {
	return r.leftCorner
}

func (r *borderPainterRelativeBorderProperties) getTop() float32 {
	return r.top
}

func (r *borderPainterRelativeBorderProperties) getLeft() float32 {
	return r.left
}

func (r *borderPainterRelativeBorderProperties) getRight() float32 {
	return r.right
}

func (r *borderPainterRelativeBorderProperties) getRotation() float64 {
	return r.rotation
}

func (r *borderPainterRelativeBorderProperties) isDimensionsSwapped() bool {
	return r.dimensionsSwapped
}

// BorderPainterPaint paints the given sides of the border. xOffset is for
// determining the starting point for patterns.
func BorderPainterPaint(
	bounds *geom.Rectangle, sides int, border *BorderPropertySet,
	ctx *RenderingContext, xOffset int) {
	if (sides&BorderPainterTop) == BorderPainterTop && border.NoTop() {
		sides -= BorderPainterTop
	}
	if (sides&BorderPainterLeft) == BorderPainterLeft && border.NoLeft() {
		sides -= BorderPainterLeft
	}
	if (sides&BorderPainterBottom) == BorderPainterBottom && border.NoBottom() {
		sides -= BorderPainterBottom
	}
	if (sides&BorderPainterRight) == BorderPainterRight && border.NoRight() {
		sides -= BorderPainterRight
	}

	//Now paint!
	if (sides&BorderPainterTop) == BorderPainterTop && border.TopColor() != FSColor(FSRGBColorTransparent) {
		borderPainterPaintBorderSide(ctx.GetOutputDevice(),
			border, bounds, BorderPainterTop, border.TopStyle(), xOffset)
	}
	if (sides&BorderPainterBottom) == BorderPainterBottom && border.BottomColor() != FSColor(FSRGBColorTransparent) {
		borderPainterPaintBorderSide(ctx.GetOutputDevice(),
			border, bounds, BorderPainterBottom, border.BottomStyle(), xOffset)
	}
	if (sides&BorderPainterLeft) == BorderPainterLeft && border.LeftColor() != FSColor(FSRGBColorTransparent) {
		borderPainterPaintBorderSide(ctx.GetOutputDevice(),
			border, bounds, BorderPainterLeft, border.LeftStyle(), xOffset)
	}
	if (sides&BorderPainterRight) == BorderPainterRight && border.RightColor() != FSColor(FSRGBColorTransparent) {
		borderPainterPaintBorderSide(ctx.GetOutputDevice(),
			border, bounds, BorderPainterRight, border.RightStyle(), xOffset)
	}
}

func borderPainterPaintBorderSide(outputDevice OutputDevice, border *BorderPropertySet,
	bounds *geom.Rectangle, currentSide int,
	borderSideStyle *IdentValue, xOffset int) {
	if borderSideStyle == IdentValueRidge || borderSideStyle == IdentValueGroove {
		var borderA *BorderPropertySet
		var borderB *BorderPropertySet
		if borderSideStyle == IdentValueRidge {
			borderA = border
			borderB = border.Darken()
		} else {
			borderA = border.Darken()
			borderB = border
		}
		borderPainterPaintBorderSideShape(
			outputDevice, bounds, borderA,
			borderB,
			0, 1, currentSide)
		borderPainterPaintBorderSideShape(
			outputDevice, bounds, borderB,
			borderA,
			1, 0.5, currentSide)
	} else if borderSideStyle == IdentValueOutset {
		borderPainterPaintBorderSideShape(outputDevice, bounds,
			border,
			border.Darken(),
			0, 1, currentSide)
	} else if borderSideStyle == IdentValueInset {
		borderPainterPaintBorderSideShape(outputDevice, bounds,
			border.Darken(),
			border,
			0, 1, currentSide)
	} else if borderSideStyle == IdentValueSolid {
		outputDevice.SetStroke(geom.NewBasicStrokeWithWidth(1.0))
		if currentSide == BorderPainterTop {
			outputDevice.SetColor(border.TopColor())
			outputDevice.Fill(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterTop, border, true, 0, 1, true))
		}
		if currentSide == BorderPainterRight {
			outputDevice.SetColor(border.RightColor())
			outputDevice.Fill(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterRight, border, true, 0, 1, true))
		}
		if currentSide == BorderPainterBottom {
			outputDevice.SetColor(border.BottomColor())
			outputDevice.Fill(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterBottom, border, true, 0, 1, true))
		}
		if currentSide == BorderPainterLeft {
			outputDevice.SetColor(border.LeftColor())
			outputDevice.Fill(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, BorderPainterLeft, border, true, 0, 1, true))
		}

	} else if borderSideStyle == IdentValueDouble {
		borderPainterPaintDoubleBorder(outputDevice, border, bounds, currentSide)
	} else {
		thickness := 0
		if currentSide == BorderPainterTop {
			thickness = calculatedStyleFloatToInt(border.Top())
		}
		if currentSide == BorderPainterBottom {
			thickness = calculatedStyleFloatToInt(border.Bottom())
		}
		if currentSide == BorderPainterRight {
			thickness = calculatedStyleFloatToInt(border.Right())
		}
		if currentSide == BorderPainterLeft {
			thickness = calculatedStyleFloatToInt(border.Left())
		}
		if borderSideStyle == IdentValueDashed {
			//outputDevice.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_OFF);
			borderPainterPaintPatternedRect(outputDevice, bounds, border, border, []float32{8.0 + float32(thickness*2), 4.0 + float32(thickness)}, currentSide, xOffset)
			//outputDevice.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);
		}
		if borderSideStyle == IdentValueDotted {
			// turn off antialiasing or the dots will be all blurry
			//outputDevice.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_OFF);
			borderPainterPaintPatternedRect(outputDevice, bounds, border, border, []float32{float32(thickness), float32(thickness)}, currentSide, xOffset)
			//outputDevice.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);
		}
	}
}

func borderPainterPaintDoubleBorder(
	outputDevice OutputDevice, border *BorderPropertySet,
	bounds *geom.Rectangle, currentSide int) {
	// draw outer border
	borderPainterPaintSolid(outputDevice, bounds, border, 0, float32(1)/3.0, currentSide)
	// draw inner border
	//paintSolid(outputDevice, bounds, border, 1, 1/3f, sides, currentSide, bevel);
	borderPainterPaintSolid(outputDevice, bounds, border, 2, float32(1)/3.0, currentSide)
}

// xOffset is for inline borders, to determine dash_phase of top and bottom.
func borderPainterPaintPatternedRect(outputDevice OutputDevice,
	bounds *geom.Rectangle, border *BorderPropertySet,
	color *BorderPropertySet, pattern []float32,
	currentSide int, xOffset int) {
	oldStroke := outputDevice.GetStroke()

	path := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, false, 0.5, 1, true)
	clip := geom.NewArea(BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, true, 0, 1, false))
	oldClip := outputDevice.GetClip()
	if oldClip != nil {
		// we need to respect the clip sent to us, get the intersection between the old and the new
		clip.Intersect(geom.NewArea(oldClip))
	}
	outputDevice.SetClip(clip)
	if currentSide == BorderPainterTop {
		outputDevice.SetColor(color.TopColor())
		outputDevice.SetStroke(geom.NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(float32(2*calculatedStyleFloatToInt(border.Top())), geom.BasicStrokeCapButt, geom.BasicStrokeJoinBevel, 0, pattern, float32(xOffset)))
		outputDevice.DrawBorderLine(
			path, BorderPainterTop, calculatedStyleFloatToInt(border.Top()), false)
	} else if currentSide == BorderPainterLeft {
		outputDevice.SetColor(color.LeftColor())
		outputDevice.SetStroke(geom.NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(float32(2*calculatedStyleFloatToInt(border.Left())), geom.BasicStrokeCapButt, geom.BasicStrokeJoinBevel, 0, pattern, 0))
		outputDevice.DrawBorderLine(
			path, BorderPainterLeft, calculatedStyleFloatToInt(border.Left()), false)
	} else if currentSide == BorderPainterRight {
		outputDevice.SetColor(color.RightColor())
		outputDevice.SetStroke(geom.NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(float32(2*calculatedStyleFloatToInt(border.Right())), geom.BasicStrokeCapButt, geom.BasicStrokeJoinBevel, 0, pattern, 0))
		outputDevice.DrawBorderLine(
			path, BorderPainterRight, calculatedStyleFloatToInt(border.Right()), false)
	} else if currentSide == BorderPainterBottom {
		outputDevice.SetColor(color.BottomColor())
		outputDevice.SetStroke(geom.NewBasicStrokeWithWidthCapJoinMiterlimitDashDashPhase(float32(2*calculatedStyleFloatToInt(border.Bottom())), geom.BasicStrokeCapButt, geom.BasicStrokeJoinBevel, 0, pattern, float32(xOffset)))
		outputDevice.DrawBorderLine(
			path, BorderPainterBottom, calculatedStyleFloatToInt(border.Bottom()), false)
	}

	outputDevice.SetClip(oldClip)
	outputDevice.SetStroke(oldStroke)
}

func borderPainterPaintBorderSideShape(outputDevice OutputDevice,
	bounds *geom.Rectangle, high *BorderPropertySet, low *BorderPropertySet,
	offset float32, scale float32, currentSide int) {
	if currentSide == BorderPainterTop {
		borderPainterPaintSolid(outputDevice, bounds, high, offset, scale, currentSide)
	} else if currentSide == BorderPainterBottom {
		borderPainterPaintSolid(outputDevice, bounds, low, offset, scale, currentSide)
	} else if currentSide == BorderPainterRight {
		borderPainterPaintSolid(outputDevice, bounds, low, offset, scale, currentSide)
	} else if currentSide == BorderPainterLeft {
		borderPainterPaintSolid(outputDevice, bounds, high, offset, scale, currentSide)
	}
}

func borderPainterPaintSolid(outputDevice OutputDevice,
	bounds *geom.Rectangle, border *BorderPropertySet,
	offset float32, scale float32, currentSide int) {

	if currentSide == BorderPainterTop {
		outputDevice.SetColor(border.TopColor())
		// draw a 1px border with a line instead of a polygon
		if calculatedStyleFloatToInt(border.Top()) == 1 {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, false, offset, scale, true)
			outputDevice.Draw(line)
		} else {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, true, offset, scale, true)
			// use polygons for borders over 1px wide
			outputDevice.Fill(line)
		}
	} else if currentSide == BorderPainterBottom {
		outputDevice.SetColor(border.BottomColor())
		if calculatedStyleFloatToInt(border.Bottom()) == 1 {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, false, offset, scale, true)
			outputDevice.Draw(line)
		} else {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, true, offset, scale, true)
			// use polygons for borders over 1px wide
			outputDevice.Fill(line)
		}
	} else if currentSide == BorderPainterRight {
		outputDevice.SetColor(border.RightColor())
		if calculatedStyleFloatToInt(border.Right()) == 1 {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, false, offset, scale, true)
			outputDevice.Draw(line)
		} else {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, true, offset, scale, true)
			// use polygons for borders over 1px wide
			outputDevice.Fill(line)
		}
	} else if currentSide == BorderPainterLeft {
		outputDevice.SetColor(border.LeftColor())
		if calculatedStyleFloatToInt(border.Left()) == 1 {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, false, offset, scale, true)
			outputDevice.Draw(line)
		} else {
			line := BorderPainterGenerateBorderShapeWithScaledOffsetWidthScaleOverlap(bounds, currentSide, border, true, offset, scale, true)
			// use polygons for borders over 1px wide
			outputDevice.Fill(line)
		}
	}
}
