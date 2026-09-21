// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/BackgroundPropertyBuilder.java

package ufo

import "strings"

// [<'background-color'> || <'background-image'> || <'background-repeat'> ||
// <'background-attachment'> || <'background-position'>] | inherit
var backgroundPropertyBuilderAll = []*CSSName{
	CSSNameBackgroundColor, CSSNameBackgroundImage, CSSNameBackgroundRepeat,
	CSSNameBackgroundAttachment, CSSNameBackgroundPosition}

type BackgroundPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewBackgroundPropertyBuilder() *BackgroundPropertyBuilder {
	b := &BackgroundPropertyBuilder{}
	b.self = b
	return b
}

func (b *BackgroundPropertyBuilder) isAppliesToBackgroundPosition(value *PropertyValue) bool {
	typ := value.GetPrimitiveType()

	if b.IsLength(value) || typ == CSSPrimitiveValueCssPercentage {
		return true
	} else if typ != CSSPrimitiveValueCssIdent {
		return false
	} else {
		ident := IdentValueValueOf(value.GetStringValue())
		return ident != nil &&
			PrimitivePropertyBuildersBackgroundPositions.Get(ident.FS_ID)
	}
}

func (b *BackgroundPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	result := b.CheckInheritAll(backgroundPropertyBuilderAll, values, origin, important, inheritAllowed)
	if result != nil {
		return result
	}

	var backgroundColor *PropertyDeclaration
	var backgroundImage *PropertyDeclaration
	var backgroundRepeat *PropertyDeclaration
	var backgroundAttachment *PropertyDeclaration
	var backgroundPosition *PropertyDeclaration

	for i := 0; i < len(values); i++ {
		value := values[i]
		b.CheckInheritAllowed(value, false)

		processingBackgroundPosition := false
		typ := value.GetPrimitiveType()
		if typ == CSSPrimitiveValueCssIdent {
			color := ConversionsGetColor(value.GetStringValue())
			if color != nil {
				if backgroundColor != nil {
					panic(NewCSSParseException("A background-color value cannot be set twice", -1))
				}

				backgroundColor = NewPropertyDeclaration(
					CSSNameBackgroundColor,
					NewPropertyValueFSColor(color),
					important, origin)
				continue
			}

			ident := b.CheckIdent(value)

			if PrimitivePropertyBuildersBackgroundRepeats.Get(ident.FS_ID) {
				if backgroundRepeat != nil {
					panic(NewCSSParseException("A background-repeat value cannot be set twice", -1))
				}

				backgroundRepeat = NewPropertyDeclaration(
					CSSNameBackgroundRepeat, value, important, origin)
			}

			if PrimitivePropertyBuildersBackgroundAttachments.Get(ident.FS_ID) {
				if backgroundAttachment != nil {
					panic(NewCSSParseException("A background-attachment value cannot be set twice", -1))
				}

				backgroundAttachment = NewPropertyDeclaration(
					CSSNameBackgroundAttachment, value, important, origin)
			}

			if ident == IdentValueTransparent {
				if backgroundColor != nil {
					panic(NewCSSParseException("A background-color value cannot be set twice", -1))
				}

				backgroundColor = NewPropertyDeclaration(
					CSSNameBackgroundColor, value, important, origin)
			}

			if ident == IdentValueNone {
				if backgroundImage != nil {
					panic(NewCSSParseException("A background-image value cannot be set twice", -1))
				}

				backgroundImage = NewPropertyDeclaration(
					CSSNameBackgroundImage, value, important, origin)
			}

			if PrimitivePropertyBuildersBackgroundPositions.Get(ident.FS_ID) {
				processingBackgroundPosition = true
			}
		} else if typ == CSSPrimitiveValueCssRgbcolor {
			if backgroundColor != nil {
				panic(NewCSSParseException("A background-color value cannot be set twice", -1))
			}

			backgroundColor = NewPropertyDeclaration(
				CSSNameBackgroundColor, value, important, origin)
		} else if typ == CSSPrimitiveValueCssUri || strings.HasPrefix(value.ToString(), IdentValueLinearGradient.AsString()) {
			if backgroundImage != nil {
				panic(NewCSSParseException("A background-image value cannot be set twice", -1))
			}

			backgroundImage = NewPropertyDeclaration(
				CSSNameBackgroundImage, value, important, origin)
		}

		if processingBackgroundPosition || b.IsLength(value) || typ == CSSPrimitiveValueCssPercentage {
			if backgroundPosition != nil {
				panic(NewCSSParseException("A background-position value cannot be set twice", -1))
			}

			v := make([]*PropertyValue, 0, 2)
			v = append(v, value)
			if i < len(values)-1 {
				next := values[i+1]
				if b.isAppliesToBackgroundPosition(next) {
					v = append(v, next)
					i++
				}
			}

			builder := CSSNameGetPropertyBuilder(CSSNameBackgroundPosition)
			backgroundPosition = builder.BuildDeclarations(
				CSSNameBackgroundPosition, v, origin, important)[0]
		}
	}

	if backgroundColor == nil {
		backgroundColor = NewPropertyDeclaration(
			CSSNameBackgroundColor, NewPropertyValueIdentValue(IdentValueTransparent), important, origin)
	}

	if backgroundImage == nil {
		backgroundImage = NewPropertyDeclaration(
			CSSNameBackgroundImage, NewPropertyValueIdentValue(IdentValueNone), important, origin)
	}

	if backgroundRepeat == nil {
		backgroundRepeat = NewPropertyDeclaration(
			CSSNameBackgroundRepeat, NewPropertyValueIdentValue(IdentValueRepeat), important, origin)
	}

	if backgroundAttachment == nil {
		backgroundAttachment = NewPropertyDeclaration(
			CSSNameBackgroundAttachment, NewPropertyValueIdentValue(IdentValueScroll), important, origin)

	}

	if backgroundPosition == nil {
		v := make([]*PropertyValue, 0, 2)
		v = append(v, NewPropertyValueFloat(CSSPrimitiveValueCssPercentage, 0.0, "0%"))
		v = append(v, NewPropertyValueFloat(CSSPrimitiveValueCssPercentage, 0.0, "0%"))
		backgroundPosition = NewPropertyDeclaration(
			CSSNameBackgroundPosition, NewPropertyValueList(propertyBuilderValueList(v)), important, origin)
	}

	return []*PropertyDeclaration{backgroundColor, backgroundImage, backgroundRepeat, backgroundAttachment, backgroundPosition}
}
