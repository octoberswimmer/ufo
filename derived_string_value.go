// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/StringValue.java

package ufo

type StringValue struct {
	DerivedValue
	stringAsArray []string
}

func NewStringValue(name *CSSName, value *PropertyValue) *StringValue {
	cssText := value.GetCssText()
	// PropertyValue.GetStringValue returns "" where Java returns null. A
	// StringValue is built from a PropertyValue of type VALUE_TYPE_STRING,
	// whose string value is never null, so "" here is the empty string.
	stringValue := value.GetStringValue()
	return &StringValue{
		DerivedValue:  NewDerivedValue(name, value.GetPrimitiveType(), &cssText, &stringValue),
		stringAsArray: value.GetStringArrayValue(),
	}
}

func (s *StringValue) AsStringArray() []string {
	return ArrayUtilCloneOrEmptyStringArray(s.stringAsArray)
}

func (s *StringValue) ToString() string {
	return s.GetStringValue()
}

func (s *StringValue) String() string {
	return s.ToString()
}
