// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/CSSParser.java
//
// The Java methods throw CSSParseException and catch it to skip the rule or
// declaration that is wrong. Here the exception is a panic with a
// *CSSParseException, and cssParserTry stands where a Java try/catch block is.
//
// Every Java parsing method declares IOException, which only Lexer.yylex
// raises. Here next() panics with a *cssParserIOException and the exported
// methods recover it: ParseStylesheet returns the error, ParseDeclaration and
// ParsePropertyValue panic with an XRRuntimeException as the Java methods
// throw a RuntimeException.
//
// The slf4j trace and debug calls of the Java class are not ported.

package ufo

import (
	"io"
	"net/url"
	"strconv"
	"strings"
)

type CSSParser struct {
	saved               *Token
	lexer               *Lexer
	errorHandler        CSSErrorHandler
	css3FeatureListener func(string)
	// uri is "" where Java has null. No code path treats a null URI and an
	// empty URI differently.
	uri string

	// namespaces maps a lower case namespace prefix to its URI. Java keeps
	// the default namespace under the null key of the same map; here it is
	// defaultNamespace, which is nil when no default namespace is declared.
	namespaces        map[string]string
	defaultNamespace  *string
	supportCMYKColors bool
}

// cssParserIOException carries an I/O error of the lexer up to the exported
// method that started the parse.
type cssParserIOException struct {
	err error
}

// cssParserTry runs body. When body panics with a *CSSParseException it
// passes the exception to handler, as a Java catch (CSSParseException e)
// block. Any other panic continues.
func cssParserTry(body func(), handler func(e *CSSParseException)) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(*CSSParseException)
			if !ok {
				panic(r)
			}
			handler(e)
		}
	}()
	body()
}

func NewCSSParser(errorHandler CSSErrorHandler) *CSSParser {
	return NewCSSParserWithCss3FeatureListener(errorHandler, func(feature string) {})
}

func NewCSSParserWithCss3FeatureListener(errorHandler CSSErrorHandler, css3FeatureListener func(string)) *CSSParser {
	return &CSSParser{
		lexer:               NewLexer(strings.NewReader("")),
		errorHandler:        errorHandler,
		css3FeatureListener: css3FeatureListener,
		namespaces:          make(map[string]string),
	}
}

func (p *CSSParser) ParseStylesheet(uri string, origin StylesheetInfoOrigin, reader io.Reader) (result *Stylesheet, err error) {
	defer func() {
		if r := recover(); r != nil {
			ioe, ok := r.(*cssParserIOException)
			if !ok {
				panic(r)
			}
			result = nil
			err = ioe.err
		}
	}()

	p.uri = uri
	p.Reset(reader)

	result = NewStylesheet(uri, origin)
	p.stylesheet(result)

	return result, nil
}

func (p *CSSParser) ParseDeclaration(origin StylesheetInfoOrigin, text string) *Ruleset {
	defer cssParserRethrowIOException()

	// XXX Set this to something more reasonable
	p.uri = "style attribute"
	p.Reset(strings.NewReader(text))

	p.skipWhitespace()

	result := NewRuleset(origin)

	cssParserTry(func() {
		p.declarationList(result, true, false, false)
	}, func(e *CSSParseException) {
		// ignore, already handled
	})

	return result
}

// cssParserRethrowIOException is deferred by the exported methods that have
// no error result. "Shouldn't" happen.
func cssParserRethrowIOException() {
	if r := recover(); r != nil {
		ioe, ok := r.(*cssParserIOException)
		if !ok {
			panic(r)
		}
		panic(NewXRRuntimeExceptionWithCause(ioe.err.Error(), ioe.err))
	}
}

// ParsePropertyValue returns nil when expr is not a valid value of cssName.
func (p *CSSParser) ParsePropertyValue(cssName *CSSName, origin StylesheetInfoOrigin, expr string) (result *PropertyValue) {
	defer cssParserRethrowIOException()

	p.uri = cssName.ToString() + " property value"
	cssParserTry(func() {
		p.Reset(strings.NewReader(expr))
		values := p.expr(
			cssName == CSSNameFontFamily ||
				cssName == CSSNameFontShorthand ||
				cssName == CSSNameFsPdfFontEncoding)

		builder := CSSNameGetPropertyBuilder(cssName)
		var props []*PropertyDeclaration
		cssParserTry(func() {
			props = builder.BuildDeclarations(cssName, values, origin, false)
		}, func(e *CSSParseException) {
			e.SetLine(p.getCurrentLine())
			panic(e)
		})

		if len(props) != 1 {
			panic(NewCSSParseException(
				"Builder created "+strconv.Itoa(len(props))+"properties, expected 1", p.getCurrentLine()))
		}

		result = props[0].GetValue()
	}, func(e *CSSParseException) {
		p.error(e, "property value", false)
		result = nil
	})
	return result
}

// stylesheet
//
//	: [ CHARSET_SYM S* STRING S* ';' ]?
//	  [S|CDO|CDC]* [ import [S|CDO|CDC]* ]*
//	  [ namespace [S|CDO|CDC]* ]*
//	  [ [ ruleset | media | page | font_face ] [S|CDO|CDC]* ]*
func (p *CSSParser) stylesheet(stylesheet *Stylesheet) {
	t := p.la()
	cssParserTry(func() {
		if t == TokenTkCharsetSym {
			cssParserTry(func() {
				t = p.next()
				p.skipWhitespace()
				t = p.next()
				if t == TokenTkString {
					/* String charset = getTokenValue(t); */

					p.skipWhitespace()
					t = p.next()
					if t != TokenTkSemicolon {
						p.push(t)
						panic(NewCSSParseExceptionToken(t, TokenTkSemicolon, p.getCurrentLine()))
					}

					// Do something
				} else {
					p.push(t)
					panic(NewCSSParseExceptionToken(t, TokenTkString, p.getCurrentLine()))
				}
			}, func(e *CSSParseException) {
				p.error(e, "@charset rule", true)
				p.recover(false, false)
			})
		}
		p.skipWhitespaceAndCdocdc()
		for {
			t = p.la()
			if t == TokenTkImportSym {
				p.importRule(stylesheet)
				p.skipWhitespaceAndCdocdc()
			} else {
				break
			}
		}
		for {
			t = p.la()
			if t == TokenTkNamespaceSym {
				p.namespace()
				p.skipWhitespaceAndCdocdc()
			} else {
				break
			}
		}
		for {
			t = p.la()
			if t == TokenTkEof {
				break
			}
			switch t.GetType() {
			case TokenTypePageSym:
				p.page(stylesheet)
			case TokenTypeMediaSym:
				p.media(stylesheet)
			case TokenTypeFontFaceSym:
				p.fontFace(stylesheet)
			case TokenTypeImportSym:
				p.next()
				p.error(NewCSSParseException("@import not allowed here", p.getCurrentLine()),
					"@import rule", true)
				p.recover(false, false)
			case TokenTypeNamespaceSym:
				p.next()
				p.error(NewCSSParseException("@namespace not allowed here", p.getCurrentLine()),
					"@namespace rule", true)
				p.recover(false, false)
			case TokenTypeAtRule:
				p.next()
				p.error(NewCSSParseException(
					"Invalid at-rule", p.getCurrentLine()), "at-rule", true)
				p.recover(false, false)
				p.ruleset(stylesheet)
			default:
				p.ruleset(stylesheet)
			}
			p.skipWhitespaceAndCdocdc()
		}
	}, func(e *CSSParseException) {
		// "shouldn't" happen
		if !e.IsCallerNotified() {
			p.error(e, "stylesheet", false)
		}
	})
}

