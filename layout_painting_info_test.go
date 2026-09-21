package ufo

import (
	"testing"

	"github.com/octoberswimmer/ufo/geom"
)

func TestPaintingInfo_translate(t *testing.T) {
	info := NewPaintingInfo(geom.NewDimension(100, 200), geom.NewRectangle(10, 20, 30, 40))
	info.Translate(5, -7)

	if got := info.GetAggregateBounds(); got.X != 15 || got.Y != 13 || got.Width != 30 || got.Height != 40 {
		t.Errorf("aggregate bounds = %v, want x=15 y=13 width=30 height=40", got)
	}
	if got := info.GetOuterMarginCorner(); got.Width != 105 || got.Height != 193 {
		t.Errorf("outer margin corner = %v, want width=105 height=193", got)
	}
}

func TestPaintingInfo_copyOfIsIndependent(t *testing.T) {
	info := NewPaintingInfo(geom.NewDimension(100, 200), geom.NewRectangle(10, 20, 30, 40))
	copied := info.CopyOf()
	copied.Translate(1, 1)
	copied.GetOuterMarginCorner().Width = 999

	if got := info.GetAggregateBounds(); got.X != 10 || got.Y != 20 {
		t.Errorf("original aggregate bounds changed to %v", got)
	}
	if got := info.GetOuterMarginCorner(); got.Width != 100 || got.Height != 200 {
		t.Errorf("original outer margin corner changed to %v", got)
	}
	if got := copied.GetAggregateBounds(); got.X != 11 || got.Y != 21 {
		t.Errorf("copied aggregate bounds = %v, want x=11 y=21", got)
	}
}
