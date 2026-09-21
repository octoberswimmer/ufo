// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/BoxBuilder.java
//
// The slf4j debug call of the Java class is not ported.

package ufo

import (
	"strconv"
	"strings"

	"github.com/octoberswimmer/ufo/dom"
)

// BoxBuilder is responsible for creating the box tree from the DOM.  This is
// mostly just a one-to-one translation from the Element to an
// InlineBox or a BlockBox (or some subclass of
// BlockBox), but the tree is reorganized according to the CSS rules.
// This includes inserting anonymous block and inline boxes, anonymous table
// content, and :before and :after content.  White
// space is also normalized at this point.  Table columns and table column groups
// are added to the table which owns them, but are not created as regular boxes.
// Floated and absolutely positioned content is always treated as inline
// content for purposes of inserting anonymous block boxes and calculating
// the kind of content contained in a given block box.
//
// The Java class has only static methods; they are the BoxBuilder... and
// boxBuilder... functions of this file. A Java method that adds to or removes
// from a list it was given takes a pointer to the slice or returns the new
// slice.

// boxBuilderUnsupportedHtml5Tags holds the HTML5 semantic/structural elements
// not supported by FlyingSaucer (HTML4/CSS2 only).
// Encountering these in input HTML may produce incorrect or unexpected rendering.
var boxBuilderUnsupportedHtml5Tags = map[string]struct{}{
	"article": {}, "aside": {}, "audio": {}, "canvas": {}, "datalist": {}, "details": {}, "dialog": {},
	"figcaption": {}, "figure": {}, "footer": {}, "header": {}, "main": {}, "mark": {}, "meter": {},
	"nav": {}, "output": {}, "picture": {}, "progress": {}, "section": {}, "summary": {},
	"template": {}, "time": {}, "track": {}, "video": {}, "wbr": {},
}

type BoxBuilderMarginDirection int

const (
	BoxBuilderMarginDirectionVertical BoxBuilderMarginDirection = iota
	BoxBuilderMarginDirectionHorizontal
)

const (
	boxBuilderContentListDocument  = 1
	boxBuilderContentListMarginBox = 2
)

func BoxBuilderCreateRootBox(c *LayoutContext, document *dom.Document) BlockBoxI {
	root := document.GetDocumentElement()

	style := c.GetSharedContext().GetStyle(root)

	var result BlockBoxI
	if style.IsTable() || style.IsInlineTable() {
		result = NewTableBox(root, style, false)
	} else {
		result = NewBlockBoxWithElementStyleAnonymous(root, style, false)
	}

	c.ResolveCounters(style)

	c.PushLayerBox(result)
	if c.IsPrint() {
		if !style.IsIdent(CSSNamePage, IdentValueAuto) {
			c.SetPageName(style.GetStringProperty(CSSNamePage))
		}
		c.GetRootLayer().AddPage(c)
	}

	return result
}

func BoxBuilderCreateChildren(c *LayoutContext, parent BlockBoxI) {

	var children []Styleable

	info := NewBoxBuilderChildBoxInfo()

	boxBuilderCreateChildrenWithBlockParentChildrenInfoInline(c, parent, parent.GetElement(), &children, info, false)

	parentIsNestingTableContent := boxBuilderIsNestingTableContent(parent.GetStyle().GetIdent(
		CSSNameDisplay))
	if !parentIsNestingTableContent && !info.IsContainsTableContent() {
		boxBuilderResolveChildren(parent, children, info)
	} else {
		children = boxBuilderStripAllWhitespace(children)
		if parentIsNestingTableContent {
			boxBuilderResolveTableContent(c, parent, children, info)
		} else {
			boxBuilderResolveChildTableContent(c, parent, children, info, IdentValueTableCell)
		}
	}
}

// BoxBuilderCreateMarginTable may return nil.
func BoxBuilderCreateMarginTable(
	c *LayoutContext,
	pageInfo *PageInfo,
	names []*MarginBoxName,
	height int,
	direction BoxBuilderMarginDirection) *TableBox {
	if !pageInfo.HasAny(names) {
		return nil
	}

	source := c.GetRootLayer().GetMaster().GetElement() // HACK

	info := NewBoxBuilderChildBoxInfo()
	pageStyle := NewEmptyStyle().DeriveStyle(pageInfo.GetPageStyle())

	tableStyle := pageStyle.DeriveStyle(
		CascadedStyleCreateLayoutStyle(
			NewPropertyDeclaration(
				CSSNameDisplay,
				NewPropertyValueIdentValue(IdentValueTable),
				true,
				StylesheetInfoOriginUser),
			NewPropertyDeclaration(
				CSSNameWidth,
				NewPropertyValueFloat(CSSPrimitiveValueCssPercentage, 100.0, "100%"),
				true,
				StylesheetInfoOriginUser),
		))
	result := boxBuilderCreateBlockBox(source, tableStyle, info, false, true).(*TableBox)
	result.SetMarginAreaRoot(true)
	result.SetChildrenContentType(BlockBoxContentTypeBlock)

	tableSectionStyle := pageStyle.CreateAnonymousStyle(IdentValueTableRowGroup)
	section := boxBuilderCreateBlockBox(source, tableSectionStyle, info, false, true).(*TableSectionBox)
	section.SetChildrenContentType(BlockBoxContentTypeBlock)

	result.AddChild(section)

	var row *TableRowBox
	if direction == BoxBuilderMarginDirectionHorizontal {
		tableRowStyle := pageStyle.CreateAnonymousStyle(IdentValueTableRow)
		row = boxBuilderCreateBlockBox(source, tableRowStyle, info, false, true).(*TableRowBox)
		row.SetChildrenContentType(BlockBoxContentTypeBlock)

		row.SetHeightOverride(height)

		section.AddChild(row)
	}

	cellCount := 0
	alwaysCreate := len(names) > 1 && direction == BoxBuilderMarginDirectionHorizontal

	for _, name := range names {
		cellStyle := pageInfo.CreateMarginBoxStyle(name, alwaysCreate)
		if cellStyle != nil {
			cell := boxBuilderCreateMarginBox(c, cellStyle, alwaysCreate)
			if cell != nil {
				if direction == BoxBuilderMarginDirectionVertical {
					tableRowStyle := pageStyle.CreateAnonymousStyle(IdentValueTableRow)
					row = boxBuilderCreateBlockBox(source, tableRowStyle, info, false, true).(*TableRowBox)
					row.SetChildrenContentType(BlockBoxContentTypeBlock)
					row.SetHeightOverride(height)
					section.AddChild(row)
				}
				row.AddChild(cell)
				cellCount++
			}
		}
	}

	if direction == BoxBuilderMarginDirectionVertical && cellCount > 0 {
		rHeight := 0
		for _, box := range section.GetChildren() {
			r := box.(*TableRowBox)
			r.SetHeightOverride(height / cellCount)
			rHeight += r.GetHeightOverride()
		}

		sectionChildren := section.GetChildren()
		for i := 0; i < len(sectionChildren) && rHeight < height; i++ {
			r := sectionChildren[i].(*TableRowBox)
			r.SetHeightOverride(r.GetHeightOverride() + 1)
			rHeight++
		}
	}

	if cellCount > 0 {
		return result
	}
	return nil
}

