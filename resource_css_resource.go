// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/resource/CSSResource.java

package ufo

import "io"

type CSSResource struct {
	AbstractResource
}

// NewCSSResource creates a CSSResource reading from stream, which may be nil.
func NewCSSResource(stream io.Reader) *CSSResource {
	r := &CSSResource{}
	r.initAbstractResourceInputStream(stream)
	return r
}
