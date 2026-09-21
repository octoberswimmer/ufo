// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/ListValue.java

package ufo

import "fmt"

var listValueNoValues = []string{}

type ListValue struct {
	DerivedValue
	// values may be nil.
	values []any
}

func NewListValue(name *CSSName, value *PropertyValue) *ListValue {
	cssText := value.GetCssText()
	return &ListValue{
		DerivedValue: NewDerivedValue(name, value.GetPrimitiveType(), &cssText, &cssText),
		values:       value.GetValues(),
	}
}

// GetValues ports the generic <T> List<T> getValues(), which casts the list
// unchecked to whatever element type the caller names. The elements are
// *PropertyValue or *CounterData; listValueGetPropertyValues and
// listValueGetCounterData do the cast for the callers in this package.
func (l *ListValue) GetValues() []any {
	return l.values
}

// listValueGetPropertyValues is getValues() read as List<PropertyValue>.
func listValueGetPropertyValues(l *ListValue) []*PropertyValue {
	if l.values == nil {
		return nil
	}
	result := make([]*PropertyValue, len(l.values))
	for i, value := range l.values {
		result[i] = value.(*PropertyValue)
	}
	return result
}

// listValueGetCounterData is getValues() read as List<CounterData>.
func listValueGetCounterData(l *ListValue) []*CounterData {
	if l.values == nil {
		return nil
	}
	result := make([]*CounterData, len(l.values))
	for i, value := range l.values {
		result[i] = value.(*CounterData)
	}
	return result
}

func (l *ListValue) AsStringArray() []string {
	if len(l.values) == 0 {
		return listValueNoValues
	}

	arr := make([]string, len(l.values))
	i := 0

	for _, value := range l.values {
		arr[i] = listValueToString(value)
		i++
	}

	return arr
}

// listValueToString is Java's value.toString().
func listValueToString(value any) string {
	if v, ok := value.(interface{ ToString() string }); ok {
		return v.ToString()
	}
	return fmt.Sprint(value)
}
