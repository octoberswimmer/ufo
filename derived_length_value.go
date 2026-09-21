// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/LengthValue.java

package ufo

import (
	"math"
	"strconv"
	"strings"
)

const (
	lengthValueMmPerCm int     = 10
	lengthValueCmPerIn float32 = 2.54
	lengthValuePtPerIn float32 = 1.0 / 72.0
	lengthValuePcPerPt float32 = 12
)

type LengthValue struct {
	DerivedValue

	// The specified length value, as a float; pulled from the CSS text
	lengthAsFloat float32

	style CalculatedStyleI

	// The specified primitive SAC data type given for this length, from the CSS text
	lengthPrimitiveType int16
}

func NewLengthValue(style CalculatedStyleI, name *CSSName, value *PropertyValue) *LengthValue {
	cssText := value.GetCssText()
	return &LengthValue{
		DerivedValue: NewDerivedValue(name, value.GetPrimitiveType(), &cssText, &cssText),

		style:               style,
		lengthAsFloat:       value.GetFloatValue(),
		lengthPrimitiveType: value.GetPrimitiveType(),
	}
}

func (l *LengthValue) AsFloat() float32 {
	return l.lengthAsFloat
}

// GetFloatProportionalTo computes a relative unit (e.g. percentage) as an absolute value, using
// the input value. Used for such properties whose parent value cannot be
// known before layout/render
//
// cssName is the name of the property. Returns the absolute value or
// computed absolute value.
func (l *LengthValue) GetFloatProportionalTo(cssName *CSSName,
	baseValue float32,
	ctx CssContext) float32 {
	return LengthValueCalcFloatProportionalValue(l.getStyle(),
		cssName,
		l.GetStringValue(),
		l.lengthAsFloat,
		l.lengthPrimitiveType,
		baseValue,
		ctx)
}

func (l *LengthValue) HasAbsoluteUnit() bool {
	return ValueConstantsIsAbsoluteUnit(l.GetCssSacUnitType())
}

func (l *LengthValue) IsDependentOnFontSize() bool {
	return l.lengthPrimitiveType == CSSPrimitiveValueCssExs ||
		l.lengthPrimitiveType == CSSPrimitiveValueCssEms
}

