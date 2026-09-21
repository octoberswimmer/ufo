// Tests for the port of
// flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/MediaRule.java.
// Flying Saucer has no JUnit test for it.

package ufo

import "testing"

func TestMediaRule_matches(t *testing.T) {
	printRule := NewMediaRule(StylesheetInfoOriginAuthor)
	printRule.AddMedium("print")
	allRule := NewMediaRule(StylesheetInfoOriginAuthor)
	allRule.AddMedium("all")
	cases := []struct {
		rule     *MediaRule
		medium   string
		expected bool
	}{
		{printRule, "print", true},
		{printRule, "PRINT", true},
		{printRule, "screen", false},
		{printRule, "ALL", true},
		{allRule, "screen", true},
		{NewMediaRule(StylesheetInfoOriginAuthor), "print", false},
	}
	for i, c := range cases {
		if actual := c.rule.Matches(c.medium); actual != c.expected {
			t.Errorf("case %d: Matches(%q) = %v", i, c.medium, actual)
		}
	}
}
