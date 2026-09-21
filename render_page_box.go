// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/render/PageBox.java

package ufo

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/octoberswimmer/ufo/dom"
	"github.com/octoberswimmer/ufo/geom"
)

var pageBoxMarginAreaDefs = []pageBoxMarginAreaI{
	newPageBoxTopLeftCorner(),
	newPageBoxTopMarginArea(),
	newPageBoxTopRightCorner(),

	newPageBoxLeftMarginArea(),
	newPageBoxRightMarginArea(),

	newPageBoxBottomLeftCorner(),
	newPageBoxBottomMarginArea(),
	newPageBoxBottomRightCorner(),
}

const pageBoxLeadingTrailingSplit = 5

type PageBox struct {
	style CalculatedStyleI

	top    int
	bottom int

	paintingTop    int
	paintingBottom int

	pageNo int

	outerPageWidth int

	// pageDimensions is nil until the first call of getPageDimensions.
	pageDimensions      *pageBoxPageDimensions
	pageDimensionsMutex sync.Mutex

	pageInfo *PageInfo

	// marginAreas has one entry per element of pageBoxMarginAreaDefs; an entry
	// is nil when the page has no margin boxes for that area.
	marginAreas []*pageBoxMarginAreaContainer

	// metadata may be nil.
	metadata *dom.Element
}

func NewPageBox(pageInfo *PageInfo, cssContext CssContext, style CalculatedStyleI, top int, pageNo int) *PageBox {
	p := &PageBox{
		marginAreas: make([]*pageBoxMarginAreaContainer, len(pageBoxMarginAreaDefs)),
	}
	p.pageInfo = pageInfo
	p.style = style
	p.outerPageWidth = p.GetWidth(cssContext)
	p.top = top
	p.bottom = top + p.GetContentHeight(cssContext)
	p.pageNo = pageNo
	return p
}

func (p *PageBox) GetWidth(cssCtx CssContext) int {
	return p.getPageDimensions(cssCtx).width
}

func (p *PageBox) GetHeight(cssCtx CssContext) int {
	return p.getPageDimensions(cssCtx).height
}

func (p *PageBox) getPageDimensions(cssCtx CssContext) *pageBoxPageDimensions {
	p.pageDimensionsMutex.Lock()
	defer p.pageDimensionsMutex.Unlock()
	if p.pageDimensions == nil {
		p.pageDimensions = p.resolvePageDimensions(cssCtx)
	}
	return p.pageDimensions
}

func (p *PageBox) resolvePageDimensions(cssCtx CssContext) *pageBoxPageDimensions {
	style := p.GetStyle()

	var width int
	if style.IsLength(CSSNameFsPageWidth) {
		width = style.GetIntPropertyProportionalTo(CSSNameFsPageWidth, 0, cssCtx)
	} else {
		width = p.resolveAutoPageWidth(cssCtx)
	}

	var height int
	if style.IsLength(CSSNameFsPageHeight) {
		height = style.GetIntPropertyProportionalTo(CSSNameFsPageHeight, 0, cssCtx)
	} else {
		height = p.resolveAutoPageHeight(cssCtx)
	}

	if style.IsIdent(CSSNameFsPageOrientation, IdentValueLandscape) {
		return &pageBoxPageDimensions{height, width}
	}
	return &pageBoxPageDimensions{width, height}
}

func (p *PageBox) isUseLetterSize() bool {
	county := pageBoxDefaultLocaleCountry()

	// Per http://en.wikipedia.org/wiki/Paper_size, letter paper is
	// a de facto standard in Canada (although the government uses
	// its own standard) and Mexico (even though it is officially an ISO
	// country)
	return county == "US" || county == "CA" || county == "MX"
}

// pageBoxDefaultLocaleCountry stands in for Locale.getDefault().getCountry().
// On Unix the JDK takes the default locale from the LC_MESSAGES locale
// category, whose value is the first non-empty one of the environment
// variables LC_ALL, LC_MESSAGES and LANG. A value has the form
// language[_COUNTRY][.encoding][@variant]; "C" and "POSIX" have no country.
func pageBoxDefaultLocaleCountry() string {
	locale := ""
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := os.Getenv(name); value != "" {
			locale = value
			break
		}
	}
	if i := strings.IndexAny(locale, ".@"); i >= 0 {
		locale = locale[:i]
	}
	i := strings.Index(locale, "_")
	if i < 0 {
		return ""
	}
	return strings.ToUpper(locale[i+1:])
}

