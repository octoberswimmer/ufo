package pdf

import (
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/dom"
)

// Flying Saucer has no unit tests of ITextReplacedElementFactory; these cover
// the branches that need no layout context or user agent.

// iTextReplacedElementFactoryTestBox is a block box with an element; every
// other BlockBoxI method panics on the nil embedded interface.
type iTextReplacedElementFactoryTestBox struct {
	ufo.BlockBoxI
	element *dom.Element
}

func (b *iTextReplacedElementFactoryTestBox) GetElement() *dom.Element { return b.element }

func iTextReplacedElementFactoryTestCreate(t *testing.T, markup string) ufo.ReplacedElement {
	t.Helper()
	doc, err := dom.ParseXMLString(markup)
	if err != nil {
		t.Fatal(err)
	}
	factory := NewITextReplacedElementFactory(NewITextOutputDevice(20))
	box := &iTextReplacedElementFactoryTestBox{element: doc.GetDocumentElement()}
	return factory.CreateReplacedElement(nil, box, nil, -1, -1)
}

func TestITextReplacedElementFactory_hiddenInputIsAnEmptyReplacedElement(t *testing.T) {
	element := iTextReplacedElementFactoryTestCreate(t, `<input type="hidden" name="a" value="b"/>`)

	empty, ok := element.(*EmptyReplacedElement)
	if !ok {
		t.Fatalf("replaced element = %T, want *EmptyReplacedElement", element)
	}
	if empty.GetIntrinsicWidth() != 1 || empty.GetIntrinsicHeight() != 1 {
		t.Errorf("intrinsic size = %dx%d, want 1x1", empty.GetIntrinsicWidth(), empty.GetIntrinsicHeight())
	}
}

func TestITextReplacedElementFactory_elementsWithoutSupportAreNotReplaced(t *testing.T) {
	for _, markup := range []string{
		`<input type="checkbox"/>`,
		`<input type="radio" name="r"/>`,
		`<input type="text"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg"/>`,
		`<img/>`,
		`<div/>`,
	} {
		if element := iTextReplacedElementFactoryTestCreate(t, markup); element != nil {
			t.Errorf("%s: replaced element = %T, want nil", markup, element)
		}
	}
}

func TestITextReplacedElementFactory_bookmarkWithoutNameHasNoAnchorName(t *testing.T) {
	element := iTextReplacedElementFactoryTestCreate(t, `<bookmark/>`)

	bookmark, ok := element.(*BookmarkElement)
	if !ok {
		t.Fatalf("replaced element = %T, want *BookmarkElement", element)
	}
	if bookmark.GetAnchorName() != nil {
		t.Errorf("anchor name = %q, want nil", *bookmark.GetAnchorName())
	}
}
