// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/BorderRadiusCorner.java

package ufo

import (
	"fmt"
	"strings"
)

var BorderRadiusCornerUndefined = NewBorderRadiusCorner(0, 0)

// borderRadiusCornerLength ports the private record BorderRadiusCorner.Length.
type borderRadiusCornerLength struct {
	value   float32
	percent bool
}

type BorderRadiusCorner struct {
	left  borderRadiusCornerLength
	right borderRadiusCornerLength
}

// TODO: FIXME the way values are passed from the CSS to the border corners really sucks, improve it

func NewBorderRadiusCorner(left float32, right float32) *BorderRadiusCorner {
	return &BorderRadiusCorner{
		left:  borderRadiusCornerLength{left, false},
		right: borderRadiusCornerLength{right, false},
	}
}

func NewBorderRadiusCornerWithFromValStyleCtx(fromVal *CSSName, style CalculatedStyleI, ctx CssContext) *BorderRadiusCorner {
	b := &BorderRadiusCorner{}
	value := style.ValueByName(fromVal)
	if lValues, ok := value.(*ListValue); ok {
		first := lValues.GetValues()[0].(*PropertyValue)
		second := first
		if len(lValues.GetValues()) > 1 {
			second = lValues.GetValues()[1].(*PropertyValue)
		}

		if fromVal == CSSNameBorderTopLeftRadius || fromVal == CSSNameBorderBottomRightRadius {
			b.right = b.calculate(fromVal, style, first, ctx)
			b.left = b.calculate(fromVal, style, second, ctx)
		} else if fromVal == CSSNameBorderTopRightRadius || fromVal == CSSNameBorderBottomLeftRadius {
			b.left = b.calculate(fromVal, style, first, ctx)
			b.right = b.calculate(fromVal, style, second, ctx)
		} else {
			panic(NewXRRuntimeException("Unknown border radius type: " + fromVal.ToString()))
		}
	} else if lv, ok := value.(*LengthValue); ok {

		if strings.Contains(lv.GetStringValue(), "%") {
			b.left = borderRadiusCornerLength{value.AsFloat() / 100.0, true}
			b.right = b.left
		} else {
			b.left = borderRadiusCornerLength{float32(calculatedStyleFloatToInt(lv.GetFloatProportionalTo(fromVal, 0, ctx))), false}
			b.right = b.left
		}
	} else {
		panic(NewXRRuntimeException("Unknown length value: " + fmt.Sprint(value)))
	}
	return b
}

func (b *BorderRadiusCorner) calculate(fromVal *CSSName, style CalculatedStyleI, value *PropertyValue, ctx CssContext) borderRadiusCornerLength {
	switch value.GetPrimitiveType() {
	case CSSPrimitiveValueCssPercentage:
		return borderRadiusCornerLength{value.GetFloatValue() / 100.0, true}
	default:
		return borderRadiusCornerLength{LengthValueCalcFloatProportionalValue(
			style,
			fromVal,
			value.GetCssText(),
			value.GetFloatValue(),
			value.GetPrimitiveType(),
			0,
			ctx),
			false,
		}
	}
}

func (b *BorderRadiusCorner) HasRadius() bool {
	return b.left.value > 0 || b.right.value > 0
}

func (b *BorderRadiusCorner) GetMaxLeft(max float32) float32 {
	if b.left.percent {
		return max * b.left.value
	}
	return min(b.left.value, max)
}

func (b *BorderRadiusCorner) GetMaxRight(max float32) float32 {
	if b.right.percent {
		return max * b.right.value
	}
	return min(b.right.value, max)
}

func (b *BorderRadiusCorner) Left() float32 {
	return b.left.value
}

func (b *BorderRadiusCorner) Right() float32 {
	return b.right.value
}
