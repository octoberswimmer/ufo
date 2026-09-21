// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/MediaRule.java

package ufo

import (
	"fmt"
	"strings"
)

type MediaRule struct {
	mediaTypes []string
	contents   []*Ruleset
	origin     StylesheetInfoOrigin
}

func NewMediaRule(origin StylesheetInfoOrigin) *MediaRule {
	return &MediaRule{origin: origin}
}

func (m *MediaRule) AddMedium(medium string) {
	m.mediaTypes = append(m.mediaTypes, medium)
}

func (m *MediaRule) Matches(medium string) bool {
	// strings.EqualFold gives the result of String.equalsIgnoreCase here: no
	// character other than the ASCII letters folds to "a" or "l".
	return strings.EqualFold(medium, "all") || stylesheetInfoContains(m.mediaTypes, "all") ||
		stylesheetInfoContains(m.mediaTypes, strings.ToLower(medium))
}

func (m *MediaRule) AddContent(ruleset *Ruleset) {
	m.contents = append(m.contents, ruleset)
}

func (m *MediaRule) GetContents() []*Ruleset {
	return m.contents
}

func (m *MediaRule) GetOrigin() StylesheetInfoOrigin {
	return m.origin
}

func (m *MediaRule) String() string {
	return fmt.Sprintf("%s{origin: %s}", "MediaRule", m.origin)
}

func (m *MediaRule) ToString() string {
	return m.String()
}
