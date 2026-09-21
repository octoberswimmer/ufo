// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/CounterPropertyBuilder.java

package ufo

// CounterPropertyBuilder is the Java abstract class CounterPropertyBuilder.
// The value returned by the abstract getDefaultValue() is the defaultValue
// field, set by the constructor of each subclass.
//
// [ <identifier> <integer>? ]+ | none | inherit
type CounterPropertyBuilder struct {
	AbstractPropertyBuilder
	defaultValue int
}

func newCounterPropertyBuilder(defaultValue int) *CounterPropertyBuilder {
	b := &CounterPropertyBuilder{defaultValue: defaultValue}
	b.self = b
	return b
}

func (b *CounterPropertyBuilder) GetDefaultValue() int {
	return b.defaultValue
}

// BuildDeclarationsWithInheritAllowed returns a PropertyValue of type
// VALUE_TYPE_LIST, but the list contains CounterData objects and not
// PropertyValue objects (XXX in the Java source).
func (b *CounterPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration {
	if len(values) == 1 {
		value := values[0]

		b.CheckInheritAllowed(value, inheritAllowed)

		if value.GetCssValueType() == CSSValueCssInherit {
			return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
		} else if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
			if value.GetCssText() == "none" {
				return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
			} else {
				data := NewCounterData(
					value.GetStringValue(),
					b.GetDefaultValue())

				return []*PropertyDeclaration{
					NewPropertyDeclaration(cssName, NewPropertyValueList(
						[]any{data}), important, origin)}
			}
		}

		panic(NewCSSParseException("The syntax of the "+cssName.ToString()+" property is invalid", -1))
	} else {
		result := []any{}
		for i := 0; i < len(values); i++ {
			value := values[i]

			if value.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
				name := value.GetStringValue()
				cValue := b.GetDefaultValue()

				if i < len(values)-1 {
					next := values[i+1]
					if next.GetPrimitiveType() == CSSPrimitiveValueCssNumber {
						b.checkNumberIsInteger(cssName, next)

						cValue = int(next.GetFloatValue())
					}

					i++
				}
				result = append(result, NewCounterData(name, cValue))
			} else {
				panic(NewCSSParseException("The syntax of the "+cssName.ToString()+" property is invalid", -1))
			}
		}

		return []*PropertyDeclaration{
			NewPropertyDeclaration(cssName, NewPropertyValueList(result), important, origin)}
	}
}

func (b *CounterPropertyBuilder) checkNumberIsInteger(cssName *CSSName, value *PropertyValue) {
	if int(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) !=
		propertyBuilderRound(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber)) {
		panic(NewCSSParseException("The value "+propertyBuilderFloatToString(value.GetFloatValueWithUnitType(CSSPrimitiveValueCssNumber))+" in "+
			cssName.ToString()+" must be an integer", -1))
	}
}

type CounterPropertyBuilderCounterReset struct {
	CounterPropertyBuilder
}

func NewCounterPropertyBuilderCounterReset() *CounterPropertyBuilderCounterReset {
	b := &CounterPropertyBuilderCounterReset{*newCounterPropertyBuilder(0)}
	b.self = b
	return b
}

type CounterPropertyBuilderCounterIncrement struct {
	CounterPropertyBuilder
}

func NewCounterPropertyBuilderCounterIncrement() *CounterPropertyBuilderCounterIncrement {
	b := &CounterPropertyBuilderCounterIncrement{*newCounterPropertyBuilder(1)}
	b.self = b
	return b
}
