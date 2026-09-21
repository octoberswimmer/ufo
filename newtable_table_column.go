// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/newtable/TableColumn.java

package ufo

import "github.com/octoberswimmer/ufo/dom"

// TableColumn is an object representing an element with display: table-column or
// display: table-column-group.
type TableColumn struct {
	element *dom.Element
	style   CalculatedStyleI

	parent *TableColumn
}

func NewTableColumn() *TableColumn {
	return &TableColumn{}
}

func NewTableColumnWithElementStyle(element *dom.Element, style CalculatedStyleI) *TableColumn {
	return &TableColumn{element: element, style: style}
}

func (t *TableColumn) GetElement() *dom.Element {
	return t.element
}

// GetPseudoElementOrClass returns "" (Java returns null).
func (t *TableColumn) GetPseudoElementOrClass() string {
	return ""
}

// GetStyle may return a nil interface.
func (t *TableColumn) GetStyle() CalculatedStyleI {
	return t.style
}

func (t *TableColumn) SetElement(e *dom.Element) {
	t.element = e
}

func (t *TableColumn) SetStyle(style CalculatedStyleI) {
	t.style = style
}

func (t *TableColumn) GetParent() *TableColumn {
	return t.parent
}

func (t *TableColumn) SetParent(parent *TableColumn) {
	t.parent = parent
}
