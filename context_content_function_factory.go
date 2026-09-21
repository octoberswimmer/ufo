// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/context/ContentFunctionFactory.java

package ufo

import "strings"

type ContentFunctionFactory struct {
	functions []ContentFunction
}

func NewContentFunctionFactory() *ContentFunctionFactory {
	f := &ContentFunctionFactory{}
	f.functions = append(f.functions, &contentFunctionFactoryPageCounterFunction{})
	f.functions = append(f.functions, &contentFunctionFactoryPagesCounterFunction{})
	f.functions = append(f.functions, &contentFunctionFactoryTargetCounterFunction{})
	f.functions = append(f.functions, &contentFunctionFactoryLeaderFunction{})
	return f
}

// LookupFunction returns the first registered function that can handle the
// given function, or nil.
func (f *ContentFunctionFactory) LookupFunction(c *LayoutContext, function *FSFunction) ContentFunction {
	for _, candidate := range f.functions {
		if candidate.CanHandle(c, function) {
			return candidate
		}
	}
	return nil
}

func (f *ContentFunctionFactory) RegisterFunction(function ContentFunction) {
	f.functions = append(f.functions, function)
}

// contentFunctionFactoryPageNumberFunction ports the abstract nested class
// ContentFunctionFactory.PageNumberFunction; the two counter functions embed
// it.
type contentFunctionFactoryPageNumberFunction struct {
}

func (p *contentFunctionFactoryPageNumberFunction) IsStatic() bool {
	return false
}

func (p *contentFunctionFactoryPageNumberFunction) Calculate(c *LayoutContext, function *FSFunction) string {
	return ""
}

func (p *contentFunctionFactoryPageNumberFunction) GetLayoutReplacementText() string {
	return "999"
}

func (p *contentFunctionFactoryPageNumberFunction) getListStyleType(function *FSFunction) *IdentValue {
	result := IdentValueDecimal

	parameters := function.GetParameters()
	if len(parameters) == 2 {
		pValue := parameters[1]
		iValue := IdentValueValueOf(pValue.GetStringValue())
		if iValue != nil {
			result = iValue
		}
	}

	return result
}

func (p *contentFunctionFactoryPageNumberFunction) isCounter(function *FSFunction, counterName string) bool {
	if function.Is("counter") {
		parameters := function.GetParameters()
		if len(parameters) == 1 || len(parameters) == 2 {
			param := parameters[0]
			if param.GetPrimitiveType() != CSSPrimitiveValueCssIdent ||
				param.GetStringValue() != counterName {
				return false
			}

			if len(parameters) == 2 {
				param = parameters[1]
				return param.GetPrimitiveType() == CSSPrimitiveValueCssIdent
			}

			return true
		}
	}

	return false
}

type contentFunctionFactoryPageCounterFunction struct {
	contentFunctionFactoryPageNumberFunction
}

func (p *contentFunctionFactoryPageCounterFunction) CalculateWithText(c *RenderingContext, function *FSFunction, text *InlineText) string {
	value := c.GetRootLayer().GetRelativePageNo(c) + 1
	return CounterFunctionCreateCounterText(p.getListStyleType(function), value)
}

func (p *contentFunctionFactoryPageCounterFunction) CanHandle(c *LayoutContext, function *FSFunction) bool {
	return c.IsPrint() && p.isCounter(function, "page")
}

type contentFunctionFactoryPagesCounterFunction struct {
	contentFunctionFactoryPageNumberFunction
}

func (p *contentFunctionFactoryPagesCounterFunction) CalculateWithText(c *RenderingContext, function *FSFunction, text *InlineText) string {
	value := c.GetRootLayer().GetRelativePageCount(c)
	return CounterFunctionCreateCounterText(p.getListStyleType(function), value)
}

func (p *contentFunctionFactoryPagesCounterFunction) CanHandle(c *LayoutContext, function *FSFunction) bool {
	return c.IsPrint() && p.isCounter(function, "pages")
}

// contentFunctionFactoryTargetCounterFunction partially implements target
// counter as specified here:
// http://www.w3.org/TR/2007/WD-css3-gcpm-20070504/#cross-references
type contentFunctionFactoryTargetCounterFunction struct {
}

func (t *contentFunctionFactoryTargetCounterFunction) IsStatic() bool {
	return false
}

func (t *contentFunctionFactoryTargetCounterFunction) CalculateWithText(c *RenderingContext, function *FSFunction, text *InlineText) string {
	uri := text.GetParent().GetElement().GetAttribute("href")
	if strings.HasPrefix(uri, "#") {
		anchor := uri[1:]
		target := c.GetBoxById(anchor)
		if target != nil {
			pageNo := c.GetRootLayer().GetRelativePageNoWithAbsY(c, target.GetAbsY())
			return CounterFunctionCreateCounterText(IdentValueDecimal, pageNo+1)
		}
	}
	return ""
}

func (t *contentFunctionFactoryTargetCounterFunction) Calculate(c *LayoutContext, function *FSFunction) string {
	return ""
}