// boxBuilderCreateMarginBox may return nil.
func boxBuilderCreateMarginBox(
	c *LayoutContext,
	cascadedStyle *CascadedStyle,
	alwaysCreate bool) *TableCellBox {
	hasContent := true

	contentDecl := cascadedStyle.PropertyByName(CSSNameContent)

	style := NewEmptyStyle().DeriveStyle(cascadedStyle)

	if style.IsDisplayNone() && !alwaysCreate {
		return nil
	}

	if style.IsIdent(CSSNameContent, IdentValueNone) ||
		style.IsIdent(CSSNameContent, IdentValueNormal) {
		hasContent = false
	}

	if style.IsAutoWidth() && !alwaysCreate && !hasContent {
		return nil
	}

	var children []Styleable

	info := NewBoxBuilderChildBoxInfoWithLayoutRunningBlocks(true)
	info.setContainsTableContent()

	element := c.GetRootLayer().GetMaster().GetElement() // XXX Doesn't make sense, but we need something here
	result := NewTableCellBox(element, style, true)

	if hasContent && !style.IsDisplayNone() {
		children = append(children, boxBuilderCreateGeneratedMarginBoxContent(
			c,
			c.GetRootLayer().GetMaster().GetElement(),
			contentDecl.GetValue(),
			style,
			info)...)

		children = boxBuilderStripAllWhitespace(children)
	}

	if len(children) == 0 && style.IsAutoWidth() && !alwaysCreate {
		return nil
	}

	boxBuilderResolveChildTableContent(c, result, children, info, IdentValueTableCell)

	return result
}

func boxBuilderResolveChildren(owner BlockBoxI, children []Styleable, info *BoxBuilderChildBoxInfo) {
	if len(children) != 0 {
		if info.IsContainsBlockLevelContent() {
			boxBuilderInsertAnonymousBlocks(owner, children, info.IsLayoutRunningBlocks())
			owner.SetChildrenContentType(BlockBoxContentTypeBlock)
		} else {
			children = WhitespaceStripperStripInlineContent(children)
			if len(children) != 0 {
				owner.SetInlineContent(children)
				owner.SetChildrenContentType(BlockBoxContentTypeInline)
			} else {
				owner.SetChildrenContentType(BlockBoxContentTypeEmpty)
			}
		}
	} else {
		owner.SetChildrenContentType(BlockBoxContentTypeEmpty)
	}
}

func boxBuilderIsAllProperTableNesting(parentDisplay *IdentValue, children []Styleable) bool {
	for _, child := range children {
		if !boxBuilderIsProperTableNesting(parentDisplay, child.GetStyle().GetIdent(CSSNameDisplay)) {
			return false
		}
	}

	return true
}

// boxBuilderResolveChildTableContent handles the situation when we find table
// content, but our parent is not table related.  For example, div -> td.
// Anonymous tables are then constructed by repeatedly pulling together
// consecutive same-table-level siblings and wrapping them in the next
// highest table level (e.g. consecutive td elements will
// be wrapped in an anonymous tr, then a tbody, and
// finally a table).
func boxBuilderResolveChildTableContent(
	c *LayoutContext, parent BlockBoxI, children []Styleable, info *BoxBuilderChildBoxInfo, target *IdentValue) {
	var childrenForAnonymous []Styleable
	var childrenWithAnonymous []Styleable

	nextUp := boxBuilderGetPreviousTableNestingLevel(target)
	for _, styleable := range children {
		if boxBuilderMatchesTableLevel(target, styleable.GetStyle().GetIdent(CSSNameDisplay)) {
			childrenForAnonymous = append(childrenForAnonymous, styleable)
		} else {
			if len(childrenForAnonymous) != 0 {
				boxBuilderCreateAnonymousTableContent(c, childrenForAnonymous[0].(BlockBoxI), nextUp,
					childrenForAnonymous, &childrenWithAnonymous)

				childrenForAnonymous = nil
			}
			childrenWithAnonymous = append(childrenWithAnonymous, styleable)
		}
	}

	if len(childrenForAnonymous) != 0 {
		boxBuilderCreateAnonymousTableContent(c, childrenForAnonymous[0].(BlockBoxI), nextUp,
			childrenForAnonymous, &childrenWithAnonymous)
	}

	if nextUp == IdentValueTable {
		boxBuilderRebalanceInlineContent(childrenWithAnonymous)
		info.setContainsBlockLevelContent()
		boxBuilderResolveChildren(parent, childrenWithAnonymous, info)
	} else {
		boxBuilderResolveChildTableContent(c, parent, childrenWithAnonymous, info, nextUp)
	}
}

