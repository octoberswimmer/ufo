// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/CssFunction.java

package ufo

type CssFunction interface {
	Evaluate() string
}
