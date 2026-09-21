// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/EmptyStyle.java

package ufo

// EmptyStyle represents the outer box to be used for evaluating positioning
// of internal boxes.
type EmptyStyle struct {
	CalculatedStyle
}

func NewEmptyStyle() *EmptyStyle {
	e := &EmptyStyle{}
	e.initCalculatedStyle(nil)
	e.self = e
	return e
}