func (p *PageBox) resolveAutoPageWidth(cssCtx CssContext) int {
	if p.isUseLetterSize() {
		return calculatedStyleFloatToInt(LengthValueCalcFloatProportionalValue(
			p.GetStyle(),
			CSSNameFsPageWidth,
			"8.5in",
			8.5,
			CSSPrimitiveValueCssIn,
			0,
			cssCtx))
	} else {
		return calculatedStyleFloatToInt(LengthValueCalcFloatProportionalValue(
			p.GetStyle(),
			CSSNameFsPageWidth,
			"210mm",
			210.0,
			CSSPrimitiveValueCssMm,
			0,
			cssCtx))
	}
}

func (p *PageBox) resolveAutoPageHeight(cssCtx CssContext) int {
	if p.isUseLetterSize() {
		return calculatedStyleFloatToInt(LengthValueCalcFloatProportionalValue(
			p.GetStyle(),
			CSSNameFsPageHeight,
			"11in",
			11.0,
			CSSPrimitiveValueCssIn,
			0,
			cssCtx))
	} else {
		return calculatedStyleFloatToInt(LengthValueCalcFloatProportionalValue(
			p.GetStyle(),
			CSSNameFsPageHeight,
			"297mm",
			297.0,
			CSSPrimitiveValueCssMm,
			0,
			cssCtx))
	}
}

func (p *PageBox) GetContentHeight(cssCtx CssContext) int {
	height := p.GetHeight(cssCtx) - p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeTop) - p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeBottom)
	if height <= 0 {
		panic(NewXRRuntimeException(
			"The content height cannot be zero or less.  Check your document margin definition."))
	}
	return height
}

func (p *PageBox) GetContentWidth(cssCtx CssContext) int {
	width := p.GetWidth(cssCtx) - p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeLeft) - p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeRight)
	if width <= 0 {
		panic(NewXRRuntimeException(
			"The content width cannot be zero or less.  Check your document margin definition."))
	}
	return width
}

func (p *PageBox) GetStyle() CalculatedStyleI {
	return p.style
}

func (p *PageBox) GetBottom() int {
	return p.bottom
}

func (p *PageBox) GetTop() int {
	return p.top
}

func (p *PageBox) GetPaintingBottom() int {
	return p.paintingBottom
}

func (p *PageBox) SetPaintingBottom(paintingBottom int) {
	p.paintingBottom = paintingBottom
}

func (p *PageBox) GetPaintingTop() int {
	return p.paintingTop
}

func (p *PageBox) SetPaintingTop(paintingTop int) {
	p.paintingTop = paintingTop
}

func (p *PageBox) GetScreenPaintingBounds(cssCtx CssContext, additionalClearance int) *geom.Rectangle {
	return geom.NewRectangle(
		additionalClearance, p.GetPaintingTop(),
		p.GetWidth(cssCtx), p.GetPaintingBottom()-p.GetPaintingTop())
}

func (p *PageBox) GetPrintPaintingBounds(cssCtx CssContext) *geom.Rectangle {
	return geom.NewRectangle(
		0, 0,
		p.GetWidth(cssCtx), p.GetHeight(cssCtx))
}

func (p *PageBox) GetPagedViewClippingBounds(cssCtx CssContext, additionalClearance int) *geom.Rectangle {
	return geom.NewRectangle(
		additionalClearance+p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeLeft),
		p.GetPaintingTop()+p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeTop),
		p.GetContentWidth(cssCtx),
		p.GetContentHeight(cssCtx))
}

func (p *PageBox) GetPrintClippingBounds(cssCtx CssContext) *geom.Rectangle {
	return geom.NewRectangle(
		p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeLeft),
		p.GetMarginBorderPadding(cssCtx, CalculatedStyleEdgeTop),
		p.GetContentWidth(cssCtx),
		p.GetContentHeight(cssCtx)-1)
}

func (p *PageBox) GetMargin(cssCtx CssContext) RectPropertySetI {
	return p.GetStyle().GetMarginRect(float32(p.outerPageWidth), cssCtx)
}

func (p *PageBox) getBorderEdge(left int, top int, cssCtx CssContext) *geom.Rectangle {
	margin := p.GetMargin(cssCtx)
	return geom.NewRectangle(left+calculatedStyleFloatToInt(margin.Left()),
		top+calculatedStyleFloatToInt(margin.Top()),
		p.GetWidth(cssCtx)-calculatedStyleFloatToInt(margin.Left())-calculatedStyleFloatToInt(margin.Right()),
		p.GetHeight(cssCtx)-calculatedStyleFloatToInt(margin.Top())-calculatedStyleFloatToInt(margin.Bottom()))
}