func boxBuilderMatchesTableLevel(target *IdentValue, value *IdentValue) bool {
	if target == IdentValueTableRowGroup {
		return boxBuilderIsTableGroupOrCaption(value)
	} else {
		return target == value
	}
}

// boxBuilderRebalanceInlineContent makes sure that any InlineBox in content
// both starts and ends within content. Used to ensure that
// it is always possible to construct anonymous blocks once an element's
// children has been distributed among anonymous table objects.
func boxBuilderRebalanceInlineContent(content []Styleable) {
	boxesByElement := map[*dom.Element]*InlineBox{}
	for _, styleable := range content {
		if iB, ok := styleable.(*InlineBox); ok {
			elem := iB.GetElement()

			if _, found := boxesByElement[elem]; !found {
				iB.SetStartsHere(true)
			}

			boxesByElement[elem] = iB
		}
	}

	for _, iB := range boxesByElement {
		iB.SetEndsHere(true)
	}
}

// boxBuilderStripSubList is
// WhitespaceStripper.stripInlineContent(content.subList(start, end)): the
// elements the stripper removes from the sublist are removed from content.
func boxBuilderStripSubList(content []Styleable, start int, end int) []Styleable {
	subList := make([]Styleable, end-start)
	copy(subList, content[start:end])
	stripped := WhitespaceStripperStripInlineContent(subList)
	if len(stripped) == len(subList) {
		return content
	}
	result := make([]Styleable, 0, len(content)-len(subList)+len(stripped))
	result = append(result, content[:start]...)
	result = append(result, stripped...)
	result = append(result, content[end:]...)
	return result
}

func boxBuilderStripAllWhitespace(content []Styleable) []Styleable {
	start := 0
	var current int
	started := false
	for current = 0; current < len(content); current++ {
		styleable := content[current]
		if !styleable.GetStyle().IsLaidOutInInlineContext() {
			if started {
				before := len(content)
				content = boxBuilderStripSubList(content, start, current)
				after := len(content)
				current -= before - after
			}
			started = false
		} else {
			if !started {
				started = true
				start = current
			}
		}
	}

	if started {
		content = boxBuilderStripSubList(content, start, current)
	}
	return content
}

// boxBuilderResolveTableContent handles the situation when our current parent
// is table related.  If everything is properly nested (e.g. a tr contains only
// td elements), nothing is done.  Otherwise, anonymous boxes
// are inserted to ensure the integrity of the table model.
func boxBuilderResolveTableContent(
	c *LayoutContext, parent BlockBoxI, children []Styleable, info *BoxBuilderChildBoxInfo) {
	parentDisplay := parent.GetStyle().GetIdent(CSSNameDisplay)
	next := boxBuilderGetNextTableNestingLevel(parentDisplay)
	if next == nil && parent.IsAnonymous() && boxBuilderContainsOrphanedTableContent(children) {
		boxBuilderResolveChildTableContent(c, parent, children, info, IdentValueTableCell)
	} else if next == nil || boxBuilderIsAllProperTableNesting(parentDisplay, children) {
		if parent.IsAnonymous() {
			boxBuilderRebalanceInlineContent(children)
		}
		boxBuilderResolveChildren(parent, children, info)
	} else {
		var childrenForAnonymous []Styleable
		var childrenWithAnonymous []Styleable
		for _, child := range children {
			childDisplay := child.GetStyle().GetIdent(CSSNameDisplay)

			if boxBuilderIsProperTableNesting(parentDisplay, childDisplay) {
				if len(childrenForAnonymous) != 0 {
					boxBuilderCreateAnonymousTableContent(c, parent, next, childrenForAnonymous,
						&childrenWithAnonymous)

					childrenForAnonymous = nil
				}
				childrenWithAnonymous = append(childrenWithAnonymous, child)
			} else {
				childrenForAnonymous = append(childrenForAnonymous, child)
			}
		}

		if len(childrenForAnonymous) != 0 {
			boxBuilderCreateAnonymousTableContent(c, parent, next, childrenForAnonymous,
				&childrenWithAnonymous)
		}

		info.setContainsBlockLevelContent()
		boxBuilderResolveChildren(parent, childrenWithAnonymous, info)
	}
}

func boxBuilderContainsOrphanedTableContent(children []Styleable) bool {
	for _, child := range children {
		display := child.GetStyle().GetIdent(CSSNameDisplay)
		if boxBuilderIsTableGroup(display) || display == IdentValueTableRow {
			return true
		}
	}

	return false
}

func boxBuilderIsParentInline(box BlockBoxI) bool {
	parentStyle := box.GetStyle().GetParent()
	return parentStyle != nil && parentStyle.IsInline()
}

func boxBuilderCreateAnonymousTableContent(c *LayoutContext, source BlockBoxI,
	next *IdentValue,
	childrenForAnonymous []Styleable,
	childrenWithAnonymous *[]Styleable) {
	nested := boxBuilderLookForBlockContent(childrenForAnonymous)
	var anonDisplay *IdentValue
	if boxBuilderIsParentInline(source) && next == IdentValueTable {
		anonDisplay = IdentValueInlineTable
	} else {
		anonDisplay = next
	}
	anonStyle := source.GetStyle().CreateAnonymousStyle(anonDisplay)
	element := source.GetElement() // XXX Doesn't really make sense, but what to do?
	anonBox := boxBuilderCreateBlockBox(element, anonStyle, nested, false, true)
	boxBuilderResolveTableContent(c, anonBox, childrenForAnonymous, nested)

	if next == IdentValueTable {
		*childrenWithAnonymous = append(*childrenWithAnonymous, boxBuilderReorderTableContent(c, anonBox.(*TableBox)))
	} else {
		*childrenWithAnonymous = append(*childrenWithAnonymous, anonBox)
	}
}

