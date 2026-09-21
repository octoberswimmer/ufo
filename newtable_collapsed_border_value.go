// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/CollapsedBorderValue.java

package ufo

// CollapsedBorderValue encapsulates all information related to a particular border side
// along with an overall precedence (e.g. cell borders take precedence over
// row borders).  It is used when comparing overlapping borders when calculating
// collapsed borders.
type CollapsedBorderValue struct {
	style      *IdentValue
	width      int
	color      FSColor
	precedence int
}

func NewCollapsedBorderValue(style *IdentValue, width int, color FSColor, precedence int) *CollapsedBorderValue {
	return &CollapsedBorderValue{
		style:      style,
		width:      width,
		color:      color,
		precedence: precedence,
	}
}

func (v *CollapsedBorderValue) Color() FSColor {
	return v.color
}

// Style may return nil.
func (v *CollapsedBorderValue) Style() *IdentValue {
	return v.style
}

func (v *CollapsedBorderValue) Width() int {
	return v.width
}

func (v *CollapsedBorderValue) WithWidth(width int) *CollapsedBorderValue {
	return NewCollapsedBorderValue(v.style, width, v.color, v.precedence)
}

func (v *CollapsedBorderValue) Precedence() int {
	return v.precedence
}

func (v *CollapsedBorderValue) Defined() bool {
	return v.style != nil
}

func (v *CollapsedBorderValue) Exists() bool {
	return v.style != nil && v.style != IdentValueNone && v.style != IdentValueHidden
}

func (v *CollapsedBorderValue) Hidden() bool {
	return v.style == IdentValueHidden
}

func CollapsedBorderValueBorderLeft(border *BorderPropertySet, precedence int) *CollapsedBorderValue {
	return NewCollapsedBorderValue(
		border.LeftStyle(), int(border.Left()), border.LeftColor(), precedence)
}

func CollapsedBorderValueBorderRight(border *BorderPropertySet, precedence int) *CollapsedBorderValue {
	return NewCollapsedBorderValue(
		border.RightStyle(), int(border.Right()), border.RightColor(), precedence)
}

func CollapsedBorderValueBorderTop(border *BorderPropertySet, precedence int) *CollapsedBorderValue {
	return NewCollapsedBorderValue(
		border.TopStyle(), int(border.Top()), border.TopColor(), precedence)
}

func CollapsedBorderValueBorderBottom(border *BorderPropertySet, precedence int) *CollapsedBorderValue {
	return NewCollapsedBorderValue(
		border.BottomStyle(), int(border.Bottom()), border.BottomColor(), precedence)
}
