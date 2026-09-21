// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/CounterFunction.java

package ufo

import (
	"strconv"
	"strings"
)

var (
	counterFunctionGreekUpperLetters = []rune("ΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩ")
	counterFunctionGreekLowerLetters = []rune("αβγδεζηθικλμνξοπρστυφχψω")
)

type CounterFunction struct {
	listStyleType *IdentValue
	counterValue  int
}

func NewCounterFunction(counterValue int, listStyleType *IdentValue) *CounterFunction {
	return &CounterFunction{counterValue: counterValue, listStyleType: listStyleType}
}

func (f *CounterFunction) Evaluate() string {
	return CounterFunctionCreateCounterText(f.listStyleType, f.counterValue)
}

func CounterFunctionCreateCounterText(listStyle *IdentValue, listCounter int) string {
	if listStyle == IdentValueLowerLatin || listStyle == IdentValueLowerAlpha {
		return strings.ToLower(counterFunctionToLatin(listCounter - 1))
	} else if listStyle == IdentValueUpperLatin || listStyle == IdentValueUpperAlpha {
		return strings.ToUpper(counterFunctionToLatin(listCounter - 1))
	} else if listStyle == IdentValueLowerRoman {
		return strings.ToLower(counterFunctionToRoman(listCounter))
	} else if listStyle == IdentValueLowerGreek {
		return counterFunctionToGreekLower(listCounter - 1)
	} else if listStyle == IdentValueUpperGreek {
		return counterFunctionToGreekUpper(listCounter - 1)
	} else if listStyle == IdentValueUpperRoman {
		return strings.ToUpper(counterFunctionToRoman(listCounter))
	} else if listStyle == IdentValueDecimalLeadingZero {
		if listCounter >= 10 {
			return strconv.Itoa(listCounter)
		}
		return "0" + strconv.Itoa(listCounter)
	} else {
		return strconv.Itoa(listCounter)
	}
}

func counterFunctionToLatin(zeroBasedIndex int) string {
	if zeroBasedIndex < 0 {
		return ""
	}
	return counterFunctionToLatin(zeroBasedIndex/26-1) + string(rune('A'+zeroBasedIndex%26))
}

func counterFunctionToGreekUpper(zeroBasedIndex int) string {
	if zeroBasedIndex < 0 {
		return ""
	}
	return counterFunctionToGreekUpper(zeroBasedIndex/24-1) + string(counterFunctionGreekUpperLetters[zeroBasedIndex%24])
}

func counterFunctionToGreekLower(zeroBasedIndex int) string {
	if zeroBasedIndex < 0 {
		return ""
	}
	return counterFunctionToGreekLower(zeroBasedIndex/24-1) + string(counterFunctionGreekLowerLetters[zeroBasedIndex%24])
}

func counterFunctionToRoman(val int) string {
	ints := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	nums := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var sb strings.Builder
	for i := 0; i < len(ints); i++ {
		count := val / ints[i]
		sb.WriteString(strings.Repeat(nums[i], max(0, count)))
		val -= ints[i] * count
	}
	return sb.String()
}
