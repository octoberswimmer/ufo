// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/Breaker.java

/*
 * Breaker.java
 * Copyright (c) 2004, 2005 Torbjoern Gannholm,
 * Copyright (c) 2005 Wisconsin Court System
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public License
 * as published by the Free Software Foundation; either version 2.1
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
 *
 */

package ufo

import (
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/octoberswimmer/ufo/dom"
)

// Breaker is a utility class that scans the text of a single inline box,
// looking for the next break point.
//
// Author of the Java class: Torbjoern Gannholm

const breakerDefaultLanguageProperty = "org.xhtmlrenderer.layout.breaker.default-language"

var (
	breakerDefaultLanguageValue string
	breakerDefaultLanguageOnce  sync.Once
)

// breakerDefaultLanguage is the Java constant DEFAULT_LANGUAGE: the system
// property org.xhtmlrenderer.layout.breaker.default-language, "en" when it is
// not set. It is read once, on first use, as Java reads it when the class is
// initialized. A system property is what ConfigurationSetProperty set, or else
// the environment variable of that name (as in Configuration).
func breakerDefaultLanguage() string {
	breakerDefaultLanguageOnce.Do(func() {
		configurationSystemPropertiesMu.Lock()
		val, ok := configurationSystemProperties[breakerDefaultLanguageProperty]
		configurationSystemPropertiesMu.Unlock()
		if !ok {
			val, ok = os.LookupEnv(breakerDefaultLanguageProperty)
		}
		if !ok {
			val = "en"
		}
		breakerDefaultLanguageValue = val
	})
	return breakerDefaultLanguageValue
}

func BreakerBreakFirstLetter(c *LayoutContext, context *LineBreakContext,
	avail int, style CalculatedStyleI) {
	font := style.GetFSFont(c)
	context.SetEnd(breakerGetFirstLetterEnd(context.GetMaster(), context.GetStart()))
	context.SetWidth(LayoutTextUtilTextWidth(c, style, font, context.GetCalculatedSubstring()))

	if context.GetWidth() > avail {
		context.SetNeedsNewLine(true)
		context.SetUnbreakable(true)
	}
}

func breakerGetFirstLetterEnd(text string, start int) int {
	letterFound := false
	end := len(text)
	for i := start; i < end; {
		currentChar, size := utf8.DecodeRuneInString(text[i:])
		if !LayoutTextUtilIsFirstLetterSeparatorChar(currentChar) {
			if letterFound {
				return i
			} else {
				letterFound = true
			}
		}
		i += size
	}
	return end
}

func BreakerBreakText(c *LayoutContext,
	context *LineBreakContext, avail int, fullLineWidth int, style CalculatedStyleI) {
	font := style.GetFSFont(c)
	whitespace := style.GetWhitespace()

	// ====== handle nowrap
	if whitespace == IdentValueNowrap {
		context.SetEnd(context.GetLast())
		context.SetWidth(LayoutTextUtilTextWidth(c, style, font, context.GetCalculatedSubstring()))
		return
	}

	//check if we should break on the next newline
	if whitespace == IdentValuePre ||
		whitespace == IdentValuePreWrap ||
		whitespace == IdentValuePreLine {
		n := strings.Index(context.GetStartSubstring(), WhitespaceStripperEol)
		if n > -1 {
			context.SetEnd(context.GetStart() + n + 1)
			context.SetWidth(LayoutTextUtilTextWidth(c, style, font, context.GetStartSubstring()[0:n]))
			context.SetNeedsNewLine(true)
			context.SetEndsOnNL(true)
		} else if whitespace == IdentValuePre {
			context.SetEnd(context.GetLast())
			context.SetWidth(LayoutTextUtilTextWidth(c, style, font, context.GetCalculatedSubstring()))
		}
	}

	//check if we may wrap
	if whitespace == IdentValuePre ||
		context.IsNeedsNewLine() && context.GetWidth() <= avail {
		return
	}

	context.SetEndsOnNL(false)

	tryToBreakAnywhere := style.GetWordBreak() == IdentValueBreakAll

	breakerDoBreakText(c, context, avail, style, tryToBreakAnywhere, fullLineWidth)
}

// BreakerGetBreakPointsProviderElement is getBreakPointsProvider(String,
// LayoutContext, Element, CalculatedStyle), the overload for an element. A nil
// element stands for Java's null.
func BreakerGetBreakPointsProviderElement(text string, c *LayoutContext, element *dom.Element, style CalculatedStyleI) BreakPointsProvider {
	return c.GetSharedContext().GetLineBreakingStrategy().GetBreakPointsProvider(text, breakerGetLanguage(c, element), style)
}

