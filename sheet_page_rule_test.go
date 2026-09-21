// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/PageRule.java.
// Flying Saucer has no JUnit test for it.

package ufo

import "testing"

func TestPageRule_applies(t *testing.T) {
	cases := []struct {
		name, pseudoPage         string
		pageName, pagePseudoPage string
		expected                 bool
	}{
		{"", "", "", "left", true},
		{"", "", "cover", "first", true},
		{"", "left", "", "left", true},
		{"", "left", "", "right", false},
		{"", "first", "", "first", true},
		{"", "first", "", "right", false},
		// The first page is assumed to be a right page.
		{"", "right", "", "first", true},
		{"", "left", "", "first", false},
		{"cover", "", "cover", "left", true},
		{"cover", "", "other", "left", false},
		{"cover", "", "", "left", false},
		{"cover", "first", "cover", "first", true},
		{"cover", "first", "cover", "right", false},
		{"cover", "right", "cover", "first", false},
	}
	for i, c := range cases {
		rule := NewPageRule(StylesheetInfoOriginAuthor, c.name, c.pseudoPage, nil, NewRuleset(StylesheetInfoOriginAuthor))
		if actual := rule.Applies(c.pageName, c.pagePseudoPage); actual != c.expected {
			t.Errorf("case %d: @page %q:%q applies to %q:%q = %v", i, c.name, c.pseudoPage, c.pageName, c.pagePseudoPage, actual)
		}
	}
}

func TestPageRule_order(t *testing.T) {
	cases := []struct {
		name, pseudoPage string
		pos              int
		expected         int64
	}{
		{"", "", 7, 1<<16 | 7},
		{"", "left", 7, 1<<16 | 7},
		{"", "first", 7, 1<<24 | 7},
		{"cover", "", 7, 1<<32 | 1<<16 | 7},
		{"cover", "first", 7, 1<<32 | 1<<24 | 7},
	}
	for i, c := range cases {
		rule := NewPageRule(StylesheetInfoOriginAuthor, c.name, c.pseudoPage, nil, NewRuleset(StylesheetInfoOriginAuthor))
		rule.SetPos(c.pos)
		if actual := rule.GetOrder(); actual != c.expected {
			t.Errorf("case %d: GetOrder() = %#x, expected %#x", i, actual, c.expected)
		}
	}
}

func TestPageRule_marginBoxesAreCopied(t *testing.T) {
	declarations := []*PropertyDeclaration{CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueBlock)}
	marginBoxes := map[*MarginBoxName][]*PropertyDeclaration{MarginBoxNameTopCenter: declarations}
	rule := NewPageRule(StylesheetInfoOriginAuthor, "", "", marginBoxes, NewRuleset(StylesheetInfoOriginAuthor))
	delete(marginBoxes, MarginBoxNameTopCenter)
	if properties := rule.GetMarginBoxProperties(MarginBoxNameTopCenter); len(properties) != 1 || properties[0] != declarations[0] {
		t.Errorf("GetMarginBoxProperties() = %v", properties)
	}
	if rule.GetMarginBoxProperties(MarginBoxNameTopLeft) != nil {
		t.Errorf("GetMarginBoxProperties() gives declarations for a box the rule does not have")
	}
}

func TestPageRule_addContentPanics(t *testing.T) {
	rule := NewPageRule(StylesheetInfoOriginAuthor, "", "", nil, NewRuleset(StylesheetInfoOriginAuthor))
	defer func() {
		if _, ok := recover().(*XRRuntimeException); !ok {
			t.Errorf("expected an *XRRuntimeException panic")
		}
	}()
	rule.AddContent(NewRuleset(StylesheetInfoOriginAuthor))
}
