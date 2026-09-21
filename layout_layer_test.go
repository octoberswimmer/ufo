package ufo

import "testing"

func TestLayer_pseudoPage(t *testing.T) {
	want := []string{"first", "left", "right", "left", "right"}
	for size, pseudoPage := range want {
		if got := layerPseudoPage(size); got != pseudoPage {
			t.Errorf("layerPseudoPage(%d) = %q, want %q", size, got, pseudoPage)
		}
	}
}

// layerTestStyle implements the CalculatedStyleI methods that the Layer
// constructors and GetZIndex call; every other method panics on the nil
// embedded interface.
type layerTestStyle struct {
	CalculatedStyleI
	positioned bool
	autoZIndex bool
	zIndex     float32
	transform  bool
}

func (s *layerTestStyle) IsPositioned() bool {
	return s.positioned
}

func (s *layerTestStyle) IsAutoZIndex() bool {
	return s.autoZIndex
}

func (s *layerTestStyle) HasTransform() bool {
	return s.transform
}

func (s *layerTestStyle) AsFloat(cssName *CSSName) float32 {
	if cssName != CSSNameZIndex {
		panic("unexpected property")
	}
	return s.zIndex
}

// layerTestBox implements the BlockBoxI methods that the Layer constructors
// call.
type layerTestBox struct {
	BlockBoxI
	box             Box
	name            string
	style           *layerTestStyle
	layer           *Layer
	containingLayer *Layer
}

func (b *layerTestBox) AsBox() *Box {
	return &b.box
}

func (b *layerTestBox) GetStyle() CalculatedStyleI {
	return b.style
}

func (b *layerTestBox) SetLayer(layer *Layer) {
	b.layer = layer
}

func (b *layerTestBox) SetContainingLayer(layer *Layer) {
	b.containingLayer = layer
}

func layerTestPositioned(name string, zIndex float32) *layerTestBox {
	return &layerTestBox{name: name, style: &layerTestStyle{positioned: true, zIndex: zIndex}}
}

func layerTestAuto(name string) *layerTestBox {
	return &layerTestBox{name: name, style: &layerTestStyle{positioned: true, autoZIndex: true}}
}

func layerTestNames(layers []*Layer) []string {
	var result []string
	for _, layer := range layers {
		result = append(result, layer.GetMaster().(*layerTestBox).name)
	}
	return result
}

func layerTestAssertNames(t *testing.T, what string, layers []*Layer, want ...string) {
	t.Helper()
	got := layerTestNames(layers)
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", what, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", what, got, want)
		}
	}
}

func TestLayer_constructorsSetLayerOnMaster(t *testing.T) {
	rootBox := &layerTestBox{name: "root", style: &layerTestStyle{autoZIndex: true}}
	root := NewLayer(rootBox)
	if rootBox.layer != root || rootBox.containingLayer != root {
		t.Errorf("NewLayer did not set the layer and the containing layer of its master")
	}
	if !root.IsStackingContext() || !root.IsRootLayer() || root.GetParent() != nil {
		t.Errorf("NewLayer(master) is not a root stacking context")
	}
	if root.GetZIndex() != 0 {
		t.Errorf("GetZIndex() with z-index: auto = %d, want 0", root.GetZIndex())
	}

	positioned := NewLayerWithParent(root, layerTestPositioned("p", 3.9))
	if !positioned.IsStackingContext() || positioned.IsRootLayer() {
		t.Errorf("a positioned box with a z-index must be a stacking context that is not the root layer")
	}
	if positioned.GetZIndex() != 3 {
		t.Errorf("GetZIndex() = %d, want 3", positioned.GetZIndex())
	}

	auto := NewLayerWithParent(root, layerTestAuto("a"))
	if auto.IsStackingContext() {
		t.Errorf("a positioned box with z-index: auto must not be a stacking context")
	}

	transformed := NewLayerWithParent(root, &layerTestBox{name: "t", style: &layerTestStyle{autoZIndex: true, transform: true}})
	if !transformed.IsStackingContext() {
		t.Errorf("a transformed box must be a stacking context")
	}
	if transformed.FindRoot() != root {
		t.Errorf("FindRoot() did not return the root layer")
	}
}

