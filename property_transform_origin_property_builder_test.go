// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/property/TransformOriginPropertyBuilderTest.java

package ufo

import (
	"strings"
	"testing"
)

func transformOriginPropertyBuilderTestIdent(name string) *PropertyValue {
	return NewPropertyValueString(CSSPrimitiveValueCssIdent, name, name)
}

func transformOriginPropertyBuilderTestPx(value float32) *PropertyValue {
	return NewPropertyValueFloat(CSSPrimitiveValueCssPx, value, propertyBuilderFloatToString(value)+"px")
}

func transformOriginPropertyBuilderTestValuesOf(result []*PropertyDeclaration) []*PropertyValue {
	values := []*PropertyValue{}
	for _, value := range result[0].GetValue().GetValues() {
		values = append(values, value.(*PropertyValue))
	}
	return values
}

func transformOriginPropertyBuilderTestAssertFloat(t *testing.T, value *PropertyValue, expected float32) {
	t.Helper()
	if value.GetFloatValue() != expected {
		t.Errorf("got %v, expected %v", value.GetFloatValue(), expected)
	}
}

func TestTransformOriginPropertyBuilder_singleLengthDefaultsVerticalToCenter(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestPx(10)}, StylesheetInfoOriginAuthor, false, true)

	values := transformOriginPropertyBuilderTestValuesOf(result)
	transformOriginPropertyBuilderTestAssertFloat(t, values[0], 10)
	if values[1].GetPrimitiveType() != CSSPrimitiveValueCssPercentage {
		t.Errorf("got type %d, expected a percentage", values[1].GetPrimitiveType())
	}
	transformOriginPropertyBuilderTestAssertFloat(t, values[1], 50)
}

func TestTransformOriginPropertyBuilder_keywordsResolveToPercentages(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("right"), transformOriginPropertyBuilderTestIdent("bottom")}, StylesheetInfoOriginAuthor, false, true)

	values := transformOriginPropertyBuilderTestValuesOf(result)
	transformOriginPropertyBuilderTestAssertFloat(t, values[0], 100)
	transformOriginPropertyBuilderTestAssertFloat(t, values[1], 100)
}

func TestTransformOriginPropertyBuilder_leftAndTopKeywordsResolveToPercent0(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("left"), transformOriginPropertyBuilderTestIdent("top")}, StylesheetInfoOriginAuthor, false, true)

	values := transformOriginPropertyBuilderTestValuesOf(result)
	transformOriginPropertyBuilderTestAssertFloat(t, values[0], 0)
	transformOriginPropertyBuilderTestAssertFloat(t, values[1], 0)
}

func TestTransformOriginPropertyBuilder_centerKeywordResolvesToPercent50(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("center")}, StylesheetInfoOriginAuthor, false, true)

	values := transformOriginPropertyBuilderTestValuesOf(result)
	transformOriginPropertyBuilderTestAssertFloat(t, values[0], 50)
	transformOriginPropertyBuilderTestAssertFloat(t, values[1], 50)
}

func TestTransformOriginPropertyBuilder_lengthAndPercentageCombination(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestPx(5), NewPropertyValueFloat(CSSPrimitiveValueCssPercentage, 25, "25%")}, StylesheetInfoOriginAuthor, false, true)

	values := transformOriginPropertyBuilderTestValuesOf(result)
	transformOriginPropertyBuilderTestAssertFloat(t, values[0], 5)
	transformOriginPropertyBuilderTestAssertFloat(t, values[1], 25)
}

func TestTransformOriginPropertyBuilder_singleVerticalKeywordImpliesCenteredHorizontalAxis(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	result := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("top")}, StylesheetInfoOriginAuthor, false, true)

	values := transformOriginPropertyBuilderTestValuesOf(result)
	transformOriginPropertyBuilderTestAssertFloat(t, values[0], 50) // horizontal: center
	transformOriginPropertyBuilderTestAssertFloat(t, values[1], 0)  // vertical: top
}

func TestTransformOriginPropertyBuilder_keywordPairsCanAppearInEitherOrder(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	topLeft := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("top"), transformOriginPropertyBuilderTestIdent("left")}, StylesheetInfoOriginAuthor, false, true)
	topLeftValues := transformOriginPropertyBuilderTestValuesOf(topLeft)
	transformOriginPropertyBuilderTestAssertFloat(t, topLeftValues[0], 0) // horizontal: left
	transformOriginPropertyBuilderTestAssertFloat(t, topLeftValues[1], 0) // vertical: top

	// "right" can only be horizontal, so "center" is pushed onto the (otherwise unspecified) vertical axis.
	centerRight := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("center"), transformOriginPropertyBuilderTestIdent("right")}, StylesheetInfoOriginAuthor, false, true)
	centerRightValues := transformOriginPropertyBuilderTestValuesOf(centerRight)
	transformOriginPropertyBuilderTestAssertFloat(t, centerRightValues[0], 100) // horizontal: right
	transformOriginPropertyBuilderTestAssertFloat(t, centerRightValues[1], 50)  // vertical: center
}

func TestTransformOriginPropertyBuilder_lengthCanBePairedWithAKeywordOnTheOtherAxis(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	lengthThenTop := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestPx(10), transformOriginPropertyBuilderTestIdent("top")}, StylesheetInfoOriginAuthor, false, true)
	lengthThenTopValues := transformOriginPropertyBuilderTestValuesOf(lengthThenTop)
	transformOriginPropertyBuilderTestAssertFloat(t, lengthThenTopValues[0], 10)
	transformOriginPropertyBuilderTestAssertFloat(t, lengthThenTopValues[1], 0)

	leftThenLength := builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("left"), transformOriginPropertyBuilderTestPx(20)}, StylesheetInfoOriginAuthor, false, true)
	leftThenLengthValues := transformOriginPropertyBuilderTestValuesOf(leftThenLength)
	transformOriginPropertyBuilderTestAssertFloat(t, leftThenLengthValues[0], 0)
	transformOriginPropertyBuilderTestAssertFloat(t, leftThenLengthValues[1], 20)
}

func TestTransformOriginPropertyBuilder_rejectsTwoHorizontalOnlyKeywords(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("left"), transformOriginPropertyBuilderTestIdent("right")}, StylesheetInfoOriginAuthor, false, true)
	})
	if !strings.Contains(e.GetMessage(), "Invalid combination") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}

func TestTransformOriginPropertyBuilder_rejectsAVerticalKeywordPairedWithALength(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestIdent("top"), transformOriginPropertyBuilderTestPx(10)}, StylesheetInfoOriginAuthor, false, true)
	})
	if !strings.Contains(e.GetMessage(), "Invalid combination") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}

func TestTransformOriginPropertyBuilder_rejectsMoreThanTwoValues(t *testing.T) {
	builder := NewTransformOriginPropertyBuilder()
	cssName := CSSNameCssProperty("transform-origin")
	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarationsWithInheritAllowed(cssName, []*PropertyValue{transformOriginPropertyBuilderTestPx(1), transformOriginPropertyBuilderTestPx(2), transformOriginPropertyBuilderTestPx(3)}, StylesheetInfoOriginAuthor, false, true)
	})
	if !strings.HasPrefix(e.GetMessage(), "Found 3 values for transform-origin") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}