// BreakerGetBreakPointsProviderText is getBreakPointsProvider(String,
// LayoutContext, Text, CalculatedStyle), the overload for a text node. A nil
// text node stands for Java's null.
func BreakerGetBreakPointsProviderText(text string, c *LayoutContext, textNode *dom.Text, style CalculatedStyleI) BreakPointsProvider {
	return c.GetSharedContext().GetLineBreakingStrategy().GetBreakPointsProvider(text, breakerGetLanguageText(c, textNode), style)
}

func breakerGetLanguage(c *LayoutContext, element *dom.Element) string {
	language := ""
	if element != nil {
		language = c.GetNamespaceHandler().GetLang(element)
	}
	if language == "" {
		language = breakerDefaultLanguage()
	}
	return language
}

func breakerGetLanguageText(c *LayoutContext, textNode *dom.Text) string {
	if textNode != nil {
		parentNode := textNode.GetParentNode()
		if element, ok := parentNode.(*dom.Element); ok {
			return breakerGetLanguage(c, element)
		}
	}
	return breakerDefaultLanguage()
}

func breakerDoBreakText(c *LayoutContext,
	context *LineBreakContext, avail int, style CalculatedStyleI,
	tryToBreakAnywhere bool, fullLineWidth int) {
	f := style.GetFSFont(c)
	currentString := context.GetStartSubstring()
	iterator := BreakerGetBreakPointsProviderText(currentString, c, context.GetTextNode(), style)
	if tryToBreakAnywhere {
		iterator = NewBreakAnywhereLineBreakStrategy(currentString)
	}
	bp := iterator.Next()
	var lastBreakPoint *BreakPoint
	right := -1
	previousWidth := 0
	previousPosition := 0
	for bp != nil && bp.GetPosition() != BreakIteratorDone {
		currentWidth := LayoutTextUtilTextWidth(c, style, f, currentString[previousPosition:bp.GetPosition()]+bp.GetHyphen())
		widthWithHyphen := previousWidth + currentWidth
		previousWidth = widthWithHyphen
		previousPosition = bp.GetPosition()
		if widthWithHyphen > avail {
			break
		}
		right = previousPosition
		lastBreakPoint = bp
		bp = iterator.Next()
	}

	// add hyphen if needed
	if bp != nil && bp.GetPosition() != BreakIteratorDone && // it fits
		right >= 0 && // some break point found
		lastBreakPoint.GetHyphen() != "" {
		master := context.GetMaster()
		at := context.GetStart() + right
		context.SetMaster(master[:at] + lastBreakPoint.GetHyphen() + master[at:])
		right += len(lastBreakPoint.GetHyphen())
	}

	if bp != nil && bp.GetPosition() == BreakIteratorDone {
		context.SetWidth(LayoutTextUtilTextWidth(c, style, f, currentString))
		context.SetEnd(len(context.GetMaster()))
		//It fits!
		return
	}

	context.SetNeedsNewLine(true)
	if right <= 0 && style.GetWordWrap() == IdentValueBreakWord {
		if !tryToBreakAnywhere {
			breakerDoBreakText(c, context, avail, style, true, fullLineWidth)
			return
		}

		if avail < fullLineWidth {
			// Float reduced avail — word may fit on next full line
			// → unbreakable: InlineBoxing's getNextLineBoxDelta moves past float
			context.SetEnd(context.GetStart() + len(currentString))
			context.SetUnbreakable(true)
			context.SetWidth(LayoutTextUtilTextWidth(c, style, f, context.GetCalculatedSubstring()))
		} else {
			// avail IS the full line width → container genuinely too narrow
			// → force a break after one code point (browser behaviour),
			// keeping surrogate pairs intact
			_, oneCodePoint := utf8.DecodeRuneInString(currentString)
			context.SetEnd(context.GetStart() + oneCodePoint)
			context.SetWidth(LayoutTextUtilTextWidth(c, style, f, currentString[0:oneCodePoint]))
		}
		return
	}

	if right > 0 { // found a place to wrap
		context.SetEnd(context.GetStart() + right)
		context.SetWidth(LayoutTextUtilTextWidth(c, style, f, context.GetMaster()[context.GetStart():context.GetStart()+right]))
		return
	}

	// unbreakable string
	context.SetEnd(context.GetStart() + len(currentString))
	context.SetUnbreakable(true)
	context.SetWidth(LayoutTextUtilTextWidth(c, style, f, context.GetCalculatedSubstring()))
}
