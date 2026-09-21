// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/TransformPropertyBuilder.java

package ufo

import "fmt"

type TransformPropertyBuilderArgumentKind int

const (
	TransformPropertyBuilderArgumentKindLengthPercentage TransformPropertyBuilderArgumentKind = iota
	TransformPropertyBuilderArgumentKindNumber
	TransformPropertyBuilderArgumentKindAngle
)

type transformPropertyBuilderFunctionSpec struct {
	minArgs      int
	maxArgs      int
	argumentKind TransformPropertyBuilderArgumentKind
}

// CSSParser lower-cases FUNCTION tokens (CSS identifiers are case-insensitive), so e.g. "skewX"
// always arrives here as "skewx".
var transformPropertyBuilderFunctions = map[string]*transformPropertyBuilderFunctionSpec{
	"matrix":     {6, 6, TransformPropertyBuilderArgumentKindNumber},
	"translate":  {1, 2, TransformPropertyBuilderArgumentKindLengthPercentage},
	"translatex": {1, 1, TransformPropertyBuilderArgumentKindLengthPercentage},
	"translatey": {1, 1, TransformPropertyBuilderArgumentKindLengthPercentage},
	"scale":      {1, 2, TransformPropertyBuilderArgumentKindNumber},
	"scalex":     {1, 1, TransformPropertyBuilderArgumentKindNumber},
	"scaley":     {1, 1, TransformPropertyBuilderArgumentKindNumber},
	"rotate":     {1, 1, TransformPropertyBuilderArgumentKindAngle},
	"skew":       {1, 2, TransformPropertyBuilderArgumentKindAngle},
	"skewx":      {1, 1, TransformPropertyBuilderArgumentKindAngle},
	"skewy":      {1, 1, TransformPropertyBuilderArgumentKindAngle},
}

// TransformPropertyBuilder parses transform: none | <transform-function>+, e.g.
// translate(10px, 10px) rotate(45deg) scale(2).
//
// Only 2D transform functions are supported, which is all that is meaningful on a flat PDF page:
// matrix, translate/translateX/translateY, scale/scaleX/scaleY, rotate, skew/skewX/skewY.
type TransformPropertyBuilder struct {
	AbstractPropertyBuilder
}

func NewTransformPropertyBuilder() *TransformPropertyBuilder {
	b := &TransformPropertyBuilder{}
	b.self = b
	return b
}

func (b *TransformPropertyBuilder) BuildDeclarationsWithInheritAllowed(
	cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin,
	important bool, inheritAllowed bool) []*PropertyDeclaration {
	if len(values) == 1 {
		value := values[0]
		b.CheckInheritAllowed(value, inheritAllowed)

		if value.GetCssValueType() == CSSValueCssInherit {
			return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
		}
		if value.GetPropertyValueType() == PropertyValueTypeValueTypeIdent && "none" == value.GetStringValue() {
			return []*PropertyDeclaration{NewPropertyDeclaration(cssName, value, important, origin)}
		}
	}

	functions := make([]*PropertyValue, 0, len(values))
	for _, value := range values {
		if value.GetPropertyValueType() != PropertyValueTypeValueTypeFunction {
			panic(NewCSSParseException(
				"Value for "+cssName.ToString()+" must be 'none' or one or more transform functions", -1))
		}
		b.checkFunction(cssName, value.GetFunction())
		functions = append(functions, value)
	}

	return []*PropertyDeclaration{NewPropertyDeclaration(cssName, NewPropertyValueList(propertyBuilderValueList(functions)), important, origin)}
}

func (b *TransformPropertyBuilder) checkFunction(cssName *CSSName, function *FSFunction) {
	spec := transformPropertyBuilderFunctions[function.GetName()]
	if spec == nil {
		panic(NewCSSParseException(
			"Function "+function.GetName()+"() is not a valid value for "+cssName.ToString(), -1))
	}

	params := function.GetParameters()
	if len(params) < spec.minArgs || len(params) > spec.maxArgs {
		panic(NewCSSParseException(
			fmt.Sprintf("%s() requires between %d and %d argument(s) but %d were given",
				function.GetName(), spec.minArgs, spec.maxArgs, len(params)), -1))
	}

	for _, param := range params {
		b.checkArgument(cssName, function.GetName(), spec.argumentKind, param)
	}
}

func (b *TransformPropertyBuilder) checkArgument(cssName *CSSName, functionName string, kind TransformPropertyBuilderArgumentKind, value *PropertyValue) {
	var valid bool
	switch kind {
	case TransformPropertyBuilderArgumentKindLengthPercentage:
		valid = b.IsLength(value) || value.GetPrimitiveType() == CSSPrimitiveValueCssPercentage
	case TransformPropertyBuilderArgumentKindNumber:
		valid = value.GetPrimitiveType() == CSSPrimitiveValueCssNumber
	case TransformPropertyBuilderArgumentKindAngle:
		valid = b.isAngle(value)
	}

	if !valid {
		panic(NewCSSParseException(
			"Invalid argument '"+value.GetCssText()+"' to "+functionName+"() in "+cssName.ToString(), -1))
	}
}

func (b *TransformPropertyBuilder) isAngle(value *PropertyValue) bool {
	typ := value.GetPrimitiveType()
	return typ == CSSPrimitiveValueCssDeg || typ == CSSPrimitiveValueCssRad || typ == CSSPrimitiveValueCssGrad
}
