// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/BoxSizingBorderBoxTest.java

package pdf

import "testing"

// Verifies that box-sizing: border-box in Flying Saucer matches browser
// (standards mode).
//
// KEY API FACTS (verified from Box.java + TableRowBox.java source):
//
//	box.getWidth()         = contentWidth + leftMBP + rightMBP   (total rendered width)
//	box.getContentWidth()  = content area only
//	box.getHeight()        = contentHeight + padding + border     (total rendered height)
//	                         set AFTER calcLayoutHeight() adds padding/border,
//	                         then OVERWRITTEN by TableRowBox.setCellHeights() with row height.
//
// WHY height assertions expect 80 not 60:
//
//	calcDimensions sets contentHeight = 60 (border-box: 80 - 10top - 10bottom)
//	calcLayoutHeight ADDS padding+border back: 60 + 10 + 10 = 80
//	setCellHeights then sets cell.setHeight(rowHeight) = 80
//	So getHeight() = 80 is CORRECT with the patch. The BUG (without patch) gives 60.
//
// IMPORTANT — min-width / max-width and table-layout:
// CSS 2.1 §17.5.2 leaves the effect of min/max-width on table cells
// undefined. Chrome and Firefox IGNORE max-width / min-width on td in
// table-layout:fixed — the column-allocated width is authoritative,
// regardless of box-sizing. In table-layout:auto (advisory widths) both
// browsers DO respect max/min-width for every box-sizing value. Section 3
// tests are therefore split into two sub-groups accordingly.
//
// NOTE — applyCSSMinMaxWidth and border-box adjustment:
//
// Both max-width and min-width in border-box mode have paddingBorderWidth
// subtracted correctly (BlockBox.getCSSMaxWidth / getCSSMinWidth):
//
//	max-width:80px, padding:20px each side →
//	  contentWidth = 80 − 40 = 40   (matches browsers)
//	  getWidth()   = 40 + 40  = 80
//	min-width:120px, padding:20px each side →
//	  contentWidth = 120 − 40 = 80  (matches browsers)
//	  getWidth()   = 80 + 40  = 120

const boxSizingBorderBoxTestResetCss = `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}
`

func boxSizingBorderBoxTestXhtml(css string, body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
        "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <style type="text/css">
` + boxSizingBorderBoxTestResetCss + css + `    </style>
  </head>
  <body>
` + body + `  </body>
</html>
`
}

func boxSizingBorderBoxTestPx(dots int, dotsPerPixel int) int {
	return dots / dotsPerPixel
}

// boxSizingBorderBoxTestMeasures are the measures of a box in pixels.
type boxSizingBorderBoxTestMeasures struct {
	width, contentWidth, height int
}

// boxSizingBorderBoxTestLayout lays out html and returns the measures of the
// box of the element with id "target".
func boxSizingBorderBoxTestLayout(t *testing.T, html string) boxSizingBorderBoxTestMeasures {
	t.Helper()
	renderer := NewITextRenderer()
	if err := renderer.SetDocumentFromString(html); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Layout(); err != nil {
		t.Fatal(err)
	}

	dpp := renderer.GetSharedContext().GetDotsPerPixel()
	target := renderer.GetSharedContext().GetBoxById("target")
	if target == nil {
		t.Fatal("no box with id target")
	}
	return boxSizingBorderBoxTestMeasures{
		width:        boxSizingBorderBoxTestPx(target.GetWidth(), dpp),
		contentWidth: boxSizingBorderBoxTestPx(target.GetContentWidth(), dpp),
		height:       boxSizingBorderBoxTestPx(target.GetHeight(), dpp),
	}
}

func boxSizingBorderBoxTestAssert(t *testing.T, what string, actual int, expected int) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s = %d, want %d", what, actual, expected)
	}
}

// 1. TABLE-CELL WIDTH
// NOTE: must use table-layout: fixed — auto layout uses calcMinMaxWidth()
// which bypasses getOuterStyleWidth() and ignores the border-box patch.
func TestBoxSizingBorderBox_TableCellWidth(t *testing.T) {
	// THE ORIGINAL BUG: border-box was ignored so getWidth() returned 110.
	//
	// CSS:  width: 100px;  padding-right: 10px;  box-sizing: border-box;
	//
	// Browser (standards mode): total width = 100 px (padding IS inside the
	// width), content width = 90 px.
	t.Run("borderBox_rightPaddingOnly", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: fixed; width: 300px; }
.cell   { display: table-cell; }
#target { width: 100px; padding-right: 10px; }
`,
			`<div class="table">
  <div class="cell" id="target">Label</div>
  <div class="cell" id="filler">Value</div>
</div>
`))
		boxSizingBorderBoxTestAssert(t, "total width (border-box) must equal CSS width of 100px", m.width, 100)
		boxSizingBorderBoxTestAssert(t, "content = CSS width (100) - right padding (10) = 90px", m.contentWidth, 90)
	})

	// CSS:  width: 100px;  padding: 10px (all sides);  box-sizing: border-box;
	//
	// Browser: total width = 100 px, content width = 80 px.
	//
	// IMPORTANT: table-layout: fixed required. Without it, auto table layout
	// uses calcMinMaxWidth() which ignores getOuterStyleWidth() for initial
	// column sizing, causing the cell to render at 120px instead of 100px.
	t.Run("borderBox_allSidesPadding", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: fixed; width: 300px; }
