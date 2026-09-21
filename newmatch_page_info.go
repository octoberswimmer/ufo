// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/newmatch/PageInfo.java

package ufo

type PageInfo struct {
	properties  []*PropertyDeclaration
	pageStyle   *CascadedStyle
	marginBoxes map[*MarginBoxName][]*PropertyDeclaration

	// xmpPropertyList is nil when the page has no -fs-pdf-xmp-metadata box.
	xmpPropertyList []*PropertyDeclaration
}

// NewPageInfo removes the -fs-pdf-xmp-metadata entry from marginBoxes, as the
// Java constructor does.
func NewPageInfo(properties []*PropertyDeclaration, pageStyle *CascadedStyle, marginBoxes map[*MarginBoxName][]*PropertyDeclaration) *PageInfo {
	p := &PageInfo{
		properties:  properties,
		pageStyle:   pageStyle,
		marginBoxes: marginBoxes,
	}
	p.xmpPropertyList = marginBoxes[MarginBoxNameFsPdfXmpMetadata]
	delete(marginBoxes, MarginBoxNameFsPdfXmpMetadata)
	return p
}

func (p *PageInfo) GetMarginBoxes() map[*MarginBoxName][]*PropertyDeclaration {
	return p.marginBoxes
}

func (p *PageInfo) GetPageStyle() *CascadedStyle {
	return p.pageStyle
}

func (p *PageInfo) GetProperties() []*PropertyDeclaration {
	return p.properties
}

func (p *PageInfo) CreateMarginBoxStyle(marginBox *MarginBoxName, alwaysCreate bool) *CascadedStyle {
	marginProps := p.marginBoxes[marginBox]

	if len(marginProps) == 0 && !alwaysCreate {
		return nil
	}

	all := make([]*PropertyDeclaration, 0, len(marginProps)+3)
	all = append(all, marginProps...)

	all = append(all, CascadedStyleCreateLayoutPropertyDeclaration(CSSNameDisplay, IdentValueTableCell))
	all = append(all, NewPropertyDeclaration(
		CSSNameVerticalAlign,
		NewPropertyValueIdentValue(marginBox.GetInitialVerticalAlign()),
		false,
		StylesheetInfoOriginUserAgent))
	all = append(all, NewPropertyDeclaration(
		CSSNameTextAlign,
		NewPropertyValueIdentValue(marginBox.GetInitialTextAlign()),
		false,
		StylesheetInfoOriginUserAgent))

	return NewCascadedStyle(all)
}

func (p *PageInfo) HasAny(marginBoxes []*MarginBoxName) bool {
	for _, marginBox := range marginBoxes {
		if _, ok := p.marginBoxes[marginBox]; ok {
			return true
		}
	}

	return false
}

func (p *PageInfo) GetXMPPropertyList() []*PropertyDeclaration {
	return p.xmpPropertyList
}
