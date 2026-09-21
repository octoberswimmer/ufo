// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/FloatManager.java

package ufo

import "github.com/octoberswimmer/ufo/geom"

// FloatManagerDirection is FloatManager.Direction.
type FloatManagerDirection int

const (
	FloatManagerDirectionLeft FloatManagerDirection = iota
	FloatManagerDirectionRight
)

// FloatManager manages all floated boxes in a given block formatting context.
// It is responsible for positioning floats and calculating clearance for
// non-floated (block) boxes.
type FloatManager struct {
	leftFloats  []*floatManagerBoxOffset
	rightFloats []*floatManagerBoxOffset
	master      BoxI
}

func NewFloatManager(master BoxI) *FloatManager {
	return &FloatManager{master: master}
}

func (f *FloatManager) FloatBox(c *LayoutContext, layer *Layer, bfc *BlockFormattingContext, box BlockBoxI) {
	if box.GetStyle().IsFloatedLeft() {
		f.position(c, bfc, box, FloatManagerDirectionLeft)
		f.save(box, layer, bfc, FloatManagerDirectionLeft)
	} else if box.GetStyle().IsFloatedRight() {
		f.position(c, bfc, box, FloatManagerDirectionRight)
		f.save(box, layer, bfc, FloatManagerDirectionRight)
	}
}

func (f *FloatManager) Clear(cssCtx CssContext, bfc *BlockFormattingContext, box BoxI) {
	if box.GetStyle().IsClearLeft() {
		f.moveClear(cssCtx, bfc, box, f.getFloats(FloatManagerDirectionLeft))
	}
	if box.GetStyle().IsClearRight() {
		f.moveClear(cssCtx, bfc, box, f.getFloats(FloatManagerDirectionRight))
	}
}

func (f *FloatManager) save(current BlockBoxI, layer *Layer, bfc *BlockFormattingContext, direction FloatManagerDirection) {
	p := bfc.GetOffset()
	f.addFloat(direction, &floatManagerBoxOffset{box: current, x: p.X, y: p.Y})
	layer.AddFloat(current)
	current.GetFloatedBoxData().SetManager(f)

	current.CalcCanvasLocation()
	current.CalcChildLocations()
}

func (f *FloatManager) position(cssCtx CssContext, bfc *BlockFormattingContext,
	current BlockBoxI, direction FloatManagerDirection) {
	f.moveAllTheWayOver(current, direction)

	f.alignToLastOpposingFloat(cssCtx, bfc, current, direction)
	f.alignToLastFloat(cssCtx, bfc, current, direction)

	if !f.fitsInContainingBlock(current) ||
		f.overlaps(cssCtx, bfc, current, f.getFloats(direction)) {
		f.moveAllTheWayOver(current, direction)
		f.moveFloatBelow(cssCtx, bfc, current, f.getFloats(direction))
	}

	if f.overlaps(cssCtx, bfc, current, f.getOpposingFloats(direction)) {
		f.moveAllTheWayOver(current, direction)
		f.moveFloatBelow(cssCtx, bfc, current, f.getFloats(direction))
		f.moveFloatBelow(cssCtx, bfc, current, f.getOpposingFloats(direction))
	}

	if current.GetStyle().IsCleared() {
		if current.GetStyle().IsClearLeft() && direction == FloatManagerDirectionLeft {
			f.moveAllTheWayOver(current, FloatManagerDirectionLeft)
		} else if current.GetStyle().IsClearRight() && direction == FloatManagerDirectionRight {
			f.moveAllTheWayOver(current, FloatManagerDirectionRight)
		}
		f.moveFloatBelow(cssCtx, bfc, current, f.getFloats(direction))
	}
}

func (f *FloatManager) getFloats(direction FloatManagerDirection) []*floatManagerBoxOffset {
	if direction == FloatManagerDirectionLeft {
		return f.leftFloats
	}
	return f.rightFloats
}

// addFloat is getFloats(direction).add(boxOffset): Java appends to the list
// that getFloats returns, which a Go slice value does not allow.
func (f *FloatManager) addFloat(direction FloatManagerDirection, boxOffset *floatManagerBoxOffset) {
	if direction == FloatManagerDirectionLeft {
		f.leftFloats = append(f.leftFloats, boxOffset)
	} else {
		f.rightFloats = append(f.rightFloats, boxOffset)
	}
}

