package ufo

import (
	"testing"

	"github.com/octoberswimmer/ufo/geom"
)

// boxRangeHelperTestDevice implements the three clip methods BoxRangeHelper
// calls; every other OutputDevice method panics on the nil embedded interface.
type boxRangeHelperTestDevice struct {
	OutputDevice
	clip geom.Shape
	log  []string
}

func (d *boxRangeHelperTestDevice) GetClip() geom.Shape {
	return d.clip
}

func (d *boxRangeHelperTestDevice) SetClip(s geom.Shape) {
	d.clip = s
	d.log = append(d.log, "setClip")
}

func (d *boxRangeHelperTestDevice) Clip(s geom.Shape) {
	if d.clip == nil {
		d.clip = s
	} else {
		d.clip = d.clip.GetBounds().Intersection(s.GetBounds())
	}
	d.log = append(d.log, "clip")
}

// boxRangeHelperTestBox implements the one BlockBoxI method BoxRangeHelper
// calls.
type boxRangeHelperTestBox struct {
	BlockBoxI
	edge *geom.Rectangle
}

func (b *boxRangeHelperTestBox) GetChildrenClipEdge(c *RenderingContext) *geom.Rectangle {
	return b.edge
}

func TestBoxRangeHelper_pushAndPopClipRegions(t *testing.T) {
	outer := &boxRangeHelperTestBox{edge: geom.NewRectangle(0, 0, 100, 100)}
	inner := &boxRangeHelperTestBox{edge: geom.NewRectangle(10, 10, 20, 20)}
	// The outer block clips content 0 to 3, the inner block content 1 to 2.
	ranges := []*BoxRangeData{
		NewBoxRangeData(outer, NewBoxRange(0, 3)),
		NewBoxRangeData(inner, NewBoxRange(1, 2)),
	}
	device := &boxRangeHelperTestDevice{}
	helper := NewBoxRangeHelper(device, ranges)

	helper.PopClipRegions(0)
	helper.PushClipRegion(nil, 0)
	if got := device.clip.GetBounds(); !got.Equals(geom.NewRectangle(0, 0, 100, 100)) {
		t.Fatalf("clip after content 0 = %v, want the outer edge", got)
	}
	if ranges[0].GetClip() != nil {
		t.Errorf("saved clip of the outer range = %v, want nil", ranges[0].GetClip())
	}

	helper.PopClipRegions(1)
	helper.PushClipRegion(nil, 1)
	if got := device.clip.GetBounds(); !got.Equals(geom.NewRectangle(10, 10, 20, 20)) {
		t.Fatalf("clip after content 1 = %v, want the inner edge", got)
	}

	helper.PopClipRegions(2)
	if got := device.clip.GetBounds(); !got.Equals(geom.NewRectangle(0, 0, 100, 100)) {
		t.Fatalf("clip after popping at 2 = %v, want the outer edge", got)
	}
	helper.PushClipRegion(nil, 2)

	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("CheckFinished did not panic while a clip region was pushed")
			}
		}()
		helper.CheckFinished()
	}()

	helper.PopClipRegions(3)
	if device.clip != nil {
		t.Errorf("clip after popping at 3 = %v, want nil", device.clip)
	}
	helper.CheckFinished()

	want := []string{"clip", "clip", "setClip", "setClip"}
	if len(device.log) != len(want) {
		t.Fatalf("device calls = %v, want %v", device.log, want)
	}
	for i := range want {
		if device.log[i] != want[i] {
			t.Fatalf("device calls = %v, want %v", device.log, want)
		}
	}
}

func TestBoxRangeHelper_emptyRangeList(t *testing.T) {
	device := &boxRangeHelperTestDevice{}
	helper := NewBoxRangeHelper(device, nil)
	helper.PopClipRegions(0)
	helper.PushClipRegion(nil, 0)
	helper.PopClipRegions(1)
	helper.CheckFinished()
	if len(device.log) != 0 {
		t.Errorf("device calls = %v, want none", device.log)
	}
}
