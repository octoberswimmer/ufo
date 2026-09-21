// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/property/TransformPropertyBuilderTest.java

package ufo

import (
	"reflect"
	"strings"
	"testing"
)

func transformPropertyBuilderTestFunction(name string, args ...*PropertyValue) *PropertyValue {
	return NewPropertyValueFSFunction(NewFSFunction(name, args))
}

func transformPropertyBuilderTestPx(value float32) *PropertyValue {
	return NewPropertyValueFloat(CSSPrimitiveValueCssPx, value, propertyBuilderFloatToString(value)+"px")
}

func transformPropertyBuilderTestDeg(value float32) *PropertyValue {
	return NewPropertyValueFloat(CSSPrimitiveValueCssDeg, value, propertyBuilderFloatToString(value)+"deg")
}

func transformPropertyBuilderTestNumber(value float32) *PropertyValue {
	return NewPropertyValueFloat(CSSPrimitiveValueCssNumber, value, propertyBuilderFloatToString(value))
}

func TestTransformPropertyBuilder_none(t *testing.T) {
	builder := NewTransformPropertyBuilder()
	cssName := CSSNameCssProperty("transform")
	none := NewPropertyValueIdentValue(IdentValueNone)

	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{none}, StylesheetInfoOriginAuthor, false, true)

	if len(result) != 1 {
		t.Fatalf("got %d declarations, expected 1", len(result))
	}
	if result[0].GetValue() != none {
		t.Errorf("the declaration value is not the value that was passed in")
	}
}

func TestTransformPropertyBuilder_singleFunction(t *testing.T) {
	builder := NewTransformPropertyBuilder()
	cssName := CSSNameCssProperty("transform")
	rotate := transformPropertyBuilderTestFunction("rotate", transformPropertyBuilderTestDeg(45))

	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{rotate}, StylesheetInfoOriginAuthor, false, true)

	if len(result) != 1 {
		t.Fatalf("got %d declarations, expected 1", len(result))
	}
	value := result[0].GetValue()
	if value.GetPropertyValueType() != PropertyValueTypeValueTypeList {
		t.Errorf("got value type %v, expected a list", value.GetPropertyValueType())
	}

	functions := value.GetValues()
	if len(functions) != 1 {
		t.Fatalf("got %d functions, expected 1", len(functions))
	}
	function := functions[0].(*PropertyValue)
	if function.GetPropertyValueType() != PropertyValueTypeValueTypeFunction {
		t.Errorf("got value type %v, expected a function", function.GetPropertyValueType())
	}
	if function.GetFunction().GetName() != "rotate" {
		t.Errorf("got function %s, expected rotate", function.GetFunction().GetName())
	}
}

func TestTransformPropertyBuilder_multipleFunctionsKeepOrder(t *testing.T) {
	builder := NewTransformPropertyBuilder()
	cssName := CSSNameCssProperty("transform")
	translate := transformPropertyBuilderTestFunction("translate", transformPropertyBuilderTestPx(10), transformPropertyBuilderTestPx(20))
	rotate := transformPropertyBuilderTestFunction("rotate", transformPropertyBuilderTestDeg(45))

	result := builder.BuildDeclarationsWithInheritAllowed(
		cssName, []*PropertyValue{translate, rotate}, StylesheetInfoOriginAuthor, false, true)

	names := []string{}
	for _, function := range result[0].GetValue().GetValues() {
		names = append(names, function.(*PropertyValue).GetFunction().GetName())
	}
	if !reflect.DeepEqual(names, []string{"translate", "rotate"}) {
		t.Errorf("got functions %v, expected [translate rotate]", names)
	}
}

func TestTransformPropertyBuilder_rejectsUnknownFunction(t *testing.T) {
	builder := NewTransformPropertyBuilder()
	cssName := CSSNameCssProperty("transform")
	unknown := transformPropertyBuilderTestFunction("rotate3d", transformPropertyBuilderTestDeg(45))

	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{unknown}, StylesheetInfoOriginAuthor, false, true)
	})
	if !strings.Contains(e.GetMessage(), "rotate3d") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}

func TestTransformPropertyBuilder_rejectsWrongArgumentCountForMatrix(t *testing.T) {
	builder := NewTransformPropertyBuilder()
	cssName := CSSNameCssProperty("transform")
	number := transformPropertyBuilderTestNumber
	matrix := transformPropertyBuilderTestFunction("matrix", number(1), number(0), number(0), number(1), number(0))

	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{matrix}, StylesheetInfoOriginAuthor, false, true)
	})
	if !strings.HasPrefix(e.GetMessage(), "matrix() requires between 6 and 6 argument(s)") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}

func TestTransformPropertyBuilder_rejectsNonAngleArgumentForRotate(t *testing.T) {
	builder := NewTransformPropertyBuilder()
	cssName := CSSNameCssProperty("transform")
	rotate := transformPropertyBuilderTestFunction("rotate", transformPropertyBuilderTestPx(45))

	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{rotate}, StylesheetInfoOriginAuthor, false, true)
	})
	if !strings.Contains(e.GetMessage(), "rotate()") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}

func TestTransformPropertyBuilder_scaleAcceptsOneOrTwoNumbers(t *testing.T) {
	builder := NewTransformPropertyBuilder()
	cssName := CSSNameCssProperty("transform")
	scale := transformPropertyBuilderTestFunction("scale", transformPropertyBuilderTestNumber(2))

	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{scale}, StylesheetInfoOriginAuthor, false, true)

	if len(result) != 1 {
		t.Fatalf("got %d declarations, expected 1", len(result))
	}
}
