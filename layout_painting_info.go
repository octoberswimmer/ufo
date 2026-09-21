// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/PaintingInfo.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// PaintingInfo is a bean which every box uses to provide its aggregate bounds
// (which may be larger than the bounds of the box itself when there is
// overhanging content) and its outer margin corner (which is used to calculate
// the size of the canvas).  The aggregate bounds calculation does not take the
// value of the overflow property into account.
type PaintingInfo struct {
	outerMarginCorner *geom.Dimension
	aggregateBounds   *geom.Rectangle
}

func NewPaintingInfo(outerMarginCorner *geom.Dimension, aggregateBounds *geom.Rectangle) *PaintingInfo {
	return &PaintingInfo{outerMarginCorner: outerMarginCorner, aggregateBounds: aggregateBounds}
}

func (p *PaintingInfo) GetAggregateBounds() *geom.Rectangle {
	return p.aggregateBounds
}

func (p *PaintingInfo) GetOuterMarginCorner() *geom.Dimension {
	return p.outerMarginCorner
}

func (p *PaintingInfo) CopyOf() *PaintingInfo {
	return NewPaintingInfo(
		geom.NewDimensionFromDimension(p.outerMarginCorner), geom.NewRectangleFromRectangle(p.aggregateBounds),
	)
}

func (p *PaintingInfo) Translate(tx int, ty int) {
	p.aggregateBounds.Translate(tx, ty)
	p.outerMarginCorner.SetSizeDouble(
		p.outerMarginCorner.GetWidth()+float64(tx), p.outerMarginCorner.GetHeight()+float64(ty))
}
