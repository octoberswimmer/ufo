// Tests for render_page_box.go. Flying Saucer has no JUnit test for PageBox.
// The expected values were printed by the Java PageBox (flying-saucer-core
// 10.6.0-SNAPSHOT) for the same @page rules with a SharedContext at 96 DPI and
// one dot per pixel, with the default locale set to en_US and to en_DE.

package ufo

import (
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo/geom"
)

func pageBoxTestNewPage(t *testing.T, css string, c CssContext) *PageBox {
	t.Helper()
	parser := NewCSSParser(CSSErrorHandlerFunc(func(uri string, message string) {
		t.Errorf("CSS error: %s", message)
	}))
	sheet, err := parser.ParseStylesheet("", StylesheetInfoOriginAuthor, strings.NewReader(css))
	if err != nil {
		t.Fatal(err)
	}
	matcher := NewMatcher(nil, nil, nil, []*Stylesheet{sheet}, "print")
	info := matcher.GetPageCascadedStyle("", "first")
	style := NewEmptyStyle().DeriveStyle(info.GetPageStyle())
	return NewPageBox(info, c, style, 100, 3)
}

func pageBoxTestContext() *RenderingContext {
	sharedContext := NewSharedContext()
	sharedContext.SetDPI(96)
	sharedContext.SetDotsPerPixel(1)
	return NewRenderingContext(sharedContext, nil, nil, nil, 0)
}

type pageBoxTestArea struct {
	dim       [2]int
	screen    [2]int
	print     [2]int
	direction BoxBuilderMarginDirection
}

type pageBoxTestCase struct {
	name          string
	locale        string
	css           string
	width, height int
	contentWidth  int
	contentHeight int
	bottom        int
	pagedViewClip [4]int
	printClip     [4]int
	borderEdge    [4]int
	// areas is in the order of pageBoxMarginAreaDefs.
	areas []pageBoxTestArea
}

func pageBoxTestRect(r *geom.Rectangle) [4]int {
	return [4]int{r.X, r.Y, r.Width, r.Height}
}

