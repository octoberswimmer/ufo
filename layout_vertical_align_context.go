// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/VerticalAlignContext.java

package ufo

// VerticalAlignContext performs the real work of vertically positioning inline
// boxes within a line (i.e. implementing the vertical-align property).  Because
// of the requirements of vertical-align: top/bottom, a VerticalAlignContext
// is actually a tree of VerticalAlignContext objects which all
// must be taken into consideration when aligning content.
type VerticalAlignContext struct {
	measurements []*InlineBoxMeasurements

	inlineTop    int
	inlineTopSet bool

	inlineBottom    int
	inlineBottomSet bool

	paintingTop    int
	paintingTopSet bool

	paintingBottom    int
	paintingBottomSet bool

	children []*verticalAlignContextChildContextData

	// nil for the root context
	parent *VerticalAlignContext
}

// NewVerticalAlignContextVerticalAlignContext is the constructor
// VerticalAlignContext(VerticalAlignContext parent).
func NewVerticalAlignContextVerticalAlignContext(parent *VerticalAlignContext) *VerticalAlignContext {
	return &VerticalAlignContext{parent: parent}
}

// NewVerticalAlignContextInlineBoxMeasurements is the constructor
// VerticalAlignContext(InlineBoxMeasurements initialMeasurements).
func NewVerticalAlignContextInlineBoxMeasurements(initialMeasurements *InlineBoxMeasurements) *VerticalAlignContext {
	v := &VerticalAlignContext{parent: nil}
	v.measurements = append(v.measurements, initialMeasurements)
	return v
}

func (v *VerticalAlignContext) moveTrackedValues(ty int) {
	if v.inlineTopSet {
		v.inlineTop += ty
	}

	if v.inlineBottomSet {
		v.inlineBottom += ty
	}

	if v.paintingTopSet {
		v.paintingTop += ty
	}

	if v.paintingBottomSet {
		v.paintingBottom += ty
	}
}

func (v *VerticalAlignContext) GetInlineBottom() int {
	return v.inlineBottom
}

func (v *VerticalAlignContext) GetInlineTop() int {
	return v.inlineTop
}

func (v *VerticalAlignContext) UpdateInlineTop(inlineTop int) {
	if !v.inlineTopSet || inlineTop < v.inlineTop {
		v.inlineTop = inlineTop
		v.inlineTopSet = true
	}
}

func (v *VerticalAlignContext) UpdatePaintingTop(paintingTop int) {
	if !v.paintingTopSet || paintingTop < v.paintingTop {
		v.paintingTop = paintingTop
		v.paintingTopSet = true
	}
}

func (v *VerticalAlignContext) UpdateInlineBottom(inlineBottom int) {
	if !v.inlineBottomSet || inlineBottom > v.inlineBottom {
		v.inlineBottom = inlineBottom
		v.inlineBottomSet = true
	}
}

func (v *VerticalAlignContext) UpdatePaintingBottom(paintingBottom int) {
	if !v.paintingBottomSet || paintingBottom > v.paintingBottom {
		v.paintingBottom = paintingBottom
		v.paintingBottomSet = true
	}
}

func (v *VerticalAlignContext) GetLineBoxHeight() int {
	return v.inlineBottom - v.inlineTop
}

func (v *VerticalAlignContext) PushMeasurements(measurements *InlineBoxMeasurements) {
	v.measurements = append(v.measurements, measurements)

	v.UpdateInlineTop(measurements.GetInlineTop())
	v.UpdateInlineBottom(measurements.GetInlineBottom())

	v.UpdatePaintingTop(measurements.GetPaintingTop())
	v.UpdatePaintingBottom(measurements.GetPaintingBottom())
}

func (v *VerticalAlignContext) GetParentMeasurements() *InlineBoxMeasurements {
	return v.measurements[len(v.measurements)-1]
}

func (v *VerticalAlignContext) PopMeasurements() {
	v.measurements = v.measurements[:len(v.measurements)-1]
}

func (v *VerticalAlignContext) GetPaintingBottom() int {
	return v.paintingBottom
}

func (v *VerticalAlignContext) GetPaintingTop() int {
	return v.paintingTop
}

func (v *VerticalAlignContext) CreateChild(root BoxI) *VerticalAlignContext {
	vaRoot := v.getRoot()
	initial := vaRoot.measurements[0]

	result := NewVerticalAlignContextVerticalAlignContext(vaRoot)
	result.PushMeasurements(initial)

	vaRoot.children = append(vaRoot.children, newVerticalAlignContextChildContextData(root, result))
	return result
}

func (v *VerticalAlignContext) getChildren() []*verticalAlignContextChildContextData {
	return v.children
}

// GetParent returns nil for the root context.
func (v *VerticalAlignContext) GetParent() *VerticalAlignContext {
	return v.parent
}

func (v *VerticalAlignContext) getRoot() *VerticalAlignContext {
	if v.parent != nil {
		return v.parent
	}
	return v
}

func (v *VerticalAlignContext) merge(context *VerticalAlignContext) {
	v.UpdateInlineBottom(context.GetInlineBottom())
	v.UpdateInlineTop(context.GetInlineTop())

	v.UpdatePaintingBottom(context.GetPaintingBottom())
	v.UpdatePaintingTop(context.GetPaintingTop())
}

func (v *VerticalAlignContext) AlignChildren() {
	children := v.getChildren()
	for _, data := range children {
		data.align()
		v.merge(data.getVerticalAlignContext())
	}
}

// verticalAlignContextChildContextData is the private class
// VerticalAlignContext.ChildContextData.
type verticalAlignContextChildContextData struct {
	root                 BoxI
	verticalAlignContext *VerticalAlignContext
}

func newVerticalAlignContextChildContextData(root BoxI, vaContext *VerticalAlignContext) *verticalAlignContextChildContextData {
	return &verticalAlignContextChildContextData{root: root, verticalAlignContext: vaContext}
}

func (d *verticalAlignContextChildContextData) getVerticalAlignContext() *VerticalAlignContext {
	return d.verticalAlignContext
}

func (d *verticalAlignContextChildContextData) moveContextContents(ty int) {
	d.moveInlineContents(d.root, ty)
}

func (d *verticalAlignContextChildContextData) moveInlineContents(box BoxI, ty int) {
	if d.canBeMoved(box) {
		box.SetY(box.GetY() + ty)
		if iB, ok := box.(*InlineLayoutBox); ok {
			for i := 0; i < iB.GetInlineChildCount(); i++ {
				child := iB.GetInlineChild(i)
				if chilBox, ok := child.(BoxI); ok {
					d.moveInlineContents(chilBox, ty)
				}
			}
		}
	}
}

func (d *verticalAlignContextChildContextData) canBeMoved(box BoxI) bool {
	vAlign := box.GetStyle().GetIdent(CSSNameVerticalAlign)
	return box == d.root ||
		!(vAlign == IdentValueTop || vAlign == IdentValueBottom)
}

func (d *verticalAlignContextChildContextData) align() {
	vAlign := d.root.GetStyle().GetIdent(CSSNameVerticalAlign)
	var delta int
	if vAlign == IdentValueTop {
		delta = d.verticalAlignContext.getRoot().GetInlineTop() -
			d.verticalAlignContext.GetInlineTop()
	} else if vAlign == IdentValueBottom {
		delta = d.verticalAlignContext.getRoot().GetInlineBottom() -
			d.verticalAlignContext.GetInlineBottom()
	} else {
		panic(NewXRRuntimeException("internal error"))
	}

	d.verticalAlignContext.moveTrackedValues(delta)
	d.moveContextContents(delta)
}
