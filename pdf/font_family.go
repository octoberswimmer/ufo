// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/FontFamily.java

package pdf

import (
	"sort"

	"github.com/octoberswimmer/ufo"
)

type FontFamily struct {
	name             string
	fontDescriptions []*FontDescription
}

// NewFontFamily is package-private in Java.
func NewFontFamily(name string) *FontFamily {
	return &FontFamily{name: name}
}

func (f *FontFamily) GetName() string {
	return f.name
}

func (f *FontFamily) GetFontDescriptions() []*FontDescription {
	return f.fontDescriptions
}

// AddFontDescription keeps the descriptions ordered by weight; descriptions
// of equal weight keep their order of insertion, as with Java's stable
// List.sort.
func (f *FontFamily) AddFontDescription(description *FontDescription) {
	f.fontDescriptions = append(f.fontDescriptions, description)
	sort.SliceStable(f.fontDescriptions, func(i, j int) bool {
		return f.fontDescriptions[i].GetWeight() < f.fontDescriptions[j].GetWeight()
	})
}

// removeFontDescriptionsIf stands for getFontDescriptions().removeIf(filter)
// in ITextFontResolver.flushFontFaceFonts: Java removes from the list that
// the getter returns, which a Go slice does not allow.
func (f *FontFamily) removeFontDescriptionsIf(filter func(*FontDescription) bool) {
	kept := f.fontDescriptions[:0]
	for _, description := range f.fontDescriptions {
		if !filter(description) {
			kept = append(kept, description)
		}
	}
	for i := len(kept); i < len(f.fontDescriptions); i++ {
		f.fontDescriptions[i] = nil
	}
	f.fontDescriptions = kept
}

// Match returns nil when the family has no font descriptions.
func (f *FontFamily) Match(desiredWeight int, style *ufo.IdentValue) *FontDescription {
	var candidates []*FontDescription

	for _, description := range f.fontDescriptions {
		if description.GetStyle() == style {
			candidates = append(candidates, description)
		}
	}

	if len(candidates) == 0 {
		if style == ufo.IdentValueItalic {
			return f.Match(desiredWeight, ufo.IdentValueOblique)
		} else if style == ufo.IdentValueOblique {
			return f.Match(desiredWeight, ufo.IdentValueNormal)
		} else {
			candidates = append(candidates, f.fontDescriptions...)
		}
	}

	result := f.findByWeight(candidates, desiredWeight, fontFamilySearchModeExact)

	if result != nil {
		return result
	} else {
		if desiredWeight <= 500 {
			return f.findByWeight(candidates, desiredWeight, fontFamilySearchModeLighterOrDarker)
		} else {
			return f.findByWeight(candidates, desiredWeight, fontFamilySearchModeDarkerOrLighter)
		}
	}
}

// fontFamilySearchMode ports the private enum FontFamily.SearchMode.
type fontFamilySearchMode int

const (
	fontFamilySearchModeExact fontFamilySearchMode = iota
	fontFamilySearchModeLighterOrDarker
	fontFamilySearchModeDarkerOrLighter
)

func (f *FontFamily) findByWeight(matches []*FontDescription, desiredWeight int, searchMode fontFamilySearchMode) *FontDescription {
	switch searchMode {
	case fontFamilySearchModeExact:
		for _, description := range matches {
			if description.GetWeight() == desiredWeight {
				return description
			}
		}
		return nil
	case fontFamilySearchModeLighterOrDarker:
		var offset int
		var description *FontDescription
		for offset = 0; offset < len(matches); offset++ {
			description = matches[offset]
			if description.GetWeight() > desiredWeight {
				break
			}
		}

		if offset > 0 && description.GetWeight() > desiredWeight {
			return matches[offset-1]
		} else {
			return description
		}
	case fontFamilySearchModeDarkerOrLighter:
		var offset int
		var description *FontDescription
		for offset = len(matches) - 1; offset >= 0; offset-- {
			description = matches[offset]
			if description.GetWeight() < desiredWeight {
				break
			}
		}

		if offset != len(matches)-1 && description != nil && description.GetWeight() < desiredWeight {
			return matches[offset+1]
		} else {
			return description
		}
	}
	return nil
}
