// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/sheet/RulesetContainer.java

package ufo

type RulesetContainer interface {
	AddContent(ruleset *Ruleset)
	GetOrigin() StylesheetInfoOrigin
}