func (f *FloatManager) getOpposingFloats(direction FloatManagerDirection) []*floatManagerBoxOffset {
	if direction == FloatManagerDirectionLeft {
		return f.rightFloats
	}
	return f.leftFloats
}

func (f *FloatManager) alignToLastFloat(cssCtx CssContext,
	bfc *BlockFormattingContext, current BlockBoxI, direction FloatManagerDirection) {

	floats := f.getFloats(direction)
	if len(floats) != 0 {
		offset := bfc.GetOffset()
		lastOffset := floats[len(floats)-1]
		last := lastOffset.box

		currentBounds := current.GetMarginEdge(cssCtx, -offset.X, -offset.Y)

		lastBounds := last.GetMarginEdge(cssCtx, -lastOffset.x, -lastOffset.y)

		moveOver := false

		if currentBounds.Y < lastBounds.Y {
			currentBounds.Translate(0, lastBounds.Y-currentBounds.Y)
			moveOver = true
		}

		if currentBounds.Y >= lastBounds.Y && currentBounds.Y < lastBounds.Y+lastBounds.Height {
			moveOver = true
		}

		if moveOver {
			if direction == FloatManagerDirectionLeft {
				currentBounds.X = lastBounds.X + last.GetWidth()
			} else if direction == FloatManagerDirectionRight {
				currentBounds.X = lastBounds.X - current.GetWidth()
			}

			currentBounds.Translate(offset.X, offset.Y)

			current.SetX(currentBounds.X)
			current.SetY(currentBounds.Y)
		}
	}
}

func (f *FloatManager) alignToLastOpposingFloat(cssCtx CssContext,
	bfc *BlockFormattingContext, current BlockBoxI, direction FloatManagerDirection) {

	floats := f.getOpposingFloats(direction)
	if len(floats) != 0 {
		offset := bfc.GetOffset()
		lastOffset := floats[len(floats)-1]

		currentBounds := current.GetMarginEdge(cssCtx, -offset.X, -offset.Y)

		lastBounds := lastOffset.box.GetMarginEdge(cssCtx,
			-lastOffset.x, -lastOffset.y)

		if currentBounds.Y < lastBounds.Y {
			currentBounds.Translate(0, lastBounds.Y-currentBounds.Y)

			currentBounds.Translate(offset.X, offset.Y)

			current.SetY(currentBounds.Y)
		}
	}
}

func (f *FloatManager) moveAllTheWayOver(current BlockBoxI, direction FloatManagerDirection) {
	if direction == FloatManagerDirectionLeft {
		current.SetX(0)
	} else if direction == FloatManagerDirectionRight {
		current.SetX(current.GetContainingBlock().GetContentWidth() - current.GetWidth())
	}
}

func (f *FloatManager) fitsInContainingBlock(current BlockBoxI) bool {
	return current.GetX() >= 0 &&
		current.GetX()+current.GetWidth() <= current.GetContainingBlock().GetContentWidth()
}

func (f *FloatManager) findLowestY(cssCtx CssContext, floats []*floatManagerBoxOffset) int {
	result := 0

	for _, floater := range floats {
		bounds := floater.box.GetMarginEdge(
			cssCtx, -floater.x, -floater.y)
		if bounds.Y+bounds.Height > result {
			result = bounds.Y + bounds.Height
		}
	}

	return result
}

func (f *FloatManager) GetClearDelta(cssCtx CssContext, bfcRelativeY int) int {
	lowestLeftY := f.findLowestY(cssCtx, f.getFloats(FloatManagerDirectionLeft))
	lowestRightY := f.findLowestY(cssCtx, f.getFloats(FloatManagerDirectionRight))

	lowestY := max(lowestLeftY, lowestRightY)

	return lowestY - bfcRelativeY
}

func (f *FloatManager) overlaps(cssCtx CssContext, bfc *BlockFormattingContext,
	current BlockBoxI, floats []*floatManagerBoxOffset) bool {
	offset := bfc.GetOffset()
	bounds := current.GetMarginEdge(cssCtx, -offset.X, -offset.Y)

	for _, floater := range floats {
		floaterBounds := floater.box.GetMarginEdge(cssCtx,
			-floater.x, -floater.y)

		if floaterBounds.Intersects(bounds) {
			return true
		}
	}

	return false
}

func (f *FloatManager) moveFloatBelow(cssCtx CssContext, bfc *BlockFormattingContext,
	current BoxI, floats []*floatManagerBoxOffset) {
	if len(floats) == 0 {
		return
	}

	offset := bfc.GetOffset()
	boxY := current.GetY() - offset.Y
	floatY := f.findLowestY(cssCtx, floats)

	if floatY-boxY > 0 {
		current.SetY(current.GetY() + floatY - boxY)
	}
}