// boxBuilderReorderTableContent reorganizes a table so that the header is the
// first row group and the footer the last.  If the table has caption boxes,
// they will be pulled out and added to an anonymous block box along with the
// table itself. If not, the table is returned.
func boxBuilderReorderTableContent(c *LayoutContext, table *TableBox) BlockBoxI {
	var topCaptions []BoxI
	var header BoxI
	var bodies []BoxI
	var footer BoxI
	var bottomCaptions []BoxI

	for _, b := range table.GetChildren() {
		display := b.GetStyle().GetIdent(CSSNameDisplay)
		if display == IdentValueTableCaption {
			side := b.GetStyle().GetIdent(CSSNameCaptionSide)
			if side == IdentValueBottom {
				bottomCaptions = append(bottomCaptions, b)
			} else { /* side == IdentValue.TOP */
				topCaptions = append(topCaptions, b)
			}
		} else if display == IdentValueTableHeaderGroup && header == nil {
			header = b
		} else if display == IdentValueTableFooterGroup && footer == nil {
			footer = b
		} else {
			bodies = append(bodies, b)
		}
	}

	table.RemoveAllChildren()
	if header != nil {
		header.(*TableSectionBox).SetHeader(true)
		table.AddChild(header)
	}
	table.AddAllChildren(bodies)
	if footer != nil {
		footer.(*TableSectionBox).SetFooter(true)
		table.AddChild(footer)
	}

	if len(topCaptions) == 0 && len(bottomCaptions) == 0 {
		return table
	} else {
		// If we have a floated table with a caption, we need to float the
		// outer anonymous box and not the table
		var anonStyle CalculatedStyleI
		if table.GetStyle().IsFloated() {
			cascadedStyle := CascadedStyleCreateLayoutStyle(
				CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueBlock),
				CascadedStyleCreateLayoutPropertyDeclaration(CSSNameFloat, table.GetStyle().GetIdent(CSSNameFloat)))

			anonStyle = table.GetStyle().DeriveStyle(cascadedStyle)
		} else {
			anonStyle = table.GetStyle().CreateAnonymousStyle(IdentValueBlock)
		}

		anonBox := NewBlockBoxWithElementStyleAnonymous(table.GetElement(), anonStyle, true)
		anonBox.SetFromCaptionedTable(true)

		anonBox.SetChildrenContentType(BlockBoxContentTypeBlock)
		anonBox.AddAllChildren(topCaptions)
		anonBox.AddChild(table)
		anonBox.AddAllChildren(bottomCaptions)

		if table.GetStyle().IsFloated() {
			anonBox.SetFloatedBoxData(NewFloatedBoxData())
			table.SetFloatedBoxData(nil)

			original := c.GetSharedContext().GetCss().GetCascadedStyle(
				table.GetElement(), false)
			modified := CascadedStyleCreateLayoutStyleWithStartingPoint(
				original,
				[]*PropertyDeclaration{
					CascadedStyleCreateLayoutPropertyDeclaration(
						CSSNameFloat, IdentValueNone),
				})
			table.SetStyle(table.GetStyle().GetParent().DeriveStyle(modified))
		}

		return anonBox
	}
}

func boxBuilderLookForBlockContent(styleables []Styleable) *BoxBuilderChildBoxInfo {
	result := NewBoxBuilderChildBoxInfo()
	for _, s := range styleables {
		if !s.GetStyle().IsLaidOutInInlineContext() {
			result.setContainsBlockLevelContent()
			break
		}
	}
	return result
}

// boxBuilderGetNextTableNestingLevel may return nil.
func boxBuilderGetNextTableNestingLevel(display *IdentValue) *IdentValue {
	if display == IdentValueTable || display == IdentValueInlineTable {
		return IdentValueTableRowGroup
	} else if boxBuilderIsTableGroup(display) {
		return IdentValueTableRow
	} else if display == IdentValueTableRow {
		return IdentValueTableCell
	} else {
		return nil
	}
}

// boxBuilderGetPreviousTableNestingLevel may return nil.
func boxBuilderGetPreviousTableNestingLevel(display *IdentValue) *IdentValue {
	if display == IdentValueTableCell {
		return IdentValueTableRow
	} else if display == IdentValueTableRow {
		return IdentValueTableRowGroup
	} else if boxBuilderIsTableGroup(display) {
		return IdentValueTable
	} else {
		return nil
	}
}

func boxBuilderIsProperTableNesting(parent *IdentValue, child *IdentValue) bool {
	return parent == IdentValueTable && boxBuilderIsTableGroupOrCaption(child) ||
		boxBuilderIsTableGroup(parent) && child == IdentValueTableRow ||
		parent == IdentValueTableRow && child == IdentValueTableCell ||
		parent == IdentValueInlineTable && boxBuilderIsTableGroup(child)
}

func boxBuilderIsTableGroup(ident *IdentValue) bool {
	return ident == IdentValueTableHeaderGroup ||
		ident == IdentValueTableRowGroup ||
		ident == IdentValueTableFooterGroup
}

func boxBuilderIsTableGroupOrCaption(child *IdentValue) bool {
	return boxBuilderIsTableGroup(child) || child == IdentValueTableCaption
}

func boxBuilderIsNestingTableContent(display *IdentValue) bool {
	return display == IdentValueTable || display == IdentValueInlineTable || display == IdentValueTableRow ||
		boxBuilderIsTableGroup(display)
}

