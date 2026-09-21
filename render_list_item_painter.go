// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/ListItemPainter.java

package ufo

// ListItemPainterPaint ports ListItemPainter.paint. ListItemPainter is a
// utility class to paint list markers (all types). See MarkerData.
func ListItemPainterPaint(c *RenderingContext, box BlockBoxI) {
	if box.GetMarkerData() == nil {
		return
	}

	markerData := box.GetMarkerData()

	if markerData.GetImageMarker() != nil {
		listItemPainterDrawImage(c, box, markerData)
	} else {
		style := box.GetStyle()
		c.GetOutputDevice().SetColor(style.GetColor())

		if markerData.GetGlyphMarker() != nil {
			listItemPainterDrawGlyph(c, box, style)
		} else if markerData.GetTextMarker() != nil {
			listItemPainterDrawText(c, box)
		}
	}
}

func listItemPainterDrawImage(c *RenderingContext, box BlockBoxI, markerData *MarkerData) {
	marker := markerData.GetImageMarker()
	img := marker.GetImage()
	x := listItemPainterGetReferenceX(c, box)
	// FIXME: findbugs possible loss of precision, cf. int / (float)2
	x += -marker.GetLayoutWidth() +
		marker.GetLayoutWidth()/2 - img.GetWidth()/2
	c.GetOutputDevice().DrawImage(img,
		x,
		listItemPainterGetListItemCenterBaseline(c, box)-img.GetHeight()/2)
}

func listItemPainterGetReferenceX(c *RenderingContext, box BlockBoxI) int {
	markerData := box.GetMarkerData()

	if markerData.GetReferenceLine() != nil {
		return markerData.GetReferenceLine().GetAbsX()
	} else {
		return box.GetAbsX() + calculatedStyleFloatToInt(box.GetMargin(c).Left())
	}
}

// Java saves the output device's antialiasing rendering hint, forces it on
// while the glyph is drawn and restores it afterwards. OutputDevice here has no
// getRenderingHint/setRenderingHint (the PDF output device implements both as
// no-ops), so those calls are dropped.
func listItemPainterDrawGlyph(c *RenderingContext, box BlockBoxI, style CalculatedStyleI) {
	// calculations for bullets
	marker := box.GetMarkerData().GetGlyphMarker()
	x := listItemPainterGetReferenceX(c, box) - marker.GetLayoutWidth()
	y := listItemPainterGetListItemCenterBaseline(c, box) - marker.GetDiameter()/2

	listStyle := style.GetIdent(CSSNameListStyleType)
	if listStyle == IdentValueDisc {
		c.GetOutputDevice().FillOval(x, y, marker.GetDiameter(), marker.GetDiameter())
	} else if listStyle == IdentValueSquare {
		c.GetOutputDevice().FillRect(x, y, marker.GetDiameter(), marker.GetDiameter())
	} else if listStyle == IdentValueCircle {
		c.GetOutputDevice().DrawOval(x, y, marker.GetDiameter(), marker.GetDiameter())
	}
}

func listItemPainterGetListItemCenterBaseline(c *RenderingContext, box BlockBoxI) int {
	childBox := listItemPainterGetFirstNestedChild(box)

	return childBox.GetAbsY() +
		childBox.GetHeight()/2 +
		calculatedStyleFloatToInt(childBox.GetMargin(c).Top())/2 -
		calculatedStyleFloatToInt(childBox.GetMargin(c).Bottom())/2 +
		calculatedStyleFloatToInt(childBox.GetPadding(c).Top())/2 -
		calculatedStyleFloatToInt(childBox.GetPadding(c).Bottom())/2
}

func listItemPainterGetFirstNestedChild(box BoxI) BoxI {
	if box.GetChildCount() > 0 {
		return listItemPainterGetFirstNestedChild(box.GetChild(0))
	} else {
		return box
	}
}

func listItemPainterDrawText(c *RenderingContext, box BlockBoxI) {
	text := box.GetMarkerData().GetTextMarker()

	x := listItemPainterGetReferenceX(c, box) - text.GetLayoutWidth()
	y := listItemPainterGetReferenceBaseline(box)

	c.GetOutputDevice().SetColor(box.GetStyle().GetColor())
	c.GetOutputDevice().SetFont(box.GetStyle().GetFSFont(c))
	c.GetTextRenderer().DrawString(c.GetOutputDevice(), text.GetText(), float32(x), float32(y))
}

func listItemPainterGetReferenceBaseline(box BlockBoxI) int {
	markerData := box.GetMarkerData()
	strutMetrics := markerData.GetStructMetrics()

	if markerData.GetReferenceLine() != nil {
		return markerData.GetReferenceLine().GetAbsY() + strutMetrics.GetBaseline()
	} else {
		return box.GetAbsY() + box.GetTy() + strutMetrics.GetBaseline()
	}
}
