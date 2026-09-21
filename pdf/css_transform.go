// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/CssTransform.java

package pdf

import (
	"math"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/geom"
)

// CssTransform builds the geom.AffineTransform for a box's CSS
// transform/transform-origin, for use by ITextOutputDevice. Percentages in
// translate() and transform-origin are resolved against the box's border box,
// matching the CSS default.

// cssTransformDegreesToRadians is the constant Math.toRadians multiplies by.
const cssTransformDegreesToRadians = 0.017453292519943295

// CssTransformToAffineTransform returns nil when the box has no transform.
func CssTransformToAffineTransform(ctx ufo.CssContext, box ufo.BoxI) *geom.AffineTransform {
	style := box.GetStyle()
	functions := style.GetTransforms()
	if len(functions) == 0 {
		return nil
	}

	bounds := box.GetPaintingBorderEdge(ctx)
	transformOrigin := style.GetTransformOrigin()
	originX := float32(bounds.X) + cssTransformResolveLength(ctx, style, transformOrigin.GetHorizontal(), float32(bounds.Width))
	originY := float32(bounds.Y) + cssTransformResolveLength(ctx, style, transformOrigin.GetVertical(), float32(bounds.Height))

	result := geom.NewAffineTransform()
	result.Translate(float64(originX), float64(originY))
	for _, function := range functions {
		result.Concatenate(CssTransformToMatrix(ctx, style, function.GetFunction(), bounds))
	}
	result.Translate(float64(-originX), float64(-originY))

	return result
}

func CssTransformToMatrix(ctx ufo.CssContext, style ufo.CalculatedStyleI, function *ufo.FSFunction, bounds *geom.Rectangle) *geom.AffineTransform {
	args := function.GetParameters()

	// CSSParser lower-cases FUNCTION tokens (CSS identifiers are case-insensitive), so e.g.
	// "skewX" always arrives here as "skewx".
	switch function.GetName() {
	case "matrix":
		return geom.NewAffineTransformWithM00M10M01M11M02M12(
			float64(cssTransformNumber(args, 0)), float64(cssTransformNumber(args, 1)), float64(cssTransformNumber(args, 2)),
			float64(cssTransformNumber(args, 3)), float64(cssTransformNumber(args, 4)), float64(cssTransformNumber(args, 5)))
	case "translate":
		var ty float32
		if len(args) == 2 {
			ty = cssTransformResolveLength(ctx, style, args[1], float32(bounds.Height))
		}
		return geom.AffineTransformGetTranslateInstance(
			float64(cssTransformResolveLength(ctx, style, args[0], float32(bounds.Width))),
			float64(ty))
	case "translatex":
		return geom.AffineTransformGetTranslateInstance(
			float64(cssTransformResolveLength(ctx, style, args[0], float32(bounds.Width))), 0)
	case "translatey":
		return geom.AffineTransformGetTranslateInstance(
			0, float64(cssTransformResolveLength(ctx, style, args[0], float32(bounds.Height))))
	case "scale":
		sx := cssTransformNumber(args, 0)
		sy := sx
		if len(args) == 2 {
			sy = cssTransformNumber(args, 1)
		}
		return geom.AffineTransformGetScaleInstance(float64(sx), float64(sy))
	case "scalex":
		return geom.AffineTransformGetScaleInstance(float64(cssTransformNumber(args, 0)), 1)
	case "scaley":
		return geom.AffineTransformGetScaleInstance(1, float64(cssTransformNumber(args, 0)))
	case "rotate":
		return geom.AffineTransformGetRotateInstance(CssTransformAngleToRadians(args[0]))
	case "skew":
		var shy float64
		if len(args) == 2 {
			shy = math.Tan(CssTransformAngleToRadians(args[1]))
		}
		return geom.AffineTransformGetShearInstance(
			math.Tan(CssTransformAngleToRadians(args[0])),
			shy)
	case "skewx":
		return geom.AffineTransformGetShearInstance(math.Tan(CssTransformAngleToRadians(args[0])), 0)
	case "skewy":
		return geom.AffineTransformGetShearInstance(0, math.Tan(CssTransformAngleToRadians(args[0])))
	default:
		return geom.NewAffineTransform()
	}
}

func cssTransformNumber(args []*ufo.PropertyValue, index int) float32 {
	return args[index].GetFloatValue()
}

func CssTransformAngleToRadians(angle *ufo.PropertyValue) float64 {
	value := angle.GetFloatValue()
	switch angle.GetPrimitiveType() {
	case ufo.CSSPrimitiveValueCssRad:
		return float64(value)
	case ufo.CSSPrimitiveValueCssGrad:
		return float64(value) * 0.9 * cssTransformDegreesToRadians
	default:
		return float64(value) * cssTransformDegreesToRadians // CSS_DEG (also covers "turn", pre-converted to degrees at parse time)
	}
}

func cssTransformResolveLength(ctx ufo.CssContext, style ufo.CalculatedStyleI, value *ufo.PropertyValue, percentageBase float32) float32 {
	return ufo.LengthValueCalcFloatProportionalValue(
		style, ufo.CSSNameTransform, value.GetCssText(), value.GetFloatValue(),
		value.GetPrimitiveType(), percentageBase, ctx)
}
