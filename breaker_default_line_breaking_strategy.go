// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/breaker/DefaultLineBreakingStrategy.java

package ufo

// DefaultLineBreakingStrategy breaks at the positions that
// UrlAwareLineBreakIterator reports.
//
// Author of the Java class: Lukas Zaruba, lukas.zaruba@gmail.com
type DefaultLineBreakingStrategy struct{}

func NewDefaultLineBreakingStrategy() *DefaultLineBreakingStrategy {
	return &DefaultLineBreakingStrategy{}
}

func (s *DefaultLineBreakingStrategy) GetBreakPointsProvider(text string, lang string, style CalculatedStyleI) BreakPointsProvider {
	var iterator BreakIteratorI = NewUrlAwareLineBreakIterator(text)
	return NewDefaultLineBreakingStrategyDefaultBreakPointsProvider(iterator)
}

// DefaultLineBreakingStrategyDefaultBreakPointsProvider is the private record
// DefaultLineBreakingStrategy.DefaultBreakPointsProvider.
type DefaultLineBreakingStrategyDefaultBreakPointsProvider struct {
	iterator BreakIteratorI
}

func NewDefaultLineBreakingStrategyDefaultBreakPointsProvider(iterator BreakIteratorI) *DefaultLineBreakingStrategyDefaultBreakPointsProvider {
	return &DefaultLineBreakingStrategyDefaultBreakPointsProvider{iterator: iterator}
}

func (p *DefaultLineBreakingStrategyDefaultBreakPointsProvider) Iterator() BreakIteratorI {
	return p.iterator
}

func (p *DefaultLineBreakingStrategyDefaultBreakPointsProvider) Next() *BreakPoint {
	next := p.iterator.Next()
	if next < 0 {
		return BreakPointGetDonePoint()
	}
	return NewBreakPoint(next)
}
