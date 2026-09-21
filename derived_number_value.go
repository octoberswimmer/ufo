// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/NumberValue.java

package ufo

type NumberValue struct {
	DerivedValue
	floatValue float32
}

func NewNumberValue(cssName *CSSName, value *PropertyValue) *NumberValue {
	cssText := value.GetCssText()
	return &NumberValue{
		DerivedValue: NewDerivedValue(cssName, value.GetPrimitiveType(), &cssText, &cssText),
		floatValue:   value.GetFloatValue(),
	}
}

func (n *NumberValue) AsFloat() float32 {
	return n.floatValue
}

func (n *NumberValue) GetFloatProportionalTo(cssName *CSSName, baseValue float32, ctx CssContext) float32 {
	return n.floatValue
}

func (n *NumberValue) HasAbsoluteUnit() bool {
	return true
}
