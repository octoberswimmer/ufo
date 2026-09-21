// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/context/StyleReference.java

package ufo

import (
	"fmt"
	"time"

	"github.com/octoberswimmer/ufo/dom"
)

type StyleReference struct {
	nsh               NamespaceHandler
	doc               *dom.Document
	stylesheetFactory *StylesheetFactoryImpl

	// matcher is the instance of our element-styles matching class. Will be
	// nil if new rules have been added since last match.
	matcher *Matcher

	uac UserAgentCallback
}

func NewStyleReference(userAgent UserAgentCallback) *StyleReference {
	return &StyleReference{
		uac:               userAgent,
		stylesheetFactory: NewStylesheetFactoryImpl(userAgent),
	}
}

// SetDocumentContext sets the documentContext attribute of the StyleReference
// object.
//
// context is the Context this StyleReference operates in; used for property
// resolution.
func (s *StyleReference) SetDocumentContext(context *SharedContext, nsh NamespaceHandler, doc *dom.Document, ui UserInterface) {
	s.nsh = nsh
	s.doc = doc
	var attRes AttributeResolver = NewStandardAttributeResolver(s.nsh, s.uac, ui)

	infos := s.getStylesheets()
	XRLogMatch("media = " + context.GetMedia())
	s.matcher = NewMatcher(
		NewDOMTreeResolver(),
		attRes,
		s.stylesheetFactory,
		s.readAndParseAll(infos, context.GetMedia()),
		context.GetMedia())
}

func (s *StyleReference) readAndParseAll(infos []*StylesheetInfo, medium string) []*Stylesheet {
	result := make([]*Stylesheet, 0, len(infos)+15)
	for _, info := range infos {
		if info.AppliesToMedia(medium) {
			sheet := s.stylesheetFactory.GetStylesheet(info)

			if sheet != nil {
				if len(sheet.GetImportRules()) != 0 {
					result = append(result, s.readAndParseAll(sheet.GetImportRules(), medium)...)
				}

				result = append(result, sheet)
			} else {
				XRLogLoad(LevelWarning, "Unable to load CSS from "+info.GetUri())
			}
		}
	}

	return result
}

func (s *StyleReference) IsHoverStyled(e *dom.Element) bool {
	return s.matcher.IsHoverStyled(e)
}

// GetCascadedPropertiesMap returns a map keyed by CSS property names (e.g.
// 'border-width'), and the assigned value. The properties should have been
// matched to the element when the Context was established for this
// StyleReference on the Document to which the Element belongs.
//
// Java returns a LinkedHashMap in the order of
// CascadedStyle.GetCascadedPropertyDeclarations; the Go map has no order, and
// a caller that needs it iterates the declarations of GetCascadedStyle.
func (s *StyleReference) GetCascadedPropertiesMap(e *dom.Element) map[string]*PropertyValue {
	cs := s.matcher.GetCascadedStyle(e, false) //this is only for debug, I think
	props := make(map[string]*PropertyValue)
	for _, pd := range cs.GetCascadedPropertyDeclarations() {
		propName := pd.GetPropertyName()
		cssName := CSSNameCssProperty(propName)
		props[propName] = cs.PropertyByName(cssName).GetValue()
	}
	return props
}

// GetPseudoElementStyle gets the pseudoElementStyle attribute of the
// StyleReference object. It returns nil when no style applies.
func (s *StyleReference) GetPseudoElementStyle(node dom.Node, pseudoElement string) *CascadedStyle {
	var e *dom.Element
	if node.GetNodeType() == dom.ElementNode {
		e = node.(*dom.Element)
	} else {
		e = node.GetParentNode().(*dom.Element)
	}
	return s.matcher.GetPECascadedStyle(e, pseudoElement)
}

// GetCascadedStyle gets the CascadedStyle for an element. This must then be
// converted in the current context to a CalculatedStyle (use
// getDerivedStyle). e may be nil.
func (s *StyleReference) GetCascadedStyle(e *dom.Element, restyle bool) *CascadedStyle {
	if e == nil {
		return CascadedStyleEmptyCascadedStyle
	}
	return s.matcher.GetCascadedStyle(e, restyle)
}

// GetPageStyle returns the style of the page; pageName may be "".
func (s *StyleReference) GetPageStyle(pageName string, pseudoPage string) *PageInfo {
	return s.matcher.GetPageCascadedStyle(pageName, pseudoPage)
}

// FlushStyleSheets flushes any stylesheet associated with this style
// reference (based on the user agent callback) that are in cache.
func (s *StyleReference) FlushStyleSheets() {
	uri := s.uac.GetBaseURL()

	if s.stylesheetFactory.ContainsStylesheet(uri) {
		s.stylesheetFactory.RemoveCachedStylesheet(uri)
		XRLogCssParse("Removing stylesheet '" + uri + "' from cache by request.")
	} else {
		XRLogCssParse("Requested removing stylesheet '" + uri + "', but it's not in cache.")
	}
}

func (s *StyleReference) FlushAllStyleSheets() {
	s.stylesheetFactory.FlushCachedStylesheets()
}

// getStylesheets gets StylesheetInfos for all stylesheets and inline styles
// associated with the current document. Default (user agent) stylesheet and
// the inline style for the current media are loaded and cached in the
// StyleSheetFactory by URI.
func (s *StyleReference) getStylesheets() []*StylesheetInfo {
	var infos []*StylesheetInfo
	st := time.Now().UnixMilli()

	if defaultStylesheet := s.nsh.GetDefaultStylesheet(); defaultStylesheet != nil {
		infos = append(infos, defaultStylesheet)
	}
	infos = append(infos, s.nsh.GetStylesheets(s.doc)...)

	// TODO: here we should also get user stylesheet from userAgent

	XRLogLoad(fmt.Sprintf("TIME: parse stylesheets in %d ms.", time.Now().UnixMilli()-st))

	return infos
}

func (s *StyleReference) GetFontFaceRules() []*FontFaceRule {
	return s.matcher.GetFontFaceRules()
}

func (s *StyleReference) SetUserAgentCallback(userAgentCallback UserAgentCallback) {
	s.uac = userAgentCallback
	s.stylesheetFactory.SetUserAgentCallback(userAgentCallback)
}

func (s *StyleReference) SetSupportCMYKColors(b bool) {
	s.stylesheetFactory.SetSupportCMYKColors(b)
}

// GetUnsupportedCssFeatures returns the unsupported CSS features the parser
// reported, in the order they were first reported.
func (s *StyleReference) GetUnsupportedCssFeatures() []string {
	return s.stylesheetFactory.GetUnsupportedCssFeatures()
}
