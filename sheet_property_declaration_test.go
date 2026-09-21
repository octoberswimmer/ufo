// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/PropertyDeclaration.java.
// Flying Saucer has no JUnit test for it.

package ufo

import (
	"strconv"
	"testing"
)

func TestPropertyDeclaration_importanceAndOrigin(t *testing.T) {
	cases := []struct {
		origin    StylesheetInfoOrigin
		important bool
		expected  int
	}{
		{StylesheetInfoOriginUserAgent, false, 1},
		{StylesheetInfoOriginUserAgent, true, 1},
		{StylesheetInfoOriginUser, false, 2},
		{StylesheetInfoOriginAuthor, false, 3},
		{StylesheetInfoOriginAuthor, true, 4},
		{StylesheetInfoOriginUser, true, 5},
	}
	for i, c := range cases {
		declaration := NewPropertyDeclaration(CSSNameDisplay, NewPropertyValueIdentValue(IdentValueBlock), c.important, c.origin)
		if actual := declaration.GetImportanceAndOrigin(); actual != c.expected {
			t.Errorf("case %d: GetImportanceAndOrigin() = %d, expected %d", i, actual, c.expected)
		}
		if actual := declaration.GetImportanceAndOrigin(); actual >= PropertyDeclarationImportanceAndOriginCount {
			t.Errorf("case %d: %d is not below the count", i, actual)
		}
	}
}

func TestPropertyDeclaration_text(t *testing.T) {
	value := NewPropertyValueIdentValue(IdentValueBlock)
	declaration := NewPropertyDeclaration(CSSNameDisplay, value, false, StylesheetInfoOriginAuthor)
	if declaration.GetPropertyName() != "display" || declaration.GetCSSName() != CSSNameDisplay || declaration.GetValue() != value {
		t.Errorf("the accessors give other values than the constructor got")
	}
	if actual := declaration.GetDeclarationStandardText(); actual != "display: block;" {
		t.Errorf("GetDeclarationStandardText() = %q", actual)
	}
	if actual := declaration.String(); actual != "display: "+value.ToString() {
		t.Errorf("String() = %q", actual)
	}
	if declaration.AsIdentValue() != IdentValueBlock {
		t.Errorf("AsIdentValue() = %v", declaration.AsIdentValue())
	}
}

func TestPropertyDeclaration_fingerprintStartsWithTheSumOfTheCharsAndTheID(t *testing.T) {
	value := NewPropertyValueIdentValue(IdentValueBlock)
	declaration := NewPropertyDeclaration(CSSNameDisplay, value, false, StylesheetInfoOriginAuthor)
	expected := strconv.Itoa(80+CSSNameDisplay.FS_ID+58) + value.GetFingerprint() + ";"
	if actual := declaration.GetFingerprint(); actual != expected {
		t.Errorf("GetFingerprint() = %q, expected %q", actual, expected)
	}
}