func (f *FloatManager) moveClear(cssCtx CssContext, bfc *BlockFormattingContext,
	current BoxI, floats []*floatManagerBoxOffset) {
	if len(floats) == 0 {
		return
	}

	// Translate from box coords to BFC coords
	offset := bfc.GetOffset()
	bounds := current.GetBorderEdge(
		current.GetX()-offset.X, current.GetY()-offset.Y, cssCtx)

	y := f.findLowestY(cssCtx, floats)

	if bounds.Y < y {
		// Translate bottom margin edge of lowest float back to box coords
		// and set the box's border edge to that value
		bounds.Y = y

		bounds.Translate(offset.X, offset.Y)

		current.SetY(bounds.Y - calculatedStyleFloatToInt(current.GetMargin(cssCtx).Top()))
	}
}

func (f *FloatManager) RemoveFloat(floater BlockBoxI) {
	f.leftFloats = f.removeFloatWithFloats(floater, f.getFloats(FloatManagerDirectionLeft))
	f.rightFloats = f.removeFloatWithFloats(floater, f.getFloats(FloatManagerDirectionRight))
}

// removeFloatWithFloats is removeFloat(BlockBox, List). It returns the list
// without the floater (Java removes through the iterator). Java's
// Box.equals is object identity, which is the identity of the root Box struct
// here.
func (f *FloatManager) removeFloatWithFloats(floater BlockBoxI, floats []*floatManagerBoxOffset) []*floatManagerBoxOffset {
	var result []*floatManagerBoxOffset
	for _, boxOffset := range floats {
		if boxOffset.box.AsBox() == floater.AsBox() {
			floater.GetFloatedBoxData().SetManager(nil)
		} else {
			result = append(result, boxOffset)
		}
	}
	return result
}

func (f *FloatManager) CalcFloatLocations() {
	f.calcFloatLocationsWithFloats(f.getFloats(FloatManagerDirectionLeft))
	f.calcFloatLocationsWithFloats(f.getFloats(FloatManagerDirectionRight))
}

func (f *FloatManager) calcFloatLocationsWithFloats(floats []*floatManagerBoxOffset) {
	for _, boxOffset := range floats {
		boxOffset.box.CalcCanvasLocation()
		boxOffset.box.CalcChildLocations()
	}
}

func (f *FloatManager) applyLineHeightHack(cssCtx CssContext, line BoxI, bounds *geom.Rectangle) {
	// this is a hack to deal with lines w/o width or height. is this valid?
	// possibly, since the line doesn't know how long it should be until it's already
	// done float adjustments
	if line.GetHeight() == 0 {
		bounds.Height = calculatedStyleFloatToInt(line.GetStyle().GetLineHeight(cssCtx))
	}
}

func (f *FloatManager) GetNextLineBoxDelta(cssCtx CssContext, bfc *BlockFormattingContext,
	line *LineBox, containingBlockContentWidth int) int {
	left := f.getFloatDistance(cssCtx, bfc, line, containingBlockContentWidth, f.leftFloats, FloatManagerDirectionLeft)
	right := f.getFloatDistance(cssCtx, bfc, line, containingBlockContentWidth, f.rightFloats, FloatManagerDirectionRight)

	var leftDelta int
	var rightDelta int

	if left.box != nil {
		leftDelta = f.calcDelta(cssCtx, line, left)
	} else {
		leftDelta = 0
	}

	if right.box != nil {
		rightDelta = f.calcDelta(cssCtx, line, right)
	} else {
		rightDelta = 0
	}

	return max(leftDelta, rightDelta)
}

func (f *FloatManager) calcDelta(cssCtx CssContext, line *LineBox, boxDistance *floatManagerBoxDistance) int {
	floated := boxDistance.box
	rect := floated.GetBorderEdge(floated.GetAbsX(), floated.GetAbsY(), cssCtx)
	bottom := rect.Y + rect.Height
	return bottom - line.GetAbsY()
}

func (f *FloatManager) GetLeftFloatDistance(cssCtx CssContext, bfc *BlockFormattingContext,
	line *LineBox, containingBlockContentWidth int) int {
	return f.getFloatDistance(cssCtx, bfc, line, containingBlockContentWidth, f.leftFloats, FloatManagerDirectionLeft).distance
}

