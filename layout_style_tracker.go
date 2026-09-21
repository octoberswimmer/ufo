// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/StyleTracker.java

package ufo

// StyleTracker is a managed list of CalculatedStyle objects.  It is used when
// keeping track of the styles which apply to a :first-line or :first-letter
// pseudo-element.
type StyleTracker struct {
	styles []*CascadedStyle
}

func NewStyleTracker() *StyleTracker {
	return &StyleTracker{}
}

func (s *StyleTracker) AddStyle(style *CascadedStyle) {
	s.styles = append(s.styles, style)
}

func (s *StyleTracker) RemoveLast() {
	if len(s.styles) != 0 {
		s.styles = s.styles[:len(s.styles)-1]
	}
}

func (s *StyleTracker) HasStyles() bool {
	return len(s.styles) != 0
}

func (s *StyleTracker) ClearStyles() {
	s.styles = nil
}

func (s *StyleTracker) DeriveAll(start CalculatedStyleI) CalculatedStyleI {
	result := start
	for _, o := range s.GetStyles() {
		result = result.DeriveStyle(o)
	}
	return result
}

func (s *StyleTracker) GetStyles() []*CascadedStyle {
	return s.styles
}

func (s *StyleTracker) CopyOf() *StyleTracker {
	result := NewStyleTracker()
	result.styles = append(result.styles, s.styles...)
	return result
}
