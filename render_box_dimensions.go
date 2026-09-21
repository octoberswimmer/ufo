// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/BoxDimensions.java

package ufo

import "fmt"

type BoxDimensions struct {
	leftMBP      int
	rightMBP     int
	contentWidth int
	height       int
}

func NewBoxDimensions(leftMBP int, rightMBP int, contentWidth int, height int) *BoxDimensions {
	return &BoxDimensions{
		leftMBP:      leftMBP,
		rightMBP:     rightMBP,
		contentWidth: contentWidth,
		height:       height,
	}
}

func (d *BoxDimensions) GetContentWidth() int {
	return d.contentWidth
}

func (d *BoxDimensions) GetHeight() int {
	return d.height
}

func (d *BoxDimensions) GetLeftMBP() int {
	return d.leftMBP
}

func (d *BoxDimensions) GetRightMBP() int {
	return d.rightMBP
}

func (d *BoxDimensions) ToString() string {
	return fmt.Sprintf("BoxDimensions{left: %d, right: %d, width: %d, height: %d}", d.leftMBP, d.rightMBP, d.contentWidth, d.height)
}

func (d *BoxDimensions) String() string {
	return d.ToString()
}