func TestLayer_collectAndSortLayers(t *testing.T) {
	root := NewLayer(&layerTestBox{name: "root", style: &layerTestStyle{autoZIndex: true}})
	add := func(parent *Layer, box *layerTestBox) *Layer {
		layer := NewLayerWithParent(parent, box)
		parent.AddChild(layer)
		return layer
	}

	add(root, layerTestPositioned("z5", 5))
	auto1 := add(root, layerTestAuto("auto1"))
	add(root, layerTestPositioned("z-2", -2))
	add(root, layerTestPositioned("z1-first", 1))
	add(root, layerTestPositioned("z0", 0))
	// Stacking contexts below a layer that is not one belong to the nearest
	// enclosing stacking context.
	add(auto1, layerTestPositioned("z1-second", 1))
	add(auto1, layerTestPositioned("z-7", -7))
	auto2 := add(auto1, layerTestAuto("auto2"))
	add(auto2, layerTestPositioned("z0-nested", 0))
	// Layers below a stacking context are not collected by its parent.
	z9 := add(root, layerTestPositioned("z9", 9))
	add(z9, layerTestPositioned("z100", 100))

	layerTestAssertNames(t, "children", root.GetChildren(), "z5", "auto1", "z-2", "z1-first", "z0", "z9")
	layerTestAssertNames(t, "negative", root.getSortedLayers(LayerWidthNegative), "z-7", "z-2")
	layerTestAssertNames(t, "zero", root.getSortedLayers(LayerWidthZero), "z0", "z0-nested")
	// The sort is stable: layers with equal z-index keep collection order.
	layerTestAssertNames(t, "positive", root.getSortedLayers(LayerWidthPositive), "z1-first", "z1-second", "z5", "z9")
	layerTestAssertNames(t, "auto", root.collectLayers(LayerWidthAuto), "auto1", "auto2")
}

func TestLayer_detach(t *testing.T) {
	root := NewLayer(&layerTestBox{name: "root", style: &layerTestStyle{autoZIndex: true}})
	a := NewLayerWithParent(root, layerTestAuto("a"))
	b := NewLayerWithParent(root, layerTestAuto("b"))
	root.AddChild(a)
	root.AddChild(b)

	before := root.GetChildren()
	a.Detach()
	layerTestAssertNames(t, "children after detach", root.GetChildren(), "b")
	layerTestAssertNames(t, "list returned before detach", before, "a", "b")

	// Detaching the root layer does nothing.
	root.Detach()

	defer func() {
		if recover() == nil {
			t.Errorf("detaching a layer twice did not panic")
		}
	}()
	a.Detach()
}

func TestLayer_noPageSequences(t *testing.T) {
	root := NewLayer(&layerTestBox{name: "root", style: &layerTestStyle{autoZIndex: true}})
	if _, ok := root.getSortedPageSequences(); ok {
		t.Fatalf("getSortedPageSequences() reported sequences before any was added")
	}
}

func TestLayer_emptyLayerPageQueriesAndFlags(t *testing.T) {
	root := NewLayer(&layerTestBox{name: "root", style: &layerTestStyle{autoZIndex: true}})
	if len(root.GetPages()) != 0 || root.GetLastPage() != nil {
		t.Errorf("a new layer has pages")
	}
	if root.CrossesPageBreak(nil, -1, 10) {
		t.Errorf("CrossesPageBreak with a negative top = true, want false")
	}
	if root.GetPage(nil, -1) != nil {
		t.Errorf("GetPage with a negative offset did not return nil")
	}

	root.SetInline(true)
	root.SetRequiresLayout(true)
	if !root.IsInline() || !root.IsRequiresLayout() {
		t.Errorf("SetInline/SetRequiresLayout were not stored")
	}
}