// import
//
//	: IMPORT_SYM S*
//	  [STRING|URI] S* [ medium [ COMMA S* medium]* ]? ';' S*
//	;
func (p *CSSParser) importRule(stylesheet *Stylesheet) {
	cssParserTry(func() {
		t := p.next()
		if t == TokenTkImportSym {
			var uri string
			mediaTypes := make([]string, 0, 1)

			p.skipWhitespace()
			t = p.next()
			switch t.GetType() {
			case TokenTypeString, TokenTypeUri:
				// java.net.URL resolves the import against the stylesheet
				// URI and, when the protocol is not one java.net.URL knows (a
				// custom protocol which the user expects to handle in the
				// user agent), java.net.URI does. net/url has no list of
				// known protocols, so one resolution covers both.
				tokenValue := p.getTokenValue(t)
				parent, err := url.Parse(stylesheet.GetURI())
				var child *url.URL
				if err == nil {
					child, err = url.Parse(tokenValue)
				}
				if err != nil {
					panic(NewCSSParseExceptionWithCause("Invalid URL, "+err.Error(), p.getCurrentLine(), err))
				}
				resolved := parent.ResolveReference(child)
				if child.Scheme == "" && resolved.Host == "" && resolved.User == nil {
					// Java writes a URL without an authority as file:/path,
					// however the stylesheet URI was written.
					resolved.OmitHost = true
				}
				uri = resolved.String()
				if parent.Opaque != "" {
					// java.net.URI resolves against an opaque URI (the
					// inline:N key of a <style> block) by returning the
					// given URI.
					uri = child.String()
				} else if parent.Scheme == "" && parent.Host == "" && child.Scheme == "" && child.Host == "" &&
					!strings.HasPrefix(parent.Path, "/") && !strings.HasPrefix(child.Path, "/") {
					// ResolveReference makes the path absolute.
					// java.net.URI keeps a relative path relative.
					uri = strings.TrimPrefix(uri, "/")
				}
				p.skipWhitespace()
				t = p.la()
				if t == TokenTkIdent {
					mediaTypes = append(mediaTypes, p.medium())
					for {
						t = p.la()
						if t == TokenTkComma {
							p.next()
							p.skipWhitespace()
							t = p.la()
							if t == TokenTkIdent {
								mediaTypes = append(mediaTypes, p.medium())
							} else {
								panic(NewCSSParseExceptionToken(
									t, TokenTkIdent, p.getCurrentLine()))
							}
						} else {
							break
						}
					}
				}
				t = p.next()
				if t == TokenTkSemicolon {
					p.skipWhitespace()
				} else {
					p.push(t)
					panic(NewCSSParseExceptionToken(
						t, TokenTkSemicolon, p.getCurrentLine()))
				}
			default:
				p.push(t)
				panic(NewCSSParseExceptionTokenArray(
					t, []*Token{TokenTkString, TokenTkUri}, p.getCurrentLine()))
			}

			if len(mediaTypes) == 0 {
				mediaTypes = append(mediaTypes, "all")
			}
			info := NewStylesheetInfo(stylesheet.GetOrigin(), uri, mediaTypes, nil)
			stylesheet.AddImportRule(info)
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkImportSym, p.getCurrentLine()))
		}
	}, func(e *CSSParseException) {
		p.error(e, "@import rule", true)
		p.recover(false, false)
	})
}

// namespace
//
//	: NAMESPACE_SYM S* [namespace_prefix S*]? [STRING|URI] S* ';' S*
//	;
//	namespace_prefix
//	: IDENT
//	;
func (p *CSSParser) namespace() {
	cssParserTry(func() {
		t := p.next()
		if t == TokenTkNamespaceSym {
			p.skipWhitespace()
			t = p.next()

			var prefix *string
			if t == TokenTkIdent {
				value := p.getTokenValue(t)
				prefix = &value
				p.skipWhitespace()
				t = p.next()
			}

			var url string
			if t == TokenTkString || t == TokenTkUri {
				url = p.getTokenValue(t)
			} else {
				panic(NewCSSParseExceptionTokenArray(
					t, []*Token{TokenTkString, TokenTkUri}, p.getCurrentLine()))
			}

			p.skipWhitespace()

			t = p.next()
			if t == TokenTkSemicolon {
				p.skipWhitespace()

				if prefix == nil {
					p.defaultNamespace = &url
				} else {
					p.namespaces[*prefix] = url
				}
			} else {
				panic(NewCSSParseExceptionToken(
					t, TokenTkSemicolon, p.getCurrentLine()))
			}
		} else {
			panic(NewCSSParseExceptionToken(t, TokenTkNamespaceSym, p.getCurrentLine()))
		}
	}, func(e *CSSParseException) {
		p.error(e, "@namespace rule", true)
		p.recover(false, false)
	})
}

// media
//
//	: MEDIA_SYM S* medium [ COMMA S* medium ]* LBRACE S* ruleset* '}' S*
//	;
func (p *CSSParser) media(stylesheet *Stylesheet) {
	t := p.next()
	cssParserTry(func() {
		if t == TokenTkMediaSym {
			mediaRule := NewMediaRule(stylesheet.GetOrigin())
			p.skipWhitespace()
			t = p.la()
			if t == TokenTkIdent {
				mediaRule.AddMedium(p.medium())
				for {
					t = p.la()
					if t == TokenTkComma {
						p.next()
						p.skipWhitespace()
						t = p.la()
						if t == TokenTkIdent {
							mediaRule.AddMedium(p.medium())
						} else {
							panic(NewCSSParseExceptionToken(t, TokenTkIdent, p.getCurrentLine()))
						}
					} else {
						break
					}
				}
				t = p.next()
				if t == TokenTkLbrace {
					p.skipWhitespace()
				LOOP:
					for {
						t = p.la()
						if t == nil {
							break
						}
						switch t.GetType() {
						case TokenTypeRbrace:
							p.next()
							break LOOP
						default:
							p.ruleset(mediaRule)
						}
					}
					p.skipWhitespace()
				} else {
					p.push(t)
					panic(NewCSSParseExceptionToken(t, TokenTkLbrace, p.getCurrentLine()))
				}
			} else {
				panic(NewCSSParseExceptionToken(t, TokenTkIdent, p.getCurrentLine()))
			}

			stylesheet.AddContentMediaRule(mediaRule)
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkMediaSym, p.getCurrentLine()))
		}
	}, func(e *CSSParseException) {
		p.error(e, "@media rule", true)
		p.recover(false, false)
	})
}

// medium
//
//	: IDENT S*
//	;
func (p *CSSParser) medium() string {
	t := p.next()
	if t == TokenTkIdent {
		result := p.getTokenValue(t)
		p.skipWhitespace()
		return result
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkIdent, p.getCurrentLine()))
	}
}

// font_face
//
//	: FONT_FACE_SYM S*
//	  '{' S* declaration [ ';' S* declaration ]* '}' S*
//	;
func (p *CSSParser) fontFace(stylesheet *Stylesheet) {
	t := p.next()
	cssParserTry(func() {
		fontFaceRule := NewFontFaceRule(stylesheet.GetOrigin())
		if t == TokenTkFontFaceSym {
			p.skipWhitespace()

			ruleset := NewRuleset(stylesheet.GetOrigin())

			p.skipWhitespace()
			t = p.next()
			if t == TokenTkLbrace {
				// Prevent runaway threads with a max loop/counter
				maxLoops := 1024 * 1024 // 1M is too much, 1K is probably too...
				i := 0
				for {
					i++
					if i >= maxLoops {
						panic(NewCSSParseExceptionToken(t, TokenTkRbrace, p.getCurrentLine()))
					}
					p.skipWhitespace()
					t = p.la()
					if t == TokenTkRbrace {
						p.next()
						p.skipWhitespace()
						break
					} else {
						p.declarationList(ruleset, false, true, true)
					}
				}
			} else {
				p.push(t)
				panic(NewCSSParseExceptionToken(t, TokenTkLbrace, p.getCurrentLine()))
			}

			fontFaceRule.AddContent(ruleset)
			stylesheet.AddFontFaceRule(fontFaceRule)
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkFontFaceSym, p.getCurrentLine()))
		}
	}, func(e *CSSParseException) {
		p.error(e, "@font-face rule", true)
		p.recover(false, false)
	})
}

// page
//
//	: PAGE_SYM S* IDENT? pseudo_page? S*
//	  '{' S* [ declaration | margin ]? [ ';' S* [ declaration | margin ]? ]* '}' S*
func (p *CSSParser) page(stylesheet *Stylesheet) {
	t := p.next()
	cssParserTry(func() {
		if t == TokenTkPageSym {
			pageName := ""
			pseudoPage := ""
			margins := make(map[*MarginBoxName][]*PropertyDeclaration)

			p.skipWhitespace()
			t = p.la()
			if t == TokenTkIdent {
				pageName = p.getTokenValue(t)
				if pageName == "auto" {
					panic(NewCSSParseException("page name may not be auto", p.getCurrentLine()))
				}
				p.next()
				t = p.la()
			}
			if t == TokenTkColon {
				pseudoPage = p.pseudoPage()
			}
			ruleset := NewRuleset(stylesheet.GetOrigin())

			p.skipWhitespace()
			t = p.next()
			if t == TokenTkLbrace {
				for {
					p.skipWhitespace()
					t = p.la()
					if t == TokenTkRbrace {
						p.next()
						p.skipWhitespace()
						break
					} else if t == TokenTkAtRule {
						for name, declarations := range p.margin(stylesheet) {
							margins[name] = declarations
						}
					} else {
						p.declarationList(ruleset, false, true, false)
					}
				}
			} else {
				p.push(t)
				panic(NewCSSParseExceptionToken(t, TokenTkLbrace, p.getCurrentLine()))
			}

			pageRule := NewPageRule(stylesheet.GetOrigin(), pageName, pseudoPage, margins, ruleset)
			stylesheet.AddContentPageRule(pageRule)
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkPageSym, p.getCurrentLine()))
		}
	}, func(e *CSSParseException) {
		p.error(e, "@page rule", true)
		p.recover(false, false)
	})
}

