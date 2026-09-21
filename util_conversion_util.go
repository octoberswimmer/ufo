// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/util/ConversionUtil.java

package ufo

// ConversionUtil holds utility methods for data conversion.

// ConversionUtilRgbToColor is not ported. The Java method converts an
// org.w3c.dom.css.RGBColor to a java.awt.Color, neither of which exists in
// this port, and its one caller is the Swing DOMInspector, which is not
// ported either.
func ConversionUtilRgbToColor(color any) any {
	panic(NewXRRuntimeException("not ported: ConversionUtil.rgbToColor"))
}
