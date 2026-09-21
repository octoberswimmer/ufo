// Tests for layout_style_tracker.go. Flying Saucer has no JUnit test for
// StyleTracker.

package ufo

import "testing"

func TestStyleTracker(t *testing.T) {
	tracker := NewStyleTracker()
	if tracker.HasStyles() {
		t.Errorf("new tracker has styles")
	}
	tracker.RemoveLast()

	upper := CascadedStyleCreateLayoutStyle(CascadedStyleCreateLayoutPropertyDeclaration(CSSNameTextTransform, IdentValueUppercase))
	smallCaps := CascadedStyleCreateLayoutStyle(CascadedStyleCreateLayoutPropertyDeclaration(CSSNameFontVariant, IdentValueSmallCaps))
	lower := CascadedStyleCreateLayoutStyle(CascadedStyleCreateLayoutPropertyDeclaration(CSSNameTextTransform, IdentValueLowercase))
	tracker.AddStyle(upper)
	tracker.AddStyle(smallCaps)
	tracker.AddStyle(lower)

	var start CalculatedStyleI = NewEmptyStyle()
	derived := tracker.DeriveAll(start)
	if !derived.IsIdent(CSSNameTextTransform, IdentValueLowercase) || !derived.IsIdent(CSSNameFontVariant, IdentValueSmallCaps) {
		t.Errorf("DeriveAll did not apply the styles in order")
	}

	copied := tracker.CopyOf()
	tracker.RemoveLast()
	if len(tracker.GetStyles()) != 2 || len(copied.GetStyles()) != 3 {
		t.Errorf("got %d styles and %d copied styles, want 2 and 3", len(tracker.GetStyles()), len(copied.GetStyles()))
	}
	derived = tracker.DeriveAll(start)
	if !derived.IsIdent(CSSNameTextTransform, IdentValueUppercase) {
		t.Errorf("RemoveLast did not remove the last style")
	}

	tracker.ClearStyles()
	if tracker.HasStyles() {
		t.Errorf("cleared tracker has styles")
	}
	if tracker.DeriveAll(start) != start {
		t.Errorf("DeriveAll without styles did not return the start style")
	}
	if !copied.HasStyles() {
		t.Errorf("clearing the tracker cleared its copy")
	}
}
