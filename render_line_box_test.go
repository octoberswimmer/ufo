package ufo

import (
	"math"
	"testing"
)

// Flying Saucer has no JUnit tests for LineBox. The expected values were
// produced by calling the private static LineBox.justificationInfo of the
// Java class over the same inputs. Each row is space count, non-space count,
// width to add, and the float bits of the non-space share, the space share,
// the resulting non-space adjustment and the resulting space adjustment.
func TestLineBoxJustificationInfoAgainstJava(t *testing.T) {
	cases := [][7]uint32{
		{0, 0, 10, 1065353216, 0, 0, 0},
		{0, 0, 10, 0, 1065353216, 0, 0},
		{0, 0, 10, 1045220557, 1061997773, 0, 0},
		{0, 1, 10, 1065353216, 0, 0, 0},
		{0, 1, 10, 0, 1065353216, 0, 0},
		{0, 1, 10, 1045220557, 1061997773, 0, 0},
		{0, 5, 13, 1065353216, 0, 1078984704, 0},
		{0, 5, 13, 0, 1065353216, 0, 0},
		{0, 5, 13, 1045220557, 1061997773, 1059481191, 0},
		{3, 12, 17, 1065353216, 0, 1069928820, 0},
		{3, 12, 17, 0, 1065353216, 0, 1085625685},
		{3, 12, 17, 1045220557, 1061997773, 1050558762, 1083248913},
		{1, 1, 1, 1065353216, 0, 0, 0},
		{1, 1, 1, 0, 1065353216, 0, 1065353216},
		{1, 1, 1, 1045220557, 1061997773, 0, 1061997773},
		{7, 40, 123, 1065353216, 0, 1078581406, 0},
		{7, 40, 123, 0, 1065353216, 0, 1099731529},
		{7, 40, 123, 1045220557, 1061997773, 1059158552, 1096870415},
		{2, 2, 999, 1065353216, 0, 1148829696, 0},
		{2, 2, 999, 0, 1065353216, 0, 1140441088},
		{2, 2, 999, 1045220557, 1061997773, 1128778957, 1137167565},
	}
	for _, tc := range cases {
		counts := NewCharCounts()
		counts.SetSpaceCount(int(tc[0]))
		counts.SetNonSpaceCount(int(tc[1]))
		info := lineBoxJustificationInfo(counts, int(tc[2]), math.Float32frombits(tc[3]), math.Float32frombits(tc[4]))
		if got := math.Float32bits(info.NonSpaceAdjust()); got != tc[5] {
			t.Errorf("%v: NonSpaceAdjust bits = %d, want %d", tc, got, tc[5])
		}
		if got := math.Float32bits(info.SpaceAdjust()); got != tc[6] {
			t.Errorf("%v: SpaceAdjust bits = %d, want %d", tc, got, tc[6])
		}
	}
}

func TestLineBoxJustifyShares(t *testing.T) {
	if got := math.Float32bits(lineBoxJustifyNonSpaceShare); got != 1045220557 {
		t.Errorf("non-space share bits = %d, want 1045220557", got)
	}
	if got := math.Float32bits(lineBoxJustifySpaceShare); got != 1061997773 {
		t.Errorf("space share bits = %d, want 1061997773", got)
	}
}