func TestPageBoxGeometry(t *testing.T) {
	h := BoxBuilderMarginDirectionHorizontal
	v := BoxBuilderMarginDirectionVertical
	cases := []pageBoxTestCase{
		{
			name:   "explicit size, margins and border",
			locale: "en_US.UTF-8",
			css:    "@page { size: 400px 300px; margin: 10px 20px 30px 40px; border: 2px solid black }",
			width:  400, height: 300, contentWidth: 336, contentHeight: 256, bottom: 356,
			pagedViewClip: [4]int{47, 1012, 336, 256},
			printClip:     [4]int{42, 12, 336, 255},
			borderEdge:    [4]int{45, 1010, 340, 260},
			areas: []pageBoxTestArea{
				{[2]int{40, 10}, [2]int{5, 1000}, [2]int{0, 0}, h},
				{[2]int{336, 10}, [2]int{45, 1000}, [2]int{40, 0}, h},
				{[2]int{20, 10}, [2]int{385, 1000}, [2]int{380, 0}, h},
				{[2]int{40, 256}, [2]int{5, 1010}, [2]int{0, 10}, v},
				// The right margin area is laid out with the width of the left
				// margin, as in Java.
				{[2]int{40, 256}, [2]int{385, 1010}, [2]int{380, 10}, v},
				{[2]int{40, 30}, [2]int{5, 1270}, [2]int{0, 270}, h},
				{[2]int{336, 30}, [2]int{45, 1270}, [2]int{40, 270}, h},
				{[2]int{20, 30}, [2]int{385, 1270}, [2]int{380, 270}, h},
			},
		},
		{
			name:   "named size, landscape",
			locale: "en_US.UTF-8",
			css:    "@page { size: a5 landscape; margin: 1in 0.5in }",
			width:  794, height: 559, contentWidth: 698, contentHeight: 367, bottom: 467,
			pagedViewClip: [4]int{53, 1096, 698, 367},
			printClip:     [4]int{48, 96, 698, 366},
			borderEdge:    [4]int{53, 1096, 698, 367},
			areas: []pageBoxTestArea{
				{[2]int{48, 96}, [2]int{5, 1000}, [2]int{0, 0}, h},
				{[2]int{698, 96}, [2]int{53, 1000}, [2]int{48, 0}, h},
				{[2]int{48, 96}, [2]int{751, 1000}, [2]int{746, 0}, h},
				{[2]int{48, 367}, [2]int{5, 1096}, [2]int{0, 96}, v},
				{[2]int{48, 367}, [2]int{751, 1096}, [2]int{746, 96}, v},
				{[2]int{48, 96}, [2]int{5, 1463}, [2]int{0, 463}, h},
				{[2]int{698, 96}, [2]int{53, 1463}, [2]int{48, 463}, h},
				{[2]int{48, 96}, [2]int{751, 1463}, [2]int{746, 463}, h},
			},
		},
		{
			name:   "auto size, letter",
			locale: "en_US.UTF-8",
			css:    "@page { margin: 2cm }",
			width:  816, height: 1056, contentWidth: 664, contentHeight: 904, bottom: 1004,
			pagedViewClip: [4]int{81, 1076, 664, 904},
			printClip:     [4]int{76, 76, 664, 903},
			borderEdge:    [4]int{81, 1076, 664, 904},
			areas: []pageBoxTestArea{
				{[2]int{76, 76}, [2]int{5, 1000}, [2]int{0, 0}, h},
				{[2]int{664, 76}, [2]int{81, 1000}, [2]int{76, 0}, h},
				{[2]int{76, 76}, [2]int{745, 1000}, [2]int{740, 0}, h},
				{[2]int{76, 904}, [2]int{5, 1076}, [2]int{0, 76}, v},
				{[2]int{76, 904}, [2]int{745, 1076}, [2]int{740, 76}, v},
				{[2]int{76, 76}, [2]int{5, 1980}, [2]int{0, 980}, h},
				{[2]int{664, 76}, [2]int{81, 1980}, [2]int{76, 980}, h},
				{[2]int{76, 76}, [2]int{745, 1980}, [2]int{740, 980}, h},
			},
		},
		{
			name:   "auto size, A4",
			locale: "en_DE.UTF-8",
			css:    "@page { margin: 2cm }",
			width:  794, height: 1123, contentWidth: 642, contentHeight: 971, bottom: 1071,
			pagedViewClip: [4]int{81, 1076, 642, 971},
			printClip:     [4]int{76, 76, 642, 970},
			borderEdge:    [4]int{81, 1076, 642, 971},
			areas: []pageBoxTestArea{
				{[2]int{76, 76}, [2]int{5, 1000}, [2]int{0, 0}, h},
				{[2]int{642, 76}, [2]int{81, 1000}, [2]int{76, 0}, h},
				{[2]int{76, 76}, [2]int{723, 1000}, [2]int{718, 0}, h},
				{[2]int{76, 971}, [2]int{5, 1076}, [2]int{0, 76}, v},
				{[2]int{76, 971}, [2]int{723, 1076}, [2]int{718, 76}, v},
				{[2]int{76, 76}, [2]int{5, 2047}, [2]int{0, 1047}, h},
				{[2]int{642, 76}, [2]int{81, 2047}, [2]int{76, 1047}, h},
				{[2]int{76, 76}, [2]int{723, 2047}, [2]int{718, 1047}, h},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("LC_ALL", tc.locale)
			c := pageBoxTestContext()
			page := pageBoxTestNewPage(t, tc.css, c)
			page.SetPaintingTop(1000)
			page.SetPaintingBottom(1000 + page.GetHeight(c))

			if page.GetWidth(c) != tc.width || page.GetHeight(c) != tc.height {
				t.Errorf("size = %d x %d, want %d x %d", page.GetWidth(c), page.GetHeight(c), tc.width, tc.height)
			}
			if page.GetContentWidth(c) != tc.contentWidth || page.GetContentHeight(c) != tc.contentHeight {
				t.Errorf("content size = %d x %d, want %d x %d", page.GetContentWidth(c), page.GetContentHeight(c), tc.contentWidth, tc.contentHeight)
			}
			if page.GetTop() != 100 || page.GetBottom() != tc.bottom {
				t.Errorf("top, bottom = %d, %d, want 100, %d", page.GetTop(), page.GetBottom(), tc.bottom)
			}
			if page.GetOuterPageWidth() != tc.width {
				t.Errorf("GetOuterPageWidth = %d, want %d", page.GetOuterPageWidth(), tc.width)
			}
			if page.GetPageNo() != 3 || !page.IsLeftPage() || page.IsRightPage() {
				t.Errorf("page no %d, left %v, right %v", page.GetPageNo(), page.IsLeftPage(), page.IsRightPage())
			}
			if got, want := pageBoxTestRect(page.GetScreenPaintingBounds(c, 5)), [4]int{5, 1000, tc.width, tc.height}; got != want {
				t.Errorf("GetScreenPaintingBounds = %v, want %v", got, want)
			}
			if got, want := pageBoxTestRect(page.GetPrintPaintingBounds(c)), [4]int{0, 0, tc.width, tc.height}; got != want {
				t.Errorf("GetPrintPaintingBounds = %v, want %v", got, want)
			}
			if got := pageBoxTestRect(page.GetPagedViewClippingBounds(c, 5)); got != tc.pagedViewClip {
				t.Errorf("GetPagedViewClippingBounds = %v, want %v", got, tc.pagedViewClip)
			}
			if got := pageBoxTestRect(page.GetPrintClippingBounds(c)); got != tc.printClip {
				t.Errorf("GetPrintClippingBounds = %v, want %v", got, tc.printClip)
			}
			if got := pageBoxTestRect(page.getBorderEdge(5, 1000, c)); got != tc.borderEdge {
				t.Errorf("getBorderEdge = %v, want %v", got, tc.borderEdge)
			}
			for i, area := range pageBoxMarginAreaDefs {
				want := tc.areas[i]
				dim := area.getLayoutDimension(c, page, page.GetMargin(c))
				if got := [2]int{dim.Width, dim.Height}; got != want.dim {
					t.Errorf("area %d: layout dimension = %v, want %v", i, got, want.dim)
				}
				screen := area.getPaintingPosition(c, page, 5, LayerPagedModePagedModeScreen)
				if got := [2]int{screen.X, screen.Y}; got != want.screen {
					t.Errorf("area %d: screen painting position = %v, want %v", i, got, want.screen)
				}
				printPos := area.getPaintingPosition(c, page, 0, LayerPagedModePagedModePrint)
				if got := [2]int{printPos.X, printPos.Y}; got != want.print {
					t.Errorf("area %d: print painting position = %v, want %v", i, got, want.print)
				}
				if area.getDirection() != want.direction {
					t.Errorf("area %d: direction = %v, want %v", i, area.getDirection(), want.direction)
				}
			}
		})
	}
}

