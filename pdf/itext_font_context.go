// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/ITextFontContext.java

package pdf

import "github.com/octoberswimmer/ufo"

type ITextFontContext struct {
}

var _ ufo.FontContext = (*ITextFontContext)(nil)

func NewITextFontContext() *ITextFontContext {
	return &ITextFontContext{}
}
