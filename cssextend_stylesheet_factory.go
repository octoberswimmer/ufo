// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/extend/StylesheetFactory.java

package ufo

import "io"

// StylesheetFactory is a factory for Cascading Style Sheets. Sheets are
// parsed using a single parser instance for all sheets. Sheets are cached by
// URI using an LRU test, but timestamp of file is not checked.
type StylesheetFactory interface {
	// Parse is parse(Reader, StylesheetInfo).
	Parse(reader io.Reader, info *StylesheetInfo) *Stylesheet

	// ParseWithUriOrigin is parse(Reader, String uri, Origin origin).
	ParseWithUriOrigin(reader io.Reader, uri string, origin StylesheetInfoOrigin) *Stylesheet

	ParseStyleDeclaration(origin StylesheetInfoOrigin, style string) *Ruleset

	GetStylesheet(si *StylesheetInfo) *Stylesheet
}
