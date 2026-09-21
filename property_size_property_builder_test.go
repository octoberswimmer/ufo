// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/parser/property/SizePropertyBuilderTest.java

package ufo

import (
	"reflect"
	"strings"
	"testing"
)

var (
	sizePropertyBuilderTestWidth     = NewPropertyValueFloat(CSSPrimitiveValueCssMm, 0, "43mm")
	sizePropertyBuilderTestHeight    = NewPropertyValueFloat(CSSPrimitiveValueCssMm, 0, "25mm")
	sizePropertyBuilderTestLandscape = NewPropertyValueFloat(CSSPrimitiveValueCssIdent, 0, "landscape")
)

// sizePropertyBuilderTestAssertDeclarations compares field by field, as the
// Java usingRecursiveFieldByFieldElementComparator does.
func sizePropertyBuilderTestAssertDeclarations(t *testing.T, result []*PropertyDeclaration, expected []*PropertyDeclaration) {
	t.Helper()
	if len(result) != len(expected) {
		t.Fatalf("got %d declarations, expected %d", len(result), len(expected))
	}
	for i := range expected {
		if !reflect.DeepEqual(result[i], expected[i]) {
			t.Errorf("declaration %d: got %v, expected %v", i, result[i].ToString(), expected[i].ToString())
		}
	}
}

func TestSizePropertyBuilder_buildDeclarationsFromOneValue(t *testing.T) {
	builder := NewSizePropertyBuilder()
	cssName := CSSNameCssProperty("size")
	width := sizePropertyBuilderTestWidth

	result := builder.BuildDeclarations(cssName, []*PropertyValue{width}, StylesheetInfoOriginUser, false)

	sizePropertyBuilderTestAssertDeclarations(t, result, []*PropertyDeclaration{
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-orientation"), NewPropertyValueIdentValue(IdentValueAuto), false, StylesheetInfoOriginUser),
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-width"), width, false, StylesheetInfoOriginUser),
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-height"), width, false, StylesheetInfoOriginUser),
	})
}

func TestSizePropertyBuilder_buildDeclarationsFromTwoValues(t *testing.T) {
	builder := NewSizePropertyBuilder()
	cssName := CSSNameCssProperty("size")
	width := sizePropertyBuilderTestWidth
	height := sizePropertyBuilderTestHeight

	result := builder.BuildDeclarations(cssName, []*PropertyValue{width, height}, StylesheetInfoOriginUserAgent, false)

	sizePropertyBuilderTestAssertDeclarations(t, result, []*PropertyDeclaration{
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-orientation"), NewPropertyValueIdentValue(IdentValueAuto), false, StylesheetInfoOriginUserAgent),
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-width"), width, false, StylesheetInfoOriginUserAgent),
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-height"), height, false, StylesheetInfoOriginUserAgent),
	})
}

func TestSizePropertyBuilder_buildDeclarationsFromThreeValues(t *testing.T) {
	builder := NewSizePropertyBuilder()
	cssName := CSSNameCssProperty("size")
	width := sizePropertyBuilderTestWidth
	height := sizePropertyBuilderTestHeight
	landscape := sizePropertyBuilderTestLandscape

	result := builder.BuildDeclarations(cssName, []*PropertyValue{width, height, landscape}, StylesheetInfoOriginUserAgent, false)

	sizePropertyBuilderTestAssertDeclarations(t, result, []*PropertyDeclaration{
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-orientation"), landscape, false, StylesheetInfoOriginUserAgent),
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-width"), width, false, StylesheetInfoOriginUserAgent),
		NewPropertyDeclaration(CSSNameCssProperty("-fs-page-height"), height, false, StylesheetInfoOriginUserAgent),
	})
}

func TestSizePropertyBuilder_declarationMustHaveAtLeastOneValue(t *testing.T) {
	builder := NewSizePropertyBuilder()
	cssName := CSSNameCssProperty("size")

	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarations(cssName, []*PropertyValue{}, StylesheetInfoOriginAuthor, false)
	})
	if !strings.HasPrefix(e.GetMessage(), "Found 0 values for size") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}

func TestSizePropertyBuilder_declarationMustHaveAtMostThreeValue(t *testing.T) {
	builder := NewSizePropertyBuilder()
	cssName := CSSNameCssProperty("size")
	width := sizePropertyBuilderTestWidth
	height := sizePropertyBuilderTestHeight
	landscape := sizePropertyBuilderTestLandscape

	e := propertyBuilderTestCatch(t, func() {
		builder.BuildDeclarations(cssName, []*PropertyValue{width, height, landscape, landscape}, StylesheetInfoOriginAuthor, false)
	})
	if !strings.HasPrefix(e.GetMessage(), "Found 4 values for size") {
		t.Errorf("unexpected message: %s", e.GetMessage())
	}
}
