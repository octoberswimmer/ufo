package ufo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// floatManagerTestDocument supplies the styles of the boxes: floated left or
// right, cleared, with and without margins.
const floatManagerTestDocument = `<html><head><style>
div { display: block; line-height: 17px; }
.l { float: left; }
.r { float: right; }
.cl { clear: left; }
.cr { clear: right; }
.cb { clear: both; }
.m { margin: 3px 5px 7px 2px; }
</style></head><body>
<div id="master"/>
<div id="s0" class="l"/><div id="s1" class="r"/><div id="s2" class="l cl"/><div id="s3" class="r cr"/>
<div id="s4" class="l cb m"/><div id="s5" class="r cl m"/><div id="s6" class="l m"/><div id="s7" class="r m"/>
<div id="s8" class="l cr"/><div id="s9" class="r cb"/>
<div id="b0"/><div id="b1" class="cl"/><div id="b2" class="cr m"/><div id="b3" class="cb m"/>
</body></html>`

var floatManagerTestSeed int64

func floatManagerTestNext(n int) int {
	floatManagerTestSeed = (floatManagerTestSeed*1103515245 + 12345) & 0x7fffffff
	return int((floatManagerTestSeed >> 8) % int64(n))
}

// TestFloatManager_matchesJava runs pseudo-random sequences of float
// placements, line queries, clears, BFC translations and float removals. The
// expected output was produced by the same sequences run against the Java
// FloatManager and BlockFormattingContext (flying-saucer-core 10.6.0-SNAPSHOT)
// with BlockBox and LineBox objects whose dimensions were set directly.
func TestFloatManager_matchesJava(t *testing.T) {
	sc, doc := layoutContextTestSharedContext(t, floatManagerTestDocument)
	styles := map[string]CalculatedStyleI{}
	for _, e := range doc.GetElementsByTagName("div") {
		styles[e.GetAttribute("id")] = sc.GetStyle(e)
	}
	var all strings.Builder
	for run := 0; run < 60; run++ {
		floatManagerTestSeed = int64(run)*7919 + 1
		floatManagerTestRun(sc, styles, run, &all)
	}
	expected, err := os.ReadFile("testdata/layout/float_manager_java.txt")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(all.String(), "\n")
	want := strings.Split(string(expected), "\n")
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			g := "<end of output>"
			if i < len(got) {
				g = got[i]
			}
			t.Fatalf("line %d differs from Java: got %q, want %q", i+1, g, want[i])
		}
	}
	if len(got) != len(want) {
		t.Fatalf("%d lines of output, Java has %d", len(got), len(want))
	}
}

func floatManagerTestRun(sc *SharedContext, styles map[string]CalculatedStyleI, run int, out *strings.Builder) {
	c := sc.NewLayoutContextInstance(nil)
	fmt.Fprintf(out, "run %d\n", run)
	master := NewBlockBoxWithElementStyleAnonymous(nil, styles["master"], false)
	cbWidth := 100 + floatManagerTestNext(400)
	master.SetContentWidth(cbWidth)
	master.SetAbsX(floatManagerTestNext(50))
	master.SetAbsY(floatManagerTestNext(50))
	master.SetContainingBlock(master)
	layer := NewLayer(master)
	c.PushLayerLayer(layer)
	bfc := NewBlockFormattingContext(master, c)
	c.PushBFC(bfc)
	var floated []*BlockBox
	lineY := 0
	steps := 5 + floatManagerTestNext(30)
	for i := 0; i < steps; i++ {
		op := floatManagerTestNext(10)
		if op < 5 {
			b := NewBlockBoxWithElementStyleAnonymous(nil, styles["s"+strconv.Itoa(floatManagerTestNext(10))], false)
			b.SetFloatedBoxData(NewFloatedBoxData())
			b.SetContainingBlock(master)
			b.SetContentWidth(10 + floatManagerTestNext(cbWidth))
			b.SetLeftMBP(floatManagerTestNext(8))
			b.SetRightMBP(floatManagerTestNext(8))
			b.SetHeight(floatManagerTestNext(60))
			b.SetY(lineY)
			c.GetBlockFormattingContext().FloatBox(c, b)
			floated = append(floated, b)
			fmt.Fprintf(out, "float x=%d y=%d ax=%d ay=%d mgr=%v\n", b.GetX(), b.GetY(), b.GetAbsX(), b.GetAbsY(),
				b.GetFloatedBoxData().GetManager() == bfc.GetFloatManager())
		} else if op < 7 {
			line := NewLineBox(master, styles["b0"])
			line.SetContainingBlock(master)
			line.SetY(lineY)
			line.SetAbsY(master.GetAbsY() + lineY + bfc.GetOffset().Y*0)
			line.SetX(floatManagerTestNext(10))
			if floatManagerTestNext(3) == 0 {
				line.SetContentWidth(0)
			} else {
				line.SetContentWidth(floatManagerTestNext(cbWidth))
			}
			if floatManagerTestNext(3) == 0 {
				line.SetHeight(0)
			} else {
				line.SetHeight(floatManagerTestNext(30))
			}
			w := floatManagerTestNext(cbWidth + 50)
			fmt.Fprintf(out, "line left=%d right=%d both=%d delta=%d\n", bfc.GetLeftFloatDistance(c, line, w),
				bfc.GetRightFloatDistance(c, line, w), bfc.GetFloatDistance(c, line, w), bfc.GetNextLineBoxDelta(c, line, w))
			lineY += floatManagerTestNext(25)
		} else if op < 8 {
			b := NewBlockBoxWithElementStyleAnonymous(nil, styles["b"+strconv.Itoa(floatManagerTestNext(4))], false)
			b.SetContainingBlock(master)
			b.SetContentWidth(cbWidth)
			b.SetHeight(floatManagerTestNext(40))
			b.SetX(floatManagerTestNext(5))
			b.SetY(lineY)
			bfc.Clear(c, b)
			fmt.Fprintf(out, "clear y=%d clearDelta=%d\n", b.GetY(), bfc.GetFloatManager().GetClearDelta(c, lineY))
			lineY = b.GetY() + b.GetHeight()
		} else if op < 9 {
			tx, ty := floatManagerTestNext(20)-5, floatManagerTestNext(20)-5
			c.Translate(tx, ty)
			fmt.Fprintf(out, "translate %s\n", bfc)
		} else if len(floated) != 0 {
			i := floatManagerTestNext(len(floated))
			b := floated[i]
			floated = append(floated[:i:i], floated[i+1:]...)
			p := bfc.GetFloatManager().GetOffset(b)
			bfc.GetFloatManager().RemoveFloat(b)
			fmt.Fprintf(out, "remove offset=%d,%d mgr=%v after=%v\n", p.X, p.Y, b.GetFloatedBoxData().GetManager() == nil,
				bfc.GetFloatManager().GetOffset(b) == nil)
		}
	}
	master.SetAbsX(master.GetAbsX() + 13)
	master.SetAbsY(master.GetAbsY() + 17)
	bfc.GetFloatManager().PerformFloatOperation(FloatManagerFloatOperationFunc(func(b BoxI) {
		fmt.Fprintf(out, "op ax=%d ay=%d\n", b.GetAbsX(), b.GetAbsY())
	}))
	bfc.GetFloatManager().CalcFloatLocations()
	for _, b := range floated {
		fmt.Fprintf(out, "loc ax=%d ay=%d\n", b.GetAbsX(), b.GetAbsY())
	}
}