func (t *contentFunctionFactoryTargetCounterFunction) GetLayoutReplacementText() string {
	return "999"
}

func (t *contentFunctionFactoryTargetCounterFunction) CanHandle(c *LayoutContext, function *FSFunction) bool {
	if c.IsPrint() && function.Is("target-counter") {
		parameters := function.GetParameters()
		if len(parameters) == 2 || len(parameters) == 3 {
			f := parameters[0].GetFunction()
			if f == nil ||
				len(f.GetParameters()) != 1 ||
				f.GetParameters()[0].GetPrimitiveType() != CSSPrimitiveValueCssIdent ||
				"href" != f.GetParameters()[0].GetStringValue() {
				return false
			}

			param := parameters[1]
			return param.GetPrimitiveType() == CSSPrimitiveValueCssIdent &&
				param.GetStringValue() == "page"
		}
	}

	return false
}

// contentFunctionFactoryLeaderFunction partially implements leaders as
// specified here:
// http://www.w3.org/TR/2007/WD-css3-gcpm-20070504/#leaders
type contentFunctionFactoryLeaderFunction struct {
}

func (l *contentFunctionFactoryLeaderFunction) IsStatic() bool {
	return false
}

func (l *contentFunctionFactoryLeaderFunction) CalculateWithText(c *RenderingContext, function *FSFunction, text *InlineText) string {
	iB := text.GetParent()
	lineBox := iB.GetLineBox()

	// There might be a target-counter function after this function.
	// Because the leader should fill up the line, we need the correct
	// width and must first compute the target-counter function.
	dynamic := false
	for _, child := range lineBox.GetChildren() {
		if child == BoxI(iB) {
			dynamic = true
		} else if inlineLayoutBox, ok := child.(*InlineLayoutBox); dynamic && ok {
			inlineLayoutBox.LookForDynamicFunctions(c)
		}
	}
	if dynamic {
		totalLineWidth := InlineBoxingPositionHorizontally(c, lineBox, 0)
		lineBox.SetContentWidth(totalLineWidth)
	}

	// Get leader value and value width
	value := l.getLeaderValue(function)

	// Compute value width using 100x string to get more precise width.
	// Otherwise, there might be a small gap on the right side. This is
	// necessary because a TextRenderer usually use double/float for width.
	tmp := strings.Repeat(value, 100)
	valueWidth := float32(LayoutTextUtilTextWidth(c, iB.GetStyle(), iB.GetStyle().GetFSFont(c), tmp)) / 100.0

	spaceWidth := LayoutTextUtilTextWidth(c, iB.GetStyle(), iB.GetStyle().GetFSFont(c), " ")

	// compute leader width and necessary count of values
	leaderWidth := iB.GetContainingBlockWidth() - iB.GetLineBox().GetWidth() + text.GetWidth()
	count := int(float32(leaderWidth-2*spaceWidth) / valueWidth)

	leaderString := " " + strings.Repeat(value, max(0, count)) + " "

	// set left margin to ensure that the leader is right aligned (for TOC)
	leaderStringWidth := LayoutTextUtilTextWidth(c, iB.GetStyle(), iB.GetStyle().GetFSFont(c), leaderString)
	iB.SetMarginLeft(c, leaderWidth-leaderStringWidth)

	return leaderString
}

func (l *contentFunctionFactoryLeaderFunction) getLeaderValue(function *FSFunction) string {
	param := function.GetParameters()[0]
	value := param.GetStringValue()
	if param.GetPrimitiveType() == CSSPrimitiveValueCssIdent {
		switch value {
		case "dotted":
			return ". "
		case "solid":
			return "_"
		case "space":
			return " "
		default:
			return value
		}
	}
	return value
}

func (l *contentFunctionFactoryLeaderFunction) Calculate(c *LayoutContext, function *FSFunction) string {
	return ""
}

func (l *contentFunctionFactoryLeaderFunction) GetLayoutReplacementText() string {
	return " . "
}

func (l *contentFunctionFactoryLeaderFunction) CanHandle(c *LayoutContext, function *FSFunction) bool {
	if c.IsPrint() && function.Is("leader") {
		parameters := function.GetParameters()
		if len(parameters) == 1 {
			param := parameters[0]
			return param.GetPrimitiveType() == CSSPrimitiveValueCssString ||
				param.GetPrimitiveType() == CSSPrimitiveValueCssIdent &&
					("dotted" == param.GetStringValue() ||
						"solid" == param.GetStringValue() ||
						"space" == param.GetStringValue())
		}
	}

	return false
}

var (
	_ ContentFunction = (*contentFunctionFactoryPageCounterFunction)(nil)
	_ ContentFunction = (*contentFunctionFactoryPagesCounterFunction)(nil)
	_ ContentFunction = (*contentFunctionFactoryTargetCounterFunction)(nil)
	_ ContentFunction = (*contentFunctionFactoryLeaderFunction)(nil)
)