func boxBuilderIsAttrFunction(function *FSFunction) bool {
	if function.Is("attr") {
		params := function.GetParameters()
		if len(params) == 1 {
			value := params[0]
			return value.GetPrimitiveType() == CSSPrimitiveValueCssIdent
		}
	}

	return false
}

func BoxBuilderIsElementFunction(function *FSFunction) bool {
	if function.Is("element") {
		params := function.GetParameters()
		if len(params) == 0 || len(params) > 2 {
			return false
		}
		value1 := params[0]
		ok := value1.GetPrimitiveType() == CSSPrimitiveValueCssIdent
		if ok && len(params) == 2 {
			value2 := params[1]
			ok = value2.GetPrimitiveType() == CSSPrimitiveValueCssIdent
		}

		return ok
	}

	return false
}

// boxBuilderMakeCounterFunction may return nil (a nil interface).
func boxBuilderMakeCounterFunction(function *FSFunction, c *LayoutContext, style CalculatedStyleI) CssFunction {
	if function.Is("counter") {
		params := function.GetParameters()
		if len(params) == 0 || len(params) > 2 {
			return nil
		}

		value := params[0]
		if value.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
			return nil
		}

		s := value.GetStringValue()
		// counter(page) and counter(pages) are handled separately
		if s == "page" || s == "pages" {
			return nil
		}

		counter := value.GetStringValue()
		listStyleType := IdentValueDecimal
		if len(params) == 2 {
			value = params[1]
			if value.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
				return nil
			}

			identValue := IdentValueValueOf(value.GetStringValue())
			if identValue != nil {
				value.SetIdentValue(identValue)
				listStyleType = identValue
			}
		}

		counterValue := c.GetCounterContext(style).GetCurrentCounterValue(counter)

		return NewCounterFunction(counterValue, listStyleType)
	} else if function.Is("counters") {
		params := function.GetParameters()
		if len(params) < 2 || len(params) > 3 {
			return nil
		}

		value := params[0]
		if value.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
			return nil
		}

		counter := value.GetStringValue()

		value = params[1]
		if value.GetPrimitiveType() != CSSPrimitiveValueCssString {
			return nil
		}

		separator := value.GetStringValue()

		listStyleType := IdentValueDecimal
		if len(params) == 3 {
			value = params[2]
			if value.GetPrimitiveType() != CSSPrimitiveValueCssIdent {
				return nil
			}

			identValue := IdentValueValueOf(value.GetStringValue())
			if identValue != nil {
				value.SetIdentValue(identValue)
				listStyleType = identValue
			}
		}

		counterValues := c.GetCounterContext(style).GetCurrentCounterValues(counter)

		return NewCountersFunction(counterValues, separator, listStyleType)
	} else {
		return nil
	}
}

func boxBuilderGetAttributeValue(attrFunc *FSFunction, e *dom.Element) string {
	value := attrFunc.GetParameters()[0]
	return e.GetAttribute(value.GetStringValue())
}

// boxBuilderCreateGeneratedContentList: peName is "" and info is nil where
// Java passes null.
func boxBuilderCreateGeneratedContentList(
	c *LayoutContext, element *dom.Element, propValue *PropertyValue,
	peName string, style CalculatedStyleI, mode int, info *BoxBuilderChildBoxInfo) []Styleable {
	values := propValue.GetValues()

	result := make([]Styleable, 0, len(values))

	for _, v := range values {
		value := v.(*PropertyValue)
		var contentFunction ContentFunction
		var function *FSFunction

		// Java's String content is null until one of the branches sets it;
		// hasContent is that null test. An attr() of a missing attribute sets
		// it to the empty string.
		content := ""
		hasContent := false

		primitiveType := value.GetPrimitiveType()
		if primitiveType == CSSPrimitiveValueCssString {
			content = value.GetStringValue()
			hasContent = true
		} else if value.GetPropertyValueType() == PropertyValueTypeValueTypeFunction {
			if mode == boxBuilderContentListDocument && boxBuilderIsAttrFunction(value.GetFunction()) {
				content = boxBuilderGetAttributeValue(value.GetFunction(), element)
				hasContent = true
			} else {
				var cFunc CssFunction
				switch mode {
				case boxBuilderContentListDocument:
					cFunc = boxBuilderMakeCounterFunction(value.GetFunction(), c, style)
				default:
					cFunc = nil
				}

				if cFunc != nil {
					//TODO: counter functions may be called with non-ordered list-style-types, e.g. disc
					content = cFunc.Evaluate()
					hasContent = true
				} else if mode == boxBuilderContentListMarginBox && BoxBuilderIsElementFunction(value.GetFunction()) {
					target := BoxBuilderGetRunningBlock(c, value)
					if target != nil {
						result = append(result, target.CopyOf())
						info.setContainsBlockLevelContent()
					}
				} else {
					contentFunction =
						c.GetContentFunctionFactory().LookupFunction(c, value.GetFunction())
					if contentFunction != nil {
						function = value.GetFunction()

						// ContentFunction returns a Go string, so a Java null
						// result of either call cannot be told from "".
						if contentFunction.IsStatic() {
							content = contentFunction.Calculate(c, function)
							hasContent = true
							contentFunction = nil
							function = nil
						} else {
							content = contentFunction.GetLayoutReplacementText()
							hasContent = true
						}
					}
				}
			}
		} else if primitiveType == CSSPrimitiveValueCssIdent {
			dv := style.ValueByName(CSSNameQuotes)

			if dv != FSDerivedValue(IdentValueNone) {
				ident := value.GetIdentValue()

				if ident == IdentValueOpenQuote {
					quotes := style.AsStringArray(CSSNameQuotes)
					content = quotes[0]
					hasContent = true
				} else if ident == IdentValueCloseQuote {
					quotes := style.AsStringArray(CSSNameQuotes)
					content = quotes[1]
					hasContent = true
				}
			}
		}

		if hasContent {
			iB := NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClass(content, nil, contentFunction, function, element, peName)
			iB.SetStartsHere(true)
			iB.SetEndsHere(true)

			result = append(result, iB)
		}
	}

	return result
}