// margin
//
//	: margin_sym S* '{' declaration [ ';' S* declaration? ]* '}' S*
//	;
func (p *CSSParser) margin(stylesheet *Stylesheet) (result map[*MarginBoxName][]*PropertyDeclaration) {
	t := p.next()
	if t != TokenTkAtRule {
		p.error(NewCSSParseExceptionToken(t, TokenTkAtRule, p.getCurrentLine()), "at rule", true)
		p.recover(true, false)
		return map[*MarginBoxName][]*PropertyDeclaration{}
	}

	name := p.getTokenValue(t)
	marginBoxName := MarginBoxNameValueOf(name)
	if marginBoxName == nil {
		p.error(NewCSSParseException(name+" is not a valid margin box name", p.getCurrentLine()), "at rule", true)
		p.recover(true, false)
		return map[*MarginBoxName][]*PropertyDeclaration{}
	}

	p.skipWhitespace()
	result = map[*MarginBoxName][]*PropertyDeclaration{}
	cssParserTry(func() {
		t = p.next()
		if t == TokenTkLbrace {
			p.skipWhitespace()
			ruleset := NewRuleset(stylesheet.GetOrigin())
			p.declarationList(ruleset, false, false, false)
			t = p.next()
			if t != TokenTkRbrace {
				p.push(t)
				panic(NewCSSParseExceptionToken(t, TokenTkRbrace, p.getCurrentLine()))
			}
			result = map[*MarginBoxName][]*PropertyDeclaration{marginBoxName: ruleset.GetPropertyDeclarations()}
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkLbrace, p.getCurrentLine()))
		}
	}, func(e *CSSParseException) {
		p.error(e, "margin box", true)
		p.recover(false, false)
	})
	return result
}

// pseudo_page
//
//	: ':' IDENT
//	;
func (p *CSSParser) pseudoPage() string {
	t := p.next()
	if t == TokenTkColon {
		t = p.next()
		if t == TokenTkIdent {
			result := p.getTokenValue(t)
			if !(result == "first" || result == "left" || result == "right") {
				panic(NewCSSParseException("Pseudo page must be one of first, left, or right", p.getCurrentLine()))
			}
			return result
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkIdent, p.getCurrentLine()))
		}
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkColon, p.getCurrentLine()))
	}
}

// operator
//
//	: '/' S* | COMMA S* | /* empty */
//	;
func (p *CSSParser) operator() {
	t := p.la()
	switch t.GetType() {
	case TokenTypeVirgule, TokenTypeComma:
		p.next()
		p.skipWhitespace()
	}
}

// combinator
//
//	: PLUS S*
//	| GREATER S*
//	| S
//	;
func (p *CSSParser) combinator() *Token {
	t := p.next()
	if t == TokenTkPlus || t == TokenTkGreater {
		p.skipWhitespace()
	} else if t != TokenTkS {
		p.push(t)
		panic(NewCSSParseExceptionTokenArray(
			t,
			[]*Token{TokenTkPlus, TokenTkGreater, TokenTkS},
			p.getCurrentLine()))
	}
	return t
}

// unary_operator
//
//	: '-' | PLUS
//	;
func (p *CSSParser) unaryOperator() int {
	t := p.next()
	if t != TokenTkMinus && t != TokenTkPlus {
		p.push(t)
		panic(NewCSSParseExceptionTokenArray(
			t, []*Token{TokenTkMinus, TokenTkPlus}, p.getCurrentLine()))
	}
	if t == TokenTkMinus {
		return -1
	} else { /* t == Token.TK_PLUS */
		return 1
	}
}

// property
//
//	: IDENT S*
//	;
func (p *CSSParser) property() string {
	t := p.next()
	var result string
	if t == TokenTkIdent {
		result = p.getTokenValue(t)
		p.skipWhitespace()
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(
			t, TokenTkIdent, p.getCurrentLine()))
	}

	return result
}

// declaration_list
//
//	: [ declaration ';' S* ]*
func (p *CSSParser) declarationList(ruleset *Ruleset, expectEOF bool, expectAtRule bool, inFontFace bool) {
	var t *Token
LOOP:
	for {
		t = p.la()
		switch t.GetType() {
		case TokenTypeSemicolon:
			p.next()
			p.skipWhitespace()
		case TokenTypeRbrace:
			break LOOP
		case TokenTypeAtRule:
			if expectAtRule {
				break LOOP
			} else {
				p.declaration(ruleset, inFontFace)
			}
		case TokenTypeEof:
			if expectEOF {
				break LOOP
			}
			p.declaration(ruleset, inFontFace)
		default:
			p.declaration(ruleset, inFontFace)
		}
	}
}

// ruleset
//
//	: selector [ COMMA S* selector ]*
//	  LBRACE S* [ declaration ';' S* ]* '}' S*
//	;
func (p *CSSParser) ruleset(container RulesetContainer) {
	cssParserTry(func() {
		ruleset := NewRuleset(container.GetOrigin())

		p.selector(ruleset)
		var t *Token
		for {
			t = p.la()
			if t == TokenTkComma {
				p.next()
				p.skipWhitespace()
				p.selector(ruleset)
			} else {
				break
			}
		}
		t = p.next()
		if t == TokenTkLbrace {
			p.skipWhitespace()
			p.declarationList(ruleset, false, false, false)
			t = p.next()
			if t == TokenTkRbrace {
				p.skipWhitespace()
			} else {
				p.push(t)
				panic(NewCSSParseExceptionToken(t, TokenTkRbrace, p.getCurrentLine()))
			}
		} else {
			p.push(t)
			panic(NewCSSParseExceptionTokenArray(
				t, []*Token{TokenTkComma, TokenTkLbrace}, p.getCurrentLine()))
		}

		if len(ruleset.GetPropertyDeclarations()) != 0 {
			container.AddContent(ruleset)
		}
	}, func(e *CSSParseException) {
		p.error(e, "ruleset", true)
		p.recover(true, false)
	})
}

// selector
//
//	: simple_selector [ combinator simple_selector ]*
//	;
func (p *CSSParser) selector(ruleset *Ruleset) {
	var selectors []*Selector
	var combinators []*Token
	selectors = append(selectors, p.simpleSelector(ruleset))
LOOP:
	for {
		t := p.la()
		switch t.GetType() {
		case TokenTypePlus, TokenTypeGreater, TokenTypeS:
			combinators = append(combinators, p.combinator())
			t = p.la()
			switch t.GetType() {
			case TokenTypeIdent, TokenTypeAsterisk, TokenTypeHash, TokenTypePeriod, TokenTypeLbracket, TokenTypeColon:
				selectors = append(selectors, p.simpleSelector(ruleset))
			default:
				panic(NewCSSParseExceptionTokenArray(t, []*Token{TokenTkIdent,
					TokenTkAsterisk, TokenTkHash, TokenTkPeriod,
					TokenTkLbracket, TokenTkColon}, p.getCurrentLine()))
			}
		default:
			break LOOP
		}
	}
	ruleset.AddFSSelector(p.mergeSimpleSelectors(selectors, combinators))
}

func (p *CSSParser) mergeSimpleSelectors(selectors []*Selector, combinators []*Token) *Selector {
	count := len(selectors)
	if count == 1 {
		return selectors[0]
	}

	lastDescendantOrChildAxis := SelectorAxisDescendantAxis
	var result *Selector
	for i := 0; i < count-1; i++ {
		first := selectors[i]
		second := selectors[i+1]
		combinator := combinators[i]

		if first.GetPseudoElement() != "" {
			panic(NewCSSParseException(
				"A simple selector with a pseudo element cannot be "+
					"combined with another simple selector", p.getCurrentLine()))
		}

		sibling := false
		if combinator == TokenTkS {
			second.SetAxis(SelectorAxisDescendantAxis)
			lastDescendantOrChildAxis = SelectorAxisDescendantAxis
		} else if combinator == TokenTkGreater {
			second.SetAxis(SelectorAxisChildAxis)
			lastDescendantOrChildAxis = SelectorAxisChildAxis
		} else if combinator == TokenTkPlus {
			first.SetAxis(SelectorAxisImmediateSiblingAxis)
			sibling = true
		}

		second.SetSpecificityB(second.GetSpecificityB() + first.GetSpecificityB())
		second.SetSpecificityC(second.GetSpecificityC() + first.GetSpecificityC())
		second.SetSpecificityD(second.GetSpecificityD() + first.GetSpecificityD())

		if !sibling {
			if result == nil {
				result = first
			}
			first.SetChainedSelector(second)
		} else {
			second.SetSiblingSelector(first)
			if result == nil || result == first {
				result = second
			}
			if i > 0 {
				for j := i - 1; j >= 0; j-- {
					selector := selectors[j]
					if selector.GetChainedSelector() == first {
						selector.SetChainedSelector(second)
						second.SetAxis(lastDescendantOrChildAxis)
						break
					}
				}
			}
		}
	}

	return result
}