.cell   { display: table-cell; }
#target { width: 100px; padding: 10px; }
`,
			`<div class="table">
  <div class="cell" id="target">Label</div>
  <div class="cell" id="filler">Value</div>
</div>
`))
		boxSizingBorderBoxTestAssert(t, "total width must equal CSS width (border-box): 100px", m.width, 100)
		boxSizingBorderBoxTestAssert(t, "content = 100 - 10 (left) - 10 (right) = 80px", m.contentWidth, 80)
	})

	// Sanity: content-box must NOT be broken by the patch.
	//
	// CSS:  width: 100px;  padding-right: 10px;  box-sizing: content-box;
	//
	// Browser: total = 110 px, content = 100 px
	t.Run("contentBox_noRegression", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: fixed; width: 300px; }
#target {
    display: table-cell;
    box-sizing: content-box;
    width: 100px;
    padding-right: 10px;
}
#filler { display: table-cell; }
`,
			`<div class="table">
  <div id="target">Label</div>
  <div id="filler">Value</div>
</div>
`))
		boxSizingBorderBoxTestAssert(t, "content-box: total = CSS width (100) + padding (10) = 110px", m.width, 110)
		boxSizingBorderBoxTestAssert(t, "content-box: content == CSS width exactly", m.contentWidth, 100)
	})
}

// 2. TABLE-CELL HEIGHT
func TestBoxSizingBorderBox_TableCellHeight(t *testing.T) {
	t.Run("borderBox_height_equalsCSS", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: fixed; width: 300px; }
.cell   { display: table-cell; }
#target { height: 80px; padding-top: 10px; padding-bottom: 10px; }
`,
			`<div class="table">
  <div class="cell" id="target">&#160;</div>
  <div class="cell" id="filler">&#160;</div>
</div>
`))
		boxSizingBorderBoxTestAssert(t, "border-box height: getHeight() must equal CSS height (80px). "+
			"If 60px, the double-subtraction bug is present.", m.height, 80)
	})

	t.Run("contentBox_legacyQuirkPreserved", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: fixed; width: 300px; }
.cell   { display: table-cell; }
#target { box-sizing: content-box; height: 80px; padding-top: 10px; padding-bottom: 10px; }
`,
			`<div class="table">
  <div class="cell" id="target">&#160;</div>
  <div class="cell" id="filler">&#160;</div>
</div>
`))
		boxSizingBorderBoxTestAssert(t, "content-box table-cell (XHTML 1.0 quirk): getHeight() == 80", m.height, 80)
	})
}

