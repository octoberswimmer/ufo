// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/style/CssKnowledge.java

package ufo

func cssKnowledgeSetOf(values ...*IdentValue) map[*IdentValue]struct{} {
	result := make(map[*IdentValue]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func cssKnowledgeContains(set map[*IdentValue]struct{}, value *IdentValue) bool {
	_, ok := set[value]
	return ok
}

var cssKnowledgeMarginsNotAllowed = cssKnowledgeSetOf(
	IdentValueTableHeaderGroup, IdentValueTableRowGroup, IdentValueTableFooterGroup,
	IdentValueTableRow, IdentValueTableCell,
)

var cssKnowledgeBordersNotAllowed = cssKnowledgeSetOf(
	IdentValueTableHeaderGroup, IdentValueTableRowGroup, IdentValueTableFooterGroup,
	IdentValueTableRow,
)

var cssKnowledgeOverflowApplicable = cssKnowledgeSetOf(
	IdentValueBlock, IdentValueListItem,
	IdentValueTable, IdentValueInlineBlock, IdentValueTableCell,
)

var cssKnowledgeMayHaveFirstLine = cssKnowledgeSetOf(
	IdentValueBlock, IdentValueListItem, IdentValueRunIn,
	IdentValueTable, IdentValueTableCell, IdentValueTableCaption,
	IdentValueInlineBlock,
)

var cssKnowledgeMayHaveFirstLetter = cssKnowledgeSetOf(
	IdentValueBlock, IdentValueListItem,
	IdentValueTableCell, IdentValueTableCaption,
	IdentValueInlineBlock,
)

var cssKnowledgeBlockEquivalents = cssKnowledgeSetOf(
	IdentValueBlock, IdentValueListItem,
	IdentValueRunIn, IdentValueInlineBlock,
	IdentValueTable, IdentValueInlineTable,
)

var cssKnowledgeLaidOutInInlineContext = cssKnowledgeSetOf(
	IdentValueInline, IdentValueInlineBlock, IdentValueInlineTable,
)

var cssKnowledgeTableSections = cssKnowledgeSetOf(
	IdentValueTableRowGroup, IdentValueTableHeaderGroup, IdentValueTableFooterGroup,
)

var cssKnowledgeUnderTableLayout = cssKnowledgeSetOf(
	IdentValueTableRowGroup, IdentValueTableHeaderGroup, IdentValueTableFooterGroup,
	IdentValueTableRow, IdentValueTableCell,
	IdentValueTableCaption, IdentValueTableColumn, IdentValueTableColumnGroup,
)
