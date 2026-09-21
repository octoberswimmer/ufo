// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/BorderPropertyBuilders.java

package ufo

// borderPropertyBuildersBorderSidePropertyBuilder is the Java abstract class
// BorderSidePropertyBuilder. The abstract getProperties() is the getProperties
// field, set by the constructor of each subclass. It stays a function so that
// the CSSName constants are read when declarations are built, as in Java, and
// not while the CSSName constants are still being initialized.
type borderPropertyBuildersBorderSidePropertyBuilder struct {
	AbstractPropertyBuilder
	getProperties func() [][]*CSSName
}

func newBorderPropertyBuildersBorderSidePropertyBuilder(getProperties func() [][]*CSSName) *borderPropertyBuildersBorderSidePropertyBuilder {
	b := &borderPropertyBuildersBorderSidePropertyBuilder{getProperties: getProperties}
	b.self = b
	return b
}

func (b *borderPropertyBuildersBorderSidePropertyBuilder) addAll(result []*PropertyDeclaration, properties []*CSSName, value *PropertyValue,
	origin StylesheetInfoOrigin, important bool) []*PropertyDeclaration {
	for _, property := range properties {
		result = append(result, NewPropertyDeclaration(
			property, value, important, origin))
	}
	return result
}

func (b *borderPropertyBuildersBorderSidePropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	props := b.getProperties()

	result := make([]*PropertyDeclaration, 0, 3)

	if len(values) == 1 && values[0].GetCssValueType() == CSSValueCssInherit {
		value := values[0]
		result = b.addAll(result, props[0], value, origin, important)
		result = b.addAll(result, props[1], value, origin, important)
		result = b.addAll(result, props[2], value, origin, important)

	} else {
		b.AssertFoundUpToValues(cssName, values, 3)
		haveBorderStyle := false
		haveBorderColor := false
		haveBorderWidth := false

		for _, value := range values {
			b.CheckInheritAllowed(value, false)
			matched := false
			borderWidth := b.convertToBorderWidth(value)
			if borderWidth != nil {
				if haveBorderWidth {
					panic(NewCSSParseException("A border width cannot be set twice", -1))
				}
				haveBorderWidth = true
				matched = true
				result = b.addAll(result, props[0], borderWidth, origin, important)
			}

			if b.isBorderStyle(value) {
				if haveBorderStyle {
					panic(NewCSSParseException("A border style cannot be set twice", -1))
				}
				haveBorderStyle = true
				matched = true
				result = b.addAll(result, props[1], value, origin, important)
			}

			borderColor := b.convertToBorderColor(value)
			if borderColor != nil {
				if haveBorderColor {
					panic(NewCSSParseException("A border color cannot be set twice", -1))
				}
				haveBorderColor = true
				matched = true
				result = b.addAll(result, props[2], borderColor, origin, important)
			}

			if !matched {
				panic(NewCSSParseException(value.GetCssText()+" is not a border width, style, or color", -1))
			}
		}

		if !haveBorderWidth {
			result = b.addAll(result, props[0], NewPropertyValueIdentValue(IdentValueFsInitialValue), origin, important)
		}

		if !haveBorderStyle {
			result = b.addAll(result, props[1], NewPropertyValueIdentValue(IdentValueFsInitialValue), origin, important)
		}

		if !haveBorderColor {
			result = b.addAll(result, props[2], NewPropertyValueIdentValue(IdentValueFsInitialValue), origin, important)
		}

	}
	return result
}

func (b *borderPropertyBuildersBorderSidePropertyBuilder) isBorderStyle(value *PropertyValue) bool {
	if value.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
		return false
	}

	ident := IdentValueValueOf(value.GetCssText())
	if ident == nil {
		return false
	}

	return PrimitivePropertyBuildersBorderStyles.Get(ident.FS_ID)
}

// convertToBorderWidth returns nil when value is not a border width.
func (b *borderPropertyBuildersBorderSidePropertyBuilder) convertToBorderWidth(value *PropertyValue) *PropertyValue {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent && !b.IsLength(value) {
		return nil
	}

	if b.IsLength(value) {
		return value
	} else {
		ident := IdentValueValueOf(value.GetStringValue())
		if ident == nil {
			return nil
		}

		if PrimitivePropertyBuildersBorderWidths.Get(ident.FS_ID) {
			return ConversionsGetBorderWidth(ident.ToString())
		} else {
			return nil
		}
	}
}

