// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/PropertyValue.java

package ufo

import (
	"fmt"
	"strconv"
	"strings"
)

// PropertyValueType ports the enum PropertyValue.Type.
type PropertyValueType int

const (
	PropertyValueTypeValueTypeNumber PropertyValueType = iota
	PropertyValueTypeValueTypeLength
	PropertyValueTypeValueTypeColor
	PropertyValueTypeValueTypeIdent
	PropertyValueTypeValueTypeString
	PropertyValueTypeValueTypeList
	PropertyValueTypeValueTypeFunction
)

var propertyValueTypeNames = [...]string{
	PropertyValueTypeValueTypeNumber:   "VALUE_TYPE_NUMBER",
	PropertyValueTypeValueTypeLength:   "VALUE_TYPE_LENGTH",
	PropertyValueTypeValueTypeColor:    "VALUE_TYPE_COLOR",
	PropertyValueTypeValueTypeIdent:    "VALUE_TYPE_IDENT",
	PropertyValueTypeValueTypeString:   "VALUE_TYPE_STRING",
	PropertyValueTypeValueTypeList:     "VALUE_TYPE_LIST",
	PropertyValueTypeValueTypeFunction: "VALUE_TYPE_FUNCTION",
}

func (t PropertyValueType) ToString() string {
	return propertyValueTypeNames[t]
}

func (t PropertyValueType) String() string {
	return propertyValueTypeNames[t]
}

// PropertyValue implements org.w3c.dom.css.CSSPrimitiveValue in Java. The
// methods of that interface are ported as methods of PropertyValue; the
// interface itself has no Go counterpart.
//
// The Java constructors are all overloads of one name. The Go constructors
// are named after the kind of value they build:
//
//	PropertyValue(short, float, String)                    NewPropertyValueFloat
//	PropertyValue(short, float, String, Token)             NewPropertyValueFloatWithOperatorToken
//	PropertyValue(FSColor)                                 NewPropertyValueFSColor
//	PropertyValue(FSColor, Token)                          NewPropertyValueFSColorWithOperatorToken
//	PropertyValue(short, String, String)                   NewPropertyValueString
//	PropertyValue(short, String, String, Token)            NewPropertyValueStringWithOperatorToken
//	PropertyValue(short, String, String, String[], Token)  NewPropertyValueStringWithStringArrayValueOperatorToken
//	PropertyValue(IdentValue)                              NewPropertyValueIdentValue
//	PropertyValue(List)                                    NewPropertyValueList
//	PropertyValue(FSFunction)                              NewPropertyValueFSFunction
//	PropertyValue(FSFunction, Token)                       NewPropertyValueFSFunctionWithOperatorToken
type PropertyValue struct {
	primitiveType int16
	cssValueType  int16

	// stringValue is "" where Java holds null.
	stringValue      string
	floatValue       float32
	stringArrayValue []string

	cssText string

	// fsColor may be nil.
	fsColor FSColor

	// identValue may be nil.
	identValue *IdentValue

	propertyValueType PropertyValueType

	// operator may be nil.
	operator *Token

	values []any
	// function may be nil.
	function *FSFunction
}

func NewPropertyValueFloat(primitiveType int16, floatValue float32, cssText string) *PropertyValue {
	return NewPropertyValueFloatWithOperatorToken(primitiveType, floatValue, cssText, nil)
}

func NewPropertyValueFloatWithOperatorToken(primitiveType int16, floatValue float32, cssText string, operatorToken *Token) *PropertyValue {
	p := &PropertyValue{}
	p.primitiveType = primitiveType
	p.floatValue = floatValue
	p.cssValueType = CSSValueCssPrimitiveValue
	p.cssText = cssText

	if primitiveType == CSSPrimitiveValueCssNumber && floatValue != 0.0 {
		p.propertyValueType = PropertyValueTypeValueTypeNumber
	} else {
		p.propertyValueType = PropertyValueTypeValueTypeLength
	}
	p.values = []any{}
	p.operator = operatorToken
	return p
}

func NewPropertyValueFSColor(color FSColor) *PropertyValue {
	return NewPropertyValueFSColorWithOperatorToken(color, nil)
}

func NewPropertyValueFSColorWithOperatorToken(color FSColor, operatorToken *Token) *PropertyValue {
	p := &PropertyValue{}
	p.primitiveType = CSSPrimitiveValueCssRgbcolor
	p.cssValueType = CSSValueCssPrimitiveValue
	p.cssText = color.ToString()
	p.fsColor = color

	p.propertyValueType = PropertyValueTypeValueTypeColor
	p.values = []any{}
	p.operator = operatorToken
	return p
}

func NewPropertyValueString(primitiveType int16, stringValue string, cssText string) *PropertyValue {
	return NewPropertyValueStringWithOperatorToken(primitiveType, stringValue, cssText, nil)
}

func NewPropertyValueStringWithOperatorToken(primitiveType int16, stringValue string, cssText string, operatorToken *Token) *PropertyValue {
	return NewPropertyValueStringWithStringArrayValueOperatorToken(primitiveType, stringValue, cssText, nil, operatorToken)
}

func NewPropertyValueStringWithStringArrayValueOperatorToken(primitiveType int16, stringValue string, cssText string, stringArrayValue []string, operatorToken *Token) *PropertyValue {
	p := &PropertyValue{}
	p.primitiveType = primitiveType
	p.stringValue = stringValue
	// Must be a case-insensitive compare since ident values aren't normalized
	// for font and font-family
	if strings.EqualFold(p.stringValue, "inherit") {
		p.cssValueType = CSSValueCssInherit
	} else {
		p.cssValueType = CSSValueCssPrimitiveValue
	}
	p.cssText = cssText

	if primitiveType == CSSPrimitiveValueCssIdent {
		p.propertyValueType = PropertyValueTypeValueTypeIdent
	} else {
		p.propertyValueType = PropertyValueTypeValueTypeString
	}
	p.stringArrayValue = ArrayUtilCloneOrEmptyStringArray(stringArrayValue)
	p.values = []any{}
	p.operator = operatorToken
	return p
}