func LengthValueCalcFloatProportionalValue(style CalculatedStyleI,
	cssName *CSSName,
	stringValue string,
	relVal float32,
	primitiveType int16,
	baseValue float32,
	ctx CssContext) float32 {

	var absVal float32 = math.SmallestNonzeroFloat32

	// NOTE: we used to cache absolute values, but have removed that to see if it
	// really makes a difference, since the calculations are so simple. In any case, for DPI-relative
	// values we shouldn't be caching, unless we also check if the DPI is changed, which
	// would seem to obviate the advantage of caching anyway.
	switch primitiveType {
	case CSSPrimitiveValueCssPx:
		absVal = relVal * float32(ctx.GetDotsPerPixel())
	case CSSPrimitiveValueCssIn:
		absVal = relVal * lengthValueCmPerIn * float32(lengthValueMmPerCm) / ctx.GetMmPerDot()
	case CSSPrimitiveValueCssCm:
		absVal = relVal * float32(lengthValueMmPerCm) / ctx.GetMmPerDot()
	case CSSPrimitiveValueCssMm:
		absVal = relVal / ctx.GetMmPerDot()
	case CSSPrimitiveValueCssPt:
		absVal = relVal * lengthValuePtPerIn * lengthValueCmPerIn * float32(lengthValueMmPerCm) / ctx.GetMmPerDot()
	case CSSPrimitiveValueCssPc:
		absVal = relVal * lengthValuePcPerPt * lengthValuePtPerIn * lengthValueCmPerIn * float32(lengthValueMmPerCm) / ctx.GetMmPerDot()
	case CSSPrimitiveValueCssEms:
		// EM is equal to font-size of element on which it is used
		// The exception is when ?em? occurs in the value of
		// the ?font-size? property itself, in which case it refers
		// to the calculated font size of the parent element
		// http://www.w3.org/TR/CSS21/fonts.html#font-size-props
		if cssName == CSSNameFontSize {
			parentFont := style.GetParent().GetFont(ctx)
			//font size and FontSize2D should be identical
			absVal = relVal * parentFont.Size() //ctx.getFontSize2D(parentFont);
		} else {
			absVal = relVal * style.GetFont(ctx).Size() //ctx.getFontSize2D(style.getFont(ctx));
		}

	case CSSPrimitiveValueCssExs:
		// To convert EMS to pixels, we need the height of the lowercase 'Xx' character in the current
		// element...
		// to the font size of the parent element (spec: 4.3.2)
		var xHeight float32
		if cssName == CSSNameFontSize {
			parentFont := style.GetParent().GetFont(ctx)
			xHeight = ctx.GetXHeight(parentFont)
		} else {
			font := style.GetFont(ctx)
			xHeight = ctx.GetXHeight(font)
		}
		absVal = relVal * xHeight

	case CSSPrimitiveValueCssPercentage:
		// percentage depends on the property this value belongs to
		if cssName == CSSNameVerticalAlign {
			baseValue = style.GetParent().GetLineHeight(ctx)
		} else if cssName == CSSNameFontSize {
			// same as with EM
			parentFont := style.GetParent().GetFont(ctx)
			baseValue = ctx.GetFontSize2D(parentFont)
		} else if cssName == CSSNameLineHeight {
			font := style.GetFont(ctx)
			baseValue = ctx.GetFontSize2D(font)
		}
		absVal = relVal / 100.0 * baseValue

	case CSSPrimitiveValueCssNumber:
		absVal = relVal
	default:
		// nothing to do, we only convert those listed above
		XRLogCascade(LevelSevere,
			"Asked to convert "+cssName.ToString()+" from relative to absolute, "+
				" don't recognize the datatype "+
				"'"+ValueConstantsStringForSACPrimitiveType(primitiveType)+"' "+
				strconv.Itoa(int(primitiveType))+"("+stringValue+")")
	}
	//assert (new Float(absVal).intValue() >= 0);

	if XRLogIsLoggingEnabled() {
		if cssName == CSSNameFontSize {
			XRLogCascade(LevelFinest, cssName.ToString()+", relative= "+
				lengthValueFloatToString(relVal)+" ("+stringValue+"), absolute= "+
				lengthValueFloatToString(absVal))
		} else {
			XRLogCascade(LevelFinest, cssName.ToString()+", relative= "+
				lengthValueFloatToString(relVal)+" ("+stringValue+"), absolute= "+
				lengthValueFloatToString(absVal)+" using base="+lengthValueFloatToString(baseValue))
		}
	}

	absVal = lengthValueRound(absVal)
	return absVal
}

// lengthValueRound is (float) Math.round((double) absVal): the nearest long,
// ties toward positive infinity, NaN to 0, saturating at the long range.
func lengthValueRound(absVal float32) float32 {
	d := float64(absVal)
	if d != d {
		return 0
	}
	r := math.Floor(d + 0.5)
	if r >= math.MaxInt64 {
		return float32(math.MaxInt64)
	}
	if r <= math.MinInt64 {
		return float32(math.MinInt64)
	}
	return float32(int64(r))
}

// lengthValueFloatToString formats a float the way Java's string
// concatenation does (Float.toString): at least one fraction digit, and
// computerized scientific notation outside [1e-3, 1e7).
func lengthValueFloatToString(f float32) string {
	d := float64(f)
	switch {
	case d != d:
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
		s := strconv.FormatFloat(d, 'f', -1, 32)
		if !strings.Contains(s, ".") {
			s += ".0"
		}
		return s
	}
	s := strconv.FormatFloat(d, 'e', -1, 32)
	mantissa, exponent, _ := strings.Cut(s, "e")
	if !strings.Contains(mantissa, ".") {
		mantissa += ".0"
	}
	exp, _ := strconv.Atoi(exponent)
	return mantissa + "E" + strconv.Itoa(exp)
}

func (l *LengthValue) getStyle() CalculatedStyleI {
	return l.style
}
