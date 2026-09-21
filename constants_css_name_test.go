// Tests of constants_css_name.go. The Java suite has no CSSNameTest; these
// check that the registry has the content and the FS_ID order of the Java
// class.

package ufo

import (
	"sort"
	"testing"
)

func TestCSSName_countAndIdOrder(t *testing.T) {
	if got := CSSNameCountCSSNames(); got != 135 {
		t.Fatalf("CSSNameCountCSSNames() = %d, want 135", got)
	}
	if got := CSSNameCountCSSPrimitiveNames(); got != 117 {
		t.Errorf("CSSNameCountCSSPrimitiveNames() = %d, want 117", got)
	}
	for id := 0; id < CSSNameCountCSSNames(); id++ {
		if got := CSSNameGetByID(id).FS_ID; got != id {
			t.Errorf("CSSNameGetByID(%d).FS_ID = %d", id, got)
		}
	}
	// First, second and last declarations of the Java class.
	if CSSNameColor.FS_ID != 0 || CSSNameBackgroundColor.FS_ID != 1 || CSSNameBoxSizing.FS_ID != 134 {
		t.Errorf("FS_ID of color, background-color, box-sizing = %d, %d, %d, want 0, 1, 134",
			CSSNameColor.FS_ID, CSSNameBackgroundColor.FS_ID, CSSNameBoxSizing.FS_ID)
	}
}

func TestCSSName_lookupByName(t *testing.T) {
	if got := CSSNameGetByPropertyName("border-top-width"); got != CSSNameBorderTopWidth {
		t.Errorf("CSSNameGetByPropertyName(border-top-width) = %v", got)
	}
	if got := CSSNameGetByPropertyName("no-such-property"); got != nil {
		t.Errorf("CSSNameGetByPropertyName(no-such-property) = %v, want nil", got)
	}
	if got := CSSNameCssProperty("-fs-page-width"); got != CSSNameFsPageWidth {
		t.Errorf("CSSNameCssProperty(-fs-page-width) = %v", got)
	}
	defer func() {
		e, ok := recover().(*XRRuntimeException)
		if !ok || e.GetMessage() != "Unknown CSS property: no-such-property" {
			t.Errorf("CSSNameCssProperty(no-such-property) panicked with %v", e)
		}
	}()
	CSSNameCssProperty("no-such-property")
}

func TestCSSName_attributes(t *testing.T) {
	if CSSNameColor.ToString() != "color" || CSSNameInitialValue(CSSNameColor) != "black" || !CSSNamePropertyInherits(CSSNameColor) {
		t.Errorf("color: name %q, initial %q, inherits %v", CSSNameColor.ToString(), CSSNameInitialValue(CSSNameColor), CSSNamePropertyInherits(CSSNameColor))
	}
	if CSSNamePropertyInherits(CSSNameOpacity) || !CSSNameIsImplemented(CSSNameOpacity) {
		t.Errorf("opacity: inherits %v, implemented %v", CSSNamePropertyInherits(CSSNameOpacity), CSSNameIsImplemented(CSSNameOpacity))
	}
	if CSSNameIsImplemented(CSSNameOutlineColor) || CSSNameGetPropertyBuilder(CSSNameOutlineColor) != nil {
		t.Errorf("outline-color: implemented %v, builder %v", CSSNameIsImplemented(CSSNameOutlineColor), CSSNameGetPropertyBuilder(CSSNameOutlineColor))
	}
	for id := 0; id < CSSNameCountCSSNames(); id++ {
		cssName := CSSNameGetByID(id)
		if CSSNameIsImplemented(cssName) != (CSSNameGetPropertyBuilder(cssName) != nil) {
			t.Errorf("%s: implemented %v, builder %v", cssName, CSSNameIsImplemented(cssName), CSSNameGetPropertyBuilder(cssName))
		}
	}
}

func TestCSSName_namesAreSorted(t *testing.T) {
	all := CSSNameAllCSS2PropertyNames()
	if len(all) != 135 || !sort.StringsAreSorted(all) {
		t.Errorf("CSSNameAllCSS2PropertyNames(): %d names, sorted %v", len(all), sort.StringsAreSorted(all))
	}
	primitive := CSSNameAllCSS2PrimitivePropertyNames()
	if len(primitive) != 117 || !sort.StringsAreSorted(primitive) {
		t.Errorf("CSSNameAllCSS2PrimitivePropertyNames(): %d names, sorted %v", len(primitive), sort.StringsAreSorted(primitive))
	}
}

func TestCSSName_compareAndEquals(t *testing.T) {
	if CSSNameColor.CompareTo(CSSNameBackgroundColor) >= 0 {
		t.Error("color does not compare below background-color")
	}
	if !CSSNameColor.Equals(CSSNameColor) || CSSNameColor.Equals(CSSNameBackgroundColor) || CSSNameColor.Equals("color") {
		t.Error("Equals does not compare by FS_ID")
	}
	if CSSNameMarginSideProperties.Top() != CSSNameMarginTop || CSSNameBorderColorProperties.Left() != CSSNameBorderLeftColor {
		t.Error("side properties do not hold the side CSSNames")
	}
}

func TestCSSName_initialDerivedValue(t *testing.T) {
	color := CSSNameColor.InitialDerivedValue()
	if color == nil || !FSRGBColorTransparent.Equals(color.AsColor()) {
		t.Errorf("initial derived value of color = %v, want the color black", color)
	}
	if display := CSSNameDisplay.InitialDerivedValue(); display == nil || display.AsIdentValue() != IdentValueInline {
		t.Errorf("initial derived value of display = %v, want inline", display)
	}
	// An initial value that starts with "=" refers to another property and is
	// not derived.
	if got := CSSNameBorderTopColor.InitialDerivedValue(); got != nil {
		t.Errorf("initial derived value of border-top-color = %v, want nil", got)
	}
	if got := CSSNameMarginShorthand.InitialDerivedValue(); got != nil {
		t.Errorf("initial derived value of margin = %v, want nil", got)
	}
	for _, propName := range CSSNameAllCSS2PrimitivePropertyNames() {
		cssName := CSSNameGetByPropertyName(propName)
		derivable := CSSNameInitialValue(cssName)[0] != '=' && CSSNameIsImplemented(cssName)
		if derivable != (cssName.InitialDerivedValue() != nil) {
			t.Errorf("%s: initial value %q, implemented %v, derived value %v",
				propName, CSSNameInitialValue(cssName), CSSNameIsImplemented(cssName), cssName.InitialDerivedValue())
		}
	}
}