// simple_selector
//
//	: typed_value [ HASH | class | attrib | pseudo ]*
//	| [ HASH | class | attrib | pseudo ]+
//	;
func (p *CSSParser) simpleSelector(ruleset *Ruleset) *Selector {
	selector := NewSelector(ruleset)

	t := p.la()
	switch t.GetType() {
	case TokenTypeAsterisk,
		TokenTypeIdent,
		TokenTypeVerticalBar:
		pair := p.typedValue(false)
		selector.SetNamespaceURI(pair.namespaceURI)
		if pair.name == nil {
			selector.SetName("")
		} else {
			selector.SetName(*pair.name)
		}

	LOOP1:
		for {
			t = p.la()
			switch t.GetType() {
			case TokenTypeHash:
				t = p.next()
				selector.AddIDCondition(p.getTokenValueWithLiteral(t, true))
			case TokenTypePeriod:
				p.classSelector(selector)
			case TokenTypeLbracket:
				p.attrib(selector)
			case TokenTypeColon:
				p.pseudo(selector)
			default:
				break LOOP1
			}
		}
	default:
		found := false
	LOOP2:
		for {
			t = p.la()
			switch t.GetType() {
			case TokenTypeHash:
				t = p.next()
				selector.AddIDCondition(p.getTokenValueWithLiteral(t, true))
				found = true
			case TokenTypePeriod:
				p.classSelector(selector)
				found = true
			case TokenTypeLbracket:
				p.attrib(selector)
				found = true
			case TokenTypeColon:
				p.pseudo(selector)
				found = true
			default:
				if !found {
					panic(NewCSSParseExceptionTokenArray(t, []*Token{TokenTkHash,
						TokenTkPeriod, TokenTkLbracket, TokenTkColon},
						p.getCurrentLine()))
				}
				break LOOP2
			}
		}
	}
	return selector
}

// type_selector
//
//	: [ namespace_prefix ]? element_name | IDENT
//	;
//	namespace_prefix
//	: [ IDENT | '*' ]? '|'
//	;
func (p *CSSParser) typedValue(matchAttribute bool) *cssParserNamespacePair {
	var prefix *string
	var name *string

	t := p.la()
	if t == TokenTkAsterisk || t == TokenTkIdent {
		p.next()
		if t == TokenTkIdent {
			value := p.getTokenValueWithLiteral(t, true)
			name = &value
		}
		t = p.la()
	} else if t == TokenTkVerticalBar {
		noNamespace := TreeResolverNoNamespace
		prefix = &noNamespace
	} else {
		panic(NewCSSParseExceptionTokenArray(
			t, []*Token{TokenTkAsterisk, TokenTkIdent, TokenTkVerticalBar},
			p.getCurrentLine()))
	}

	if t == TokenTkVerticalBar {
		p.next()
		t = p.next()
		if t == TokenTkAsterisk || t == TokenTkIdent {
			if prefix == nil {
				prefix = name
			}
			if t == TokenTkIdent {
				value := p.getTokenValueWithLiteral(t, true)
				name = &value
			}
		} else {
			panic(NewCSSParseExceptionTokenArray(
				t, []*Token{TokenTkAsterisk, TokenTkIdent}, p.getCurrentLine()))
		}
	}

	var namespaceURI *string
	if prefix != nil && *prefix != TreeResolverNoNamespace {
		value, ok := p.namespaces[strings.ToLower(*prefix)]
		if !ok {
			panic(NewCSSParseException("There is no namespace with prefix "+*prefix+" defined",
				p.getCurrentLine()))
		}
		namespaceURI = &value
	} else if prefix == nil && !matchAttribute {
		namespaceURI = p.defaultNamespace
	}

	if matchAttribute && name == nil {
		panic(NewCSSParseException("An attribute name is required", p.getCurrentLine()))
	}

	return &cssParserNamespacePair{namespaceURI: namespaceURI, name: name}
}

// class
//
//	: '.' IDENT
//	;
func (p *CSSParser) classSelector(selector *Selector) {
	t := p.next()
	if t == TokenTkPeriod {
		t = p.next()
		if t == TokenTkIdent {
			selector.AddClassCondition(p.getTokenValueWithLiteral(t, true))
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkIdent, p.getCurrentLine()))
		}
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkPeriod, p.getCurrentLine()))
	}
}

// attrib
//
//	: '[' S* [ namespace_prefix ]? IDENT S*
//	      [ [ PREFIXMATCH |
//	          SUFFIXMATCH |
//	          SUBSTRINGMATCH |
//	          '=' |
//	          INCLUDES |
//	          DASHMATCH ] S* [ IDENT | STRING ] S*
//	      ]? ']'
//	;
func (p *CSSParser) attrib(selector *Selector) {
	t := p.next()
	if t == TokenTkLbracket {
		p.skipWhitespace()
		t = p.la()
		if t == TokenTkIdent || t == TokenTkAsterisk || t == TokenTkVerticalBar {
			existenceMatch := true
			pair := p.typedValue(true)
			attrNamespaceURI := pair.namespaceURI
			// typedValue(true) panics when there is no name.
			attrName := *pair.name
			p.skipWhitespace()
			t = p.la()
			switch t.GetType() {
			case TokenTypeEquals,
				TokenTypeIncludes,
				TokenTypeDashmatch,
				TokenTypePrefixmatch,
				TokenTypeSuffixmatch,
				TokenTypeSubstringmatch:
				existenceMatch = false
				selectorType := p.next()
				p.skipWhitespace()
				t = p.next()
				if t == TokenTkIdent || t == TokenTkString {
					value := p.getTokenValueWithLiteral(t, true)
					switch selectorType.GetType() {
					case TokenTypeEquals:
						selector.AddAttributeEqualsCondition(attrNamespaceURI, attrName, value)
					case TokenTypeDashmatch:
						selector.AddAttributeMatchesFirstPartCondition(attrNamespaceURI, attrName, value)
					case TokenTypeIncludes:
						selector.AddAttributeMatchesListCondition(attrNamespaceURI, attrName, value)
					case TokenTypePrefixmatch:
						selector.AddAttributePrefixCondition(attrNamespaceURI, attrName, value)
					case TokenTypeSuffixmatch:
						selector.AddAttributeSuffixCondition(attrNamespaceURI, attrName, value)
					case TokenTypeSubstringmatch:
						selector.AddAttributeSubstringCondition(attrNamespaceURI, attrName, value)
					}
					p.skipWhitespace()
				} else {
					p.push(t)
					panic(NewCSSParseExceptionTokenArray(t,
						[]*Token{TokenTkIdent, TokenTkString},
						p.getCurrentLine()))
				}
				p.skipWhitespace()
				t = p.la()
			}
			if existenceMatch {
				selector.AddAttributeExistsCondition(attrNamespaceURI, attrName)
			}
			if t == TokenTkRbracket {
				p.next()
			} else {
				panic(NewCSSParseExceptionTokenArray(t, []*Token{TokenTkEquals,
					TokenTkIncludes, TokenTkDashmatch, TokenTkPrefixmatch,
					TokenTkSuffixmatch, TokenTkSubstringmatch, TokenTkRbracket},
					p.getCurrentLine()))
			}
		} else {
			panic(NewCSSParseExceptionTokenArray(
				t, []*Token{TokenTkIdent, TokenTkAsterisk}, p.getCurrentLine()))
		}
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkLbracket, p.getCurrentLine()))
	}
}

