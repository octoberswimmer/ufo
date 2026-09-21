// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BlockFormattingContext.java

package ufo

import (
	"strconv"

	"github.com/octoberswimmer/ufo/geom"
)

// BlockFormattingContext represents a block formatting context as defined in
// the CSS spec. Its main purpose is to provide BFC relative coordinates for a
// FloatManager. This coordinate space is used when positioning floats and
// calculating the amount of space floated boxes take up at a given y position.
//
// NOTE: The Translate method must be called when a block box in the normal
// flow is moved (i.e. its static position changes)
type BlockFormattingContext struct {
	x int
	y int

	persistentBFC *PersistentBFC
}

func NewBlockFormattingContext(block BlockBoxI, c *LayoutContext) *BlockFormattingContext {
	return &BlockFormattingContext{
		persistentBFC: NewPersistentBFC(block, c),
	}
}

func (b *BlockFormattingContext) GetOffset() *geom.Point {
	return geom.NewPoint(b.x, b.y)
}

func (b *BlockFormattingContext) Translate(x int, y int) {
	b.x -= x
	b.y -= y
}

func (b *BlockFormattingContext) GetFloatManager() *FloatManager {
	return b.persistentBFC.GetFloatManager()
}

func (b *BlockFormattingContext) GetLeftFloatDistance(cssCtx CssContext, line *LineBox, containingBlockWidth int) int {
	return b.GetFloatManager().GetLeftFloatDistance(cssCtx, b, line, containingBlockWidth)
}

func (b *BlockFormattingContext) GetRightFloatDistance(cssCtx CssContext, line *LineBox, containingBlockWidth int) int {
	return b.GetFloatManager().GetRightFloatDistance(cssCtx, b, line, containingBlockWidth)
}

func (b *BlockFormattingContext) GetFloatDistance(cssCtx CssContext, line *LineBox, containingBlockWidth int) int {
	return b.GetLeftFloatDistance(cssCtx, line, containingBlockWidth) +
		b.GetRightFloatDistance(cssCtx, line, containingBlockWidth)
}

func (b *BlockFormattingContext) GetNextLineBoxDelta(cssCtx CssContext, line *LineBox, containingBlockWidth int) int {
	return b.GetFloatManager().GetNextLineBoxDelta(cssCtx, b, line, containingBlockWidth)
}

func (b *BlockFormattingContext) FloatBox(c *LayoutContext, floated BlockBoxI) {
	b.GetFloatManager().FloatBox(c, c.GetLayer(), b, floated)
}

func (b *BlockFormattingContext) Clear(c *LayoutContext, current BoxI) {
	b.GetFloatManager().Clear(c, b, current)
}

func (b *BlockFormattingContext) String() string {
	return "BlockFormattingContext: (" + strconv.Itoa(b.x) + "," + strconv.Itoa(b.y) + ")"
}

func (b *BlockFormattingContext) ToString() string {
	return b.String()
}
