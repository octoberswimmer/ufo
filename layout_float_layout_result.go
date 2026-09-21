// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/FloatLayoutResult.java

package ufo

// FloatLayoutResult is a bean containing the result of laying out a floated
// block.  If the floated block can't fit on the current line, it will be
// marked pending with the result that it will be laid out again once the line
// has been saved.
//
// FIXME: This class can go away
type FloatLayoutResult struct {
	pending bool
	block   BlockBoxI
}

func NewFloatLayoutResult(pending bool, block BlockBoxI) *FloatLayoutResult {
	return &FloatLayoutResult{pending: pending, block: block}
}

func (f *FloatLayoutResult) IsPending() bool {
	return f.pending
}

func (f *FloatLayoutResult) GetBlock() BlockBoxI {
	return f.block
}
