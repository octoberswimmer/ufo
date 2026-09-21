// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/derived/RectPropertySet.java

package ufo

// RectPropertySetI lists the non-private methods of RectPropertySet, which
// BorderPropertySet extends.
type RectPropertySetI interface {
	AsRectPropertySet() *RectPropertySet

	ToString() string
	Top() float32
	Right() float32
	Bottom() float32
	Left() float32
	GetLeftRightDiff() float32
	Height() float32
	Width() float32
	SetTop(top float32)
	SetRight(right float32)
	SetBottom(bottom float32)
	SetLeft(left float32)
	Reset()
	CopyOf() RectPropertySetI
	IsAllZeros() bool
	HasNegativeValues() bool
	ResetNegativeValues() RectPropertySetI
}

var RectPropertySetAllZeros = NewRectPropertySet(0, 0, 0, 0)

// RectPropertySet represents a set of CSS properties that together define
// some rectangular area, and per-side thickness.
//
// No method of RectPropertySet calls a method that BorderPropertySet
// overrides, so the struct holds no self reference.
type RectPropertySet struct {
	top    float32
	right  float32
	bottom float32
	left   float32
}

func NewRectPropertySet(
	top float32,
	right float32,
	bottom float32,
	left float32,
) *RectPropertySet {
	return &RectPropertySet{top: top, right: right, bottom: bottom, left: left}
}

func RectPropertySetNewInstance(
	style CalculatedStyleI,
	sideProperties *CSSNameCSSSideProperties,
	cbWidth float32,
	ctx CssContext,
) RectPropertySetI {
	// HACK isLengthValue is part of margin auto hack
	var top, right, bottom, left float32
	if style.IsLengthOrNumber(sideProperties.Top()) {
		top = style.GetFloatPropertyProportionalHeight(sideProperties.Top(), cbWidth, ctx)
	}
	if style.IsLengthOrNumber(sideProperties.Right()) {
		right = style.GetFloatPropertyProportionalWidth(sideProperties.Right(), cbWidth, ctx)
	}
	if style.IsLengthOrNumber(sideProperties.Bottom()) {
		bottom = style.GetFloatPropertyProportionalHeight(sideProperties.Bottom(), cbWidth, ctx)
	}
	if style.IsLengthOrNumber(sideProperties.Left()) {
		left = style.GetFloatPropertyProportionalWidth(sideProperties.Left(), cbWidth, ctx)
	}
	return NewRectPropertySet(top, right, bottom, left)
}

func (r *RectPropertySet) AsRectPropertySet() *RectPropertySet {
	return r
}

func (r *RectPropertySet) ToString() string {
	return "RectPropertySet[top=" + lengthValueFloatToString(r.top) +
		",right=" + lengthValueFloatToString(r.right) +
		",bottom=" + lengthValueFloatToString(r.bottom) +
		",left=" + lengthValueFloatToString(r.left) + "]"
}

func (r *RectPropertySet) String() string {
	return r.ToString()
}

func (r *RectPropertySet) Top() float32 {
	return r.top
}

func (r *RectPropertySet) Right() float32 {
	return r.right
}

func (r *RectPropertySet) Bottom() float32 {
	return r.bottom
}

func (r *RectPropertySet) Left() float32 {
	return r.left
}

func (r *RectPropertySet) GetLeftRightDiff() float32 {
	return r.left - r.right
}

func (r *RectPropertySet) Height() float32 {
	return r.top + r.bottom
}

func (r *RectPropertySet) Width() float32 {
	return r.left + r.right
}

func (r *RectPropertySet) SetTop(top float32) {
	r.top = top
}

func (r *RectPropertySet) SetRight(right float32) {
	r.right = right
}

func (r *RectPropertySet) SetBottom(bottom float32) {
	r.bottom = bottom
}

func (r *RectPropertySet) SetLeft(left float32) {
	r.left = left
}

func (r *RectPropertySet) Reset() {
	r.SetRight(0)
	r.SetLeft(0)
	r.SetTop(0)
	r.SetBottom(0)
}

func (r *RectPropertySet) CopyOf() RectPropertySetI {
	return NewRectPropertySet(r.top, r.right, r.bottom, r.left)
}

func (r *RectPropertySet) IsAllZeros() bool {
	return r.top == 0.0 && r.right == 0.0 && r.bottom == 0.0 && r.left == 0.0
}

func (r *RectPropertySet) HasNegativeValues() bool {
	return r.top < 0 || r.right < 0 || r.bottom < 0 || r.left < 0
}

func (r *RectPropertySet) ResetNegativeValues() RectPropertySetI {
	return NewRectPropertySet(rectPropertySetMax(0, r.Top()), rectPropertySetMax(0, r.Right()), rectPropertySetMax(0, r.Bottom()), rectPropertySetMax(0, r.Left()))
}

// rectPropertySetMax is Math.max(float, float): NaN if either argument is
// NaN, and positive zero is greater than negative zero.
func rectPropertySetMax(a float32, b float32) float32 {
	return max(a, b)
}