func NewPropertyValueIdentValue(ident *IdentValue) *PropertyValue {
	p := &PropertyValue{}
	p.primitiveType = CSSPrimitiveValueCssIdent
	p.stringValue = ident.ToString()
	if p.stringValue == "inherit" {
		p.cssValueType = CSSValueCssInherit
	} else {
		p.cssValueType = CSSValueCssPrimitiveValue
	}
	p.cssText = ident.ToString()

	p.propertyValueType = PropertyValueTypeValueTypeIdent
	p.identValue = ident
	p.values = []any{}
	return p
}

func NewPropertyValueList(values []any) *PropertyValue {
	p := &PropertyValue{}
	p.primitiveType = CSSPrimitiveValueCssUnknown // HACK
	p.cssValueType = CSSValueCssCustom
	p.cssText = propertyValueListToString(values) // HACK

	p.values = values
	p.propertyValueType = PropertyValueTypeValueTypeList
	return p
}

func NewPropertyValueFSFunction(function *FSFunction) *PropertyValue {
	return NewPropertyValueFSFunctionWithOperatorToken(function, nil)
}

func NewPropertyValueFSFunctionWithOperatorToken(function *FSFunction, operatorToken *Token) *PropertyValue {
	p := &PropertyValue{}
	p.primitiveType = CSSPrimitiveValueCssUnknown
	p.cssValueType = CSSValueCssCustom
	p.cssText = function.ToString()

	p.function = function
	p.propertyValueType = PropertyValueTypeValueTypeFunction
	p.values = []any{}
	p.operator = operatorToken
	return p
}

// propertyValueListToString formats values as java.util.List.toString does:
// "[a, b]", each element through its ToString method.
func propertyValueListToString(values []any) string {
	var result strings.Builder
	result.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			result.WriteString(", ")
		}
		switch v := value.(type) {
		case interface{ ToString() string }:
			result.WriteString(v.ToString())
		case nil:
			result.WriteString("null")
		default:
			result.WriteString(fmt.Sprint(v))
		}
	}
	result.WriteByte(']')
	return result.String()
}

func (p *PropertyValue) GetCounterValue() any {
	panic(NewXRRuntimeException("Unsupported operation: getCounterValue"))
}

// GetFloatValueWithUnitType ports getFloatValue(short unitType), which
// ignores the unit type.
func (p *PropertyValue) GetFloatValueWithUnitType(unitType int16) float32 {
	return p.floatValue
}

func (p *PropertyValue) GetFloatValue() float32 {
	return p.floatValue
}

func (p *PropertyValue) GetPrimitiveType() int16 {
	return p.primitiveType
}

func (p *PropertyValue) GetRGBColorValue() any {
	panic(NewXRRuntimeException("Unsupported operation: getRGBColorValue"))
}

func (p *PropertyValue) GetRectValue() any {
	panic(NewXRRuntimeException("Unsupported operation: getRectValue"))
}

// GetStringValue returns "" where Java returns null.
func (p *PropertyValue) GetStringValue() string {
	return p.stringValue
}

func (p *PropertyValue) SetFloatValue(unitType int16, floatValue float32) {
	panic(NewXRRuntimeException("Unsupported operation: setFloatValue"))
}

func (p *PropertyValue) SetStringValue(stringType int16, stringValue string) {
	panic(NewXRRuntimeException("Unsupported operation: setStringValue"))
}

func (p *PropertyValue) GetCssText() string {
	return p.cssText
}

func (p *PropertyValue) GetCssValueType() int16 {
	return p.cssValueType
}

func (p *PropertyValue) SetCssText(cssText string) {
	panic(NewXRRuntimeException("Unsupported operation: setCssText"))
}

// GetFSColor may return nil.
func (p *PropertyValue) GetFSColor() FSColor {
	return p.fsColor
}

// GetIdentValue may return nil.
func (p *PropertyValue) GetIdentValue() *IdentValue {
	return p.identValue
}

func (p *PropertyValue) SetIdentValue(identValue *IdentValue) {
	p.identValue = identValue
}

func (p *PropertyValue) GetPropertyValueType() PropertyValueType {
	return p.propertyValueType
}

// GetOperator may return nil.
func (p *PropertyValue) GetOperator() *Token {
	return p.operator
}

func (p *PropertyValue) GetStringArrayValue() []string {
	return ArrayUtilCloneOrEmptyStringArray(p.stringArrayValue)
}

func (p *PropertyValue) ToString() string {
	return p.cssText
}

func (p *PropertyValue) String() string {
	return p.cssText
}

// GetValues returns the elements of a list value. The Java method is generic
// in the element type and casts unchecked; callers here assert the element
// type themselves. The returned slice is a copy, in place of Java's
// unmodifiable view.
func (p *PropertyValue) GetValues() []any {
	result := make([]any, len(p.values))
	copy(result, p.values)
	return result
}

// GetFunction may return nil.
func (p *PropertyValue) GetFunction() *FSFunction {
	return p.function
}

func (p *PropertyValue) GetFingerprint() string {
	switch p.GetPropertyValueType() {
	case PropertyValueTypeValueTypeIdent:
		if p.identValue == nil {
			p.identValue = IdentValueGetByIdentString(p.GetStringValue())
		}
		return "I" + strconv.Itoa(p.identValue.FS_ID)
	default:
		return p.GetCssText()
	}
}