// 3. MIN-WIDTH / MAX-WIDTH with border-box
//
//	A. table-layout: auto  — applyCSSMinMaxWidth IS called.
//	B. table-layout: fixed — applyCSSMinMaxWidth NOT called; col width wins.
func TestBoxSizingBorderBox_MinMaxWidth(t *testing.T) {
	// A. table-layout: auto

	// CSS:  width: 200px;  max-width: 80px;  padding: 20px each side;
	//       box-sizing: border-box;  table-layout: auto
	//
	// applyCSSMinMaxWidth correctly applies border-box adjustment for
	// max-width: contentWidth = max-width − paddingBorderWidth = 80 − 40 =
	// 40px, getWidth() = 40 (content) + 40 (padding) = 80px.
	//
	// Regression guard: if applyCSSMinMaxWidth is NOT called (isBorderBox()
	// guard regression), the cell expands to its full auto-layout allocation
	// (≫ 80px) because max-width is silently ignored for content-box cells.
	t.Run("borderBox_maxWidth_autoLayout", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: auto; width: 500px; }
.cell   { display: table-cell; }
#target { width: 200px; max-width: 80px; padding-left: 20px; padding-right: 20px; }
`,
			`<div class="table">
  <div class="cell" id="target">x</div>
  <div class="cell" id="filler">&#160;</div>
</div>
`))
		// applyCSSMinMaxWidth is called (auto-layout) and correctly applies
		// border-box adjustment: total = max-width = 80px.
		// Regression: if skipped (isBorderBox() guard), total ≫ 80px.
		boxSizingBorderBoxTestAssert(t, "auto-layout, max-width:80px (border-box) → total must be 80px", m.width, 80)
		// content = max-width (80) − left pad (20) − right pad (20) = 40px
		boxSizingBorderBoxTestAssert(t, "content = max-width (80) − left pad (20) − right pad (20) = 40px", m.contentWidth, 40)
	})

	// CSS:  width: 30px;  min-width: 120px;  padding: 20px;
	//       box-sizing: border-box;
	//
	// Browser: min-width is border-box → total ≥ 120px, content = 120-40 = 80px.
	t.Run("borderBox_minWidth_autoLayout", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: auto; width: 500px; }
.cell   { display: table-cell; }
#target { width: 30px; min-width: 120px; padding-left: 20px; padding-right: 20px; }
`,
			`<div class="table">
  <div class="cell" id="target">x</div>
  <div class="cell" id="filler">&#160;</div>
</div>
`))
		// getCSSMinWidth now correctly subtracts paddingBorderWidth (40) for
		// border-box. content floor = 120 − 40 = 80; total = 80 + 40 = 120.
		// Matches Chrome + Firefox.
		boxSizingBorderBoxTestAssert(t, "auto-layout, min-width:120px (border-box) → total must be 120px", m.width, 120)
		boxSizingBorderBoxTestAssert(t, "content = min-width (120) − left pad (20) − right pad (20) = 80px", m.contentWidth, 80)
	})

	// B. table-layout: fixed

	// table-layout:fixed, border-box: col-allocated width (200px) must win
	// over max-width:80px.
	t.Run("borderBox_maxWidth_fixedLayout_colWidthWins", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: fixed; width: 500px; }
.cell   { display: table-cell; }
#target { width: 200px; max-width: 80px; padding-left: 20px; padding-right: 20px; }
`,
			`<div class="table">
  <div class="cell" id="target">x</div>
  <div class="cell" id="filler">&#160;</div>
</div>
`))
		boxSizingBorderBoxTestAssert(t, "fixed-layout, border-box: col-allocated width (200px) must win over max-width:80px", m.width, 200)
	})

	// table-layout:fixed, border-box: col-allocated width (80px) must win
	// over min-width:200px.
	t.Run("borderBox_minWidth_fixedLayout_colWidthWins", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`.table  { display: table; table-layout: fixed; width: 500px; }
.cell   { display: table-cell; }
#target { width: 80px; min-width: 200px; padding-left: 10px; padding-right: 10px; }
`,
			`<div class="table">
  <div class="cell" id="target">x</div>
  <div class="cell" id="filler">&#160;</div>
</div>
`))
		boxSizingBorderBoxTestAssert(t, "fixed-layout, border-box: col-allocated width (80px) must win over min-width:200px", m.width, 80)
	})
}

// 4. DISPLAY: BLOCK — no regression
func TestBoxSizingBorderBox_RegularBlock(t *testing.T) {
	t.Run("blockBorderBox_unaffected", func(t *testing.T) {
		m := boxSizingBorderBoxTestLayout(t, boxSizingBorderBoxTestXhtml(
			`#target { display: block; width: 100px; padding-left: 10px; padding-right: 10px; }
`,
			`<div id="target">x</div>
`))
		boxSizingBorderBoxTestAssert(t, "width", m.width, 100)
		boxSizingBorderBoxTestAssert(t, "content width", m.contentWidth, 80)
	})
}
