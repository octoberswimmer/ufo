// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/FSColor.java

package ufo

type FSColor interface {
	LightenColor() FSColor
	DarkenColor() FSColor
	// ToString returns the CSS text of the color (Java Object.toString, which
	// PropertyValue uses as the cssText of a color value).
	ToString() string
}
