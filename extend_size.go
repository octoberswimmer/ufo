// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/Size.java

package ufo

import "fmt"

// Size is the Java record Size(int width, int height).
type Size struct {
	width  int
	height int
}

func NewSize(width int, height int) *Size {
	return &Size{width: width, height: height}
}

func (s *Size) Width() int {
	return s.width
}

func (s *Size) Height() int {
	return s.height
}

// Scale returns the size scaled to the target width and height. A target of
// -1 is computed from the other target, keeping the aspect ratio. The
// receiver is returned when neither target is positive or the size does not
// change.
func (s *Size) Scale(width int, height int) *Size {
	if width > 0 || height > 0 {
		targetWidth := width
		targetHeight := height

		if targetWidth == -1 {
			targetWidth = int(float64(s.Width()) * (float64(targetHeight) / float64(s.Height())))
		}

		if targetHeight == -1 {
			targetHeight = int(float64(s.Height()) * (float64(targetWidth) / float64(s.Width())))
		}

		if s.Width() != targetWidth || s.Height() != targetHeight {
			return NewSize(targetWidth, targetHeight)
		}
	}
	return s
}

// Equals ports the record's generated equals.
func (s *Size) Equals(other *Size) bool {
	if other == nil {
		return false
	}
	return s.width == other.width && s.height == other.height
}

// String ports the record's generated toString.
func (s *Size) String() string {
	return fmt.Sprintf("Size[width=%d, height=%d]", s.width, s.height)
}

func (s *Size) ToString() string {
	return s.String()
}
