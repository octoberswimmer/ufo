// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BoxRangeHelper.java

package ufo

type BoxRangeHelper struct {
	clipRegionStack []*BoxRangeData

	outputDevice OutputDevice
	rangeList    []*BoxRangeData

	rangeIndex int
	// current is nil once every range has been pushed.
	current *BoxRangeData
}

func NewBoxRangeHelper(outputDevice OutputDevice, rangeList []*BoxRangeData) *BoxRangeHelper {
	h := &BoxRangeHelper{
		outputDevice: outputDevice,
		rangeList:    rangeList,
	}

	if len(rangeList) != 0 {
		h.current = rangeList[0]
	}
	return h
}

func (h *BoxRangeHelper) CheckFinished() {
	if len(h.clipRegionStack) != 0 {
		panic(NewXRRuntimeException("internal error"))
	}
}

func (h *BoxRangeHelper) PushClipRegion(c *RenderingContext, contentIndex int) {
	for h.current != nil && h.current.GetRange().GetStart() == contentIndex {
		h.current.SetClip(h.outputDevice.GetClip())
		h.clipRegionStack = append(h.clipRegionStack, h.current)

		// A nil *geom.Rectangle is passed on as a nil Shape.
		if clipEdge := h.current.GetBox().GetChildrenClipEdge(c); clipEdge != nil {
			h.outputDevice.Clip(clipEdge)
		} else {
			h.outputDevice.Clip(nil)
		}

		if h.rangeIndex == len(h.rangeList)-1 {
			h.current = nil
		} else {
			h.rangeIndex++
			h.current = h.rangeList[h.rangeIndex]
		}
	}
}

func (h *BoxRangeHelper) PopClipRegions(contentIndex int) {
	for len(h.clipRegionStack) != 0 {
		data := h.clipRegionStack[len(h.clipRegionStack)-1]
		if data.GetRange().GetEnd() == contentIndex {
			h.outputDevice.SetClip(data.GetClip())
			h.clipRegionStack = h.clipRegionStack[:len(h.clipRegionStack)-1]
		} else {
			break
		}
	}
}
