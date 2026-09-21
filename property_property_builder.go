// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/property/PropertyBuilder.java

package ufo

// PropertyBuilder builds the PropertyDeclaration objects for one CSS property.
//
// Java declares the values as List<? extends CSSPrimitiveValue>; every caller
// passes PropertyValue objects and most builders cast to PropertyValue, so the
// Go signature takes []*PropertyValue.
type PropertyBuilder interface {
	// BuildDeclarationsWithInheritAllowed builds a list of PropertyDeclaration
	// objects for the CSS property cssName.
	BuildDeclarationsWithInheritAllowed(cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool, inheritAllowed bool) []*PropertyDeclaration

	BuildDeclarations(cssName *CSSName, values []*PropertyValue, origin StylesheetInfoOrigin, important bool) []*PropertyDeclaration
}
