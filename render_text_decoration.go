// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/TextDecoration.java

package ufo

type TextDecoration struct {
	identValue *IdentValue
	offset     int
	thickness  int
}

func NewTextDecoration(identValue *IdentValue) *TextDecoration {
	return &TextDecoration{identValue: identValue}
}

func (t *TextDecoration) GetOffset() int {
	return t.offset
}

func (t *TextDecoration) SetOffset(offset int) {
	t.offset = offset
}

func (t *TextDecoration) GetThickness() int {
	return t.thickness
}

func (t *TextDecoration) SetThickness(thickness int) {
	if thickness == 0 {
		t.thickness = 1
	} else {
		t.thickness = thickness
	}
}

func (t *TextDecoration) GetIdentValue() *IdentValue {
	return t.identValue
}
