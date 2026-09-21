// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BreakAtLineContext.java

package ufo

type BreakAtLineContext struct {
	block BlockBoxI
	line  int
}

func NewBreakAtLineContext(block BlockBoxI, line int) *BreakAtLineContext {
	return &BreakAtLineContext{block: block, line: line}
}

func (b *BreakAtLineContext) GetBlock() BlockBoxI {
	return b.block
}

func (b *BreakAtLineContext) GetLine() int {
	return b.line
}
