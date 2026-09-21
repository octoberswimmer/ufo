// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/BorderSpacingPropertyBuilder.java

package ufo

var borderSpacingPropertyBuilderAll = []*CSSName{
	CSSNameFsBorderSpacingHorizontal, CSSNameFsBorderSpacingVertical}

type BorderSpacingPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewBorderSpacingPropertyBuilder() *BorderSpacingPropertyBuilder {
	b := &BorderSpacingPropertyBuilder{}
	b.self = b
	return b
}

func (b *BorderSpacingPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	result := b.CheckInheritAll(borderSpacingPropertyBuilderAll, values, origin, important, inheritAllowed)
	if result != nil {
		return result
	}

	b.AssertFoundUpToValues(CSSNameBorderSpacing, values, 2)

	var horizontalSpacing *PropertyDeclaration
	var verticalSpacing *PropertyDeclaration

	if len(values) == 1 {
		value := values[0]
		b.CheckLengthType(cssName, value)
		if value.GetFloatValue() < 0.0 {
			panic(NewCSSParseException("border-spacing may not be negative", -1))
		}
		horizontalSpacing = NewPropertyDeclaration(
			CSSNameFsBorderSpacingHorizontal, value, important, origin)
		verticalSpacing = NewPropertyDeclaration(
			CSSNameFsBorderSpacingVertical, value, important, origin)
	} else { /* len(values) == 2 */
		horizontal := values[0]
		b.CheckLengthType(cssName, horizontal)
		if horizontal.GetFloatValue() < 0.0 {
			panic(NewCSSParseException("border-spacing may not be negative", -1))
		}
		horizontalSpacing = NewPropertyDeclaration(
			CSSNameFsBorderSpacingHorizontal, horizontal, important, origin)

		vertical := values[1]
		b.CheckLengthType(cssName, vertical)
		if vertical.GetFloatValue() < 0.0 {
			panic(NewCSSParseException("border-spacing may not be negative", -1))
		}
		verticalSpacing = NewPropertyDeclaration(
			CSSNameFsBorderSpacingVertical, vertical, important, origin)
	}

	return []*PropertyDeclaration{horizontalSpacing, verticalSpacing}
}
