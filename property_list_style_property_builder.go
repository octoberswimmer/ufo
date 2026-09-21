// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/ListStylePropertyBuilder.java

package ufo

var listStylePropertyBuilderAll = []*CSSName{
	CSSNameListStyleType, CSSNameListStylePosition, CSSNameListStyleImage}

type ListStylePropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewListStylePropertyBuilder() *ListStylePropertyBuilder {
	b := &ListStylePropertyBuilder{}
	b.self = b
	return b
}

func (b *ListStylePropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	inherited := b.CheckInheritAll(listStylePropertyBuilderAll, values, origin, important, inheritAllowed)
	if inherited != nil {
		return inherited
	}

	var listStyleType *PropertyDeclaration
	var listStylePosition *PropertyDeclaration
	var listStyleImage *PropertyDeclaration

	for _, value := range values {
		b.CheckInheritAllowed(value, false)
		typ := value.GetPrimitiveType()
		if typ == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)

			if ident == IdentValueNone {
				if listStyleType == nil {
					listStyleType = NewPropertyDeclaration(
						CSSNameListStyleType, value, important, origin)
				}

				if listStyleImage == nil {
					listStyleImage = NewPropertyDeclaration(
						CSSNameListStyleImage, value, important, origin)
				}
			} else if PrimitivePropertyBuildersListStylePositions.Get(ident.FS_ID) {
				if listStylePosition != nil {
					panic(NewCSSParseException("A list-style-position value cannot be set twice", -1))
				}

				listStylePosition = NewPropertyDeclaration(
					CSSNameListStylePosition, value, important, origin)
			} else if PrimitivePropertyBuildersListStyleTypes.Get(ident.FS_ID) {
				if listStyleType != nil {
					panic(NewCSSParseException("A list-style-type value cannot be set twice", -1))
				}

				listStyleType = NewPropertyDeclaration(
					CSSNameListStyleType, value, important, origin)
			}
		} else if typ == CSSPrimitiveValueCssUri {
			if listStyleImage != nil {
				panic(NewCSSParseException("A list-style-image value cannot be set twice", -1))
			}

			listStyleImage = NewPropertyDeclaration(
				CSSNameListStyleImage, value, important, origin)
		}
	}

	result := make([]*PropertyDeclaration, 0, 3)
	if listStyleType != nil {
		result = append(result, listStyleType)
	}
	if listStylePosition != nil {
		result = append(result, listStylePosition)
	}
	if listStyleImage != nil {
		result = append(result, listStyleImage)
	}

	return result
}