func (p *CSSParser) addPseudoClassOrElement(t *Token, selector *Selector) {
	value := p.getTokenValue(t)
	switch value {
	case "link":
		selector.AddLinkCondition()
	case "visited":
		selector.SetPseudoClass(SelectorVisitedPseudoclass)
	case "hover":
		selector.SetPseudoClass(SelectorHoverPseudoclass)
	case "focus":
		selector.SetPseudoClass(SelectorFocusPseudoclass)
	case "active":
		selector.SetPseudoClass(SelectorActivePseudoclass)
	case "first-child":
		selector.AddFirstChildCondition()
	case "even":
		selector.AddEvenChildCondition()
	case "odd":
		selector.AddOddChildCondition()
	case "last-child":
		selector.AddLastChildCondition()
	case "first-line", "first-letter", "before", "after":
		selector.SetPseudoElement(value)
	default:
		panic(NewCSSParseException(value+" is not a recognized pseudo-class", p.getCurrentLine()))
	}
}

func (p *CSSParser) addPseudoClassOrElementFunction(t *Token, selector *Selector) {
	f0 := []rune(p.getTokenValue(t))
	f := string(f0[:len(f0)-1])

	switch f {
	case "lang":
		p.skipWhitespace()
		t = p.next()
		if t == TokenTkIdent {
			lang := p.getTokenValue(t)
			selector.AddLangCondition(lang)
			p.skipWhitespace()
			t = p.next()
		} else {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkIdent, p.getCurrentLine()))
		}
	case "nth-child":
		var number strings.Builder
		for {
			t = p.next()
			if !(t != nil && (t == TokenTkIdent || t == TokenTkS || t == TokenTkNumber || t == TokenTkDimension || t == TokenTkPlus || t == TokenTkMinus)) {
				break
			}
			number.WriteString(p.getTokenValue(t))
		}

		cssParserTry(func() {
			selector.AddNthChildCondition(number.String())
		}, func(e *CSSParseException) {
			e.SetLine(p.getCurrentLine())
			p.push(t)
			panic(e)
		})
	case "has":
		result := p.parseHasPseudoClass(selector.GetRuleset())
		selector.AddHasCondition(result.relativeSelectors, result.specificityB, result.specificityC, result.specificityD)
		t = result.lastToken
	default:
		p.push(t)
		panic(NewCSSParseException(f+" is not a valid function in this context", p.getCurrentLine()))
	}

	if t != TokenTkRparen {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkRparen, p.getCurrentLine()))
	}
}

func (p *CSSParser) parseHasPseudoClass(ruleset *Ruleset) *cssParserHasPseudoClassResult {
	p.skipWhitespace()
	var relativeSelectors []*SelectorHasRelativeSelector
	maxSpecificityB := 0
	maxSpecificityC := 0
	maxSpecificityD := 0

	for {
		t := p.la()
		if t == TokenTkRparen {
			if len(relativeSelectors) == 0 {
				panic(NewCSSParseException("The :has() pseudo-class requires a non-empty selector argument", p.getCurrentLine()))
			}
			return &cssParserHasPseudoClassResult{relativeSelectors, maxSpecificityB, maxSpecificityC, maxSpecificityD, p.next()}
		}

		if t == TokenTkComma {
			panic(NewCSSParseException("The :has() pseudo-class does not allow empty selectors", p.getCurrentLine()))
		}

		parsed := p.parseRelativeSelectorInHas(ruleset)
		relativeSelectors = append(relativeSelectors, parsed.relativeSelector)
		maxSpecificityB = max(maxSpecificityB, parsed.specificityB)
		maxSpecificityC = max(maxSpecificityC, parsed.specificityC)
		maxSpecificityD = max(maxSpecificityD, parsed.specificityD)

		p.skipWhitespace()
		t = p.next()
		if t == TokenTkComma {
			p.skipWhitespace()
		} else if t == TokenTkRparen {
			return &cssParserHasPseudoClassResult{relativeSelectors, maxSpecificityB, maxSpecificityC, maxSpecificityD, t}
		} else {
			p.push(t)
			panic(NewCSSParseExceptionTokenArray(t, []*Token{TokenTkComma, TokenTkRparen}, p.getCurrentLine()))
		}
	}
}

func (p *CSSParser) parseRelativeSelectorInHas(ruleset *Ruleset) *cssParserParsedRelativeSelector {
	var axes []SelectorAxis
	var selectors []*Selector

	currentAxis := SelectorAxisDescendantAxis
	t := p.la()
	if t == TokenTkPlus || t == TokenTkGreater {
		currentAxis = p.combinatorToAxis(p.combinator())
	}

	selectors = append(selectors, p.simpleSelector(ruleset))
	axes = append(axes, currentAxis)

	specificityB := selectors[0].GetSpecificityB()
	specificityC := selectors[0].GetSpecificityC()
	specificityD := selectors[0].GetSpecificityD()

	for {
		t = p.la()
		if t == TokenTkRparen || t == TokenTkComma {
			break
		}
		if t != TokenTkPlus && t != TokenTkGreater && t != TokenTkS {
			panic(NewCSSParseExceptionTokenArray(
				t,
				[]*Token{TokenTkPlus, TokenTkGreater, TokenTkS, TokenTkComma, TokenTkRparen},
				p.getCurrentLine()))
		}

		axis := p.combinatorToAxis(p.combinator())
		t = p.la()
		if !p.isSimpleSelectorStart(t) {
			panic(NewCSSParseExceptionTokenArray(
				t,
				[]*Token{TokenTkIdent, TokenTkAsterisk, TokenTkHash, TokenTkPeriod, TokenTkLbracket, TokenTkColon},
				p.getCurrentLine()))
		}
		next := p.simpleSelector(ruleset)
		selectors = append(selectors, next)
		axes = append(axes, axis)
		specificityB += next.GetSpecificityB()
		specificityC += next.GetSpecificityC()
		specificityD += next.GetSpecificityD()
	}

	return &cssParserParsedRelativeSelector{NewSelectorHasRelativeSelector(axes, selectors), specificityB, specificityC, specificityD}
}

func (p *CSSParser) combinatorToAxis(combinator *Token) SelectorAxis {
	if combinator == TokenTkPlus {
		return SelectorAxisImmediateSiblingAxis
	}
	if combinator == TokenTkGreater {
		return SelectorAxisChildAxis
	}
	return SelectorAxisDescendantAxis
}

func (p *CSSParser) isSimpleSelectorStart(t *Token) bool {
	return t == TokenTkIdent || t == TokenTkAsterisk || t == TokenTkHash ||
		t == TokenTkPeriod || t == TokenTkLbracket || t == TokenTkColon ||
		t == TokenTkVerticalBar
}

type cssParserParsedRelativeSelector struct {
	relativeSelector *SelectorHasRelativeSelector
	specificityB     int
	specificityC     int
	specificityD     int
}

type cssParserHasPseudoClassResult struct {
	relativeSelectors []*SelectorHasRelativeSelector
	specificityB      int
	specificityC      int
	specificityD      int
	lastToken         *Token
}

func (p *CSSParser) addPseudoElement(t *Token, selector *Selector) {
	value := p.getTokenValue(t)
	switch value {
	case "first-line", "first-letter", "before", "after":
		selector.SetPseudoElement(value)
	default:
		panic(NewCSSParseException(value+" is not a recognized pseudo-element", p.getCurrentLine()))
	}
}

// pseudo
//
//	: ':' ':'? [ IDENT | FUNCTION S* IDENT? S* ')' ]
//	;
func (p *CSSParser) pseudo(selector *Selector) {
	t := p.next()
	if t == TokenTkColon {
		t = p.next()
		switch t.GetType() {
		case TokenTypeColon:
			t = p.next()
			p.addPseudoElement(t, selector)
		case TokenTypeIdent:
			p.addPseudoClassOrElement(t, selector)
		case TokenTypeFunction:
			p.addPseudoClassOrElementFunction(t, selector)
		default:
			p.push(t)
			panic(NewCSSParseExceptionTokenArray(t,
				[]*Token{TokenTkIdent, TokenTkFunction}, p.getCurrentLine()))
		}
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkColon, p.getCurrentLine()))
	}
}

func (p *CSSParser) checkCSSName(cssName *CSSName, propertyName string) bool {
	if cssName == nil {
		p.checkForUnsupportedCssProperties(propertyName)
		p.errorHandler.Error(
			p.uri,
			propertyName+" is an unrecognized CSS property at line "+
				strconv.Itoa(p.getCurrentLine())+". Ignoring declaration.")
		return false
	}

	if !CSSNameIsImplemented(cssName) {
		p.errorHandler.Error(
			p.uri,
			propertyName+" is not implemented at line "+
				strconv.Itoa(p.getCurrentLine())+". Ignoring declaration.")
		return false
	}

	builder := CSSNameGetPropertyBuilder(cssName)
	if builder == nil {
		p.errorHandler.Error(
			p.uri,
			"(bug) No property builder defined for "+propertyName+
				" at line "+strconv.Itoa(p.getCurrentLine())+". Ignoring declaration.")
		return false
	}

	return true
}

