// Tests for the ports of Stylesheet.java, Ruleset.java and FontFaceRule.java
// in flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/. Flying
// Saucer has no JUnit test for them.

package ufo

import "testing"

func TestStylesheet_contentsKeepTheOrderOfAddition(t *testing.T) {
	sheet := NewStylesheet("a.css", StylesheetInfoOriginUser)
	ruleset := NewRuleset(StylesheetInfoOriginUser)
	mediaRule := NewMediaRule(StylesheetInfoOriginUser)
	pageRule := NewPageRule(StylesheetInfoOriginUser, "", "", nil, NewRuleset(StylesheetInfoOriginUser))
	sheet.AddContentMediaRule(mediaRule)
	sheet.AddContent(ruleset)
	sheet.AddContentPageRule(pageRule)
	contents := sheet.GetContents()
	if len(contents) != 3 || contents[0] != any(mediaRule) || contents[1] != any(ruleset) || contents[2] != any(pageRule) {
		t.Errorf("GetContents() = %v", contents)
	}
	importRule := NewStylesheetInfo(StylesheetInfoOriginUser, "b.css", []string{"all"}, nil)
	sheet.AddImportRule(importRule)
	if rules := sheet.GetImportRules(); len(rules) != 1 || rules[0] != importRule {
		t.Errorf("GetImportRules() = %v", rules)
	}
	if sheet.GetURI() != "a.css" || sheet.GetOrigin() != StylesheetInfoOriginUser {
		t.Errorf("GetURI() or GetOrigin() gives another value than the constructor got")
	}
	if actual := sheet.String(); actual != "Stylesheet{uri:a.css, origin: USER}" {
		t.Errorf("String() = %q", actual)
	}
	var _ RulesetContainer = sheet
	var _ RulesetContainer = mediaRule
	var _ RulesetContainer = pageRule
	var _ RulesetContainer = NewFontFaceRule(StylesheetInfoOriginUser)
}

func TestRuleset_propertiesAndSelectors(t *testing.T) {
	ruleset := NewRuleset(StylesheetInfoOriginAuthor)
	display := CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueBlock)
	width := CascadedStyleCreateLayoutPropertyDeclaration(CSSNameWidth, IdentValueNone)
	color := CascadedStyleCreateLayoutPropertyDeclaration(CSSNameColor, IdentValueNone)
	ruleset.AddProperty(display)
	ruleset.AddAllProperties([]*PropertyDeclaration{width, color})
	declarations := ruleset.GetPropertyDeclarations()
	if len(declarations) != 3 || declarations[0] != display || declarations[1] != width || declarations[2] != color {
		t.Errorf("GetPropertyDeclarations() = %v", declarations)
	}
	selector := NewSelector(ruleset)
	selector.SetName("p")
	ruleset.AddFSSelector(selector)
	if selectors := ruleset.GetFSSelectors(); len(selectors) != 1 || selectors[0] != selector || selector.GetRuleset() != ruleset {
		t.Errorf("GetFSSelectors() = %v", selectors)
	}
}

func TestFontFaceRule_has(t *testing.T) {
	rule := NewFontFaceRule(StylesheetInfoOriginAuthor)
	ruleset := NewRuleset(StylesheetInfoOriginAuthor)
	ruleset.AddProperty(CascadedStyleCreateLayoutPropertyDeclaration(CSSNameFontFamily, IdentValueNone))
	ruleset.AddProperty(CascadedStyleCreateLayoutPropertyDeclaration(CSSNameFontStyle, IdentValueNone))
	rule.AddContent(ruleset)
	if !rule.HasFontFamily() || !rule.HasFontStyle() || rule.HasFontWeight() {
		t.Errorf("HasFontFamily, HasFontStyle, HasFontWeight = %v, %v, %v", rule.HasFontFamily(), rule.HasFontStyle(), rule.HasFontWeight())
	}
	defer func() {
		if _, ok := recover().(*XRRuntimeException); !ok {
			t.Errorf("expected an *XRRuntimeException panic for the second rule set")
		}
	}()
	rule.AddContent(ruleset)
}