func TestPageBoxMarginAreaNames(t *testing.T) {
	want := [][]*MarginBoxName{
		{MarginBoxNameTopLeftCorner},
		{MarginBoxNameTopLeft, MarginBoxNameTopCenter, MarginBoxNameTopRight},
		{MarginBoxNameTopRightCorner},
		{MarginBoxNameLeftTop, MarginBoxNameLeftMiddle, MarginBoxNameLeftBottom},
		{MarginBoxNameRightTop, MarginBoxNameRightMiddle, MarginBoxNameRightBottom},
		{MarginBoxNameBottomLeftCorner},
		{MarginBoxNameBottomLeft, MarginBoxNameBottomCenter, MarginBoxNameBottomRight},
		{MarginBoxNameBottomRightCorner},
	}
	if len(pageBoxMarginAreaDefs) != len(want) {
		t.Fatalf("%d margin areas, want %d", len(pageBoxMarginAreaDefs), len(want))
	}
	for i, area := range pageBoxMarginAreaDefs {
		got := area.getMarginBoxNames()
		if len(got) != len(want[i]) {
			t.Errorf("area %d: names %v, want %v", i, got, want[i])
			continue
		}
		for j := range got {
			if got[j] != want[i][j] {
				t.Errorf("area %d: names %v, want %v", i, got, want[i])
				break
			}
		}
	}
	// The first five areas are exported as leading text, the rest as trailing
	// text.
	if pageBoxLeadingTrailingSplit != 5 {
		t.Errorf("pageBoxLeadingTrailingSplit = %d", pageBoxLeadingTrailingSplit)
	}
}

func TestPageBoxContentHeightNotPositive(t *testing.T) {
	t.Setenv("LC_ALL", "en_US.UTF-8")
	defer func() {
		e, ok := recover().(*XRRuntimeException)
		if !ok || e.Error() != "The content height cannot be zero or less.  Check your document margin definition." {
			t.Errorf("panic = %v", e)
		}
	}()
	pageBoxTestNewPage(t, "@page { size: 400px 300px; margin: 150px 10px }", pageBoxTestContext())
}

func TestPageBoxContentWidthNotPositive(t *testing.T) {
	t.Setenv("LC_ALL", "en_US.UTF-8")
	c := pageBoxTestContext()
	page := pageBoxTestNewPage(t, "@page { size: 400px 300px; margin: 10px 200px }", c)
	defer func() {
		e, ok := recover().(*XRRuntimeException)
		if !ok || e.Error() != "The content width cannot be zero or less.  Check your document margin definition." {
			t.Errorf("panic = %v", e)
		}
	}()
	page.GetContentWidth(c)
}

func TestPageBoxDefaultLocaleCountry(t *testing.T) {
	cases := []struct {
		lcAll, lcMessages, lang string
		want                    string
	}{
		{"", "", "", ""},
		{"", "", "C", ""},
		{"", "", "POSIX", ""},
		{"", "", "en_US.UTF-8", "US"},
		{"", "", "fr_CA", "CA"},
		{"", "", "es_MX.ISO-8859-1", "MX"},
		{"", "", "de_DE@euro", "DE"},
		{"", "en_GB.UTF-8", "en_US.UTF-8", "GB"},
		{"ja_JP.UTF-8", "en_GB.UTF-8", "en_US.UTF-8", "JP"},
		{"", "", "de", ""},
	}
	for _, tc := range cases {
		t.Setenv("LC_ALL", tc.lcAll)
		t.Setenv("LC_MESSAGES", tc.lcMessages)
		t.Setenv("LANG", tc.lang)
		if got := pageBoxDefaultLocaleCountry(); got != tc.want {
			t.Errorf("LC_ALL=%q LC_MESSAGES=%q LANG=%q: country %q, want %q", tc.lcAll, tc.lcMessages, tc.lang, got, tc.want)
		}
	}
}
