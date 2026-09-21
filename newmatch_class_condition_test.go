// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/newmatch/ClassConditionTest.java

package ufo

import "testing"

func classConditionTestCondition() *ConditionClassCondition {
	return ConditionCreateClassCondition("active").(*ConditionClassCondition)
}

func assertContainsClassName(t *testing.T, classAttribute string, expected bool) {
	t.Helper()
	if actual := classConditionTestCondition().ContainsClassName(classAttribute); actual != expected {
		t.Errorf("ContainsClassName(%q) = %v, expected %v", classAttribute, actual, expected)
	}
}

func TestClassCondition_classAttributeEqualsExpectedClassName(t *testing.T) {
	assertContainsClassName(t, "active", true)
	assertContainsClassName(t, "activex", false)
}

func TestClassCondition_classAttributeContainsExpectedClassName(t *testing.T) {
	assertContainsClassName(t, "foo active bar", true)
	assertContainsClassName(t, "foo active ", true)
	assertContainsClassName(t, " active bar", true)
	assertContainsClassName(t, " active ", true)
}

func TestClassCondition_classAttributeStartsWithExpectedClassName(t *testing.T) {
	assertContainsClassName(t, "active foo bar", true)
	assertContainsClassName(t, " active foo bar", true)
	assertContainsClassName(t, "activex foo bar", false)
	assertContainsClassName(t, "inactive foo bar", false)
}

func TestClassCondition_classAttributeEndsWithExpectedClassName(t *testing.T) {
	assertContainsClassName(t, "foo bar active", true)
	assertContainsClassName(t, "foo bar active ", true)
	assertContainsClassName(t, "foo bar inactive", false)
	assertContainsClassName(t, "foo bar activeactive", false)
	assertContainsClassName(t, "foo bar active-active", false)
	assertContainsClassName(t, "foo bar active_active", false)
	assertContainsClassName(t, "foo bar activex", false)
}

func TestClassCondition_classAttributeContainsSimilarClassNames(t *testing.T) {
	assertContainsClassName(t, "activex active inactive", true)
	assertContainsClassName(t, "activex _active inactive", false)
}

func TestClassCondition_classNotMatches(t *testing.T) {
	assertContainsClassName(t, "", false)
	assertContainsClassName(t, "_active", false)
	assertContainsClassName(t, "active_", false)
	assertContainsClassName(t, "act ive", false)
}
