// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/QuotesPropertyBuilder.java

package ufo

import "strings"

type QuotesPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewQuotesPropertyBuilder() *QuotesPropertyBuilder {
	b := &QuotesPropertyBuilder{}
	b.self = b
	return b
}

func (b *QuotesPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	if len(values) == 1 {
		value := values[0]
		if value.GetCssValueType() == CSSValueCssInherit {
			return []*PropertyDeclaration{}
		} else if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			ident := b.CheckIdent(value)
			if ident == IdentValueNone {
				return []*PropertyDeclaration{
					NewPropertyDeclaration(CSSNameQuotes, value, important, origin)}
			}
		}
	}

	if len(values)%2 == 1 {
		// The message ends with the values as Java's List.toString() prints
		// them.
		texts := make([]string, 0, len(values))
		for _, value := range values {
			texts = append(texts, value.ToString())
		}
		panic(NewCSSParseException(
			"Mismatched quotes ["+strings.Join(texts, ", ")+"]", -1))
	}

	resultValues := b.getStringValues(values)

	if len(resultValues) != 0 {
		list := make([]any, 0, len(resultValues))
		for _, resultValue := range resultValues {
			list = append(list, resultValue)
		}
		return []*PropertyDeclaration{
			NewPropertyDeclaration(CSSNameQuotes, NewPropertyValueList(list), important, origin)}
	} else {
		return []*PropertyDeclaration{}
	}
}

// getStringValues checks each value in turn (operator, then type) before
// moving to the next one, which is the order in which the Java stream
// evaluates its peek stages.
func (b *QuotesPropertyBuilder) getStringValues(values []*PropertyValue) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		b.assertNoOperator(value)
		b.assertValueIsString(value)
		result = append(result, value.GetStringValue())
	}
	return result
}

func (b *QuotesPropertyBuilder) assertNoOperator(cssPrimitiveValue *PropertyValue) {
	if cssPrimitiveValue.GetOperator() != nil {
		panic(NewCSSParseException(
			"Found unexpected operator, "+cssPrimitiveValue.GetOperator().GetExternalName(), -1))
	}
}

func (b *QuotesPropertyBuilder) assertValueIsString(value *PropertyValue) {
	typ := value.GetPrimitiveType()
	if typ == CSSPrimitiveValueCssUri {
		panic(NewCSSParseException("URI is not allowed here", -1))
	} else if value.GetPropertyValueType() == PropertyValueTypeValueTypeFunction {
		panic(NewCSSParseException("Function "+value.GetFunction().GetName()+" is not allowed here", -1))
	} else if typ == CSSPrimitiveValueCssIdent {
		panic(NewCSSParseException("Identifier is not a valid value for the quotes property", -1))
	} else if typ != CSSPrimitiveValueCssString {
		panic(NewCSSParseException(value.GetCssText()+" is not a value value for the quotes property", -1))
	}
}
