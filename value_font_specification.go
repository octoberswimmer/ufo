// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/value/FontSpecification.java

package ufo

// FontSpecification ports a Java record.
//
// User: tobe
// Date: 2005-jun-23
type FontSpecification struct {
	size       float32
	fontWeight *IdentValue
	families   []string
	fontStyle  *IdentValue
	variant    *IdentValue
}

func NewFontSpecification(size float32, fontWeight *IdentValue, families []string, fontStyle *IdentValue, variant *IdentValue) *FontSpecification {
	return &FontSpecification{
		size:       size,
		fontWeight: fontWeight,
		families:   families,
		fontStyle:  fontStyle,
		variant:    variant,
	}
}

func (f *FontSpecification) Size() float32 {
	return f.size
}

func (f *FontSpecification) FontWeight() *IdentValue {
	return f.fontWeight
}

func (f *FontSpecification) Families() []string {
	return f.families
}

func (f *FontSpecification) FontStyle() *IdentValue {
	return f.fontStyle
}

func (f *FontSpecification) Variant() *IdentValue {
	return f.variant
}
