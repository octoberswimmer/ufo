// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/resource/AbstractResource.java

package ufo

import (
	"io"
	"time"
)

// AbstractResource is the base of the resource classes. It is embedded by
// value and initialized with initAbstractResource or
// initAbstractResourceInputStream, which port the two Java constructors.
type AbstractResource struct {
	inputSource     *InputSource
	createTimeStamp int64
}

func (r *AbstractResource) initAbstractResource(source *InputSource) {
	r.inputSource = source
	r.createTimeStamp = time.Now().UnixMilli()
}

func (r *AbstractResource) initAbstractResourceInputStream(is io.Reader) {
	r.initAbstractResource(InputSourcesFromStream(is))
}

func (r *AbstractResource) GetResourceInputSource() *InputSource {
	return r.inputSource
}

func (r *AbstractResource) GetResourceLoadTimeStamp() int64 {
	return r.createTimeStamp
}
