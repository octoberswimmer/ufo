// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/DerivedValue.java

package ufo

// DerivedValue is the abstract base of the derived value classes. It is a
// struct to embed; the methods a subclass has to override panic.
type DerivedValue struct {
	asString       string
	cssSacUnitType int16
}

// NewDerivedValue ports the protected constructor. cssStringValue is nullable
// in Java; nil stands for null.
func NewDerivedValue(name *CSSName, cssSACUnitType int16, cssText *string, cssStringValue *string) DerivedValue {
	d := DerivedValue{cssSacUnitType: cssSACUnitType}

	if cssText == nil {
		stringValue := "null"
		if cssStringValue != nil {
			stringValue = *cssStringValue
		}
		panic(NewXRRuntimeException(
			"CSSValue for '" + name.ToString() + "' is null after " +
				"resolving CSS identifier for value '" + stringValue + "'"))
	}
	d.asString = d.deriveStringValue(*cssText, cssStringValue)
	return d
}

func (d *DerivedValue) deriveStringValue(cssText string, cssStringValue *string) string {
	switch d.cssSacUnitType {
	case CSSPrimitiveValueCssIdent, CSSPrimitiveValueCssString, CSSPrimitiveValueCssUri, CSSPrimitiveValueCssAttr:
		if cssStringValue != nil {
			return *cssStringValue
		}
		return cssText
	default:
		return cssText
	}
}

// GetStringValue returns the getCssText() or getStringValue(), depending.
func (d *DerivedValue) GetStringValue() string {
	return d.asString
}

// IsDeclaredInherit is always false: a value declared INHERIT should always
// be IdentValueInherit, not a DerivedValue.
func (d *DerivedValue) IsDeclaredInherit() bool {
	return false
}

func (d *DerivedValue) GetCssSacUnitType() int16 {
	return d.cssSacUnitType
}

func (d *DerivedValue) IsAbsoluteUnit() bool {
	return ValueConstantsIsAbsoluteUnit(d.cssSacUnitType)
}

func (d *DerivedValue) AsFloat() float32 {
	panic(NewXRRuntimeException("asFloat() needs to be overridden in subclass."))
}

func (d *DerivedValue) AsColor() FSColor {
	panic(NewXRRuntimeException("asColor() needs to be overridden in subclass."))
}

func (d *DerivedValue) GetFloatProportionalTo(cssName *CSSName, baseValue float32, ctx CssContext) float32 {
	panic(NewXRRuntimeException("getFloatProportionalTo() needs to be overridden in subclass."))
}

func (d *DerivedValue) AsString() string {
	return d.GetStringValue()
}

func (d *DerivedValue) AsStringArray() []string {
	panic(NewXRRuntimeException("asStringArray() needs to be overridden in subclass."))
}

func (d *DerivedValue) AsIdentValue() *IdentValue {
	panic(NewXRRuntimeException("asIdentValue() needs to be overridden in subclass."))
}

func (d *DerivedValue) HasAbsoluteUnit() bool {
	panic(NewXRRuntimeException("hasAbsoluteUnit() needs to be overridden in subclass."))
}

func (d *DerivedValue) IsIdent() bool {
	return false
}

func (d *DerivedValue) IsDependentOnFontSize() bool {
	return false
}
