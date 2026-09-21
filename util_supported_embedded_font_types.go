// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/util/SupportedEmbeddedFontTypes.java

package ufo

import "strings"

type SupportedEmbeddedFontTypes struct {
	name       string
	TypeString string
	Extension  string
}

var (
	SupportedEmbeddedFontTypesOtf = &SupportedEmbeddedFontTypes{name: "OTF", TypeString: "font/otf", Extension: ".otf"}
	SupportedEmbeddedFontTypesTtf = &SupportedEmbeddedFontTypes{name: "TTF", TypeString: "font/ttf", Extension: ".ttf"}
)

// SupportedEmbeddedFontTypesValues returns the constants in the order of
// their Java declaration.
func SupportedEmbeddedFontTypesValues() []*SupportedEmbeddedFontTypes {
	return []*SupportedEmbeddedFontTypes{SupportedEmbeddedFontTypesOtf, SupportedEmbeddedFontTypesTtf}
}

func (t *SupportedEmbeddedFontTypes) Name() string {
	return t.name
}

func (t *SupportedEmbeddedFontTypes) String() string {
	return t.name
}

func SupportedEmbeddedFontTypesIsSupported(uri string) bool {
	for _, t := range SupportedEmbeddedFontTypesValues() {
		if strings.Contains(uri, t.TypeString) {
			return true
		}
	}
	return false
}

func SupportedEmbeddedFontTypesGetExtension(uri string) string {
	for _, t := range SupportedEmbeddedFontTypesValues() {
		if strings.Contains(uri, t.TypeString) {
			return t.Extension
		}
	}
	return ""
}
