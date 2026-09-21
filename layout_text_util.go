// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/TextUtil.java
//
// The functions are named LayoutTextUtil... because util/TextUtil already
// owns the TextUtil... names in this package.

package ufo

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// layoutTextUtilToLowerCase is String.toLowerCase(Locale.ROOT): the full
// Unicode case mapping, including the final sigma rule.
func layoutTextUtilToLowerCase(text string) string {
	return cases.Lower(language.Und).String(text)
}

// layoutTextUtilToUpperCase is String.toUpperCase(Locale.ROOT): the full
// Unicode case mapping, in which one character can become several.
func layoutTextUtilToUpperCase(text string) string {
	return cases.Upper(language.Und).String(text)
}

func LayoutTextUtilTransformText(text string, style CalculatedStyleI) string {
	transform := style.GetIdent(CSSNameTextTransform)
	fontVariant := style.GetIdent(CSSNameFontVariant)
	return LayoutTextUtilTransformTextWithTransformFontVariant(text, transform, fontVariant)
}

func LayoutTextUtilTransformTextWithTransformFontVariant(text string, transform *IdentValue, fontVariant *IdentValue) string {
	if transform == IdentValueLowercase {
		text = layoutTextUtilToLowerCase(text)
	}
	if transform == IdentValueUppercase {
		text = layoutTextUtilToUpperCase(text)
	}
	if transform == IdentValueCapitalize {
		text = layoutTextUtilCapitalizeWords(text)
	}

	if fontVariant == IdentValueSmallCaps {
		text = layoutTextUtilToUpperCase(text)
	}
	return text
}

func LayoutTextUtilTransformFirstLetterText(text string, style CalculatedStyleI) string {
	if text != "" {
		transform := style.GetIdent(CSSNameTextTransform)
		fontVariant := style.GetIdent(CSSNameFontVariant)
		for i := 0; i < len(text); {
			currentChar, width := utf8.DecodeRuneInString(text[i:])
			if !LayoutTextUtilIsFirstLetterSeparatorChar(currentChar) {
				// Java works on UTF-16 chars: a character outside the BMP is
				// seen as a surrogate, which Character.toLowerCase and
				// Character.toUpperCase leave unchanged.
				if transform == IdentValueLowercase {
					if currentChar <= 0xFFFF {
						currentChar = unicode.ToLower(currentChar)
					}
					text = LayoutTextUtilReplaceChar(text, currentChar, i)
				} else if transform == IdentValueUppercase || transform == IdentValueCapitalize || fontVariant == IdentValueSmallCaps {
					if currentChar <= 0xFFFF {
						currentChar = unicode.ToUpper(currentChar)
					}
					text = LayoutTextUtilReplaceChar(text, currentChar, i)
				}
				break
			}
			i += width
		}
	}
	return text
}

// LayoutTextUtilReplaceChar replaces the character at the specified index by
// another. index is the byte offset at which the character starts. An index
// outside the text returns the text unchanged.
func LayoutTextUtilReplaceChar(text string, newChar rune, index int) string {
	if index < 0 || index >= len(text) {
		return text
	}
	_, width := utf8.DecodeRuneInString(text[index:])
	var b strings.Builder
	b.Grow(len(text))
	b.WriteString(text[:index])
	b.WriteRune(newChar)
	b.WriteString(text[index+width:])
	return b.String()
}

// LayoutTextUtilIsFirstLetterSeparatorChar takes a rune where Java takes a
// UTF-16 char. Java sees a character outside the BMP as a surrogate, whose
// general category is not one of the separator categories, so such a rune is
// never a separator.
func LayoutTextUtilIsFirstLetterSeparatorChar(c rune) bool {
	if c > 0xFFFF {
		return false
	}
	return unicode.In(c,
		unicode.Ps, // START_PUNCTUATION
		unicode.Pe, // END_PUNCTUATION
		unicode.Pi, // INITIAL_QUOTE_PUNCTUATION
		unicode.Pf, // FINAL_QUOTE_PUNCTUATION
		unicode.Po, // OTHER_PUNCTUATION
		unicode.Zs, // SPACE_SEPARATOR
	)
}

// LayoutTextUtilTextWidth measures the width of text in dots, including any
// letter-spacing from style (applied after each character).
// The spacing contribution is rounded up so that a width accumulated from
// substring measurements never understates the width of the whole run.
func LayoutTextUtilTextWidth(c CssContext, style CalculatedStyleI, font FSFont, text string) int {
	width := c.GetTextRenderer().GetWidth(c.GetFontContext(), font, text)
	letterSpacing := style.LetterSpacing(c)
	if letterSpacing == 0.0 {
		return width
	}
	return width + int(math.Ceil(float64(letterSpacing*float32(utf8.RuneCountInString(text)))))
}

func layoutTextUtilCapitalizeWords(text string) string {
	if text == "" {
		return text
	}

	result := layoutTextUtilDoCapitalizeWords(text)
	if layoutTextUtilUtf16Length(result) != layoutTextUtilUtf16Length(text) {
		UuPString("error! to strings arent the same length = -" + result + "-" + text + "-")
	}
	return result
}

// layoutTextUtilUtf16Length is Java's String.length().
func layoutTextUtilUtf16Length(text string) int {
	n := 0
	for _, r := range text {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func layoutTextUtilDoCapitalizeWords(text string) string {
	var sb strings.Builder
	cap := true
	for i := 0; i < len(text); {
		r, width := utf8.DecodeRuneInString(text[i:])
		ch := text[i : i+width]
		// Java uppercases one UTF-16 char at a time, which leaves each half
		// of a surrogate pair, and so every character outside the BMP,
		// unchanged.
		if cap && r <= 0xFFFF {
			sb.WriteString(layoutTextUtilToUpperCase(ch))
		} else {
			sb.WriteString(ch)
		}
		cap = ch == " "
		i += width
	}
	return sb.String()
}