// declaration
//
//	: property ':' S* expr prio?
//	;
func (p *CSSParser) declaration(ruleset *Ruleset, inFontFace bool) {
	cssParserTry(func() {
		t := p.la()
		if t == TokenTkIdent {
			propertyName := p.property()
			cssName := CSSNameGetByPropertyName(propertyName)

			valid := p.checkCSSName(cssName, propertyName)

			t = p.next()
			if t == TokenTkColon {
				p.skipWhitespace()

				values := p.expr(
					cssName == CSSNameFontFamily ||
						cssName == CSSNameFontShorthand ||
						cssName == CSSNameFsPdfFontEncoding)
				important := false

				t = p.la()
				if t == TokenTkImportantSym {
					p.prio()
					important = true
				}

				t = p.la()
				if !(t == TokenTkSemicolon || t == TokenTkRbrace || t == TokenTkEof) {
					panic(NewCSSParseExceptionTokenArray(
						t,
						[]*Token{TokenTkSemicolon, TokenTkRbrace},
						p.getCurrentLine()))
				}

				if valid {
					p.checkForUnsupportedDisplayValues(cssName, values)
					cssParserTry(func() {
						builder := CSSNameGetPropertyBuilder(cssName)
						ruleset.AddAllProperties(builder.BuildDeclarationsWithInheritAllowed(
							cssName, values, ruleset.GetOrigin(), important, !inFontFace))
					}, func(e *CSSParseException) {
						e.SetLine(p.getCurrentLine())
						p.error(e, "declaration", true)
					})
				}
			} else {
				p.push(t)
				panic(NewCSSParseExceptionToken(t, TokenTkColon, p.getCurrentLine()))
			}
		} else {
			panic(NewCSSParseExceptionToken(t, TokenTkIdent, p.getCurrentLine()))
		}
	}, func(e *CSSParseException) {
		p.error(e, "declaration", true)
		p.recover(false, true)
	})
}

// prio
//
//	: IMPORTANT_SYM S*
//	;
func (p *CSSParser) prio() {
	t := p.next()
	if t == TokenTkImportantSym {
		p.skipWhitespace()
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkImportantSym, p.getCurrentLine()))
	}
}

// expr
//
//	: term [ operator term ]*
//	;
func (p *CSSParser) expr(literal bool) []*PropertyValue {
	result := make([]*PropertyValue, 0, 10)
	result = append(result, p.term(literal, nil))
LOOP:
	for {
		t := p.la()
		operator := false
		var operatorToken *Token
		switch t.GetType() {
		case TokenTypeVirgule,
			TokenTypeComma:
			operatorToken = t
			p.operator()
			t = p.la()
			operator = true
		}
		switch t.GetType() {
		case TokenTypePlus,
			TokenTypeMinus,
			TokenTypeNumber,
			TokenTypePercentage,
			TokenTypePx,
			TokenTypeCm,
			TokenTypeMm,
			TokenTypeIn,
			TokenTypePt,
			TokenTypePc,
			TokenTypeEms,
			TokenTypeExs,
			TokenTypeAngle,
			TokenTypeDimension,
			TokenTypeTime,
			TokenTypeFreq,
			TokenTypeString,
			TokenTypeIdent,
			TokenTypeUri,
			TokenTypeHash,
			TokenTypeFunction:
			result = append(result, p.term(literal, operatorToken))
		default:
			if operator {
				panic(NewCSSParseExceptionTokenArray(t, []*Token{
					TokenTkNumber, TokenTkPlus, TokenTkMinus,
					TokenTkPercentage, TokenTkPx, TokenTkEms, TokenTkExs,
					TokenTkPc, TokenTkMm, TokenTkCm, TokenTkIn, TokenTkPt,
					TokenTkAngle, TokenTkDimension, TokenTkTime, TokenTkFreq, TokenTkString,
					TokenTkIdent, TokenTkUri, TokenTkHash, TokenTkFunction},
					p.getCurrentLine()))
			} else {
				break LOOP
			}
		}
	}

	return result
}

func (p *CSSParser) extractNumber(t *Token) string {
	token := p.getTokenValue(t)

	offset := 0
	ch := []rune(token)
	for _, c := range ch {
		if c < '0' || c > '9' {
			break
		}
		offset++
	}
	if ch[offset] == '.' {
		offset++

		for i := offset; i < len(ch); i++ {
			c := ch[i]
			if c < '0' || c > '9' {
				break
			}
			offset++
		}
	}

	return string(ch[:offset])
}

func (p *CSSParser) extractUnit(t *Token) string {
	s := p.extractNumber(t)
	return p.getTokenValue(t)[len(s):]
}

func (p *CSSParser) sign(sign int) string {
	if sign < 0 {
		return "-"
	}
	return ""
}

// cssParserParseFloat is Float.parseFloat for the text of a num token.
func cssParserParseFloat(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		panic(NewXRRuntimeException("For input string: \"" + s + "\""))
	}
	return float32(f)
}

