// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/SizePropertyBuilder.java

package ufo

var sizePropertyBuilderAll = []*CSSName{CSSNameFsPageOrientation, CSSNameFsPageHeight, CSSNameFsPageWidth}
var sizePropertyBuilderPageOrientations = map[string]struct{}{"landscape": {}, "portrait": {}}
var sizePropertyBuilderAuto = NewPropertyValueIdentValue(IdentValueAuto)

type SizePropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewSizePropertyBuilder() *SizePropertyBuilder {
	b := &SizePropertyBuilder{}
	b.self = b
	return b
}

func (b *SizePropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	result := make([]*PropertyDeclaration, 0, 3)
	b.AssertFoundUpToValues(cssName, values, 3)

	if len(values) == 1 {
		value := values[0]

		b.CheckInheritAllowed(value, inheritAllowed)

		if value.GetCssValueType() == CSSValueCssInherit {
			return b.CheckInheritAll(sizePropertyBuilderAll, values, origin, important, inheritAllowed)
		} else if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			pageSize := PageSizeGetPageSize(value.GetStringValue())
			if pageSize != nil {
				result = sizePropertyBuilderAddPageSize(result, pageSize.Width(), pageSize.Height(), origin, important)
				return result
			}

			ident := b.CheckIdent(value)
			if ident == IdentValueLandscape || ident == IdentValuePortrait {
				result = sizePropertyBuilderAddPageSizeWithOrientation(result, value, sizePropertyBuilderAuto, sizePropertyBuilderAuto, origin, important)
				return result
			} else if ident == IdentValueAuto {
				result = sizePropertyBuilderAddPageSizeWithOrientation(result, value, value, value, origin, important)
				return result
			} else {
				panic(NewCSSParseException("Identifier "+ident.ToString()+" is not a valid value for "+cssName.ToString(), -1))
			}
		} else if b.IsLength(value) {
			result = sizePropertyBuilderAddPageSize(result, value, value, origin, important)
			return result
		} else {
			panic(NewCSSParseException("Value for "+cssName.ToString()+" must be a length or identifier", -1))
		}
	} else if len(values) == 2 {
		value1 := values[0]
		value2 := values[1]

		b.CheckInheritAllowed(value2, false)

		if b.IsLength(value1) && b.IsLength(value2) {
			result = sizePropertyBuilderAddPageSize(result, value1, value2, origin, important)
			return result
		} else if value1.GetPrimitiveType() == CSSPrimitiveValueCssIdent &&
			value2.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			if _, ok := sizePropertyBuilderPageOrientations[value2.GetStringValue()]; ok {
				temp := value1
				value1 = value2
				value2 = temp
			}

			sizePropertyBuilderValidatePageOrientation(value1)

			pageSize := PageSizeGetPageSize(value2.GetStringValue())
			if pageSize == nil {
				panic(NewCSSParseException("Value "+value2.ToString()+" is not a valid page size", -1))
			}

			result = sizePropertyBuilderAddPageSizeWithOrientation(result, value1, pageSize.Width(), pageSize.Height(), origin, important)
			return result
		} else {
			panic(NewCSSParseException("Invalid value for size property", -1))
		}
	} else if len(values) == 3 {
		value1 := values[0]
		value2 := values[1]
		value3 := values[2]

		b.CheckInheritAllowed(value3, false)

		if b.IsLength(value1) && b.IsLength(value2) && value3.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			sizePropertyBuilderValidatePageOrientation(value3)
			result = sizePropertyBuilderAddPageSizeWithOrientation(result, value3, value1, value2, origin, important)
			return result
		} else {
			panic(NewCSSParseException("Size property parsing error", -1))
		}
	} else {
		panic(NewCSSParseException("Invalid value count for size property", -1))
	}
}

// sizePropertyBuilderAddPageSize and
// sizePropertyBuilderAddPageSizeWithOrientation return result with the three
// declarations appended; Java appends to the list it is given.
func sizePropertyBuilderAddPageSize(result []*PropertyDeclaration, width *PropertyValue, height *PropertyValue,
	origin StylesheetInfoOrigin, important bool) []*PropertyDeclaration {
	return sizePropertyBuilderAddPageSizeWithOrientation(result, sizePropertyBuilderAuto, width, height, origin, important)
}

func sizePropertyBuilderAddPageSizeWithOrientation(result []*PropertyDeclaration, orientation *PropertyValue,
	width *PropertyValue, height *PropertyValue,
	origin StylesheetInfoOrigin, important bool) []*PropertyDeclaration {
	if width != nil {
		sizePropertyBuilderValidatePageDimension(width)
	}
	if height != nil {
		sizePropertyBuilderValidatePageDimension(height)
	}

	result = append(result, NewPropertyDeclaration(CSSNameFsPageOrientation, orientation, important, origin))
	result = append(result, NewPropertyDeclaration(CSSNameFsPageWidth, width, important, origin))
	result = append(result, NewPropertyDeclaration(CSSNameFsPageHeight, height, important, origin))
	return result
}

func sizePropertyBuilderValidatePageDimension(value *PropertyValue) {
	if value.GetFloatValue() < 0.0 {
		panic(NewCSSParseException("A page dimension may not be negative: "+propertyBuilderFloatToString(value.GetFloatValue()), -1))
	}
}

func sizePropertyBuilderValidatePageOrientation(orientation *PropertyValue) {
	if _, ok := sizePropertyBuilderPageOrientations[orientation.ToString()]; !ok {
		panic(NewCSSParseException("Value "+orientation.ToString()+" is not a valid page orientation", -1))
	}
}
