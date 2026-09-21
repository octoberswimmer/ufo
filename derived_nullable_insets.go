// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/NullableInsets.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// NullableInsets ports the record NullableInsets; each side is nil when it is
// not set.
type NullableInsets struct {
	top    *int
	left   *int
	bottom *int
	right  *int
}

func NewNullableInsets(top *int, left *int, bottom *int, right *int) *NullableInsets {
	return &NullableInsets{top: top, left: left, bottom: bottom, right: right}
}

func (n *NullableInsets) Top() *int {
	return n.top
}

func (n *NullableInsets) Left() *int {
	return n.left
}

func (n *NullableInsets) Bottom() *int {
	return n.bottom
}

func (n *NullableInsets) Right() *int {
	return n.right
}

func (n *NullableInsets) WithDefaults(defaults *geom.Insets) *geom.Insets {
	top := n.atLeast(n.Top(), defaults.Top)
	left := n.atLeast(n.Left(), defaults.Left)
	bottom := n.atLeast(n.Bottom(), defaults.Bottom)
	right := n.atLeast(n.Right(), defaults.Right)
	return geom.NewInsets(top, left, bottom, right)
}

func (n *NullableInsets) atLeast(value *int, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	return max(defaultValue, *value)
}
