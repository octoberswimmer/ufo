// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/extend/FontResolver.java

package ufo

type FontResolver interface {
	// ResolveFont returns nil when no font matches the specification.
	ResolveFont(renderingContext *SharedContext, spec *FontSpecification) FSFont
	FlushCache()
}
