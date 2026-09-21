// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/CharCounts.java

package ufo

// CharCounts holds the number of space and non-space characters of a line
// which is being justified.
type CharCounts struct {
	spaceCount    int
	nonSpaceCount int
}

func NewCharCounts() *CharCounts {
	return &CharCounts{}
}

func (c *CharCounts) GetSpaceCount() int {
	return c.spaceCount
}

func (c *CharCounts) SetSpaceCount(spaceCount int) {
	c.spaceCount = spaceCount
}

func (c *CharCounts) GetNonSpaceCount() int {
	return c.nonSpaceCount
}

func (c *CharCounts) SetNonSpaceCount(nonSpaceCount int) {
	c.nonSpaceCount = nonSpaceCount
}