// BoxBuilderGetRunningBlock may return nil (a nil interface).
func BoxBuilderGetRunningBlock(c *LayoutContext, value *PropertyValue) BlockBoxI {
	params := value.GetFunction().GetParameters()
	ident := params[0].GetStringValue()
	var position *PageElementPosition
	if len(params) == 2 {
		position = PageElementPositionByIdent(params[1].GetStringValue())
	}
	if position == nil {
		position = PageElementPositionFirst
	}
	return c.GetRootDocumentLayer().GetRunningBlock(ident, c.GetPage(), position)
}

func boxBuilderInsertGeneratedContent(
	c *LayoutContext, element *dom.Element, parentStyle CalculatedStyleI,
	peName string, children *[]Styleable, info *BoxBuilderChildBoxInfo) {
	peStyle := c.GetCss().GetPseudoElementStyle(element, peName)
	if peStyle != nil {
		contentDecl := peStyle.PropertyByName(CSSNameContent)
		counterResetDecl := peStyle.PropertyByName(CSSNameCounterReset)
		counterIncrDecl := peStyle.PropertyByName(CSSNameCounterIncrement)

		var calculatedStyle CalculatedStyleI
		if contentDecl != nil || counterResetDecl != nil || counterIncrDecl != nil {
			calculatedStyle = parentStyle.DeriveStyle(peStyle)
			if calculatedStyle.IsDisplayNone() {
				return
			}
			if calculatedStyle.IsIdent(CSSNameContent, IdentValueNone) {
				return
			}
			if calculatedStyle.IsIdent(CSSNameContent, IdentValueNormal) && (peName == "before" || peName == "after") {
				return
			}

			if calculatedStyle.IsTable() || calculatedStyle.IsTableRow() || calculatedStyle.IsTableSection() {
				newPeStyle :=
					CascadedStyleCreateLayoutStyleWithStartingPoint(peStyle, []*PropertyDeclaration{
						CascadedStyleCreateLayoutPropertyDeclaration(
							CSSNameDisplay,
							IdentValueBlock),
					})
				calculatedStyle = parentStyle.DeriveStyle(newPeStyle)
			}
			c.ResolveCounters(calculatedStyle)
		}

		if contentDecl != nil {
			propValue := contentDecl.GetValue()
			*children = append(*children, boxBuilderCreateGeneratedContent(c, element, peName, calculatedStyle,
				propValue, info)...)
		}
	}
}

func boxBuilderCreateGeneratedContent(
	c *LayoutContext, element *dom.Element, peName string,
	style CalculatedStyleI, property *PropertyValue, info *BoxBuilderChildBoxInfo) []Styleable {
	if style.IsDisplayNone() || style.IsIdent(CSSNameDisplay, IdentValueTableColumn) ||
		style.IsIdent(CSSNameDisplay, IdentValueTableColumnGroup) {
		return nil
	}

	inlineBoxes := boxBuilderCreateGeneratedContentList(
		c, element, property, peName, style, boxBuilderContentListDocument, nil)

	if style.IsInline() {
		for _, inlineBox := range inlineBoxes {
			iB := inlineBox.(*InlineBox)
			iB.SetStyle(style)
			iB.ApplyTextTransform()
		}
		return inlineBoxes
	} else {
		anon := style.CreateAnonymousStyle(IdentValueInline)
		for _, inlineBox := range inlineBoxes {
			iB := inlineBox.(*InlineBox)
			iB.SetStyle(anon)
			iB.ApplyTextTransform()
			iB.SetElement(nil)
		}

		result := boxBuilderCreateBlockBox(element, style, info, true, false)
		result.SetInlineContent(inlineBoxes)
		result.SetChildrenContentType(BlockBoxContentTypeInline)
		result.SetPseudoElementOrClass(peName)

		if !style.IsLaidOutInInlineContext() {
			info.setContainsBlockLevelContent()
		}

		return []Styleable{result}
	}
}

func boxBuilderCreateGeneratedMarginBoxContent(
	c *LayoutContext, element *dom.Element, property *PropertyValue,
	style CalculatedStyleI, info *BoxBuilderChildBoxInfo) []Styleable {
	result := boxBuilderCreateGeneratedContentList(
		c, element, property, "", style, boxBuilderContentListMarginBox, info)

	anon := style.CreateAnonymousStyle(IdentValueInline)
	for _, s := range result {
		if iB, ok := s.(*InlineBox); ok {
			iB.SetElement(nil)
			iB.SetStyle(anon)
			iB.ApplyTextTransform()
		}
	}

	return result
}

func boxBuilderCreateBlockBox(
	source *dom.Element, style CalculatedStyleI, info *BoxBuilderChildBoxInfo, generated bool, anonymous bool) BlockBoxI {
	if style.IsFloated() && !(style.IsAbsolute() || style.IsFixed()) {
		var result BlockBoxI
		if style.IsTable() || style.IsInlineTable() {
			result = NewTableBox(source, style, anonymous)
		} else {
			result = NewBlockBoxWithElementStyleAnonymous(source, style, anonymous)
		}
		result.SetFloatedBoxData(NewFloatedBoxData())
		return result
	} else if style.IsSpecifiedAsBlock() {
		return NewBlockBoxWithElementStyleAnonymous(source, style, anonymous)
	} else if !generated && (style.IsTable() || style.IsInlineTable()) {
		return NewTableBox(source, style, anonymous)
	} else if style.IsTableCell() {
		info.setContainsTableContent()
		return NewTableCellBox(source, style, anonymous)
	} else if !generated && style.IsTableRow() {
		info.setContainsTableContent()
		return NewTableRowBox(source, style, anonymous)
	} else if !generated && style.IsTableSection() {
		info.setContainsTableContent()
		return NewTableSectionBox(source, style, anonymous)
	} else if style.IsTableCaption() {
		info.setContainsTableContent()
		return NewBlockBoxWithElementStyleAnonymous(source, style, anonymous)
	} else {
		return NewBlockBoxWithElementStyleAnonymous(source, style, anonymous)
	}
}

