// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/MultiColumnLayoutTest.java

package pdf

import (
	"math"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/pdf/writer"
)

var multiColumnLayoutTestLeftColumnMarkers = []string{
	"COL-ALPHA", "COL-BETA", "COL-GAMMA", "COL-DELTA", "COL-EPSILON", "COL-ZETA"}
var multiColumnLayoutTestRightColumnMarkers = []string{
	"COL-ETA", "COL-THETA", "COL-IOTA", "COL-KAPPA", "COL-LAMBDA", "COL-MU"}

func TestMultiColumnLayout_multiColumnCssParsesAndRendersIntoPdf(t *testing.T) {
	bytes := testUtilsFromClasspathResource(t, "page-with-columns.html")
	pdf := testUtilsPrintFile(t, bytes, "page-with-columns.pdf")
	pdf.containsText(t, "COL-ALPHA", "COL-BETA", "COL-GAMMA", "COL-DELTA")

	allMarkers := append(append([]string{}, multiColumnLayoutTestLeftColumnMarkers...), multiColumnLayoutTestRightColumnMarkers...)
	xByMarker := multiColumnLayoutTestExtractMarkerXCoordinates(t, pdf.doc, allMarkers)

	if len(xByMarker) != 12 {
		t.Fatalf("found %d markers, want 12: %v", len(xByMarker), xByMarker)
	}

	leftBand := multiColumnLayoutTestAverageX(xByMarker, multiColumnLayoutTestLeftColumnMarkers)
	rightBand := multiColumnLayoutTestAverageX(xByMarker, multiColumnLayoutTestRightColumnMarkers)
	maxLeft := math.Inf(-1)
	for _, m := range multiColumnLayoutTestLeftColumnMarkers {
		maxLeft = math.Max(maxLeft, xByMarker[m])
	}
	minRight := math.Inf(1)
	for _, m := range multiColumnLayoutTestRightColumnMarkers {
		minRight = math.Min(minRight, xByMarker[m])
	}
	if !(maxLeft < minRight-20) {
		t.Errorf("max x of the left column %v is not less than min x of the right column %v - 20", maxLeft, minRight)
	}
	if !(leftBand < rightBand-20) {
		t.Errorf("left band %v is not less than right band %v - 20", leftBand, rightBand)
	}

	for _, m := range multiColumnLayoutTestLeftColumnMarkers {
		if math.Abs(xByMarker[m]-leftBand) > 8 {
			t.Errorf("%s at x %v, not within 8 of %v", m, xByMarker[m], leftBand)
		}
	}
	for _, m := range multiColumnLayoutTestRightColumnMarkers {
		if math.Abs(xByMarker[m]-rightBand) > 8 {
			t.Errorf("%s at x %v, not within 8 of %v", m, xByMarker[m], rightBand)
		}
	}
}

func multiColumnLayoutTestAverageX(xByMarker map[string]float64, markers []string) float64 {
	sum := 0.0
	for _, m := range markers {
		sum += xByMarker[m]
	}
	return sum / float64(len(markers))
}

// multiColumnLayoutTestExtractMarkerXCoordinates stands for the
// PDFTextStripper subclass of the Java test, which records the smallest x of
// the characters of each marker: the x where the first string that starts
// with the marker is drawn.
func multiColumnLayoutTestExtractMarkerXCoordinates(t *testing.T, document *writer.ReadDocument, markers []string) map[string]float64 {
	t.Helper()
	positions := map[string]float64{}
	for i := 0; i < document.NumPages(); i++ {
		runs, err := document.PageTextRuns(i)
		if err != nil {
			t.Fatal(err)
		}
		for _, run := range runs {
			for _, marker := range markers {
				if _, ok := positions[marker]; ok {
					continue
				}
				if strings.HasPrefix(run.Text, marker) {
					positions[marker] = run.X
				}
			}
		}
	}
	return positions
}
