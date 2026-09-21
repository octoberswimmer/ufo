// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/LineBreakingStrategy.java

package ufo

// LineBreakingStrategy creates the BreakPointsProvider for a text.
//
// Author of the Java interface: Lukas Zaruba, lukas.zaruba@gmail.com
type LineBreakingStrategy interface {
	GetBreakPointsProvider(text string, lang string, style CalculatedStyleI) BreakPointsProvider
}
