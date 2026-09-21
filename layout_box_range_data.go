// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BoxRangeData.java

package ufo

import (
	"fmt"

	"github.com/octoberswimmer/ufo/geom"
)

type BoxRangeData struct {
	box      BlockBoxI
	boxRange *BoxRange

	// clip is nil until SetClip is called, and stays nil when the output
	// device had no clip at that time.
	clip geom.Shape
}

func NewBoxRangeData(box BlockBoxI, boxRange *BoxRange) *BoxRangeData {
	return &BoxRangeData{box: box, boxRange: boxRange}
}

func (d *BoxRangeData) GetBox() BlockBoxI {
	return d.box
}

func (d *BoxRangeData) GetRange() *BoxRange {
	return d.boxRange
}

// GetClip may return nil.
func (d *BoxRangeData) GetClip() geom.Shape {
	return d.clip
}

func (d *BoxRangeData) SetClip(clip geom.Shape) {
	d.clip = clip
}

func (d *BoxRangeData) ToString() string {
	box := "null"
	if d.box != nil {
		box = d.box.ToString()
	}
	clip := "null"
	if d.clip != nil {
		clip = fmt.Sprint(d.clip)
	}
	return fmt.Sprintf("[range= %s, box=%s, clip=%s]", d.boxRange.ToString(), box, clip)
}

func (d *BoxRangeData) String() string {
	return d.ToString()
}
