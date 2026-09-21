// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/ContentPropertyBuilder.java

package ufo

import "strings"

type ContentPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewContentPropertyBuilder() *ContentPropertyBuilder {
	b := &ContentPropertyBuilder{}
	b.self = b
	return b
}

func (b *ContentPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	if len(values) == 1 {
		value := values[0]
		if value.GetCssValueType() == CSSValueCssInherit {
			return []*PropertyDeclaration{}
		} else if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			if ident == IdentValueNone || ident == IdentValueNormal {
				return []*PropertyDeclaration{
					NewPropertyDeclaration(CSSNameContent, value, important, origin)}
			}
		}
	}

	resultValues := []*PropertyValue{}
	for _, value := range values {
		if value.GetOperator() != nil {
			panic(NewCSSParseException(
				"Found unexpected operator, "+value.GetOperator().GetExternalName(), -1))
		}

		typ := value.GetPrimitiveType()
		if typ == CSSPrimitiveValueCssUri {
			continue
		} else if typ == CSSPrimitiveValueCssString {
			resultValues = append(resultValues, value)
		} else if value.GetPropertyValueType() == PropertyValueTypeValueTypeFunction {
			if !b.isFunctionAllowed(value.GetFunction()) {
				panic(NewCSSParseException(
					"Function "+value.GetFunction().GetName()+" is not allowed here", -1))
			}
			resultValues = append(resultValues, value)
		} else if typ == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			if ident == IdentValueOpenQuote || ident == IdentValueCloseQuote ||
				ident == IdentValueNoCloseQuote || ident == IdentValueNoOpenQuote {
				resultValues = append(resultValues, value)
			} else {
				panic(NewCSSParseException(
					"Identifier "+ident.ToString()+" is not a valid value for the content property", -1))
			}
		} else {
			panic(NewCSSParseException(
				value.GetCssText()+" is not a value value for the content property", -1))
		}
	}

	if len(resultValues) != 0 {
		return []*PropertyDeclaration{
			NewPropertyDeclaration(CSSNameContent, NewPropertyValueList(propertyBuilderValueList(resultValues)), important, origin)}
	} else {
		return []*PropertyDeclaration{}
	}
}

func (b *ContentPropertyBuilder) isFunctionAllowed(function *FSFunction) bool {
	return function.Is("attr") || function.Is("counter") || function.Is("counters") ||
		function.Is("element") || strings.HasPrefix(function.GetName(), "-fs") || function.Is("target-counter") ||
		function.Is("leader")
}