// term
//
//	: unary_operator?
//	  [ NUMBER S* | PERCENTAGE S* | LENGTH S* | EMS S* | EXS S* | ANGLE S* |
//	    TIME S* | FREQ S* ]
//	| STRING S* | IDENT S* | URI S* | hexcolor | function
//	;
func (p *CSSParser) term(literal bool, operatorToken *Token) *PropertyValue {
	sign := 1
	t := p.la()
	if t == TokenTkPlus || t == TokenTkMinus {
		sign = p.unaryOperator()
		t = p.la()
	}
	var result *PropertyValue
	switch t.GetType() {
	case TokenTypeAngle:
		unit := p.extractUnit(t)
		var primitiveType int16
		switch unit {
		case "deg":
			primitiveType = CSSPrimitiveValueCssDeg
		case "rad":
			primitiveType = CSSPrimitiveValueCssRad
		case "grad":
			primitiveType = CSSPrimitiveValueCssGrad
		default:
			panic(NewCSSParseException("Unsupported CSS unit "+unit, p.getCurrentLine()))
		}

		result = NewPropertyValueFloat(primitiveType,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t))

		p.next()
		p.skipWhitespace()

	// "turn" isn't a distinct DOM CSSPrimitiveValue unit, so convert it to degrees eagerly.
	case TokenTypeDimension:
		unit := p.extractUnit(t)
		if unit != "turn" {
			panic(NewCSSParseException("Unsupported CSS unit "+unit, p.getCurrentLine()))
		}

		result = NewPropertyValueFloat(CSSPrimitiveValueCssDeg,
			float32(sign)*cssParserParseFloat(p.extractNumber(t))*360,
			p.sign(sign)+p.getTokenValue(t))

		p.next()
		p.skipWhitespace()

	case TokenTypeTime, TokenTypeFreq:
		panic(NewCSSParseException("Unsupported CSS unit "+p.extractUnit(t), p.getCurrentLine()))
	case TokenTypeNumber:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssNumber,
			float32(sign)*cssParserParseFloat(p.getTokenValue(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypePercentage:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssPercentage,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeEms:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssEms,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeExs:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssExs,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypePx:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssPx,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeCm:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssCm,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeMm:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssMm,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeIn:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssIn,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypePt:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssPt,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypePc:
		result = NewPropertyValueFloatWithOperatorToken(
			CSSPrimitiveValueCssPc,
			float32(sign)*cssParserParseFloat(p.extractNumber(t)),
			p.sign(sign)+p.getTokenValue(t),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeString:
		s := p.getTokenValue(t)
		result = NewPropertyValueStringWithOperatorToken(
			CSSPrimitiveValueCssString,
			s,
			p.getRawTokenValue(),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeIdent:
		value := p.getTokenValueWithLiteral(t, literal)
		result = NewPropertyValueStringWithOperatorToken(
			CSSPrimitiveValueCssIdent,
			value,
			value,
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeUri:
		result = NewPropertyValueStringWithOperatorToken(
			CSSPrimitiveValueCssUri,
			p.getTokenValue(t),
			p.getRawTokenValue(),
			operatorToken)
		p.next()
		p.skipWhitespace()
	case TokenTypeHash:
		result = p.hexcolor(operatorToken)
	case TokenTypeFunction:
		result = p.function(operatorToken)
	default:
		panic(NewCSSParseExceptionTokenArray(t, []*Token{TokenTkNumber,
			TokenTkPercentage, TokenTkPx, TokenTkEms, TokenTkExs,
			TokenTkPc, TokenTkMm, TokenTkCm, TokenTkIn, TokenTkPt,
			TokenTkAngle, TokenTkTime, TokenTkFreq, TokenTkString,
			TokenTkIdent, TokenTkUri, TokenTkHash, TokenTkFunction},
			p.getCurrentLine()))
	}
	return result
}

// function
//
//	: FUNCTION S* expr ')' S*
//	;
func (p *CSSParser) function(operatorToken *Token) *PropertyValue {
	var result *PropertyValue
	t := p.next()
	if t == TokenTkFunction {
		f := p.getTokenValue(t)
		p.skipWhitespace()
		params := p.expr(false)
		t = p.next()
		if t != TokenTkRparen {
			p.push(t)
			panic(NewCSSParseExceptionToken(t, TokenTkRparen, p.getCurrentLine()))
		}

		if f == "rgb(" || f == "rgba(" {
			result = NewPropertyValueFSColorWithOperatorToken(p.createRGBColorFromFunction(params), operatorToken)
		} else if f == "cmyk(" {
			if !p.IsSupportCMYKColors() {
				panic(NewCSSParseException(
					"The current output device does not support CMYK colors", p.getCurrentLine()))
			}
			//in accordance to http://www.w3.org/TR/css3-gcpm/#cmyk-colors
			result = NewPropertyValueFSColorWithOperatorToken(p.createCMYKColorFromFunction(params), operatorToken)
		} else {
			name := []rune(f)
			result = NewPropertyValueFSFunctionWithOperatorToken(NewFSFunction(string(name[:len(name)-1]), params), operatorToken)
		}

		p.skipWhitespace()
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkFunction, p.getCurrentLine()))
	}

	return result
}

func (p *CSSParser) createCMYKColorFromFunction(params []*PropertyValue) *FSCMYKColor {
	if len(params) != 4 {
		panic(NewCSSParseException(
			"The cmyk() function must have exactly four parameters",
			p.getCurrentLine()))
	}

	var colorComponents [4]float32

	for i := 0; i < len(params); i++ {
		colorComponents[i] = p.parseCMYKColorComponent(params[i], i+1) //Warning on the truncation?
	}

	return NewFSCMYKColor(colorComponents[0], colorComponents[1], colorComponents[2], colorComponents[3])
}

func (p *CSSParser) parseCMYKColorComponent(value *PropertyValue, paramNo int) float32 {
	primitiveType := value.GetPrimitiveType()
	var result float32
	if primitiveType == CSSPrimitiveValueCssNumber {
		result = value.GetFloatValue()
	} else if primitiveType == CSSPrimitiveValueCssPercentage {
		result = value.GetFloatValue() / 100.0
	} else {
		panic(NewCSSParseException(
			"Parameter "+strconv.Itoa(paramNo)+" to the cmyk() function is "+
				"not a number or a percentage", p.getCurrentLine()))
	}

	if result < 0.0 || result > 1.0 {
		panic(NewCSSParseException(
			"Parameter "+strconv.Itoa(paramNo)+" to the cmyk() function must be between zero and one", p.getCurrentLine()))
	}

	return result
}

func (p *CSSParser) createRGBColorFromFunction(params []*PropertyValue) *FSRGBColor {
	if len(params) != 3 && len(params) != 4 {
		panic(NewCSSParseException(
			"The rgb() function must have three or four parameters",
			p.getCurrentLine()))
	}

	red := int(p.calculateColor(params, 0))
	green := int(p.calculateColor(params, 1))
	blue := int(p.calculateColor(params, 2))
	var alpha float32 = 1
	if len(params) >= 4 {
		alpha = p.calculateColor(params, 3)
	}

	return NewFSRGBColorWithAlpha(red, green, blue, alpha)
}

func (p *CSSParser) calculateColor(params []*PropertyValue, index int) float32 {
	value := params[index]
	primitiveType := p.validateType(index, value)

	var f float32
	switch primitiveType {
	case CSSPrimitiveValueCssPercentage:
		f = value.GetFloatValue() / 100 * 255
	default:
		f = value.GetFloatValue()
	}

	if f < 0 {
		return 0
	} else if f > 255 {
		return 255
	} else {
		return f
	}
}

func (p *CSSParser) validateType(index int, value *PropertyValue) int16 {
	primitiveType := value.GetPrimitiveType()
	if primitiveType != CSSPrimitiveValueCssPercentage && primitiveType != CSSPrimitiveValueCssNumber {
		panic(NewCSSParseException(
			"Parameter "+strconv.Itoa(index+1)+" to the rgb() function is "+
				"not a number or percentage", p.getCurrentLine()))
	}

	if primitiveType != CSSPrimitiveValueCssNumber && index == 3 {
		panic(NewCSSParseException(
			"Parameter alpha to the rgba() function is "+
				"not a number", p.getCurrentLine()))
	}
	return primitiveType
}

// There is a constraint on the color that it must
// have either 3 or 6 hex-digits (i.e., [0-9a-fA-F])
// after the "#"; e.g., "#000" is OK, but "#abcd" is not.
//
//	hexcolor
//	  : HASH S*
//	  ;
func (p *CSSParser) hexcolor(operatorToken *Token) *PropertyValue {
	var result *PropertyValue
	t := p.next()
	if t == TokenTkHash {
		s := []rune(p.getTokenValue(t))
		if len(s) != 3 && len(s) != 6 || !p.isHexString(s) {
			p.push(t)
			panic(NewCSSParseException("#"+string(s)+" is not a valid color definition", p.getCurrentLine()))
		}
		var color *FSRGBColor
		if len(s) == 3 {
			color = NewFSRGBColor(
				p.convertToIntegerWithHexchar2(s[0], s[0]),
				p.convertToIntegerWithHexchar2(s[1], s[1]),
				p.convertToIntegerWithHexchar2(s[2], s[2]))
		} else { /* s.length == 6 */
			color = NewFSRGBColor(
				p.convertToIntegerWithHexchar2(s[0], s[1]),
				p.convertToIntegerWithHexchar2(s[2], s[3]),
				p.convertToIntegerWithHexchar2(s[4], s[5]))
		}
		result = NewPropertyValueFSColorWithOperatorToken(color, operatorToken)
		p.skipWhitespace()
	} else {
		p.push(t)
		panic(NewCSSParseExceptionToken(t, TokenTkHash, p.getCurrentLine()))
	}

	return result
}

func (p *CSSParser) isHexString(s []rune) bool {
	for i := 0; i < len(s); i++ {
		if !cssParserIsHexChar(s[i]) {
			return false
		}
	}
	return true
}

func (p *CSSParser) convertToIntegerWithHexchar2(hexchar1 rune, hexchar2 rune) int {
	result := p.convertToInteger(hexchar1)
	result <<= 4
	result |= p.convertToInteger(hexchar2)
	return result
}

func (p *CSSParser) convertToInteger(hexchar1 rune) int {
	if hexchar1 >= '0' && hexchar1 <= '9' {
		return int(hexchar1 - '0')
	} else if hexchar1 >= 'a' && hexchar1 <= 'f' {
		return int(hexchar1-'a') + 10
	} else { /* if (hexchar1 >= 'A' && hexchar1 <= 'F') */
		return int(hexchar1-'A') + 10
	}
}

func (p *CSSParser) skipWhitespace() {
	var t *Token
	for {
		t = p.next()
		if t != TokenTkS {
			break
		}
		// skip
	}
	p.push(t)
}

func (p *CSSParser) skipWhitespaceAndCdocdc() {
	var t *Token
	for {
		t = p.next()
		if !(t == TokenTkS || t == TokenTkCdo || t == TokenTkCdc) {
			break
		}
	}
	p.push(t)
}

func (p *CSSParser) next() *Token {
	if p.saved != nil {
		result := p.saved
		p.saved = nil
		return result
	} else {
		result, err := p.lexer.Yylex()
		if err != nil {
			panic(&cssParserIOException{err: err})
		}
		return result
	}
}

func (p *CSSParser) push(t *Token) {
	if p.saved != nil {
		panic(NewXRRuntimeException("saved must be null"))
	}
	p.saved = t
}

func (p *CSSParser) la() *Token {
	result := p.next()
	p.push(result)
	return result
}

func (p *CSSParser) error(e *CSSParseException, what string, rethrowEOF bool) {
	if !e.IsCallerNotified() {
		message := e.GetMessage() + " Skipping " + what + "."
		p.errorHandler.Error(p.uri, message)
	}
	e.SetCallerNotified(true)
	if e.IsEOF() && rethrowEOF {
		panic(e)
	}
}

func (p *CSSParser) recover(needBlock bool, stopBeforeBlockClose bool) {
	braces := 0
	foundBlock := false
LOOP:
	for {
		t := p.next()
		if t == TokenTkEof {
			return
		}
		switch t.GetType() {
		case TokenTypeLbrace:
			foundBlock = true
			braces++
		case TokenTypeRbrace:
			if braces == 0 {
				if stopBeforeBlockClose {
					p.push(t)
					break LOOP
				}
			} else {
				braces--
				if braces == 0 {
					break LOOP
				}
			}
		case TokenTypeSemicolon:
			if braces == 0 && (!needBlock || foundBlock) {
				break LOOP
			}
		}
	}
	p.skipWhitespace()
}

func (p *CSSParser) Reset(r io.Reader) {
	p.saved = nil
	clear(p.namespaces)
	p.defaultNamespace = nil
	p.lexer.Yyreset(r)
	p.lexer.SetYyLine(0)
}

func (p *CSSParser) getRawTokenValue() string {
	return p.lexer.Yytext()
}

func (p *CSSParser) getTokenValue(t *Token) string {
	return p.getTokenValueWithLiteral(t, false)
}

func (p *CSSParser) getTokenValueWithLiteral(t *Token, literal bool) string {
	switch t.GetType() {
	case TokenTypeString:
		return cssParserProcessEscapes([]rune(p.lexer.Yytext()), 1, p.lexer.Yylength()-1)
	case TokenTypeHash:
		return cssParserProcessEscapes([]rune(p.lexer.Yytext()), 1, p.lexer.Yylength())
	case TokenTypeUri:
		ch := []rune(p.lexer.Yytext())
		start := 4
		for ch[start] == '\t' || ch[start] == '\r' ||
			ch[start] == '\n' || ch[start] == '\f' {
			start++
		}
		if ch[start] == '\'' || ch[start] == '"' {
			start++
		}
		end := len(ch) - 2
		for ch[end] == '\t' || ch[end] == '\r' ||
			ch[end] == '\n' || ch[end] == '\f' {
			end--
		}
		if ch[end] == '\'' || ch[end] == '"' {
			end--
		}

		uriResult := cssParserProcessEscapes(ch, start, end+1)

		// Relative URIs are resolved relative to CSS file, not XHTML file
		if p.isRelativeURI(uriResult) && p.uri != "" {
			uriResult = p.getPartBeforeLastSlash(p.uri) + uriResult
		} else if p.isServerRelativeURI(uriResult) && p.uri != "" && p.isHierarchicalAbsoluteUri(p.uri) {
			uriResult = p.getPartBeforeFirstSlash(p.uri) + uriResult
		}

		return uriResult
	case TokenTypeAtRule,
		TokenTypeIdent,
		TokenTypeFunction:
		start := 0
		count := p.lexer.Yylength()
		if t.GetType() == TokenTypeAtRule {
			start++
		}
		result := cssParserProcessEscapes([]rune(p.lexer.Yytext()), start, count)
		if !literal {
			result = strings.ToLower(result)
		}
		return result
	default:
		return p.lexer.Yytext()
	}
}

// cssParserParseURI is new java.net.URI(uri) as far as this class uses it.
// ok is false where the Java constructor throws URISyntaxException: net/url
// accepts characters that java.net.URI does not, and a scheme with nothing
// after it.
func cssParserParseURI(uri string) (u *url.URL, ok bool) {
	for _, c := range uri {
		if c <= ' ' || c == 0x7F || strings.ContainsRune("\"<>\\^`{|}", c) {
			return nil, false
		}
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, false
	}
	if u.Scheme != "" && uri[len(u.Scheme)+1:] == "" {
		return nil, false
	}
	return u, true
}

func (p *CSSParser) getPartBeforeFirstSlash(uri string) string {
	u, ok := cssParserParseURI(uri)
	if !ok {
		panic(NewXRRuntimeException("Invalid uri: " + uri))
	}
	// Keep the scheme delimiter when there is no authority (e.g. file:/path),
	// otherwise server-relative rewrite yields "file/..." instead of "file:/...".
	if u.Host == "" && u.User == nil {
		return u.Scheme + ":"
	}
	authority := u.Host
	if u.User != nil {
		authority = u.User.String() + "@" + authority
	}
	return u.Scheme + "://" + authority
}

func (p *CSSParser) getPartBeforeLastSlash(uri string) string {
	lastSlash := strings.LastIndex(uri, "/")
	if lastSlash == -1 {
		return ""
	}
	return uri[:lastSlash+1]
}

func (p *CSSParser) isRelativeURI(uri string) bool {
	return uri != "" && !strings.HasPrefix(uri, "/") && !p.isAbsoluteUri(uri)
}

func (p *CSSParser) isServerRelativeURI(uri string) bool {
	return strings.HasPrefix(uri, "/") && !p.isAbsoluteUri(uri)
}

// isHierarchicalAbsoluteUri reports whether uri is an absolute hierarchical
// URI suitable as a base for resolving server-relative resource paths (/...).
//
// Opaque absolute URIs such as the synthetic inline:N keys used for <style>
// blocks must not participate in path rewriting — otherwise
// url('/assets/font.ttf') becomes inline/assets/font.ttf.
func (p *CSSParser) isHierarchicalAbsoluteUri(uri string) bool {
	u, ok := cssParserParseURI(uri)
	if !ok {
		return false
	}
	return u.Scheme != "" && u.Opaque == ""
}

func (p *CSSParser) isAbsoluteUri(uri string) bool {
	if strings.HasPrefix(uri, "http:") || strings.HasPrefix(uri, "https:") ||
		strings.HasPrefix(uri, "data:") || strings.HasPrefix(uri, "blob:") {
		return true
	}
	u, ok := cssParserParseURI(uri)
	if !ok {
		return false
	}
	return u.Scheme != ""
}

func (p *CSSParser) getCurrentLine() int {
	return p.lexer.YyLine()
}

func cssParserIsHexChar(c rune) bool {
	return c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f'
}

func cssParserProcessEscapes(ch []rune, start int, end int) string {
	var result strings.Builder
	result.Grow(len(ch) + 10)

	for i := start; i < end; i++ {
		c := ch[i]

		if c == '\\' {
			// eat escaped newlines and handle te\st == test situations
			if i < end-2 && ch[i+1] == '\r' && ch[i+2] == '\n' {
				i += 2
				continue
			} else {
				if i+1 < len(ch) && (ch[i+1] == '\n' || ch[i+1] == '\r' || ch[i+1] == '\f') {
					i++
					continue
				} else if i+1 >= len(ch) {
					// process \ escaped (\\)
					result.WriteRune(c)
					continue
				} else if !cssParserIsHexChar(ch[i+1]) {
					continue
				}
			}

			// Unicode escapes
			i++
			current := i
			for i < end && cssParserIsHexChar(ch[i]) && i-current < 6 {
				i++
			}

			cvalue, err := strconv.ParseInt(string(ch[current:i]), 16, 32)
			if err != nil {
				panic(NewXRRuntimeException("For input string: \"" + string(ch[current:i]) + "\""))
			}
			if cvalue < 0xFFFF {
				result.WriteRune(rune(cvalue))
			}

			i--

			if i < end-2 && ch[i+1] == '\r' && ch[i+2] == '\n' {
				i += 2
			} else if i < end-1 &&
				(ch[i+1] == ' ' || ch[i+1] == '\t' ||
					ch[i+1] == '\n' || ch[i+1] == '\r' ||
					ch[i+1] == '\f') {
				i++
			}
		} else {
			result.WriteRune(c)
		}
	}

	return result.String()
}

func (p *CSSParser) IsSupportCMYKColors() bool {
	return p.supportCMYKColors
}

func (p *CSSParser) SetSupportCMYKColors(b bool) {
	p.supportCMYKColors = b
}

// cssParserNamespacePair holds nil where the Java record holds null: a nil
// namespaceURI matches any namespace, a nil name is the universal selector.
type cssParserNamespacePair struct {
	namespaceURI *string
	name         *string
}

func (p *CSSParser) checkForUnsupportedCssProperties(propertyName string) {
	if _, ok := CSSNameUnsupportedCss3Properties[propertyName]; ok {
		p.css3FeatureListener(propertyName)
	}
}

func (p *CSSParser) checkForUnsupportedDisplayValues(cssName *CSSName, values []*PropertyValue) {
	if cssName == CSSNameDisplay {
		for _, v := range values {
			if v.GetPropertyValueType() != PropertyValueTypeValueTypeIdent {
				continue
			}
			displayValue := v.GetStringValue()
			if _, ok := CSSNameUnsupportedCss3DisplayValues[displayValue]; ok {
				p.css3FeatureListener("display: " + displayValue)
				break
			}
		}
	}
}
