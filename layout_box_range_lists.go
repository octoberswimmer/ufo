// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BoxRangeLists.java

package ufo

// BoxRangeLists holds the clip ranges of the block content list and of the
// inline content list that BoxCollector fills. Java callers insert into the
// lists that getBlock() and getInline() return; a Go slice returned by value
// cannot be grown by the caller, so BoxCollector inserts through
// insertBlock and insertInline instead.
type BoxRangeLists struct {
	block  []*BoxRangeData
	inline []*BoxRangeData
}

func NewBoxRangeLists() *BoxRangeLists {
	return &BoxRangeLists{}
}

func (l *BoxRangeLists) GetBlock() []*BoxRangeData {
	return l.block
}

func (l *BoxRangeLists) GetInline() []*BoxRangeData {
	return l.inline
}

// insertBlock is getBlock().add(index, data).
func (l *BoxRangeLists) insertBlock(index int, data *BoxRangeData) {
	l.block = boxRangeListsInsert(l.block, index, data)
}

// insertInline is getInline().add(index, data).
func (l *BoxRangeLists) insertInline(index int, data *BoxRangeData) {
	l.inline = boxRangeListsInsert(l.inline, index, data)
}

func boxRangeListsInsert(list []*BoxRangeData, index int, data *BoxRangeData) []*BoxRangeData {
	if index < 0 || index > len(list) {
		panic(NewXRRuntimeException("Index out of range"))
	}
	list = append(list, nil)
	copy(list[index+1:], list[index:])
	list[index] = data
	return list
}
