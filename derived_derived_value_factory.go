// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/DerivedValueFactory.java

package ufo

import "sync"

// Java uses an unsynchronized static HashMap; concurrent writes to a Go map
// abort the program, so the map is guarded by a mutex.
var derivedValueFactoryCachedColors = map[string]FSDerivedValue{}
var derivedValueFactoryCachedColorsMutex sync.Mutex

// DerivedValueFactoryNewDerivedValue accepts a nil style unless the value is
// inherit.
func DerivedValueFactoryNewDerivedValue(
	style CalculatedStyleI, cssName *CSSName, value *PropertyValue) FSDerivedValue {
	if value.GetCssValueType() == CSSValueCssInherit {
		return style.GetParent().ValueByName(cssName)
	}
	switch value.GetPropertyValueType() {
	case PropertyValueTypeValueTypeLength:
		return NewLengthValue(style, cssName, value)
	case PropertyValueTypeValueTypeIdent:
		return derivedValueFactoryGetIdentValue(value)
	case PropertyValueTypeValueTypeString:
		return NewStringValue(cssName, value)
	case PropertyValueTypeValueTypeNumber:
		return NewNumberValue(cssName, value)
	case PropertyValueTypeValueTypeColor:
		return derivedValueFactoryGetColor(cssName, value, value.GetCssText())
	case PropertyValueTypeValueTypeList:
		return NewListValue(cssName, value)
	case PropertyValueTypeValueTypeFunction:
		return NewFunctionValue(cssName, value)
	}
	panic(NewXRRuntimeException("Unknown property value type: " + value.GetPropertyValueType().ToString()))
}

func derivedValueFactoryGetIdentValue(value *PropertyValue) *IdentValue {
	if ident := value.GetIdentValue(); ident != nil {
		return ident
	}
	return IdentValueGetByIdentString(value.GetStringValue())
}

func derivedValueFactoryGetColor(cssName *CSSName, value *PropertyValue, cssText string) FSDerivedValue {
	derivedValueFactoryCachedColorsMutex.Lock()
	defer derivedValueFactoryCachedColorsMutex.Unlock()
	color, ok := derivedValueFactoryCachedColors[cssText]
	if !ok {
		color = NewColorValue(cssName, value)
		derivedValueFactoryCachedColors[cssText] = color
	}
	return color
}
