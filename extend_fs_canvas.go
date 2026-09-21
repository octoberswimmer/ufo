// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/FSCanvas.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

type FSCanvas interface {
	GetFixedRectangle() *geom.Rectangle

	GetX() int

	GetY() int
}
