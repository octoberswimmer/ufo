// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/BreakPointsProvider.java

package ufo

// BreakPointsProvider supplies the break points of a text one after another.
//
// Author of the Java interface: Lukas Zaruba, lukas.zaruba@gmail.com
type BreakPointsProvider interface {
	// Next returns the next breaking point if available.
	// If there are no more breaking points, it returns a BreakPoint with
	// position == BreakIteratorDone (-1).
	Next() *BreakPoint
}
