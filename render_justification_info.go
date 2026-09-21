// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/JustificationInfo.java

package ufo

// JustificationInfo is the record JustificationInfo(float nonSpaceAdjust,
// float spaceAdjust): the width added after each non-space character and after
// each space character of a justified line.
type JustificationInfo struct {
	nonSpaceAdjust float32
	spaceAdjust    float32
}

func NewJustificationInfo(nonSpaceAdjust float32, spaceAdjust float32) *JustificationInfo {
	return &JustificationInfo{nonSpaceAdjust: nonSpaceAdjust, spaceAdjust: spaceAdjust}
}

func (j *JustificationInfo) NonSpaceAdjust() float32 {
	return j.nonSpaceAdjust
}

func (j *JustificationInfo) SpaceAdjust() float32 {
	return j.spaceAdjust
}
