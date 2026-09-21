// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/FontPropertyBuilder.java

package ufo

import "strings"

// [ [ <'font-style'> || <'font-variant'> || <'font-weight'> ]? <'font-size'> [ / <'line-height'> ]? <'font-family'> ]
var fontPropertyBuilderAll = []*CSSName{
	CSSNameFontStyle, CSSNameFontVariant, CSSNameFontWeight,
	CSSNameFontSize, CSSNameLineHeight, CSSNameFontFamily}

type FontPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewFontPropertyBuilder() *FontPropertyBuilder {
	b := &FontPropertyBuilder{}
	b.self = b
	return b
}

func (b *FontPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	result := b.CheckInheritAll(fontPropertyBuilderAll, values, origin, important, inheritAllowed)
	if result != nil {
		return result
	}

	var fontStyle *PropertyDeclaration
	var fontVariant *PropertyDeclaration
	var fontWeight *PropertyDeclaration
	var fontSize *PropertyDeclaration
	var lineHeight *PropertyDeclaration
	var fontFamily *PropertyDeclaration

	keepGoing := false

	// i is the cursor of the Java ListIterator: next() reads values[i] and
	// advances, previous() steps back.
	i := 0
	for i < len(values) {
		value := values[i]
		i++
		typ := value.GetPrimitiveType()
		if typ == CSSPrimitiveValueCssIdent {
			// The parser will have given us ident values as they appear
			// (case-wise) in the CSS text since we might be creating
			// a font-family list out of them.  Here we want the normalized
			// (lowercase) version though.
			lowerCase := strings.ToLower(value.GetStringValue())
			value = NewPropertyValueString(CSSPrimitiveValueCssIdent, lowerCase, lowerCase)
			ident := b.CheckIdent(value)
			if ident == IdentValueNormal { // skip to avoid double set false positives
				continue
			}
			if PrimitivePropertyBuildersFontStyles.Get(ident.FS_ID) {
				if fontStyle != nil {
					panic(NewCSSParseException("font-style cannot be set twice", -1))
				}
				fontStyle = NewPropertyDeclaration(CSSNameFontStyle, value, important, origin)
			} else if PrimitivePropertyBuildersFontVariants.Get(ident.FS_ID) {
				if fontVariant != nil {
					panic(NewCSSParseException("font-variant cannot be set twice", -1))
				}
				fontVariant = NewPropertyDeclaration(CSSNameFontVariant, value, important, origin)
			} else if PrimitivePropertyBuildersFontWeights.Get(ident.FS_ID) {
				if fontWeight != nil {
					panic(NewCSSParseException("font-weight cannot be set twice", -1))
				}
				fontWeight = NewPropertyDeclaration(CSSNameFontWeight, value, important, origin)
			} else {
				keepGoing = true
				break
			}
		} else if typ == CSSPrimitiveValueCssNumber && value.GetFloatValue() > 0 {
			if fontWeight != nil {
				panic(NewCSSParseException("font-weight cannot be set twice", -1))
			}

			weight := ConversionsGetNumericFontWeight(value.GetFloatValue())
			if weight == nil {
				panic(NewCSSParseException(value.ToString()+" is not a valid font weight", -1))
			}

			replacement := NewPropertyValueString(
				CSSPrimitiveValueCssIdent, weight.ToString(), weight.ToString())
			replacement.SetIdentValue(weight)

			fontWeight = NewPropertyDeclaration(CSSNameFontWeight, replacement, important, origin)
		} else {
			keepGoing = true
			break
		}
	}

	if keepGoing {
		i--
		value := values[i]
		i++

		if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			lowerCase := strings.ToLower(value.GetStringValue())
			value = NewPropertyValueString(CSSPrimitiveValueCssIdent, lowerCase, lowerCase)
		}

		fontSizeBuilder := CSSNameGetPropertyBuilder(CSSNameFontSize)
		l := fontSizeBuilder.BuildDeclarations(
			CSSNameFontSize, []*PropertyValue{value}, origin, important)

		fontSize = l[0]

		if i < len(values) {
			value = values[i]
			i++
			if value.GetOperator() == TokenTkVirgule {
				lineHeightBuilder := CSSNameGetPropertyBuilder(CSSNameLineHeight)
				l = lineHeightBuilder.BuildDeclarations(
					CSSNameLineHeight, []*PropertyValue{value}, origin, important)
				lineHeight = l[0]
			} else {
				i--
			}
		}

		if i < len(values) {
			families := []*PropertyValue{}
			for i < len(values) {
				families = append(families, values[i])
				i++
			}
			fontFamilyBuilder := CSSNameGetPropertyBuilder(CSSNameFontFamily)
			l = fontFamilyBuilder.BuildDeclarations(
				CSSNameFontFamily, families, origin, important)
			fontFamily = l[0]
		}
	}

	if fontStyle == nil {
		fontStyle = NewPropertyDeclaration(
			CSSNameFontStyle, NewPropertyValueIdentValue(IdentValueNormal), important, origin)
	}

	if fontVariant == nil {
		fontVariant = NewPropertyDeclaration(
			CSSNameFontVariant, NewPropertyValueIdentValue(IdentValueNormal), important, origin)
	}

	if fontWeight == nil {
		fontWeight = NewPropertyDeclaration(
			CSSNameFontWeight, NewPropertyValueIdentValue(IdentValueNormal), important, origin)
	}

	if fontSize == nil {
		panic(NewCSSParseException("A font-size value is required", -1))
	}

	if lineHeight == nil {
		lineHeight = NewPropertyDeclaration(
			CSSNameLineHeight, NewPropertyValueIdentValue(IdentValueNormal), important, origin)
	}

	// XXX font-family should be reset too (although, does this really make sense?)

	result = make([]*PropertyDeclaration, 0, len(fontPropertyBuilderAll))
	result = append(result, fontStyle)
	result = append(result, fontVariant)
	result = append(result, fontWeight)
	result = append(result, fontSize)
	result = append(result, lineHeight)
	if fontFamily != nil {
		result = append(result, fontFamily)
	}

	return result
}
