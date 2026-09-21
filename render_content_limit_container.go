// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/ContentLimitContainer.java

package ufo

import (
	"strconv"
	"strings"
)

type ContentLimitContainer struct {
	parent        *ContentLimitContainer
	initialPageNo int
	contentLimits []*ContentLimit

	lastPage *PageBox
}

// NewContentLimitContainer creates a container. parent may be nil.
func NewContentLimitContainer(parent *ContentLimitContainer, c *LayoutContext, startAbsY int) *ContentLimitContainer {
	l := &ContentLimitContainer{}
	// As in Java, the parent is assigned after the page lookup, so this lookup
	// reads and writes the last page of the new container, not of the
	// outermost ancestor.
	l.initialPageNo = l.GetPage(c, startAbsY).GetPageNo()
	l.parent = parent
	return l
}

func (l *ContentLimitContainer) GetInitialPageNo() int {
	return l.initialPageNo
}

func (l *ContentLimitContainer) GetLastPageNo() int {
	return l.initialPageNo + len(l.contentLimits) - 1
}

// GetContentLimit may return nil.
func (l *ContentLimitContainer) GetContentLimit(pageNo int) *ContentLimit {
	return l.getContentLimitWithAddAsNeeded(pageNo, false)
}

func (l *ContentLimitContainer) getContentLimitWithAddAsNeeded(pageNo int, addAsNeeded bool) *ContentLimit {
	if addAsNeeded {
		for len(l.contentLimits) < pageNo-l.initialPageNo+1 {
			l.contentLimits = append(l.contentLimits, NewContentLimit())
		}
	}

	target := pageNo - l.initialPageNo
	if target >= 0 && target < len(l.contentLimits) {
		return l.contentLimits[pageNo-l.initialPageNo]
	} else {
		return nil
	}
}

func (l *ContentLimitContainer) UpdateTop(c *LayoutContext, absY int) {
	page := l.GetPage(c, absY)

	l.getContentLimitWithAddAsNeeded(page.GetPageNo(), true).UpdateTop(absY)

	parent := l.GetParent()
	if parent != nil {
		parent.UpdateTop(c, absY)
	}
}

func (l *ContentLimitContainer) UpdateBottom(c *LayoutContext, absY int) {
	page := l.GetPage(c, absY)

	l.getContentLimitWithAddAsNeeded(page.GetPageNo(), true).UpdateBottom(absY)

	parent := l.GetParent()
	if parent != nil {
		parent.UpdateBottom(c, absY)
	}
}

func (l *ContentLimitContainer) GetPage(c *LayoutContext, absY int) *PageBox {
	var page *PageBox
	last := l.getLastPage()
	if last != nil && absY >= last.GetTop() && absY < last.GetBottom() {
		page = last
	} else {
		page = c.GetRootLayer().GetPage(c, absY)
		l.setLastPage(page)
	}
	return page
}

func (l *ContentLimitContainer) getLastPage() *PageBox {
	c := l
	for c.GetParent() != nil {
		c = c.GetParent()
	}
	return c.lastPage
}

func (l *ContentLimitContainer) setLastPage(page *PageBox) {
	c := l
	for c.GetParent() != nil {
		c = c.GetParent()
	}
	c.lastPage = page
}

// GetParent may return nil.
func (l *ContentLimitContainer) GetParent() *ContentLimitContainer {
	return l.parent
}

func (l *ContentLimitContainer) IsContainsMultiplePages() bool {
	return len(l.contentLimits) > 1
}

func (l *ContentLimitContainer) ToString() string {
	limits := make([]string, len(l.contentLimits))
	for i, limit := range l.contentLimits {
		limits[i] = limit.ToString()
	}
	return "[initialPageNo=" + strconv.Itoa(l.initialPageNo) + ", limits=[" + strings.Join(limits, ", ") + "]]"
}

func (l *ContentLimitContainer) String() string {
	return l.ToString()
}
