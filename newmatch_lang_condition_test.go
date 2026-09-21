// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/newmatch/LangConditionTest.java

package ufo

import "testing"

func assertLangMatches(t *testing.T, langAttribute string, expected bool) {
	t.Helper()
	condition := ConditionCreateLangCondition("et").(*ConditionLangCondition)
	if actual := condition.MatchesString(langAttribute); actual != expected {
		t.Errorf("MatchesString(%q) = %v, expected %v", langAttribute, actual, expected)
	}
}

func TestLangCondition_langAttributeEqualsExpectedLanguage(t *testing.T) {
	assertLangMatches(t, "et", true)
}

func TestLangCondition_countryPartIsIgnored(t *testing.T) {
	assertLangMatches(t, "et-EE", true)
	assertLangMatches(t, "et-FI", true)
}

func TestLangCondition_langIsCaseInsensitive(t *testing.T) {
	assertLangMatches(t, "ET-EE", true)
	assertLangMatches(t, "ET-EE", true)
}

func TestLangCondition_langAttributeNotEqualToExpectedLanguage(t *testing.T) {
	assertLangMatches(t, "en", false)
	assertLangMatches(t, "en-ET", false)
}
