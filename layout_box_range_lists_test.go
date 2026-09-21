package ufo

import "testing"

func TestBoxRangeLists_insertKeepsListOrder(t *testing.T) {
	lists := NewBoxRangeLists()
	if len(lists.GetBlock()) != 0 || len(lists.GetInline()) != 0 {
		t.Fatalf("new lists are not empty")
	}

	a := NewBoxRangeData(nil, NewBoxRange(0, 1))
	b := NewBoxRangeData(nil, NewBoxRange(1, 2))
	c := NewBoxRangeData(nil, NewBoxRange(2, 3))

	// BoxCollector inserts the range of an enclosing block before the ranges
	// of the blocks inside it, which were added first.
	lists.insertBlock(0, b)
	lists.insertBlock(1, c)
	lists.insertBlock(0, a)

	got := lists.GetBlock()
	if len(got) != 3 || got[0] != a || got[1] != b || got[2] != c {
		t.Errorf("block list = %v, want [a b c]", got)
	}

	lists.insertInline(0, c)
	lists.insertInline(0, a)
	lists.insertInline(1, b)
	got = lists.GetInline()
	if len(got) != 3 || got[0] != a || got[1] != b || got[2] != c {
		t.Errorf("inline list = %v, want [a b c]", got)
	}
}

func TestBoxRangeLists_insertOutOfRangePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("insertBlock(1, ...) into an empty list did not panic")
		}
	}()
	NewBoxRangeLists().insertBlock(1, NewBoxRangeData(nil, NewBoxRange(0, 1)))
}