func (f *FloatManager) GetRightFloatDistance(cssCtx CssContext, bfc *BlockFormattingContext,
	line *LineBox, containingBlockContentWidth int) int {
	return f.getFloatDistance(cssCtx, bfc, line, containingBlockContentWidth, f.rightFloats, FloatManagerDirectionRight).distance
}

func (f *FloatManager) getFloatDistance(cssCtx CssContext, bfc *BlockFormattingContext,
	line *LineBox, containingBlockContentWidth int,
	floatsList []*floatManagerBoxOffset, direction FloatManagerDirection) *floatManagerBoxDistance {
	if len(floatsList) == 0 {
		return &floatManagerBoxDistance{box: nil, distance: 0}
	}

	offset := bfc.GetOffset()
	lineBounds := line.GetMarginEdge(cssCtx, -offset.X, -offset.Y)
	lineBounds.Width = containingBlockContentWidth

	var farthestOver int
	if direction == FloatManagerDirectionLeft {
		farthestOver = lineBounds.X
	} else {
		farthestOver = lineBounds.X + lineBounds.Width
	}

	f.applyLineHeightHack(cssCtx, line, lineBounds)
	var farthestOverBox BlockBoxI
	for _, floater := range floatsList {
		fr := floater.box.GetMarginEdge(cssCtx, -floater.x, -floater.y)
		if lineBounds.Intersects(fr) {
			if direction == FloatManagerDirectionLeft && fr.X+fr.Width > farthestOver {
				farthestOver = fr.X + fr.Width
			} else if direction == FloatManagerDirectionRight && fr.X < farthestOver {
				farthestOver = fr.X
			}
			farthestOverBox = floater.box
		}
	}

	if direction == FloatManagerDirectionLeft {
		return &floatManagerBoxDistance{box: farthestOverBox, distance: farthestOver - lineBounds.X}
	} else {
		return &floatManagerBoxDistance{box: farthestOverBox, distance: lineBounds.X + lineBounds.Width - farthestOver}
	}
}

func (f *FloatManager) GetMaster() BoxI {
	return f.master
}

// GetOffset returns nil when the floater is not managed by this manager.
func (f *FloatManager) GetOffset(floater BlockBoxI) *geom.Point {
	// FIXME inefficient (but probably doesn't matter)
	if floater.GetStyle().IsFloatedLeft() {
		return f.getOffsetWithFloats(floater, f.getFloats(FloatManagerDirectionLeft))
	}
	return f.getOffsetWithFloats(floater, f.getFloats(FloatManagerDirectionRight))
}

func (f *FloatManager) getOffsetWithFloats(floater BlockBoxI, floats []*floatManagerBoxOffset) *geom.Point {
	for _, boxOffset := range floats {
		box := boxOffset.box

		if box.AsBox() == floater.AsBox() {
			return geom.NewPoint(boxOffset.x, boxOffset.y)
		}
	}

	return nil
}

func (f *FloatManager) performFloatOperationWithFloats(op FloatManagerFloatOperation, floats []*floatManagerBoxOffset) {
	for _, boxOffset := range floats {
		box := boxOffset.box

		box.SetAbsX(box.GetX() + f.GetMaster().GetAbsX() - boxOffset.x)
		box.SetAbsY(box.GetY() + f.GetMaster().GetAbsY() - boxOffset.y)

		op.Operate(box)
	}
}

func (f *FloatManager) PerformFloatOperation(op FloatManagerFloatOperation) {
	f.performFloatOperationWithFloats(op, f.getFloats(FloatManagerDirectionLeft))
	f.performFloatOperationWithFloats(op, f.getFloats(FloatManagerDirectionRight))
}

// floatManagerBoxOffset is the record FloatManager.BoxOffset.
type floatManagerBoxOffset struct {
	box BlockBoxI
	x   int
	y   int
}

// floatManagerBoxDistance is the record FloatManager.BoxDistance; box may be
// nil.
type floatManagerBoxDistance struct {
	box      BlockBoxI
	distance int
}

// FloatManagerFloatOperation is FloatManager.FloatOperation.
type FloatManagerFloatOperation interface {
	Operate(floater BoxI)
}

// FloatManagerFloatOperationFunc adapts a function to
// FloatManagerFloatOperation, for the places where Java passes a lambda.
type FloatManagerFloatOperationFunc func(floater BoxI)

func (f FloatManagerFloatOperationFunc) Operate(floater BoxI) {
	f(floater)
}
