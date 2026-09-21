// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/BackgroundPosition.java

package ufo

type BackgroundPosition struct {
	horizontal *PropertyValue
	vertical   *PropertyValue
}

func NewBackgroundPosition(horizontal *PropertyValue, vertical *PropertyValue) *BackgroundPosition {
	return &BackgroundPosition{horizontal: horizontal, vertical: vertical}
}

func (b *BackgroundPosition) GetHorizontal() *PropertyValue {
	return b.horizontal
}

func (b *BackgroundPosition) GetVertical() *PropertyValue {
	return b.vertical
}
