// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/TransformOriginPropertyBuilder.java

package ufo

// TransformOriginPropertyBuilder parses transform-origin, mirroring the (2-value) background-position grammar:
// a single length/percentage/keyword, or a horizontal/vertical pair. A lone top/bottom
// keyword implies a centered horizontal axis, and keyword pairs may appear in either order
// (e.g. top left and left top are equivalent).
type TransformOriginPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewTransformOriginPropertyBuilder() *TransformOriginPropertyBuilder {
	b := &TransformOriginPropertyBuilder{}
	b.self = b
	return b
}

func (b *TransformOriginPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin,
	important bool, inheritAllowed bool) []*PropertyDeclaration {
	b.AssertFoundUpToValues(cssName, values, 2)

	first := values[0]
	var second *PropertyValue
	if len(values) == 2 {
		second = values[1]
	}

	b.CheckInheritAllowed(first, inheritAllowed)
	if second == nil && first.GetCssValueType() == CSSValueCssInherit {
		return []*PropertyDeclaration{NewPropertyDeclaration(cssName, first, important, origin)}
	}
	if second != nil {
		b.CheckInheritAllowed(second, false)
	}

	b.CheckIdentLengthOrPercentType(cssName, first)
	if second == nil {
		if first.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
			return b.twoValues(cssName, first, b.percent(50), important, origin)
		}
	} else {
		b.CheckIdentLengthOrPercentType(cssName, second)
	}

	var firstIdent *IdentValue
	if first.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
		firstIdent = b.checkKeyword(cssName, first)
	}
	var secondIdent *IdentValue
	if second == nil {
		secondIdent = IdentValueCenter
	} else {
		if second.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			secondIdent = b.checkKeyword(cssName, second)
		}
	}

	if firstIdent == nil && secondIdent == nil {
		return b.twoValues(cssName, first, second, important, origin)
	} else if firstIdent != nil && secondIdent != nil {
		if firstIdent == IdentValueTop || firstIdent == IdentValueBottom || secondIdent == IdentValueLeft || secondIdent == IdentValueRight {
			swap := firstIdent
			firstIdent = secondIdent
			secondIdent = swap
		}
		b.checkAxisCombination(cssName, firstIdent, secondIdent)
		return b.twoValues(cssName, b.percentForIdent(firstIdent), b.percentForIdent(secondIdent), important, origin)
	} else if firstIdent != nil {
		b.checkAxisCombination(cssName, firstIdent, nil)
		return b.twoValues(cssName, b.percentForIdent(firstIdent), second, important, origin)
	} else {
		b.checkAxisCombination(cssName, nil, secondIdent)
		return b.twoValues(cssName, first, b.percentForIdent(secondIdent), important, origin)
	}
}

func (b *TransformOriginPropertyBuilder) checkAxisCombination(cssName *CSSName, firstIdent *IdentValue, secondIdent *IdentValue) {
	if firstIdent == IdentValueTop || firstIdent == IdentValueBottom || secondIdent == IdentValueLeft || secondIdent == IdentValueRight {
		panic(NewCSSParseException("Invalid combination of keywords in "+cssName.ToString(), -1))
	}
}

func (b *TransformOriginPropertyBuilder) checkKeyword(cssName *CSSName, value *PropertyValue) *IdentValue {
	ident := b.CheckIdent(value)
	if ident != IdentValueLeft && ident != IdentValueRight && ident != IdentValueTop && ident != IdentValueBottom && ident != IdentValueCenter {
		panic(NewCSSParseException("Invalid keyword '"+ident.ToString()+"' for "+cssName.ToString(), -1))
	}
	return ident
}

func (b *TransformOriginPropertyBuilder) percentForIdent(ident *IdentValue) *PropertyValue {
	if ident == IdentValueLeft || ident == IdentValueTop {
		return b.percent(0)
	}
	if ident == IdentValueRight || ident == IdentValueBottom {
		return b.percent(100)
	}
	return b.percent(50)
}

func (b *TransformOriginPropertyBuilder) twoValues(
	cssName *CSSName, horizontal *PropertyValue, vertical *PropertyValue, important bool, origin StylesheetInfoOrigin) []*PropertyDeclaration {
	return []*PropertyDeclaration{NewPropertyDeclaration(
		cssName, NewPropertyValueList(propertyBuilderValueList([]*PropertyValue{horizontal, vertical})), important, origin)}
}

func (b *TransformOriginPropertyBuilder) percent(percent float32) *PropertyValue {
	return NewPropertyValueFloat(CSSPrimitiveValueCssPercentage, percent, propertyBuilderFloatToString(percent)+"%")
}
