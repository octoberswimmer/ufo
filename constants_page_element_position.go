// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/constants/PageElementPosition.java

package ufo

// PageElementPosition ports a Java enum with a field, so the constants are
// pointers.
type PageElementPosition struct {
	ident string
}

var (
	PageElementPositionStart      = &PageElementPosition{ident: "start"}
	PageElementPositionFirst      = &PageElementPosition{ident: "first"}
	PageElementPositionLast       = &PageElementPosition{ident: "last"}
	PageElementPositionLastExcept = &PageElementPosition{ident: "last-except"}
)

var pageElementPositionAll = map[string]*PageElementPosition{
	PageElementPositionStart.ident:      PageElementPositionStart,
	PageElementPositionFirst.ident:      PageElementPositionFirst,
	PageElementPositionLast.ident:       PageElementPositionLast,
	PageElementPositionLastExcept.ident: PageElementPositionLastExcept,
}

// PageElementPositionByIdent may return nil.
func PageElementPositionByIdent(ident string) *PageElementPosition {
	return pageElementPositionAll[ident]
}

func (p *PageElementPosition) ToString() string {
	return p.ident
}

func (p *PageElementPosition) String() string {
	return p.ident
}
