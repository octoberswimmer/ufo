// Tests of constants_ident_value.go, constants_margin_box_name.go and
// constants_page_element_position.go. The Java suite has no tests of these
// classes.

package ufo

import "testing"

func TestIdentValue_idOrderAndLookup(t *testing.T) {
	if got := IdentValueGetIdentCount(); got != 161 {
		t.Errorf("IdentValueGetIdentCount() = %d, want 161", got)
	}
	// First and last declarations of the Java class.
	if IdentValueAbsolute.FS_ID != 0 || IdentValueManual.FS_ID != 160 {
		t.Errorf("FS_ID of absolute, manual = %d, %d, want 0, 160", IdentValueAbsolute.FS_ID, IdentValueManual.FS_ID)
	}
	if IdentValueGetByIdentString("table-row-group") != IdentValueTableRowGroup {
		t.Error("IdentValueGetByIdentString(table-row-group) is not IdentValueTableRowGroup")
	}
	if IdentValueValueOf("100") != IdentValueFontWeight100 || IdentValueValueOf("no-such-ident") != nil {
		t.Error("IdentValueValueOf does not look up by ident string")
	}
	if !IdentValueLooksLikeIdent("auto") || IdentValueLooksLikeIdent("Auto") {
		t.Error("IdentValueLooksLikeIdent is not an exact lookup")
	}
	defer func() {
		e, ok := recover().(*XRRuntimeException)
		if !ok || e.GetMessage() != "Ident named no-such-ident has no IdentValue instance assigned to it." {
			t.Errorf("IdentValueGetByIdentString(no-such-ident) panicked with %v", e)
		}
	}()
	IdentValueGetByIdentString("no-such-ident")
}

func TestIdentValue_asDerivedValue(t *testing.T) {
	var value FSDerivedValue = IdentValueInherit
	if !value.IsDeclaredInherit() || !value.IsIdent() || value.AsString() != "inherit" || value.AsIdentValue() != IdentValueInherit {
		t.Error("IdentValueInherit does not behave as the inherit ident")
	}
	if IdentValueAuto.IsDeclaredInherit() || IdentValueAuto.IsDependentOnFontSize() {
		t.Error("IdentValueAuto is declared inherit or dependent on font size")
	}
}

func TestMarginBoxName_valueOf(t *testing.T) {
	if MarginBoxNameTopLeftCorner.FS_ID != 0 || MarginBoxNameFsPdfXmpMetadata.FS_ID != 16 {
		t.Errorf("FS_ID of top-left-corner, -fs-pdf-xmp-metadata = %d, %d, want 0, 16",
			MarginBoxNameTopLeftCorner.FS_ID, MarginBoxNameFsPdfXmpMetadata.FS_ID)
	}
	box := MarginBoxNameValueOf("left-bottom")
	if box != MarginBoxNameLeftBottom || box.GetInitialTextAlign() != IdentValueCenter || box.GetInitialVerticalAlign() != IdentValueBottom {
		t.Errorf("MarginBoxNameValueOf(left-bottom) = %v", box)
	}
	if MarginBoxNameValueOf("no-such-box") != nil {
		t.Error("MarginBoxNameValueOf(no-such-box) is not nil")
	}
}

func TestPageElementPosition_byIdent(t *testing.T) {
	if PageElementPositionByIdent("last-except") != PageElementPositionLastExcept || PageElementPositionByIdent("middle") != nil {
		t.Error("PageElementPositionByIdent does not look up by ident string")
	}
	if PageElementPositionFirst.ToString() != "first" {
		t.Errorf("PageElementPositionFirst.ToString() = %q", PageElementPositionFirst.ToString())
	}
}
