// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/FloatedBoxData.java

package ufo

// FloatedBoxData is a bean containing additional information used by floated
// boxes. The marginFromSibling property contains the margin from our
// previous inflow block level sibling (if it exists). It is necessary to
// correctly position the box when collapsing vertical margins.
type FloatedBoxData struct {
	drawingLayer      *Layer
	manager           *FloatManager
	marginFromSibling int
}

func NewFloatedBoxData() *FloatedBoxData {
	return &FloatedBoxData{}
}

// GetDrawingLayer may return nil.
func (f *FloatedBoxData) GetDrawingLayer() *Layer {
	return f.drawingLayer
}

func (f *FloatedBoxData) SetDrawingLayer(drawingLayer *Layer) {
	f.drawingLayer = drawingLayer
}

// GetManager may return nil.
func (f *FloatedBoxData) GetManager() *FloatManager {
	return f.manager
}

func (f *FloatedBoxData) SetManager(manager *FloatManager) {
	f.manager = manager
}

func (f *FloatedBoxData) GetMarginFromSibling() int {
	return f.marginFromSibling
}

func (f *FloatedBoxData) SetMarginFromSibling(marginFromSibling int) {
	f.marginFromSibling = marginFromSibling
}
