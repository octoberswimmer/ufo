// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/BackgroundSize.java

package ufo

type BackgroundSize struct {
	contain  bool
	cover    bool
	bothAuto bool

	width  *PropertyValue
	height *PropertyValue
}

func NewBackgroundSizeWithContainCoverBothAuto(contain bool, cover bool, bothAuto bool) *BackgroundSize {
	return &BackgroundSize{contain: contain, cover: cover, bothAuto: bothAuto}
}

func NewBackgroundSize(width *PropertyValue, height *PropertyValue) *BackgroundSize {
	return &BackgroundSize{width: width, height: height}
}

func (b *BackgroundSize) IsContain() bool {
	return b.contain
}

func (b *BackgroundSize) IsCover() bool {
	return b.cover
}

func (b *BackgroundSize) IsBothAuto() bool {
	return b.bothAuto
}

func (b *BackgroundSize) GetWidth() *PropertyValue {
	return b.width
}

func (b *BackgroundSize) GetHeight() *PropertyValue {
	return b.height
}
