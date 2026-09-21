// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/InlineChild.java

package ufo

// InlineChild is the Java marker interface (no methods) of the objects an
// InlineLayoutBox holds. Its implementations are InlineText, InlineLayoutBox
// and BlockBox (with BlockBox's subclasses). It has no methods here either,
// so the Go compiler does not restrict what is stored in it;
// InlineLayoutBox.AddInlineChild panics for any other type, as Java does.
type InlineChild interface {
}
