// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/ColorValue.java

package ufo

type ColorValue struct {
	DerivedValue
	// color may be nil.
	color FSColor
}

func NewColorValue(name *CSSName, value *PropertyValue) *ColorValue {
	cssText := value.GetCssText()
	return &ColorValue{
		DerivedValue: NewDerivedValue(name, value.GetPrimitiveType(), &cssText, &cssText),

		color: value.GetFSColor(),
	}
}

// AsColor returns the value as a Color, if it is a color.
func (c *ColorValue) AsColor() FSColor {
	return c.color
}
