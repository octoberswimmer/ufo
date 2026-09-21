// Ported from flying-saucer-pdf/src/test/java/org/xhtmlrenderer/pdf/bug/FloatedTableCellTest.java

package pdf

import (
	"testing"

	"github.com/octoberswimmer/ufo"
)

// Reproducible example for https://github.com/flyingsaucerproject/flyingsaucer/issues/276

func TestFloatedTableCell_tableWithFloatedCell(t *testing.T) {
	page := ` <html>
	<body>
		<table>
			<tbody>
				<tr>
					<td id="cell-1" style="float:left;">first cell {float:left}</td>
					<td id="cell-2">second cell</td>
				</tr>
			</tbody>
		</table>
	</body>
</html>`

	renderer := NewITextRenderer()
	result, err := renderer.CreatePDFDocument(ufo.XMLResourceLoadString(page).GetDocument())
	if err != nil {
		t.Fatal(err)
	}
	pdf := testUtilsPrintFile(t, result, "table-with-floated-cell.pdf")
	pdf.containsText(t, "first cell", "second cell")
}

func TestFloatedTableCell_tableCell(t *testing.T) {
	page := `<table><tr><td style="float:left;">first cell {float:left;}</td></tr></table>`

	renderer := NewITextRenderer()
	result, err := renderer.CreatePDFDocument(ufo.XMLResourceLoadString(page).GetDocument())
	if err != nil {
		t.Fatal(err)
	}

	pdf := testUtilsPrintFile(t, result, "table-cell.pdf")
	pdf.containsText(t, "first cell")
}