func (p *PageBox) PaintBorder(c *RenderingContext, additionalClearance int, mode LayerPagedMode) {
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = p.GetPaintingTop()
	case LayerPagedModePagedModePrint:
		top = 0
	}
	c.GetOutputDevice().PaintBorderWithStyleEdgeSides(c,
		p.GetStyle(),
		p.getBorderEdge(additionalClearance, top, c),
		BorderPainterAll)
}

func (p *PageBox) PaintBackground(c *RenderingContext, additionalClearance int, mode LayerPagedMode) {
	var bounds *geom.Rectangle
	switch mode {
	case LayerPagedModePagedModeScreen:
		bounds = p.GetScreenPaintingBounds(c, additionalClearance)
	case LayerPagedModePagedModePrint:
		bounds = p.GetPrintPaintingBounds(c)
	}
	c.GetOutputDevice().PaintBackgroundWithStyleBoundsBgImageContainerBorder(c, p.GetStyle(), bounds, bounds, p.GetStyle().GetBorder(c))
}

func (p *PageBox) PaintMarginAreas(c *RenderingContext, additionalClearance int, mode LayerPagedMode) {
	for i := 0; i < len(pageBoxMarginAreaDefs); i++ {
		container := p.marginAreas[i]
		if container != nil {
			table := container.table
			pt := container.area.getPaintingPosition(c, p, additionalClearance, mode)

			c.GetOutputDevice().Translate(float64(pt.X), float64(pt.Y))
			table.GetLayer().Paint(c)
			c.GetOutputDevice().Translate(float64(-pt.X), float64(-pt.Y))
		}
	}
}

func (p *PageBox) GetPageNo() int {
	return p.pageNo
}

func (p *PageBox) GetOuterPageWidth() int {
	return p.outerPageWidth
}

func (p *PageBox) GetMarginBorderPadding(cssCtx CssContext, edge *CalculatedStyleEdge) int {
	return p.GetStyle().GetMarginBorderPadding(cssCtx, p.GetOuterPageWidth(), edge)
}

func (p *PageBox) GetPageInfo() *PageInfo {
	return p.pageInfo
}

// GetMetadata may return nil.
func (p *PageBox) GetMetadata() *dom.Element {
	return p.metadata
}

func (p *PageBox) Layout(c *LayoutContext) {
	c.SetPage(p)
	p.retrievePageMetadata(c)
	p.layoutMarginAreas(c)
}

// HACK Would much prefer to do this in ITextRenderer or ITextOutputDevice
// but given the existing API, this is about the only place it can be done
func (p *PageBox) retrievePageMetadata(c *LayoutContext) {
	props := p.GetPageInfo().GetXMPPropertyList()
	if len(props) != 0 {
		for _, decl := range props {
			if decl.GetCSSName() == CSSNameContent {
				value := decl.GetValue()
				values := value.GetValues()
				if len(values) == 1 {
					funcVal := values[0].(*PropertyValue)
					if funcVal.GetPropertyValueType() == PropertyValueTypeValueTypeFunction {
						function := funcVal.GetFunction()
						if BoxBuilderIsElementFunction(function) {
							metadata := BoxBuilderGetRunningBlock(c, funcVal)
							if metadata != nil {
								p.metadata = metadata.GetElement()
							}
						}
					}
				}
				break
			}
		}
	}
}

func (p *PageBox) layoutMarginAreas(c *LayoutContext) {
	margin := p.GetMargin(c)
	for i := 0; i < len(pageBoxMarginAreaDefs); i++ {
		area := pageBoxMarginAreaDefs[i]

		dim := area.getLayoutDimension(c, p, margin)
		table := BoxBuilderCreateMarginTable(
			c, p.pageInfo,
			area.getMarginBoxNames(),
			int(dim.GetHeight()),
			area.getDirection())
		if table != nil {
			table.SetContainingBlock(NewMarginBox(geom.NewRectangleWidthHeight(int(dim.GetWidth()), int(dim.GetHeight()))))
			func() {
				defer c.SetNoPageBreak(0)
				c.SetNoPageBreak(1)

				c.ReInit(false)
				c.PushLayerBox(table)
				c.GetRootLayer().AddPage(c)

				table.Layout(c)

				c.PopLayer()
			}()
			p.marginAreas[i] = &pageBoxMarginAreaContainer{area, table}
		}
	}
}

func (p *PageBox) IsLeftPage() bool {
	return p.pageNo%2 != 0
}

func (p *PageBox) IsRightPage() bool {
	return p.pageNo%2 == 0
}

