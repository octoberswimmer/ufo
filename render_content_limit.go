// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/ContentLimit.java

package ufo

import "strconv"

const ContentLimitUndefined = -1

type ContentLimit struct {
	top    int
	bottom int
}

func NewContentLimit() *ContentLimit {
	return &ContentLimit{top: ContentLimitUndefined, bottom: ContentLimitUndefined}
}

func (l *ContentLimit) GetTop() int {
	return l.top
}

func (l *ContentLimit) UpdateTop(top int) {
	if l.top == ContentLimitUndefined || top < l.top {
		l.top = top
	}
}

func (l *ContentLimit) GetBottom() int {
	return l.bottom
}

func (l *ContentLimit) UpdateBottom(bottom int) {
	if l.bottom == ContentLimitUndefined || bottom > l.bottom {
		l.bottom = bottom
	}
}

func (l *ContentLimit) ToString() string {
	return "[top=" + strconv.Itoa(l.top) + ", bottom=" + strconv.Itoa(l.bottom) + "]"
}

func (l *ContentLimit) String() string {
	return l.ToString()
}