// convertToBorderColor returns nil when value is not a border color.
func (b *borderPropertyBuildersBorderSidePropertyBuilder) convertToBorderColor(value *PropertyValue) *PropertyValue {
	typ := value.GetPrimitiveType()
	if typ != CSSPrimitiveValueCssIdent && typ != CSSPrimitiveValueCssRgbcolor {
		return nil
	}

	if typ != CSSPrimitiveValueCssRgbcolor {
		color := ConversionsGetColor(value.GetStringValue())
		if color != nil {
			return NewPropertyValueFSColor(color)
		}

		ident := IdentValueValueOf(value.GetCssText())
		if ident == nil || ident != IdentValueTransparent {
			return nil
		}
	}
	return value
}

type BorderPropertyBuildersBorderTop struct {
	borderPropertyBuildersBorderSidePropertyBuilder
}

func NewBorderPropertyBuildersBorderTop() *BorderPropertyBuildersBorderTop {
	b := &BorderPropertyBuildersBorderTop{*newBorderPropertyBuildersBorderSidePropertyBuilder(func() [][]*CSSName {
		return [][]*CSSName{
			{CSSNameBorderTopWidth},
			{CSSNameBorderTopStyle},
			{CSSNameBorderTopColor}}
	})}
	b.self = b
	return b
}

type BorderPropertyBuildersBorderRight struct {
	borderPropertyBuildersBorderSidePropertyBuilder
}

func NewBorderPropertyBuildersBorderRight() *BorderPropertyBuildersBorderRight {
	b := &BorderPropertyBuildersBorderRight{*newBorderPropertyBuildersBorderSidePropertyBuilder(func() [][]*CSSName {
		return [][]*CSSName{
			{CSSNameBorderRightWidth},
			{CSSNameBorderRightStyle},
			{CSSNameBorderRightColor}}
	})}
	b.self = b
	return b
}

type BorderPropertyBuildersBorderBottom struct {
	borderPropertyBuildersBorderSidePropertyBuilder
}

func NewBorderPropertyBuildersBorderBottom() *BorderPropertyBuildersBorderBottom {
	b := &BorderPropertyBuildersBorderBottom{*newBorderPropertyBuildersBorderSidePropertyBuilder(func() [][]*CSSName {
		return [][]*CSSName{
			{CSSNameBorderBottomWidth},
			{CSSNameBorderBottomStyle},
			{CSSNameBorderBottomColor}}
	})}
	b.self = b
	return b
}

type BorderPropertyBuildersBorderLeft struct {
	borderPropertyBuildersBorderSidePropertyBuilder
}

func NewBorderPropertyBuildersBorderLeft() *BorderPropertyBuildersBorderLeft {
	b := &BorderPropertyBuildersBorderLeft{*newBorderPropertyBuildersBorderSidePropertyBuilder(func() [][]*CSSName {
		return [][]*CSSName{
			{CSSNameBorderLeftWidth},
			{CSSNameBorderLeftStyle},
			{CSSNameBorderLeftColor}}
	})}
	b.self = b
	return b
}

type BorderPropertyBuildersBorder struct {
	borderPropertyBuildersBorderSidePropertyBuilder
}

func NewBorderPropertyBuildersBorder() *BorderPropertyBuildersBorder {
	b := &BorderPropertyBuildersBorder{*newBorderPropertyBuildersBorderSidePropertyBuilder(func() [][]*CSSName {
		return [][]*CSSName{
			{
				CSSNameBorderTopWidth, CSSNameBorderRightWidth,
				CSSNameBorderBottomWidth, CSSNameBorderLeftWidth},
			{
				CSSNameBorderTopStyle, CSSNameBorderRightStyle,
				CSSNameBorderBottomStyle, CSSNameBorderLeftStyle},
			{
				CSSNameBorderTopColor, CSSNameBorderRightColor,
				CSSNameBorderBottomColor, CSSNameBorderLeftColor}}
	})}
	b.self = b
	return b
}
