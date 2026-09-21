// Ported from flying-saucer-core/src/test/java/org/xhtmlrenderer/css/newmatch/ConditionTest.java

package ufo

import (
	"reflect"
	"testing"
)

func TestCondition_nthChildParsing(t *testing.T) {
	cases := []struct {
		input    string
		expected *ConditionNthChildCondition
	}{
		{"33", NewConditionNthChildCondition(0, 33)},
		{"even", NewConditionNthChildCondition(2, 0)},
		{"odd", NewConditionNthChildCondition(2, 1)},
		{"22n+33", NewConditionNthChildCondition(22, 33)},
		{"-5n", NewConditionNthChildCondition(-5, 0)},
		{"-5n+7", NewConditionNthChildCondition(-5, 7)},
		{"-5n-2", NewConditionNthChildCondition(-5, -2)},
		{"+5n-1", NewConditionNthChildCondition(5, -1)},
	}
	for _, c := range cases {
		actual := ConditionCreateNthChildCondition(c.input)
		if !reflect.DeepEqual(actual, Condition(c.expected)) {
			t.Errorf("ConditionCreateNthChildCondition(%q) = %+v, expected %+v", c.input, actual, c.expected)
		}
	}
}

func TestCondition_nthChildMatching(t *testing.T) {
	cases := []struct {
		input                   string
		expectedMatchingIndices []int
		description             string
	}{
		{"1", []int{1}, "the 1st element"},
		{"2", []int{2}, "the 2nd element"},
		{"10", []int{10}, "the 10th element"},
		{"11", []int{11}, "the 11th element"},
		{"odd", []int{1, 3, 5, 7, 9, 11}, "odd elements (2*n+1)"},
		{"even", []int{2, 4, 6, 8, 10}, "even elements (2*n)"},
		{"5n", []int{5, 10}, "power of 5"},
		{"n+4", []int{4, 5, 6, 7, 8, 9, 10, 11}, "the fourth and all following elements"},
		{"4n+1", []int{1, 5, 9}, "4*0+1, 4*1+1, 4*2+1"},
		{"n+3", []int{3, 4, 5, 6, 7, 8, 9, 10, 11}, "TODO"},
		{"-n+3", []int{1, 2, 3}, "the first three list items"},
		{"-n+4", []int{1, 2, 3, 4}, "the first four list items"},
		{"-3n+2", []int{2}, "only 2-3*0"},
		{"-3n+7", []int{1, 4, 7}, "7-3*2, 7-3*1, 7-3*0"},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			condition := ConditionCreateNthChildCondition(c.input).(*ConditionNthChildCondition)
			var actual []int
			for index := 1; index < 12; index++ {
				if condition.MatchesInt(index) {
					actual = append(actual, index)
				}
			}
			if !reflect.DeepEqual(actual, c.expectedMatchingIndices) {
				t.Errorf("%s: matched %v, expected %v", c.description, actual, c.expectedMatchingIndices)
			}
		})
	}
}

// The remaining tests have no JUnit counterpart.

func TestCondition_nthChildParsingAcceptsSpacesAndUpperCase(t *testing.T) {
	actual := ConditionCreateNthChildCondition("  2N + 1 ")
	if !reflect.DeepEqual(actual, Condition(NewConditionNthChildCondition(2, 1))) {
		t.Errorf("got %+v", actual)
	}
}

func TestCondition_nthChildParsingRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{"", "n+", "2n+1x", "foo", "2 n"} {
		t.Run(input, func(t *testing.T) {
			defer func() {
				recovered := recover()
				exception, ok := recovered.(*CSSParseException)
				if !ok {
					t.Fatalf("expected a *CSSParseException panic, got %v", recovered)
				}
				if exception.GetLine() != -1 {
					t.Errorf("line = %d, expected -1", exception.GetLine())
				}
			}()
			ConditionCreateNthChildCondition(input)
		})
	}
}

func TestCondition_split(t *testing.T) {
	cases := []struct {
		input    string
		ch       byte
		expected []string
	}{
		{"abc", ' ', []string{"abc"}},
		{"", ' ', []string{""}},
		{"a b", ' ', []string{"a", "b"}},
		{"  a   b  ", ' ', []string{"a", "b"}},
		{"en-US-x", '-', []string{"en", "US", "x"}},
		{"-en", '-', []string{"en"}},
		{"--", '-', nil},
	}
	for _, c := range cases {
		if actual := conditionSplit(c.input, c.ch); !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("conditionSplit(%q, %q) = %#v, expected %#v", c.input, c.ch, actual, c.expected)
		}
	}
}

func TestCondition_equalsIgnoreCase(t *testing.T) {
	cases := []struct {
		a, b     string
		expected bool
	}{
		{"et", "ET", true},
		{"et", "Et", true},
		{"et", "en", false},
		{"et", "et-EE", false},
		// Character.toUpperCase maps both 'ı' and 'i' to 'I'.
		{"ı", "i", true},
		{"", "", true},
	}
	for _, c := range cases {
		if actual := conditionEqualsIgnoreCase(c.a, c.b); actual != c.expected {
			t.Errorf("conditionEqualsIgnoreCase(%q, %q) = %v, expected %v", c.a, c.b, actual, c.expected)
		}
	}
}

func TestCondition_isWhitespace(t *testing.T) {
	for _, r := range []rune{' ', '\t', '\n', '\u000B', '\f', '\r', '\u001C', '\u001F', ' ', ' ', '　'} {
		if !conditionIsWhitespace(r) {
			t.Errorf("conditionIsWhitespace(%U) = false, expected true", r)
		}
	}
	for _, r := range []rune{'a', '_', '-', ' ', ' ', ' ', '\u0085'} {
		if conditionIsWhitespace(r) {
			t.Errorf("conditionIsWhitespace(%U) = true, expected false", r)
		}
	}
}