func (p *PageBox) ExportLeadingText(c *RenderingContext, writer io.Writer) error {
	for i := 0; i < pageBoxLeadingTrailingSplit; i++ {
		container := p.marginAreas[i]
		if container != nil {
			if err := container.table.ExportText(c, writer); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *PageBox) ExportTrailingText(c *RenderingContext, writer io.Writer) error {
	for i := pageBoxLeadingTrailingSplit; i < len(p.marginAreas); i++ {
		container := p.marginAreas[i]
		if container != nil {
			if err := container.table.ExportText(c, writer); err != nil {
				return err
			}
		}
	}
	return nil
}

// pageBoxPageDimensions ports the private record PageBox.PageDimensions.
type pageBoxPageDimensions struct {
	width  int
	height int
}

// pageBoxMarginAreaContainer ports the private record
// PageBox.MarginAreaContainer.
type pageBoxMarginAreaContainer struct {
	area  pageBoxMarginAreaI
	table *TableBox
}

// pageBoxMarginAreaI lists the methods of the private abstract class
// PageBox.MarginArea.
type pageBoxMarginAreaI interface {
	getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension

	getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point

	getMarginBoxNames() []*MarginBoxName

	getDirection() BoxBuilderMarginDirection
}

// pageBoxMarginArea ports the private abstract class PageBox.MarginArea. None
// of its methods calls an abstract method, so it holds no self.
type pageBoxMarginArea struct {
	marginBoxNames []*MarginBoxName
}

func (m *pageBoxMarginArea) getMarginBoxNames() []*MarginBoxName {
	return m.marginBoxNames
}

func (m *pageBoxMarginArea) getDirection() BoxBuilderMarginDirection {
	return BoxBuilderMarginDirectionHorizontal
}

type pageBoxTopLeftCorner struct {
	pageBoxMarginArea
}

func newPageBoxTopLeftCorner() *pageBoxTopLeftCorner {
	return &pageBoxTopLeftCorner{pageBoxMarginArea{[]*MarginBoxName{MarginBoxNameTopLeftCorner}}}
}

func (a *pageBoxTopLeftCorner) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(calculatedStyleFloatToInt(margin.Left()), calculatedStyleFloatToInt(margin.Top()))
}

func (a *pageBoxTopLeftCorner) getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingTop()
	case LayerPagedModePagedModePrint:
		top = 0
	}
	return geom.NewPoint(additionalClearance, top)
}

type pageBoxTopRightCorner struct {
	pageBoxMarginArea
}

func newPageBoxTopRightCorner() *pageBoxTopRightCorner {
	return &pageBoxTopRightCorner{pageBoxMarginArea{[]*MarginBoxName{MarginBoxNameTopRightCorner}}}
}

func (a *pageBoxTopRightCorner) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(calculatedStyleFloatToInt(margin.Right()), calculatedStyleFloatToInt(margin.Top()))
}

func (a *pageBoxTopRightCorner) getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	left := additionalClearance + page.GetWidth(c) - calculatedStyleFloatToInt(page.GetMargin(c).Right())
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingTop()
	case LayerPagedModePagedModePrint:
		top = 0
	}
	return geom.NewPoint(left, top)
}

type pageBoxBottomRightCorner struct {
	pageBoxMarginArea
}

func newPageBoxBottomRightCorner() *pageBoxBottomRightCorner {
	return &pageBoxBottomRightCorner{pageBoxMarginArea{[]*MarginBoxName{MarginBoxNameBottomRightCorner}}}
}

func (a *pageBoxBottomRightCorner) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(calculatedStyleFloatToInt(margin.Right()), calculatedStyleFloatToInt(margin.Bottom()))
}

func (a *pageBoxBottomRightCorner) getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	left := additionalClearance + page.GetWidth(c) - calculatedStyleFloatToInt(page.GetMargin(c).Right())
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingBottom() - calculatedStyleFloatToInt(page.GetMargin(c).Bottom())
	case LayerPagedModePagedModePrint:
		top = page.GetHeight(c) - calculatedStyleFloatToInt(page.GetMargin(c).Bottom())
	}
	return geom.NewPoint(left, top)
}

type pageBoxBottomLeftCorner struct {
	pageBoxMarginArea
}

func newPageBoxBottomLeftCorner() *pageBoxBottomLeftCorner {
	return &pageBoxBottomLeftCorner{pageBoxMarginArea{[]*MarginBoxName{MarginBoxNameBottomLeftCorner}}}
}

func (a *pageBoxBottomLeftCorner) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(calculatedStyleFloatToInt(margin.Left()), calculatedStyleFloatToInt(margin.Bottom()))
}