func boxBuilderAddColumns(c *LayoutContext, table *TableBox, parent *TableColumn) {
	sharedContext := c.GetSharedContext()

	working := parent.GetElement().GetFirstChild()
	found := false
	for working != nil {
		if working.GetNodeType() == dom.ElementNode {
			element := working.(*dom.Element)
			style := sharedContext.GetStyle(element)

			if style.IsIdent(CSSNameDisplay, IdentValueTableColumn) {
				found = true
				col := NewTableColumnWithElementStyle(element, style)
				col.SetParent(parent)
				table.AddStyleColumn(col)
			}
		}
		working = working.GetNextSibling()
	}
	if !found {
		table.AddStyleColumn(parent)
	}
}

func boxBuilderAddColumnOrColumnGroup(
	c *LayoutContext, table *TableBox, e *dom.Element, style CalculatedStyleI) {
	if style.IsIdent(CSSNameDisplay, IdentValueTableColumn) {
		table.AddStyleColumn(NewTableColumnWithElementStyle(e, style))
	} else { /* style.isIdent(CSSName.DISPLAY, IdentValue.TABLE_COLUMN_GROUP) */
		boxBuilderAddColumns(c, table, NewTableColumnWithElementStyle(e, style))
	}
}

// boxBuilderCreateInlineBox: node may be nil.
func boxBuilderCreateInlineBox(
	text string, parent *dom.Element, parentStyle CalculatedStyleI, node *dom.Text) *InlineBox {

	var result *InlineBox
	_, parentIsRoot := parent.GetParentNode().(*dom.Document)
	if parentStyle.IsInline() && !parentIsRoot {
		result = NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClassStyle(text, node, nil, nil, parent, "", parentStyle)
	} else {
		result = NewInlineBoxWithContentFunctionFunctionElementPseudoElementOrClassStyle(text, node, nil, nil, nil, "", parentStyle.CreateAnonymousStyle(IdentValueInline))
	}
	result.ApplyTextTransform()

	return result
}

// boxBuilderCreateChildrenWithBlockParentChildrenInfoInline is the private
// six-argument overload of createChildren. blockParent is a nil interface
// where Java passes null.
func boxBuilderCreateChildrenWithBlockParentChildrenInfoInline(
	c *LayoutContext, blockParent BlockBoxI, parent *dom.Element,
	children *[]Styleable, info *BoxBuilderChildBoxInfo, inline bool) {
	sharedContext := c.GetSharedContext()

	parentStyle := sharedContext.GetStyle(parent)

	boxBuilderInsertGeneratedContent(c, parent, parentStyle, "before", children, info)

	needStartText := inline
	needEndText := inline
	var previousIB *InlineBox
	// Java's do/while evaluates its condition, which advances working to the
	// next sibling, after a continue as well; the post statement does the same.
	for working := parent.GetFirstChild(); working != nil; working = working.GetNextSibling() {
		var child Styleable
		nodeType := working.GetNodeType()
		if nodeType == dom.ElementNode {
			element := working.(*dom.Element)
			boxBuilderCheckForUnsupportedTags(element, sharedContext)

			style := sharedContext.GetStyle(element)
			if style.IsDisplayNone() {
				continue
			}

			c.ResolveCountersWithStartIndex(style, boxBuilderParseStartIndex(working))

			if style.IsIdent(CSSNameDisplay, IdentValueTableColumn) ||
				style.IsIdent(CSSNameDisplay, IdentValueTableColumnGroup) {
				if blockParent != nil &&
					(blockParent.GetStyle().IsTable() || blockParent.GetStyle().IsInlineTable()) {
					table := blockParent.(*TableBox)
					boxBuilderAddColumnOrColumnGroup(c, table, element, style)
				}

				continue
			}

			if style.IsInline() {
				if needStartText {
					needStartText = false
					iB := boxBuilderCreateInlineBox("", parent, parentStyle, nil)
					iB.SetStartsHere(true)
					iB.SetEndsHere(false)
					*children = append(*children, iB)
					previousIB = iB
				}
				boxBuilderCreateChildrenWithBlockParentChildrenInfoInline(c, nil, element, children, info, true)
				if inline {
					if previousIB != nil {
						previousIB.SetEndsHere(false)
					}
					needEndText = true
				}
			} else {
				block := boxBuilderCreateBlockBox(element, style, info, false, false)
				if style.IsListItem() {
					block.SetListCounter(c.GetCounterContext(style).GetCurrentCounterValue("list-item"))
				}

				if style.IsTable() || style.IsInlineTable() {
					table := block.(*TableBox)
					table.EnsureChildren(c)

					block = boxBuilderReorderTableContent(c, table)
				}

				if !info.IsContainsBlockLevelContent() &&
					!style.IsLaidOutInInlineContext() {
					info.setContainsBlockLevelContent()
				}

				if block.GetStyle().MayHaveFirstLine() {
					block.SetFirstLineStyle(c.GetCss().GetPseudoElementStyle(element, "first-line"))
				}
				if block.GetStyle().MayHaveFirstLetter() {
					block.SetFirstLetterStyle(c.GetCss().GetPseudoElementStyle(element, "first-letter"))
				}
				//I think we need to do this to evaluate counters correctly
				block.EnsureChildren(c)
				child = block
			}
		} else if nodeType == dom.TextNode || nodeType == dom.CDATASectionNode {
			needStartText = false
			needEndText = false

			textNode := working.(*dom.Text)

			iB := boxBuilderCreateInlineBox(textNode.GetData(), parent, parentStyle, textNode)
			child = iB

			iB.SetEndsHere(true)
			if previousIB == nil {
				iB.SetStartsHere(true)
			} else {
				previousIB.SetEndsHere(false)
			}
			previousIB = iB
		} else if nodeType == dom.EntityReferenceNode {
			// The dom package has no entity reference type (its parsers expand
			// entities); the text comes from the Node interface.
			iB := boxBuilderCreateInlineBox(working.GetTextContent(), parent, parentStyle, nil)
			child = iB

			iB.SetEndsHere(true)
			if previousIB == nil {
				iB.SetStartsHere(true)
			} else {
				previousIB.SetEndsHere(false)
			}
			previousIB = iB
		}

		if child != nil {
			*children = append(*children, child)
		}
	}
	if needStartText || needEndText {
		iB := boxBuilderCreateInlineBox("", parent, parentStyle, nil)
		iB.SetStartsHere(needStartText)
		iB.SetEndsHere(needEndText)
		*children = append(*children, iB)
	}
	boxBuilderInsertGeneratedContent(c, parent, parentStyle, "after", children, info)
}

