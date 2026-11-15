package pdf

import (
	"errors"
	"strings"
	"testing"

	"github.com/octoberswimmer/ufo"
	"github.com/octoberswimmer/ufo/pdf/fonts"
	"github.com/octoberswimmer/ufo/pdf/writer"
)

func fallbackTestDescription(t *testing.T) (*FontDescription, *writer.BaseFont, *writer.BaseFont) {
	t.Helper()
	helvetica, err := writer.CreateFont(writer.BaseFontHelvetica, writer.BaseFontCp1252, writer.BaseFontNotEmbedded)
	if err != nil {
		t.Fatal(err)
	}
	bundled, ok := fonts.Lookup(fonts.WenQuanYiMicroHei)
	if !ok {
		t.Fatal("WenQuanYi Micro Hei is not bundled")
	}
	fallback, err := bundled.BaseFont()
	if err != nil {
		t.Fatal(err)
	}
	description := NewFontDescriptionWithStyleWeight(helvetica, ufo.IdentValueNormal, 400)
	description.SetFallbackFont(bundled.BaseFont)
	return description, helvetica, fallback
}

// The bundled font has the scripts its notice lists.
func TestBundledFontCoverage(t *testing.T) {
	_, _, fallback := fallbackTestDescription(t)
	for _, r := range "Aé日こテ한販ПΔ№" {
		if !fallback.CharExists(r) {
			t.Errorf("WenQuanYi Micro Hei has no glyph for %q", r)
		}
	}
	for _, r := range "אعก" {
		if fallback.CharExists(r) {
			t.Errorf("WenQuanYi Micro Hei has a glyph for %q; NOTICE.md says it has none", r)
		}
	}
}

func TestFontRunsSplitATextByTheFontThatShowsEachCharacter(t *testing.T) {
	description, helvetica, fallback := fallbackTestDescription(t)
	// Hebrew is in neither font and stays with the description's font,
	// which leaves it out.
	runs := description.fontRuns("Total 税込 ¥1 אב Привет")
	var got []string
	for _, run := range runs {
		name := "helvetica"
		if run.font == fallback {
			name = "fallback"
		} else if run.font != helvetica {
			t.Fatalf("unexpected font %v", run.font)
		}
		got = append(got, name+":"+run.text)
	}
	want := "helvetica:Total |fallback:税込|helvetica: ¥1 אב |fallback:Привет"
	if strings.Join(got, "|") != want {
		t.Errorf("runs = %q, want %q", strings.Join(got, "|"), want)
	}

	plain := NewFontDescriptionWithStyleWeight(helvetica, ufo.IdentValueNormal, 400)
	if runs := plain.fontRuns("Total 税込"); len(runs) != 1 || runs[0].font != helvetica || runs[0].text != "Total 税込" {
		t.Errorf("a description without a fallback font splits its text: %v", runs)
	}
}

// The width of a text is the same whether it is measured whole or in parts,
// which line breaking relies on.
func TestGetWidthAddsTheWidthsOfBothFonts(t *testing.T) {
	description, helvetica, fallback := fallbackTestDescription(t)
	font := NewITextFSFont(description, 1000)
	renderer := NewITextTextRenderer()
	whole := renderer.GetWidth(nil, font, "Total 税込")
	want := int(helvetica.GetWidthPoint("Total ", 1000) + fallback.GetWidthPoint("税込", 1000))
	if whole != want {
		t.Errorf("GetWidth = %d, want %d", whole, want)
	}
	if latin := renderer.GetWidth(nil, font, "Total "); latin != int(helvetica.GetWidthPoint("Total ", 1000)) {
		t.Errorf("GetWidth of Windows-1252 text = %d, want the width in Helvetica", latin)
	}
}

func TestFallbackFontThatCannotBeLoadedStopsTheRendering(t *testing.T) {
	description, _, _ := fallbackTestDescription(t)
	failure := errors.New("no such font")
	description.SetFallbackFont(func() (*writer.BaseFont, error) { return nil, failure })
	defer func() {
		err, ok := recover().(error)
		if !ok || !errors.Is(err, failure) {
			t.Errorf("recovered %v, want the error of the fallback font", err)
		}
	}()
	description.fontRuns("税")
}
