// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/Html5Support.java

package ufo

import (
	"fmt"
	"strings"
)

// Html5SupportLogUnsupportedFeatures takes both sets as slices in insertion
// order: the Java callers pass LinkedHashSets, whose order appears in the
// message. The Java class logs through slf4j at warn level; here the messages
// go to the general XRLog channel at LevelWarning.
func Html5SupportLogUnsupportedFeatures(tags []string, cssFeatures []string) {
	if len(tags) != 0 {
		formatted := make([]string, 0, len(tags))
		for _, tag := range tags {
			formatted = append(formatted, fmt.Sprintf("<%s>", tag))
		}
		XRLogGeneral(LevelWarning, fmt.Sprintf(
			"Encountered HTML5 elements which are not supported by FlyingSaucer: %s. Rendering may be incorrect.",
			strings.Join(formatted, ", ")))
	}

	if len(cssFeatures) != 0 {
		formatted := make([]string, 0, len(cssFeatures))
		for _, feature := range cssFeatures {
			formatted = append(formatted, "\""+feature+"\"")
		}
		XRLogGeneral(LevelWarning, fmt.Sprintf(
			"Encountered CSS3 features not supported by FlyingSaucer: %s. Rendering may be incorrect.",
			strings.Join(formatted, ", ")))
	}
}
