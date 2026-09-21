// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/resource/Resource.java

package ufo

type Resource interface {
	// GetResourceInputSource returns nil when the resource has no input
	// source.
	GetResourceInputSource() *InputSource

	// GetResourceLoadTimeStamp returns the creation time of the resource in
	// milliseconds since the Unix epoch.
	GetResourceLoadTimeStamp() int64
}

var (
	_ Resource = (*CSSResource)(nil)
	_ Resource = (*ImageResource)(nil)
	_ Resource = (*XMLResource)(nil)
)