func boxBuilderCheckForUnsupportedTags(element *dom.Element, sharedContext *SharedContext) {
	tagName := element.GetLocalName()
	if tagName == "" {
		tagName = element.GetTagName()
	}
	if tagName != "" {
		if _, unsupported := boxBuilderUnsupportedHtml5Tags[strings.ToLower(tagName)]; unsupported {
			sharedContext.AddUnsupportedTag(tagName)
		}
	}
}

// boxBuilderParseStartIndex may return nil.
func boxBuilderParseStartIndex(node dom.Node) *int {
	switch node.GetNodeName() {
	case "ol":
		return boxBuilderParseAttribute(node, "start")
	case "li":
		return boxBuilderParseAttribute(node, "value")
	default:
		return nil
	}
}

// boxBuilderParseAttribute may return nil. The value is parsed as
// Integer.parseInt does: an optional sign and decimal digits within the int
// range, and the subtraction wraps as Java's int does.
func boxBuilderParseAttribute(node dom.Node, attributeName string) *int {
	element, ok := node.(*dom.Element)
	if ok && element.HasAttribute(attributeName) {
		attributeValue := element.GetAttribute(attributeName)
		parsed, err := strconv.ParseInt(attributeValue, 10, 32)
		if err == nil {
			result := int(int32(parsed) - 1)
			return &result
		}
	}
	return nil
}

func boxBuilderInsertAnonymousBlocks(parent BoxI, children []Styleable, layoutRunningBlocks bool) {
	var inline []Styleable

	var parents []*InlineBox
	var savedParents []*InlineBox

	for _, child := range children {
		if child.GetStyle().IsLaidOutInInlineContext() &&
			!(layoutRunningBlocks && child.GetStyle().IsRunning()) {
			inline = append(inline, child)

			if child.GetStyle().IsInline() {
				iB := child.(*InlineBox)
				if iB.IsStartsHere() {
					parents = append(parents, iB)
				}
				if iB.IsEndsHere() {
					if len(parents) == 0 {
						// Deque.removeLast on an empty deque
						panic(NewXRRuntimeException("NoSuchElementException"))
					}
					parents = parents[:len(parents)-1]
				}
			}
		} else {
			if len(inline) != 0 {
				boxBuilderCreateAnonymousBlock(parent, inline, savedParents)
				inline = nil
				savedParents = make([]*InlineBox, len(parents))
				copy(savedParents, parents)
			}
			parent.AddChild(child.(BoxI))
		}
	}

	boxBuilderCreateAnonymousBlock(parent, inline, savedParents)
}

func boxBuilderCreateAnonymousBlock(parent BoxI, inline []Styleable, savedParents []*InlineBox) {
	inline = WhitespaceStripperStripInlineContent(inline)
	if len(inline) != 0 {
		anonymousBox := NewAnonymousBlockBox(parent.GetElement(),
			parent.GetStyle().CreateAnonymousStyle(IdentValueBlock),
			savedParents, inline,
		)

		parent.AddChild(anonymousBox)
	}
}

type BoxBuilderChildBoxInfo struct {
	containsBlockLevelContent bool
	containsTableContent      bool
	layoutRunningBlocks       bool
}

func NewBoxBuilderChildBoxInfo() *BoxBuilderChildBoxInfo {
	return NewBoxBuilderChildBoxInfoWithLayoutRunningBlocks(false)
}

func NewBoxBuilderChildBoxInfoWithLayoutRunningBlocks(layoutRunningBlocks bool) *BoxBuilderChildBoxInfo {
	return &BoxBuilderChildBoxInfo{layoutRunningBlocks: layoutRunningBlocks}
}

func (i *BoxBuilderChildBoxInfo) IsContainsBlockLevelContent() bool {
	return i.containsBlockLevelContent
}

func (i *BoxBuilderChildBoxInfo) setContainsBlockLevelContent() {
	i.containsBlockLevelContent = true
}

func (i *BoxBuilderChildBoxInfo) IsContainsTableContent() bool {
	return i.containsTableContent
}

func (i *BoxBuilderChildBoxInfo) setContainsTableContent() {
	i.containsTableContent = true
}

func (i *BoxBuilderChildBoxInfo) IsLayoutRunningBlocks() bool {
	return i.layoutRunningBlocks
}
