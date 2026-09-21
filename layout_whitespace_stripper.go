// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/WhitespaceStripper.java

package ufo

import (
	"regexp"
	"strings"
)

const (
	WhitespaceStripperSpace = " "
	WhitespaceStripperEol   = "\n"
	WhitespaceStripperEolc  = '\n'
)

// Java's \s (without UNICODE_CHARACTER_CLASS) is [ \t\n\x0B\f\r]; RE2's \s
// leaves out \x0B, so the classes are written out.
var (
	WhitespaceStripperLinefeedSpaceCollapse       = regexp.MustCompile("[ \\t\\n\\x0B\\f\\r]+\\n[ \\t\\n\\x0B\\f\\r]+")
	WhitespaceStripperLinefeedToSpace             = regexp.MustCompile("\\n")
	WhitespaceStripperTabToSpace                  = regexp.MustCompile("\\t")
	WhitespaceStripperSpaceCollapse               = regexp.MustCompile("(?: )+")
	WhitespaceStripperSpaceBeforeLinefeedCollapse = regexp.MustCompile("[ \\t\\x0B\\f\\r]\\n")
)

// WhitespaceStripperStripInlineContent strips whitespace early in inline
// content generation. This can be done because "whitespage" does not ally to
// :first-line and :first-letter. For dynamic pseudo-classes we are allowed to
// choose which properties apply.
//
// The Java method removes elements from the list it is given. Here the
// stripped list is returned; the argument's backing array is not modified.
func WhitespaceStripperStripInlineContent(inlineContent []Styleable) []Styleable {
	collapse := false
	allWhitespace := true

	for _, node := range inlineContent {
		if node.GetStyle().IsInline() {
			iB := node.(*InlineBox)
			collapseNext := whitespaceStripperStripWhitespace(iB, collapse)
			if !iB.IsRemovableWhitespace() {
				allWhitespace = false
			}

			collapse = collapseNext
		} else {
			if !whitespaceStripperCanCollapseThrough(node) {
				allWhitespace = false
				collapse = false
			}
		}
	}

	if allWhitespace {
		return whitespaceStripperStripTextContent(inlineContent)
	}
	return inlineContent
}

func whitespaceStripperCanCollapseThrough(styleable Styleable) bool {
	style := styleable.GetStyle()
	return style.IsFloated() || style.IsAbsolute() || style.IsFixed() || style.IsRunning()
}

func whitespaceStripperStripTextContent(stripped []Styleable) []Styleable {
	onlyAnonymous := true
	for _, node := range stripped {
		if node.GetStyle().IsInline() {
			iB := node.(*InlineBox)
			if iB.GetElement() != nil {
				onlyAnonymous = false
			}

			iB.TruncateText()
		}
	}

	if onlyAnonymous {
		result := make([]Styleable, 0, len(stripped))
		for _, node := range stripped {
			if !node.GetStyle().IsInline() {
				result = append(result, node)
			}
		}
		return result
	}
	return stripped
}

// whitespaceStripperStripWhitespace strips all whitespace from the text
// according to the CSS 2.1 spec on whitespace handling. It accounts for the
// different whitespace settings like normal, nowrap, pre, etc
//
// It returns whether the next leading space should collapse or not.
func whitespaceStripperStripWhitespace(iB *InlineBox, collapseLeading bool) bool {

	whitespace := iB.GetStyle().GetIdent(CSSNameWhiteSpace)

	text := iB.GetText()

	text = whitespaceStripperCollapseWhitespace(iB, whitespace, text, collapseLeading)

	collapseNext := strings.HasSuffix(text, WhitespaceStripperSpace) &&
		(whitespace == IdentValueNormal || whitespace == IdentValueNowrap || whitespace == IdentValuePre)

	iB.SetText(text)
	if whitespaceStripperTrimIsEmpty(text) {
		if whitespace == IdentValueNormal || whitespace == IdentValueNowrap {
			iB.SetRemovableWhitespace(true)
		} else if whitespace == IdentValuePre {
			iB.SetRemovableWhitespace(false) //actually unnecessary, is set to this by default
		} else if !strings.Contains(text, WhitespaceStripperEol) { //and whitespace.equals("pre-line"), the only one left
			iB.SetRemovableWhitespace(true)
		}
	}
	if text == "" {
		return collapseLeading
	}
	return collapseNext
}

// whitespaceStripperTrimIsEmpty is text.trim().isEmpty(): String.trim removes
// every character up to and including U+0020 from both ends.
func whitespaceStripperTrimIsEmpty(text string) bool {
	for _, r := range text {
		if r > ' ' {
			return false
		}
	}
	return true
}

func whitespaceStripperCollapseWhitespace(iB *InlineBox, whitespace *IdentValue, text string, collapseLeading bool) string {
	if whitespace == IdentValueNormal || whitespace == IdentValueNowrap {
		text = WhitespaceStripperLinefeedSpaceCollapse.ReplaceAllLiteralString(text, WhitespaceStripperEol)
	} else if whitespace == IdentValuePre {
		text = WhitespaceStripperSpaceBeforeLinefeedCollapse.ReplaceAllLiteralString(text, WhitespaceStripperEol)
	}

	if whitespace == IdentValueNormal || whitespace == IdentValueNowrap {
		text = WhitespaceStripperLinefeedToSpace.ReplaceAllLiteralString(text, WhitespaceStripperSpace)
		text = WhitespaceStripperTabToSpace.ReplaceAllLiteralString(text, WhitespaceStripperSpace)
		text = WhitespaceStripperSpaceCollapse.ReplaceAllLiteralString(text, WhitespaceStripperSpace)
	} else if whitespace == IdentValuePre || whitespace == IdentValuePreWrap {
		tabSize := calculatedStyleFloatToInt(iB.GetStyle().AsFloat(CSSNameTabSize))
		tabs := strings.Repeat(" ", tabSize)
		text = WhitespaceStripperTabToSpace.ReplaceAllLiteralString(text, tabs)
	} else if whitespace == IdentValuePreLine {
		text = WhitespaceStripperTabToSpace.ReplaceAllLiteralString(text, WhitespaceStripperSpace)
		text = WhitespaceStripperSpaceCollapse.ReplaceAllLiteralString(text, WhitespaceStripperSpace)
	}

	if whitespace == IdentValueNormal || whitespace == IdentValueNowrap {
		// collapse first space against prev inline
		if strings.HasPrefix(text, WhitespaceStripperSpace) &&
			collapseLeading {
			text = text[1:]
		}
	}

	return text
}