func (a *pageBoxBottomLeftCorner) getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingBottom() - calculatedStyleFloatToInt(page.GetMargin(c).Bottom())
	case LayerPagedModePagedModePrint:
		top = page.GetHeight(c) - calculatedStyleFloatToInt(page.GetMargin(c).Bottom())
	}
	return geom.NewPoint(additionalClearance, top)
}

type pageBoxLeftMarginArea struct {
	pageBoxMarginArea
}

func newPageBoxLeftMarginArea() *pageBoxLeftMarginArea {
	return &pageBoxLeftMarginArea{pageBoxMarginArea{[]*MarginBoxName{
		MarginBoxNameLeftTop,
		MarginBoxNameLeftMiddle,
		MarginBoxNameLeftBottom}}}
}

func (a *pageBoxLeftMarginArea) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(calculatedStyleFloatToInt(margin.Left()), page.GetContentHeight(c))
}

func (a *pageBoxLeftMarginArea) getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingTop() + calculatedStyleFloatToInt(page.GetMargin(c).Top())
	case LayerPagedModePagedModePrint:
		top = calculatedStyleFloatToInt(page.GetMargin(c).Top())
	}
	return geom.NewPoint(additionalClearance, top)
}

func (a *pageBoxLeftMarginArea) getDirection() BoxBuilderMarginDirection {
	return BoxBuilderMarginDirectionVertical
}

type pageBoxRightMarginArea struct {
	pageBoxMarginArea
}

func newPageBoxRightMarginArea() *pageBoxRightMarginArea {
	return &pageBoxRightMarginArea{pageBoxMarginArea{[]*MarginBoxName{
		MarginBoxNameRightTop,
		MarginBoxNameRightMiddle,
		MarginBoxNameRightBottom}}}
}

// getLayoutDimension uses the left margin for the width, as the Java method
// does.
func (a *pageBoxRightMarginArea) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(calculatedStyleFloatToInt(margin.Left()), page.GetContentHeight(c))
}

func (a *pageBoxRightMarginArea) getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	left := additionalClearance + page.GetWidth(c) - calculatedStyleFloatToInt(page.GetMargin(c).Right())
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingTop() + calculatedStyleFloatToInt(page.GetMargin(c).Top())
	case LayerPagedModePagedModePrint:
		top = calculatedStyleFloatToInt(page.GetMargin(c).Top())
	}
	return geom.NewPoint(left, top)
}

func (a *pageBoxRightMarginArea) getDirection() BoxBuilderMarginDirection {
	return BoxBuilderMarginDirectionVertical
}

type pageBoxTopMarginArea struct {
	pageBoxMarginArea
}

func newPageBoxTopMarginArea() *pageBoxTopMarginArea {
	return &pageBoxTopMarginArea{pageBoxMarginArea{[]*MarginBoxName{
		MarginBoxNameTopLeft,
		MarginBoxNameTopCenter,
		MarginBoxNameTopRight}}}
}

func (a *pageBoxTopMarginArea) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(page.GetContentWidth(c), calculatedStyleFloatToInt(margin.Top()))
}

func (a *pageBoxTopMarginArea) getPaintingPosition(c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	left := additionalClearance + calculatedStyleFloatToInt(page.GetMargin(c).Left())
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingTop()
	case LayerPagedModePagedModePrint:
		top = 0
	}
	return geom.NewPoint(left, top)
}

type pageBoxBottomMarginArea struct {
	pageBoxMarginArea
}

func newPageBoxBottomMarginArea() *pageBoxBottomMarginArea {
	return &pageBoxBottomMarginArea{pageBoxMarginArea{[]*MarginBoxName{
		MarginBoxNameBottomLeft,
		MarginBoxNameBottomCenter,
		MarginBoxNameBottomRight}}}
}

func (a *pageBoxBottomMarginArea) getLayoutDimension(c CssContext, page *PageBox, margin RectPropertySetI) *geom.Dimension {
	return geom.NewDimension(page.GetContentWidth(c), calculatedStyleFloatToInt(margin.Bottom()))
}

func (a *pageBoxBottomMarginArea) getPaintingPosition(
	c *RenderingContext, page *PageBox, additionalClearance int, mode LayerPagedMode) *geom.Point {
	left := additionalClearance + calculatedStyleFloatToInt(page.GetMargin(c).Left())
	var top int
	switch mode {
	case LayerPagedModePagedModeScreen:
		top = page.GetPaintingBottom() - calculatedStyleFloatToInt(page.GetMargin(c).Bottom())
	case LayerPagedModePagedModePrint:
		top = page.GetHeight(c) - calculatedStyleFloatToInt(page.GetMargin(c).Bottom())
	}
	return geom.NewPoint(left, top)
}
