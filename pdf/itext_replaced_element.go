// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextReplacedElement.java

package pdf

import "github.com/octoberswimmer/ufo"

type ITextReplacedElement interface {
	ufo.ReplacedElement
	Paint(c *ufo.RenderingContext, outputDevice *ITextOutputDevice, box ufo.BlockBoxI)
}
