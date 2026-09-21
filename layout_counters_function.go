// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/CountersFunction.java

package ufo

import "strings"

type CountersFunction struct {
	listStyleType *IdentValue
	counterValues []int
	separator     string
}

func NewCountersFunction(counterValues []int, separator string, listStyleType *IdentValue) *CountersFunction {
	return &CountersFunction{listStyleType: listStyleType, counterValues: counterValues, separator: separator}
}

func (f *CountersFunction) Evaluate() string {
	var sb strings.Builder
	for i, value := range f.counterValues {
		sb.WriteString(CounterFunctionCreateCounterText(f.listStyleType, value))
		if i < len(f.counterValues)-1 {
			sb.WriteString(f.separator)
		}
	}
	return sb.String()
}
