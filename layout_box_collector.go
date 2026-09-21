// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BoxCollector.java

package ufo

import (
	"fmt"

	"github.com/octoberswimmer/ufo/geom"
)

// BoxCollector collects boxes which intersect a given clip region.  If
// available, aggregate bounds information will be used.  Block and inline
// content are added to separate lists as they are painted in separate render
// phases.
//
// The Java methods append to the List<Box> arguments, so the Go methods take
// pointers to the slices. Box identity (Java ==) is compared through AsBox(),
// which gives the same pointer no matter which interface value holds the box.
type BoxCollector struct{}

func NewBoxCollector() *BoxCollector {
	return &BoxCollector{}
}

func (bc *BoxCollector) Collect(
	c CssContext, clip geom.Shape, layer *Layer,
	blockContent *[]BoxI, inlineContent *[]BoxI, rangeLists *BoxRangeLists) {
	if layer.IsInline() {
		bc.collectInlineLayer(c, clip, layer, blockContent, inlineContent, rangeLists)
	} else {
		bc.CollectWithContainer(c, clip, layer, layer.GetMaster(), blockContent, inlineContent, rangeLists)
	}
}

func (bc *BoxCollector) IntersectsAny(
	c CssContext, clip geom.Shape, master BoxI) bool {
	return bc.intersectsAnyWithContainer(c, clip, master, master)
}

func (bc *BoxCollector) collectInlineLayer(
	c CssContext, clip geom.Shape, layer *Layer,
	blockContent *[]BoxI, inlineContent *[]BoxI, rangeLists *BoxRangeLists) {
	iB := layer.GetMaster().(*InlineLayoutBox)
	content := iB.GetElementWithContent()

	for _, b := range content {
		if b.Intersects(c, clip) {
			if _, ok := b.(*InlineLayoutBox); ok {
				*inlineContent = append(*inlineContent, b)
			} else if bb, ok := b.(BlockBoxI); ok {
				if bb.IsInline() {
					if bc.IntersectsAny(c, clip, b) {
						*inlineContent = append(*inlineContent, bb)
					}
				} else {
					bc.CollectWithContainer(c, clip, layer, bb, blockContent, inlineContent, rangeLists)
				}
			} else {
				panic(NewXRRuntimeException(fmt.Sprintf("Unexpected element type: %T", b)))
			}
		}
	}
}

func (bc *BoxCollector) intersectsAggregateBounds(clip geom.Shape, box BoxI) bool {
	if clip == nil {
		return true
	}
	info := box.GetPaintingInfo()
	if info == nil {
		return false
	}
	bounds := info.GetAggregateBounds()
	return clip.Intersects(bounds)
}

func (bc *BoxCollector) CollectWithContainer(
	c CssContext, clip geom.Shape, layer *Layer, container BoxI,
	blockContent *[]BoxI, inlineContent *[]BoxI, rangeLists *BoxRangeLists) {
	if layer != container.GetContainingLayer() {
		return
	}

	_, isBlock := container.(BlockBoxI)

	blockStart := 0
	inlineStart := 0
	blockRangeStart := 0
	inlineRangeStart := 0
	if isBlock {
		blockStart = len(*blockContent)
		inlineStart = len(*inlineContent)

		blockRangeStart = len(rangeLists.GetBlock())
		inlineRangeStart = len(rangeLists.GetInline())
	}

	if lineBox, ok := container.(*LineBox); ok {
		if bc.intersectsAggregateBounds(clip, container) ||
			container.GetPaintingInfo() == nil && container.Intersects(c, clip) {
			*inlineContent = append(*inlineContent, container)
			lineBox.AddAllChildrenWithLayer(inlineContent, layer)
		}
	} else {
		intersectsAggregateBounds := bc.intersectsAggregateBounds(clip, container)
		if container.GetLayer() == nil || !isBlock {
			if intersectsAggregateBounds ||
				container.GetPaintingInfo() == nil && container.Intersects(c, clip) {
				*blockContent = append(*blockContent, container)
				if container.GetStyle().IsTable() { // HACK
					if renderingContext, ok := c.(*RenderingContext); ok {
						table := container.(*TableBox)
						if table.HasContentLimitContainer() {
							table.UpdateHeaderFooterPosition(renderingContext)
						}
					}
				}
			}
		}

		if container.GetPaintingInfo() == nil || intersectsAggregateBounds {
			if container.GetLayer() == nil || container.AsBox() == layer.GetMaster().AsBox() {
				for i := 0; i < container.GetChildCount(); i++ {
					child := container.GetChild(i)
					bc.CollectWithContainer(c, clip, layer, child, blockContent, inlineContent, rangeLists)
				}
			}
		}
	}

	bc.saveRangeData(
		c, container, blockContent, inlineContent,
		rangeLists, isBlock, blockStart, inlineStart,
		blockRangeStart, inlineRangeStart)
}

func (bc *BoxCollector) saveRangeData(
	c CssContext, container BoxI, blockContent *[]BoxI, inlineContent *[]BoxI,
	rangeLists *BoxRangeLists, isBlock bool, blockStart int, inlineStart int,
	blockRangeStart int, inlineRangeStart int) {
	renderingContext, isRenderingContext := c.(*RenderingContext)
	if isBlock && isRenderingContext {
		blockBox := container.(BlockBoxI)
		if blockBox.IsNeedsClipOnPaint(renderingContext) {
			blockEnd := len(*blockContent)
			if blockStart != blockEnd {
				boxRange := NewBoxRange(blockStart, blockEnd)
				rangeLists.insertBlock(blockRangeStart, NewBoxRangeData(blockBox, boxRange))
			}

			inlineEnd := len(*inlineContent)
			if inlineStart != inlineEnd {
				boxRange := NewBoxRange(inlineStart, inlineEnd)
				rangeLists.insertInline(inlineRangeStart, NewBoxRangeData(blockBox, boxRange))
			}
		}
	}
}

func (bc *BoxCollector) intersectsAnyWithContainer(
	c CssContext, clip geom.Shape,
	master BoxI, container BoxI) bool {
	if _, ok := container.(*LineBox); ok {
		return container.Intersects(c, clip)
	} else {
		_, isBlock := container.(BlockBoxI)
		if container.GetLayer() == nil || !isBlock {
			if container.Intersects(c, clip) {
				return true
			}
		}

		if container.GetLayer() == nil || container.AsBox() == master.AsBox() {
			for i := 0; i < container.GetChildCount(); i++ {
				child := container.GetChild(i)
				possibleResult := bc.intersectsAnyWithContainer(c, clip, master, child)
				if possibleResult {
					return true
				}
			}
		}
	}

	return false
}
