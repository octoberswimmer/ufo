// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/ColumnsPropertyBuilder.java

package ufo

type ColumnsPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewColumnsPropertyBuilder() *ColumnsPropertyBuilder {
	b := &ColumnsPropertyBuilder{}
	b.self = b
	return b
}

func (b *ColumnsPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	b.AssertFoundUpToValues(cssName, values, 2)

	if len(values) == 1 && values[0].GetCssValueType() == CSSValueCssInherit {
		b.CheckInheritAllowed(values[0], inheritAllowed)
		return []*PropertyDeclaration{
			NewPropertyDeclaration(CSSNameColumnWidth, values[0], important, origin),
			NewPropertyDeclaration(CSSNameColumnCount, values[0], important, origin)}
	}

	columnWidth := NewPropertyValueIdentValue(IdentValueAuto)
	columnCount := NewPropertyValueIdentValue(IdentValueAuto)
	widthSpecified := false
	countSpecified := false

	for _, value := range values {
		b.CheckInheritAllowed(value, false)
		typ := value.GetPrimitiveType()

		if typ == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			if ident != IdentValueAuto {
				panic(NewCSSParseException("Only auto is allowed as an identifier in columns", -1))
			}
			if !widthSpecified {
				columnWidth = value
				widthSpecified = true
			} else if !countSpecified {
				columnCount = value
				countSpecified = true
			} else {
				panic(NewCSSParseException("Duplicate values in columns shorthand", -1))
			}
		} else if b.IsLength(value) {
			if widthSpecified {
				panic(NewCSSParseException("Column width specified more than once", -1))
			}
			if value.GetFloatValue() < 0.0 {
				panic(NewCSSParseException("column-width may not be negative", -1))
			}
			columnWidth = value
			widthSpecified = true
		} else if typ == CSSPrimitiveValueCssNumber &&
			int(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) == propertyBuilderRound(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) {
			if countSpecified {
				panic(NewCSSParseException("Column count specified more than once", -1))
			}
			if value.GetFloatValue() < 1.0 {
				panic(NewCSSParseException("column-count must be at least 1", -1))
			}
			columnCount = value
			countSpecified = true
		} else {
			panic(NewCSSParseException("Invalid value in columns shorthand", -1))
		}
	}

	result := make([]*PropertyDeclaration, 0, 2)
	result = append(result, NewPropertyDeclaration(CSSNameColumnWidth, columnWidth, important, origin))
	result = append(result, NewPropertyDeclaration(CSSNameColumnCount, columnCount, important, origin))
	return result
}
