// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/extend/ContentFunction.java

package ufo

// ContentFunction is the interface for objects which implement a function
// which creates content (e.g. counter(pages))
type ContentFunction interface {
	// IsStatic reports whether the function value can change at render time.
	// If true, Calculate will be called.
	// If false, CalculateWithText will be called.
	IsStatic() bool

	// Calculate is calculate(LayoutContext, FSFunction).
	Calculate(c *LayoutContext, function *FSFunction) string

	// CalculateWithText is calculate(RenderingContext, FSFunction, InlineText).
	CalculateWithText(c *RenderingContext, function *FSFunction, text *InlineText) string

	// GetLayoutReplacementText: if a function value can change at render time
	// (i.e. IsStatic returns false) use this text as an approximation at
	// layout.
	GetLayoutReplacementText() string

	CanHandle(c *LayoutContext, function *FSFunction) bool
}
