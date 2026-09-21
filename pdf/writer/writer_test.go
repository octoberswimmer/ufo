package writer

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const testTrueTypeFont = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"

type testDoc struct {
	t   *testing.T
	buf bytes.Buffer
	doc *Document
	w   *PdfWriter
}

func newTestDoc(t *testing.T, width, height float32, compress bool) *testDoc {
	t.Helper()
	td := &testDoc{t: t}
	td.doc = NewDocument(NewRectangle(0, 0, width, height), 0, 0, 0, 0)
	w, err := PdfWriterGetInstance(td.doc, &td.buf)
	if err != nil {
		t.Fatal(err)
	}
	if !compress {
		w.SetCompressionLevel(PdfStreamNoCompression)
	}
	td.w = w
	td.doc.Open()
	return td
}

func (td *testDoc) close() *ReadDocument {
	td.t.Helper()
	if err := td.doc.Close(); err != nil {
		td.t.Fatal(err)
	}
	rd, err := ReadPDF(td.buf.Bytes())
	if err != nil {
		td.t.Fatalf("reading the written file: %v", err)
	}
	return rd
}

func helvetica(t *testing.T) *BaseFont {
	t.Helper()
	bf, err := CreateFont(BaseFontHelvetica, BaseFontCp1252, BaseFontNotEmbedded)
	if err != nil {
		t.Fatal(err)
	}
	return bf
}

func showAt(cb *PdfContentByte, bf *BaseFont, size, x, y float32, text string) {
	cb.BeginText()
	cb.SetFontAndSize(bf, size)
	cb.SetTextMatrix(1, 0, 0, 1, x, y)
	cb.ShowText(text)
	cb.EndText()
}

// writeTemp writes the PDF to a file for the external tools.
func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func runQpdfCheck(t *testing.T, data []byte) {
	t.Helper()
	qpdf, err := exec.LookPath("qpdf")
	if err != nil {
		t.Log("qpdf not found; skipping qpdf --check")
		return
	}
	out, err := exec.Command(qpdf, "--check", writeTemp(t, data)).CombinedOutput()
	if err != nil {
		t.Fatalf("qpdf --check failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "WARNING") {
		t.Fatalf("qpdf --check reported warnings:\n%s", out)
	}
}

func runPdftotext(t *testing.T, data []byte) (string, bool) {
	t.Helper()
	tool, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Log("pdftotext not found; skipping")
		return "", false
	}
	cmd := exec.Command(tool, "-enc", "UTF-8", writeTemp(t, data), "-")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("pdftotext failed: %v\n%s", err, stderr.String())
	}
	if stderr.Len() > 0 {
		t.Fatalf("pdftotext reported errors:\n%s", stderr.String())
	}
	return string(out), true
}

func runPdfinfo(t *testing.T, data []byte) (string, bool) {
	t.Helper()
	tool, err := exec.LookPath("pdfinfo")
	if err != nil {
		t.Log("pdfinfo not found; skipping")
		return "", false
	}
	cmd := exec.Command(tool, "-enc", "UTF-8", "-l", "10", writeTemp(t, data))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("pdfinfo failed: %v\n%s", err, stderr.String())
	}
	if stderr.Len() > 0 {
		t.Fatalf("pdfinfo reported errors:\n%s", stderr.String())
	}
	return string(out), true
}

func TestFormatNumber(t *testing.T) {
	cases := map[float64]string{
		0: "0", 1: "1", -1: "-1", 0.5: "0.5", 27.336000442504883: "27.34", 1.00004: "1", 1.00005: "1",
		-0.00001: "0", 1234567: "1234567", 0.21256: "0.21256", math.NaN(): "0", math.Inf(1): "0",
	}
	for in, want := range cases {
		if got := formatNumber(in); got != want {
			t.Errorf("formatNumber(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestTwoPagesOfDifferentSizes(t *testing.T) {
	td := newTestDoc(t, 595, 842, true)
	bf := helvetica(t)
	showAt(td.w.GetDirectContent(), bf, 12, 72, 770, "First page")
	if got := td.w.GetPageNumber(); got != 1 {
		t.Errorf("GetPageNumber on the first page = %d, want 1", got)
	}
	td.doc.SetPageSize(NewRectangle(0, 0, 300.5, 200))
	if !td.doc.NewPage() {
		t.Fatal("NewPage reported false for a page with content")
	}
	if got := td.w.GetPageNumber(); got != 2 {
		t.Errorf("GetPageNumber on the second page = %d, want 2", got)
	}
	showAt(td.w.GetDirectContent(), bf, 12, 20, 150, "Second page")
	// A page break on a page without content adds no page.
	td.doc.NewPage()
	if td.doc.NewPage() {
		t.Error("NewPage reported true for an empty page")
	}

	rd := td.close()
	if rd.NumPages() != 2 {
		t.Fatalf("NumPages = %d, want 2", rd.NumPages())
	}
	if w, h := rd.PageSize(0); w != 595 || h != 842 {
		t.Errorf("page 0 size = %v x %v, want 595 x 842", w, h)
	}
	if w, h := rd.PageSize(1); w != 300.5 || h != 200 {
		t.Errorf("page 1 size = %v x %v, want 300.5 x 200", w, h)
	}
	if rd.Version() != "1.4" {
		t.Errorf("Version = %q, want 1.4", rd.Version())
	}
	for i, want := range []string{"First page", "Second page"} {
		got, err := rd.PageText(i)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("PageText(%d) = %q, want %q", i, got, want)
		}
	}
	runs, err := rd.PageTextRuns(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0] != (ReadTextRun{Text: "Second page", X: 20, Y: 150}) {
		t.Errorf("PageTextRuns(1) = %+v, want [{Second page 20 150}]", runs)
	}
	text, err := rd.Text()
	if err != nil {
		t.Fatal(err)
	}
	if text != "First page\nSecond page\n" {
		t.Errorf("Text = %q", text)
	}

	runQpdfCheck(t, td.buf.Bytes())
	if out, ok := runPdftotext(t, td.buf.Bytes()); ok {
		if !strings.Contains(out, "First page") || !strings.Contains(out, "Second page") {
			t.Errorf("pdftotext output = %q", out)
		}
	}
	if out, ok := runPdfinfo(t, td.buf.Bytes()); ok {
		if !regexp.MustCompile(`Pages:\s+2\b`).MatchString(out) {
			t.Errorf("pdfinfo does not report 2 pages:\n%s", out)
		}
		if !regexp.MustCompile(`Page\s+1 size:\s+595 x 842 pts`).MatchString(out) ||
			!regexp.MustCompile(`Page\s+2 size:\s+300\.5 x 200 pts`).MatchString(out) {
			t.Errorf("pdfinfo page sizes:\n%s", out)
		}
	}
}

func TestPdfVersionAndNoPages(t *testing.T) {
	var buf bytes.Buffer
	doc := NewDocument(NewRectangle(0, 0, 100, 100), 0, 0, 0, 0)
	w, err := PdfWriterGetInstance(doc, &buf)
	if err != nil {
		t.Fatal(err)
	}
	w.SetPdfVersion(PdfWriterVersion17)
	doc.Open()
	err = doc.Close()
	var de *DocumentException
	if !errors.As(err, &de) || de.Message != "The document has no pages." {
		t.Fatalf("Close of an empty document = %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-1.7\n")) {
		t.Errorf("header = %q", buf.Bytes()[:9])
	}
}

func TestHelveticaMetrics(t *testing.T) {
	bf := helvetica(t)
	if got := bf.GetWidthString("Hello"); got != 2278 {
		t.Errorf("GetWidthString(Hello) = %d, want 2278", got)
	}
	if got, want := bf.GetWidthPoint("Hello", 12), float32(27.336); math.Abs(float64(got-want)) > 1e-4 {
		t.Errorf("GetWidthPoint(Hello, 12) = %v, want %v", got, want)
	}
	for c, want := range map[rune]int{'H': 722, 'e': 556, 'l': 222, 'o': 556, ' ': 278, 0xa0: 278, 0x20ac: 556, 0x2014: 1000, 0x0416: 0} {
		if got := bf.GetWidth(c); got != want {
			t.Errorf("GetWidth(%q) = %d, want %d", c, got, want)
		}
	}
	if got := bf.GetKerning('A', 'V'); got != -70 {
		t.Errorf("GetKerning(A, V) = %d, want -70", got)
	}
	if got := bf.GetKerning('A', 'A'); got != 0 {
		t.Errorf("GetKerning(A, A) = %d, want 0", got)
	}
	// AV: 667 + 667 - 70 = 1264.
	if got, want := bf.GetWidthPointKerned("AV", 10), float32(12.64); math.Abs(float64(got-want)) > 1e-4 {
		t.Errorf("GetWidthPointKerned(AV, 10) = %v, want %v", got, want)
	}
	descriptors := map[int]float32{
		BaseFontAscent: 718, BaseFontAwtAscent: 718, BaseFontDescent: -207, BaseFontAwtDescent: -207,
		BaseFontCapheight: 718, BaseFontItalicangle: 0,
		BaseFontBboxllx: -166, BaseFontBboxlly: -225, BaseFontBboxurx: 1000, BaseFontBboxury: 931,
		BaseFontAwtLeading: 0, BaseFontAwtMaxadvance: 1166,
		BaseFontUnderlinePosition: -100, BaseFontUnderlineThickness: 50,
		BaseFontStrikethroughPosition: 0, BaseFontStrikethroughThickness: 0,
	}
	for key, want := range descriptors {
		if got := bf.GetFontDescriptor(key, 1000); got != want {
			t.Errorf("GetFontDescriptor(%d, 1000) = %v, want %v", key, got, want)
		}
	}
	if bf.GetXHeight() != 523 || bf.GetStdVW() != 88 {
		t.Errorf("XHeight, StdVW = %d, %d, want 523, 88", bf.GetXHeight(), bf.GetStdVW())
	}
	if bf.GetPostscriptFontName() != "Helvetica" || bf.GetFamilyFontName()[0][3] != "Helvetica" || bf.GetFullFontName()[0][3] != "Helvetica" {
		t.Errorf("names = %q, %v, %v", bf.GetPostscriptFontName(), bf.GetFamilyFontName(), bf.GetFullFontName())
	}
	if !bf.CharExists('é') || !bf.CharExists(0x20ac) || bf.CharExists(0x0416) {
		t.Error("CharExists does not follow Cp1252")
	}
}

func TestAllStandardFontsLoad(t *testing.T) {
	for name := range base14Fonts {
		bf, err := CreateFont(name, BaseFontCp1252, false)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if bf.GetPostscriptFontName() != name {
			t.Errorf("%s: PostScript name %q", name, bf.GetPostscriptFontName())
		}
		if bf.GetWidth('a') == 0 {
			t.Errorf("%s: width of 'a' is 0", name)
		}
	}
	oblique, _ := CreateFont(BaseFontHelveticaOblique, BaseFontCp1252, false)
	if got := oblique.GetFontDescriptor(BaseFontItalicangle, 12); got != -12 {
		t.Errorf("Helvetica-Oblique ItalicAngle = %v, want -12", got)
	}
	courier, _ := CreateFont(BaseFontCourier, BaseFontCp1252, false)
	if !courier.IsFixedPitch() || courier.GetWidth('i') != 600 || courier.HasKernPairs() {
		t.Error("Courier is not read as a fixed pitch font of width 600 without kerning")
	}
	symbol, _ := CreateFont(BaseFontSymbol, BaseFontCp1252, false)
	// Code 0x61 of Symbol is alpha, 631 wide.
	if !symbol.IsFontSpecific() || symbol.GetWidth('a') != 631 {
		t.Errorf("Symbol: fontSpecific %v, width of code 0x61 = %d, want 631", symbol.IsFontSpecific(), symbol.GetWidth('a'))
	}
	if _, err := CreateFont("NoSuchFont", BaseFontCp1252, false); err == nil {
		t.Error("CreateFont(NoSuchFont) did not fail")
	}
}

func TestWinAnsiTextRoundTrip(t *testing.T) {
	td := newTestDoc(t, 400, 400, false)
	bf := helvetica(t)
	cb := td.w.GetDirectContent()
	text := "Caf\u00e9 (na\u00efve) \\ 5\u20ac \u2014 \u201cquoted\u201d"
	showAt(cb, bf, 10, 20, 350, text)
	// A character outside Cp1252 is left out.
	showAt(cb, bf, 10, 20, 330, "a\u0416b")

	// Two runs on one line with a gap, then one directly after the other.
	cb.BeginText()
	cb.SetFontAndSize(bf, 10)
	cb.SetTextMatrix(1, 0, 0, 1, 20, 300)
	cb.ShowText("left")
	cb.SetTextMatrix(1, 0, 0, 1, 120, 300)
	cb.ShowText("right")
	x := 120 + bf.GetWidthPoint("right", 10)
	cb.SetTextMatrix(1, 0, 0, 1, x, 300)
	cb.ShowText("most")
	cb.EndText()

	// A TJ array: a small adjustment joins, a large one separates.
	cb.BeginText()
	cb.SetFontAndSize(bf, 10)
	cb.SetTextMatrix(1, 0, 0, 1, 20, 280)
	array := NewPdfTextArray()
	array.AddString("ke")
	array.AddNumber(40)
	array.AddString("rn")
	array.AddNumber(-600)
	array.AddString("gap")
	cb.ShowTextArray(array)
	cb.EndText()

	rd := td.close()
	got, err := rd.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	want := text + "\nab\nleft rightmost\nkern gap"
	if got != want {
		t.Errorf("PageText =\n%q\nwant\n%q", got, want)
	}
	content, _ := rd.PageContent(0)
	if !bytes.Contains(content, []byte("[(ke)40(rn)-600(gap)]TJ")) {
		t.Errorf("TJ array not written as expected:\n%s", content)
	}
	if fonts := rd.PageFonts(0); len(fonts) != 1 || fonts[0] != "Helvetica" {
		t.Errorf("PageFonts = %v", fonts)
	}
	runQpdfCheck(t, td.buf.Bytes())
	if out, ok := runPdftotext(t, td.buf.Bytes()); ok {
		for _, s := range []string{text, "left", "rightmost"} {
			if !strings.Contains(out, s) {
				t.Errorf("pdftotext output lacks %q:\n%s", s, out)
			}
		}
	}
}

func TestShowTextKerned(t *testing.T) {
	td := newTestDoc(t, 200, 200, false)
	cb := td.w.GetDirectContent()
	cb.BeginText()
	cb.SetFontAndSize(helvetica(t), 12)
	cb.SetTextMatrix(1, 0, 0, 1, 10, 100)
	cb.ShowTextKerned("AVA")
	cb.EndText()
	rd := td.close()
	content, _ := rd.PageContent(0)
	if !bytes.Contains(content, []byte("[(A)70(V)80(A)]TJ")) {
		t.Errorf("kerned text:\n%s", content)
	}
	if got, _ := rd.PageText(0); got != "AVA" {
		t.Errorf("PageText = %q", got)
	}
}

// TestShowTextArrayFormat checks the TJ operand against what OpenPDF writes
// for the same PdfTextArray: adjacent numbers are added up, a sum of zero
// removes the number, and a space separates a number only from a number
// before it.
func TestShowTextArrayFormat(t *testing.T) {
	cases := []struct {
		build func(a *PdfTextArray)
		want  string
	}{
		{func(a *PdfTextArray) {
			a.AddString("W")
			a.AddNumber(-250)
			a.AddString("i")
			a.AddNumber(-250)
			a.AddString("d")
		}, "[(W)-250(i)-250(d)]TJ\n"},
		{func(a *PdfTextArray) {
			a.AddNumber(10)
			a.AddNumber(5)
			a.AddString("a")
			a.AddString("b")
		}, "[15(ab)]TJ\n"},
		{func(a *PdfTextArray) {
			a.AddString("a")
			a.AddNumber(10)
			a.AddNumber(-10)
			a.AddString("b")
		}, "[(a)(b)]TJ\n"},
	}
	for _, c := range cases {
		td := newTestDoc(t, 200, 200, false)
		cb := td.w.GetDirectContent()
		cb.BeginText()
		cb.SetFontAndSize(helvetica(t), 12)
		start := len(cb.ToPdf())
		array := NewPdfTextArray()
		c.build(array)
		cb.ShowTextArray(array)
		if got := string(cb.ToPdf()[start:]); got != c.want {
			t.Errorf("ShowTextArray wrote %q, want %q", got, c.want)
		}
		cb.EndText()
		td.close()
	}
}

func TestShowTextWithoutFontPanics(t *testing.T) {
	td := newTestDoc(t, 200, 200, false)
	defer func() {
		r := recover()
		if _, ok := r.(*DocumentException); !ok {
			t.Errorf("recovered %v, want a *DocumentException", r)
		}
	}()
	td.w.GetDirectContent().ShowText("x")
}

func TestTrueTypeRoundTrip(t *testing.T) {
	if _, err := os.Stat(testTrueTypeFont); err != nil {
		t.Skipf("TrueType font %s is not available", testTrueTypeFont)
	}
	bf, err := CreateFont(testTrueTypeFont, BaseFontIdentityH, BaseFontEmbedded)
	if err != nil {
		t.Fatal(err)
	}
	if bf.GetPostscriptFontName() != "DejaVuSans" {
		t.Errorf("PostScript name = %q", bf.GetPostscriptFontName())
	}
	family := bf.GetFamilyFontName()
	if len(family) == 0 || family[0][3] != "DejaVu Sans" {
		t.Errorf("family names = %v", family)
	}
	if full := bf.GetFullFontName(); len(full) == 0 || full[0][3] != "DejaVu Sans" {
		t.Errorf("full names = %v", full)
	}
	if bf.GetOS2WeightClass() != 400 || bf.IsItalic() || bf.IsBold() || bf.IsFixedPitch() {
		t.Errorf("weight %d italic %v bold %v fixed %v", bf.GetOS2WeightClass(), bf.IsItalic(), bf.IsBold(), bf.IsFixedPitch())
	}
	// DejaVu Sans: unitsPerEm 2048, 'A' advance 1401 -> 1401*1000/2048 = 684.
	if got := bf.GetWidth('A'); got != 684 {
		t.Errorf("GetWidth(A) = %d, want 684", got)
	}
	if !bf.CharExists('Ж') || bf.CharExists(0x4e2d) {
		t.Error("CharExists: want Ж present and U+4E2D absent in DejaVu Sans")
	}
	if a, d := bf.GetFontDescriptor(BaseFontBboxury, 10), bf.GetFontDescriptor(BaseFontBboxlly, 10); a <= 0 || d >= 0 {
		t.Errorf("BBOXURY %v, BBOXLLY %v", a, d)
	}
	if a, d := bf.GetFontDescriptor(BaseFontAscent, 10), bf.GetFontDescriptor(BaseFontDescent, 10); a <= 0 || d >= 0 {
		t.Errorf("ASCENT %v, DESCENT %v", a, d)
	}
	if up, ut := bf.GetFontDescriptor(BaseFontUnderlinePosition, 10), bf.GetFontDescriptor(BaseFontUnderlineThickness, 10); up >= 0 || ut <= 0 {
		t.Errorf("underline position %v thickness %v", up, ut)
	}
	if sp, st := bf.GetFontDescriptor(BaseFontStrikethroughPosition, 10), bf.GetFontDescriptor(BaseFontStrikethroughThickness, 10); sp <= 0 || st <= 0 {
		t.Errorf("strikethrough position %v thickness %v", sp, st)
	}
	os2, ok := bf.GetTables()["OS/2"]
	if !ok {
		t.Fatal("GetTables has no OS/2 table")
	}
	data := bf.GetFontData()
	if weight := int(data[os2[0]+4])<<8 | int(data[os2[0]+5]); weight != bf.GetOS2WeightClass() {
		t.Errorf("usWeightClass read through GetTables = %d, want %d", weight, bf.GetOS2WeightClass())
	}
	if bf.GetKerning('A', 'V') >= 0 {
		t.Errorf("GetKerning(A, V) = %d, want a negative value from the kern table", bf.GetKerning('A', 'V'))
	}

	td := newTestDoc(t, 500, 300, true)
	cb := td.w.GetDirectContent()
	lines := []string{"Привет, мир!", "Ελληνικά 𝛼 → ∑", "Zażółć gęślą jaźń"}
	for i, line := range lines {
		showAt(cb, bf, 14, 30, float32(250-30*i), line)
	}
	showAt(cb, helvetica(t), 14, 30, 100, "Latin in Helvetica")
	rd := td.close()
	got, err := rd.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	// U+1D6FC is not in DejaVu Sans and is left out.
	want := "Привет, мир!\nΕλληνικά  → ∑\nZażółć gęślą jaźń\nLatin in Helvetica"
	if bf.CharExists(0x1d6fc) {
		want = strings.Join(lines, "\n") + "\nLatin in Helvetica"
	}
	if got != want {
		t.Errorf("PageText =\n%q\nwant\n%q", got, want)
	}
	if fonts := rd.PageFonts(0); len(fonts) != 2 || fonts[0] != "DejaVuSans" || fonts[1] != "Helvetica" {
		t.Errorf("PageFonts = %v", fonts)
	}

	runQpdfCheck(t, td.buf.Bytes())
	if out, ok := runPdftotext(t, td.buf.Bytes()); ok {
		for _, s := range []string{"Привет, мир!", "Ελληνικά", "Zażółć gęślą jaźń", "Latin in Helvetica"} {
			if !strings.Contains(out, s) {
				t.Errorf("pdftotext output lacks %q:\n%s", s, out)
			}
		}
	}
	if tool, err := exec.LookPath("pdffonts"); err == nil {
		out, err := exec.Command(tool, writeTemp(t, td.buf.Bytes())).CombinedOutput()
		if err != nil {
			t.Fatalf("pdffonts: %v\n%s", err, out)
		}
		if !regexp.MustCompile(`DejaVuSans\s+CID TrueType\s+Identity-H\s+yes`).Match(out) {
			t.Errorf("pdffonts does not report an embedded CID TrueType font:\n%s", out)
		}
	}
}

func TestTrueTypeFromBytes(t *testing.T) {
	data, err := os.ReadFile(testTrueTypeFont)
	if err != nil {
		t.Skipf("TrueType font %s is not available", testTrueTypeFont)
	}
	bf, err := CreateFontWithBytes("anything.ttf", BaseFontIdentityH, true, false, data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bf.GetFontType() != BaseFontFontTypeTtuni || !bf.IsEmbedded() {
		t.Errorf("font type %d embedded %v", bf.GetFontType(), bf.IsEmbedded())
	}
	if _, err := CreateFontWithBytes("bad.ttf", BaseFontIdentityH, true, false, []byte("not a font"), nil); err == nil {
		t.Error("CreateFontWithBytes accepted bytes that are not a font")
	}
}

func TestTrueTypeCollectionMember(t *testing.T) {
	data, err := os.ReadFile(testTrueTypeFont)
	if err != nil {
		t.Skipf("TrueType font %s is not available", testTrueTypeFont)
	}
	// A collection with one member: the ttcf header, then the font with its
	// table offsets moved by the header length.
	header := []byte{'t', 't', 'c', 'f', 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 16}
	ttc := append(append([]byte{}, header...), data...)
	numTables := int(data[4])<<8 | int(data[5])
	for i := 0; i < numTables; i++ {
		rec := 16 + 12 + 16*i + 8
		off := uint32(ttc[rec])<<24 | uint32(ttc[rec+1])<<16 | uint32(ttc[rec+2])<<8 | uint32(ttc[rec+3])
		off += 16
		ttc[rec], ttc[rec+1], ttc[rec+2], ttc[rec+3] = byte(off>>24), byte(off>>16), byte(off>>8), byte(off)
	}
	names, err := EnumerateTTCNamesBytes(ttc)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "DejaVu Sans" {
		t.Errorf("EnumerateTTCNamesBytes = %v", names)
	}
	bf, err := CreateFontWithBytes("fonts.ttc,0", BaseFontIdentityH, true, false, ttc, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bf.GetWidth('A') != 684 {
		t.Errorf("GetWidth(A) = %d, want 684", bf.GetWidth('A'))
	}
	if bytes.HasPrefix(bf.GetFontData(), []byte("ttcf")) {
		t.Error("the embedded data is the collection, not the member font")
	}
	if _, err := CreateFontWithBytes("fonts.ttc,3", BaseFontIdentityH, true, false, ttc, nil); err == nil {
		t.Error("an index past the end of the collection was accepted")
	}
}

func testJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 6), uint8(y * 8), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func testPNGWithAlpha(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	for x := 0; x < 4; x++ {
		img.SetNRGBA(x, 0, color.NRGBA{255, 0, 0, uint8(x * 85)})
		img.SetNRGBA(x, 1, color.NRGBA{0, 0, 255, 255})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestImages(t *testing.T) {
	jpegData := testJPEG(t)
	jpg, err := ImageGetInstance(jpegData)
	if err != nil {
		t.Fatal(err)
	}
	if jpg.GetWidth() != 40 || jpg.GetHeight() != 30 || jpg.GetColorComponents() != 3 || !jpg.IsJpeg() {
		t.Errorf("JPEG: %v x %v, %d components", jpg.GetWidth(), jpg.GetHeight(), jpg.GetColorComponents())
	}
	pngImg, err := ImageGetInstance(testPNGWithAlpha(t))
	if err != nil {
		t.Fatal(err)
	}
	if pngImg.GetWidth() != 4 || pngImg.GetHeight() != 2 || !pngImg.HasAlpha() {
		t.Errorf("PNG: %v x %v, alpha %v", pngImg.GetWidth(), pngImg.GetHeight(), pngImg.HasAlpha())
	}

	// Scaling as ITextFSImage.scale does it: a copy with its own size.
	scaled := ImageGetInstanceFromImage(jpg)
	scaled.ScaleAbsolute(80, 60)
	if scaled.GetPlainWidth() != 80 || scaled.GetScaledHeight() != 60 || jpg.GetPlainWidth() != 40 {
		t.Errorf("scaled %v x %v, original %v", scaled.GetPlainWidth(), scaled.GetScaledHeight(), jpg.GetPlainWidth())
	}
	scaled.ScaleToFit(20, 20)
	if scaled.GetScaledWidth() != 20 || scaled.GetScaledHeight() != 15 {
		t.Errorf("ScaleToFit gave %v x %v", scaled.GetScaledWidth(), scaled.GetScaledHeight())
	}

	td := newTestDoc(t, 300, 300, true)
	cb := td.w.GetDirectContent()
	if err := cb.AddImage(jpg, 80, 0, 0, 60, 10, 200); err != nil {
		t.Fatal(err)
	}
	if err := cb.AddImage(scaled, 20, 0, 0, 15, 100, 200); err != nil {
		t.Fatal(err)
	}
	if err := cb.AddImage(pngImg, 40, 0, 0, 20, 10, 100); err != nil {
		t.Fatal(err)
	}
	pngImg.SetAbsolutePosition(150, 100)
	if err := cb.AddImageAbsolute(pngImg); err != nil {
		t.Fatal(err)
	}
	rd := td.close()
	images, err := rd.PageImages(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 2 {
		t.Fatalf("the page has %d image XObjects, want 2 (the scaled copy shares one)", len(images))
	}
	j, p := images[0], images[1]
	if j.Filter != "DCTDecode" || j.Width != 40 || j.Height != 30 || j.ColorSpace != "DeviceRGB" || j.HasSMask || !bytes.Equal(j.Data, jpegData) {
		t.Errorf("JPEG XObject: %+v (data equal: %v)", *j, bytes.Equal(j.Data, jpegData))
	}
	wantPixels := []byte{255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255}
	if p.Filter != "FlateDecode" || p.Width != 4 || p.Height != 2 || p.ColorSpace != "DeviceRGB" || !p.HasSMask || !bytes.Equal(p.Data, wantPixels) {
		t.Errorf("PNG XObject: %+v", *p)
	}
	content, _ := rd.PageContent(0)
	if !bytes.Contains(content, []byte("q 80 0 0 60 10 200 cm\n/Im1 Do Q\n")) || !bytes.Contains(content, []byte("q 4 0 0 2 150 100 cm\n/Im2 Do Q\n")) {
		t.Errorf("image operators:\n%s", content)
	}

	runQpdfCheck(t, td.buf.Bytes())
	if tool, err := exec.LookPath("pdfimages"); err == nil {
		out, err := exec.Command(tool, "-list", writeTemp(t, td.buf.Bytes())).CombinedOutput()
		if err != nil {
			t.Fatalf("pdfimages: %v\n%s", err, out)
		}
		for _, re := range []string{`image\s+40\s+30\s+rgb\s+3\s+8\s+jpeg`, `image\s+4\s+2\s+rgb\s+3\s+8\s+image`, `smask\s+4\s+2\s+gray\s+1\s+8\s+image`} {
			if !regexp.MustCompile(re).Match(out) {
				t.Errorf("pdfimages -list lacks %s:\n%s", re, out)
			}
		}
	}
}

func TestImageDPI(t *testing.T) {
	// JFIF APP0 with 300 dpi, then a baseline frame header for 2 x 1, 1 component.
	data := []byte{0xff, 0xd8,
		0xff, 0xe0, 0, 16, 'J', 'F', 'I', 'F', 0, 1, 1, 1, 0x01, 0x2c, 0x01, 0x2c, 0, 0,
		0xff, 0xc0, 0, 11, 8, 0, 1, 0, 2, 1, 1, 0x11, 0,
		0xff, 0xd9}
	img, err := ImageGetInstance(data)
	if err != nil {
		t.Fatal(err)
	}
	if img.GetDpiX() != 300 || img.GetDpiY() != 300 || img.GetWidth() != 2 || img.GetHeight() != 1 || img.GetColorComponents() != 1 {
		t.Errorf("dpi %d x %d, size %v x %v, components %d", img.GetDpiX(), img.GetDpiY(), img.GetWidth(), img.GetHeight(), img.GetColorComponents())
	}
	if _, err := ImageGetInstance([]byte("GARBAGE DATA")); err == nil {
		t.Error("ImageGetInstance accepted unrecognized data")
	}
}

func TestPathsClipAndStateBalance(t *testing.T) {
	td := newTestDoc(t, 200, 200, false)
	cb := td.w.GetDirectContent()
	cb.SaveState()
	cb.SetLineWidth(0.5)
	cb.SetLineCap(PdfContentByteLineCapRound)
	cb.SetLineJoin(PdfContentByteLineJoinBevel)
	cb.SetMiterLimit(10)
	cb.SetLineDash([]float32{3, 1.5}, 0)
	cb.Rectangle(10, 10, 100, 80)
	cb.Clip()
	cb.NewPath()
	cb.SaveState()
	cb.ConcatCTM(1, 0, 0, 1, 5, 5)
	cb.SetColorFill(NewRGBColor(255, 0, 0))
	cb.SetColorStroke(NewCMYKColor(0, 1, 0.5, 0.25))
	cb.MoveTo(20, 20)
	cb.LineTo(60, 20)
	cb.CurveTo(70, 20, 80, 30, 80, 40)
	cb.ClosePath()
	cb.FillStroke()
	cb.MoveTo(0, 0)
	cb.LineTo(1, 1)
	cb.EoClip()
	cb.NewPath()
	cb.RestoreState()
	cb.SetColorFill(NewRGBColorWithAlpha(0, 0, 255, 128))
	cb.Rectangle(30, 30, 10, 10)
	cb.EoFill()
	cb.SetLiteral("[]0 d\n")
	cb.MoveTo(0, 0)
	cb.LineTo(10, 10)
	cb.Stroke()
	cb.RestoreState()

	rd := td.close()
	content, err := rd.PageContent(0)
	if err != nil {
		t.Fatal(err)
	}
	want := "q\n0.5 w\n1 J\n2 j\n10 M\n[3 1.5] 0 d\n10 10 100 80 re\nW\nn\nq\n1 0 0 1 5 5 cm\n1 0 0 rg\n0 1 0.5 0.25 K\n" +
		"20 20 m\n60 20 l\n70 20 80 30 80 40 c\nh\nB\n0 0 m\n1 1 l\nW*\nn\nQ\n/GS1 gs\n0 0 1 rg\n30 30 10 10 re\nf*\n[]0 d\n0 0 m\n10 10 l\nS\nQ\n"
	if string(content) != want {
		t.Errorf("content =\n%s\nwant\n%s", content, want)
	}
	fields := strings.Fields(string(content))
	q, bigQ := 0, 0
	for _, f := range fields {
		switch f {
		case "q":
			q++
		case "Q":
			bigQ++
		}
	}
	if q != 2 || bigQ != 2 {
		t.Errorf("q count %d, Q count %d, want 2 and 2", q, bigQ)
	}
	if !bytes.Contains(td.buf.Bytes(), []byte("/ca 0.50196")) {
		t.Error("the ExtGState for the half transparent fill was not written")
	}
	runQpdfCheck(t, td.buf.Bytes())
}

func TestUnbalancedStatePanics(t *testing.T) {
	td := newTestDoc(t, 200, 200, false)
	cb := td.w.GetDirectContent()
	func() {
		defer func() {
			if _, ok := recover().(*DocumentException); !ok {
				t.Error("RestoreState without SaveState did not panic with a *DocumentException")
			}
		}()
		cb.RestoreState()
	}()
	cb.Reset()
	cb.SaveState()
	defer func() {
		if _, ok := recover().(*DocumentException); !ok {
			t.Error("ending a page with an open SaveState did not panic with a *DocumentException")
		}
	}()
	td.doc.NewPage()
}

func TestOpacityGStatesAreShared(t *testing.T) {
	td := newTestDoc(t, 200, 200, false)
	cb := td.w.GetDirectContent()
	for i := 0; i < 3; i++ {
		gs := NewPdfGState()
		gs.SetBlendMode(PdfGStateBmNormal)
		gs.SetFillOpacity(0.25)
		cb.SetGState(gs)
		cb.Rectangle(0, 0, 10, 10)
		cb.Fill()
	}
	td.close()
	if n := bytes.Count(td.buf.Bytes(), []byte("/Type/ExtGState")); n != 1 {
		t.Errorf("%d ExtGState objects written, want 1", n)
	}
	if !bytes.Contains(td.buf.Bytes(), []byte("/BM/Normal/ca 0.25")) {
		t.Error("ExtGState entries not written")
	}
}

func TestTemplates(t *testing.T) {
	td := newTestDoc(t, 300, 300, true)
	cb := td.w.GetDirectContent()
	tpl := cb.CreateTemplate(100, 40)
	cb.AddTemplate(tpl, 1, 0, 0, 1, 50, 200)
	cb.AddTemplateAt(tpl, 50, 100)
	// The content is added after the template was drawn.
	tpl.SaveState()
	tpl.SetColorFill(NewGrayColor(0.5))
	tpl.Rectangle(0, 0, 100, 40)
	tpl.Fill()
	tpl.RestoreState()
	showAt(&tpl.PdfContentByte, helvetica(t), 10, 5, 15, "in template")
	showAt(cb, helvetica(t), 10, 50, 50, "on page")

	rd := td.close()
	got, err := rd.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	if want := "in template\nin template\non page"; got != want {
		t.Errorf("PageText = %q, want %q", got, want)
	}
	if n := bytes.Count(td.buf.Bytes(), []byte("/Subtype/Form")); n != 1 {
		t.Errorf("%d form XObjects written, want 1", n)
	}
	runQpdfCheck(t, td.buf.Bytes())
	if out, ok := runPdftotext(t, td.buf.Bytes()); ok {
		if strings.Count(out, "in template") != 2 || !strings.Contains(out, "on page") {
			t.Errorf("pdftotext output = %q", out)
		}
	}
}

func TestOutlinesLinksAndNamedDestinations(t *testing.T) {
	td := newTestDoc(t, 400, 600, true)
	w := td.w
	bf := helvetica(t)
	showAt(w.GetDirectContent(), bf, 12, 50, 550, "Chapter one")

	// The default destination of ITextOutputDevice.initializePage.
	defaultDest := NewPdfDestinationWithParameter(PdfDestinationFith, 600)
	defaultDest.AddPage(w.GetPageReference(1))

	// A link to a page that does not exist yet, and a URI link.
	dest := NewPdfDestinationWithLeftTopZoom(PdfDestinationXyz, 0, 480.5, 0)
	dest.AddPage(w.GetPageReference(2))
	action := NewPdfAction()
	action.Put(PdfNameS, PdfNameGoto)
	action.Put(PdfNameD, dest)
	annot := NewPdfAnnotation(w, 50, 540, 150, 560, action)
	annot.Put(PdfNameSubtype, PdfNameLink)
	annot.Put(PdfNameF, NewPdfNumber(PdfAnnotationFlagsPrint))
	annot.SetBorderStyle(NewPdfBorderDictionary(0, 0))
	annot.SetBorder(NewPdfBorderArray(0, 0, 0))
	w.AddAnnotation(annot)
	w.AddAnnotation(NewPdfAnnotation(w, 50, 500, 150, 520, NewPdfActionWithUrl("https://example.com/a?b=(c)")))

	td.doc.NewPage()
	showAt(w.GetDirectContent(), bf, 12, 50, 550, "Chapter two")

	root := w.GetRootOutline()
	one := NewPdfOutline(root, defaultDest, "Chapter one")
	sectionDest := NewPdfDestinationWithLeftTopZoom(PdfDestinationXyz, 0, 300, 0)
	sectionDest.AddPage(w.GetPageReference(1))
	section := NewPdfOutline(one, sectionDest, "Section 1.1 \u2013 \u0416")
	NewPdfOutline(section, sectionDest, "Deep")
	twoDest := NewPdfDestinationWithLeftTopZoom(PdfDestinationXyz, 0, 600, 0)
	twoDest.AddPage(w.GetPageReference(2))
	NewPdfOutlineWithOpen(root, twoDest, "Chapter two", false)

	// Named destinations as ITextOutputDevice.writeNamedDestinations.
	names := NewPdfArray()
	names.Add(NewPdfStringWithEncoding("anchor", PdfObjectTextUnicode))
	named := NewPdfDestinationWithLeftTopZoom(PdfDestinationXyz, 0, 123, 0)
	named.AddPage(w.GetPageReference(2))
	obj, err := w.AddToBody(named)
	if err != nil {
		t.Fatal(err)
	}
	names.Add(obj.GetIndirectReference())
	nameTree := NewPdfDictionary()
	nameTree.Put(PdfNameNames, names)
	treeObj, err := w.AddToBody(nameTree)
	if err != nil {
		t.Fatal(err)
	}
	namesDict := NewPdfDictionary()
	namesDict.Put(PdfNameDests, treeObj.GetIndirectReference())
	namesObj, err := w.AddToBody(namesDict)
	if err != nil {
		t.Fatal(err)
	}
	w.GetExtraCatalog().Put(PdfNameNames, namesObj.GetIndirectReference())
	w.GetExtraCatalog().Put(PdfNameLang, NewPdfString("en"))
	w.SetViewerPreferences(PdfWriterDisplayDocTitle)

	rd := td.close()

	outlines := rd.Outlines()
	if len(outlines) != 2 {
		t.Fatalf("%d top-level outlines, want 2", len(outlines))
	}
	o1, o2 := outlines[0], outlines[1]
	if o1.Title != "Chapter one" || o1.Dest == nil || o1.Dest.Page != 0 || o1.Dest.Type != "FitH" || o1.Dest.Top != 600 || !o1.Open {
		t.Errorf("first outline: %+v dest %+v", *o1, o1.Dest)
	}
	if len(o1.Children) != 1 || o1.Children[0].Title != "Section 1.1 \u2013 \u0416" || o1.Children[0].Dest.Top != 300 || o1.Children[0].Dest.Type != "XYZ" {
		t.Fatalf("children of the first outline: %+v", o1.Children)
	}
	if deep := o1.Children[0].Children; len(deep) != 1 || deep[0].Title != "Deep" {
		t.Errorf("nested outline: %+v", deep)
	}
	if o2.Title != "Chapter two" || o2.Dest == nil || o2.Dest.Page != 1 || o2.Dest.Top != 600 || len(o2.Children) != 0 {
		t.Errorf("second outline: %+v dest %+v", *o2, o2.Dest)
	}

	links := rd.LinkAnnotations(0)
	if len(links) != 2 {
		t.Fatalf("%d links on page 0, want 2", len(links))
	}
	if l := links[0]; l.Rect != [4]float64{50, 540, 150, 560} || l.Dest == nil || l.Dest.Page != 1 || l.Dest.Type != "XYZ" || l.Dest.Top != 480.5 || l.URI != "" {
		t.Errorf("GoTo link: %+v dest %+v", *l, l.Dest)
	}
	if l := links[1]; l.URI != "https://example.com/a?b=(c)" || l.Dest != nil || l.Rect != [4]float64{50, 500, 150, 520} {
		t.Errorf("URI link: %+v", *l)
	}
	if len(rd.LinkAnnotations(1)) != 0 {
		t.Error("page 1 has links")
	}

	dests := rd.NamedDestinations()
	if d := dests["anchor"]; d == nil || d.Page != 1 || d.Top != 123 {
		t.Errorf("named destinations: %+v", dests)
	}
	catalog := rd.Catalog()
	if catalog["Lang"] != "en" || catalog["PageMode"] != "UseOutlines" {
		t.Errorf("catalog: %v", catalog)
	}
	if !bytes.Contains(td.buf.Bytes(), []byte("/ViewerPreferences<</DisplayDocTitle true>>")) {
		t.Error("viewer preferences not written")
	}

	runQpdfCheck(t, td.buf.Bytes())
	if tool, err := exec.LookPath("mutool"); err == nil {
		out, err := exec.Command(tool, "show", writeTemp(t, td.buf.Bytes()), "outline").CombinedOutput()
		if err != nil {
			t.Fatalf("mutool show outline: %v\n%s", err, out)
		}
		for _, s := range []string{"Chapter one", "Section 1.1", "Deep", "Chapter two"} {
			if !bytes.Contains(out, []byte(s)) {
				t.Errorf("mutool outline lacks %q:\n%s", s, out)
			}
		}
	}
}

func TestInfoDictionary(t *testing.T) {
	td := newTestDoc(t, 200, 200, true)
	info := td.w.GetInfo()
	info.Put(PdfNameProducer, NewPdfString("the producer"))
	info.Put(PdfNameCreator, NewPdfString("the creator (v1)"))
	td.w.GetDirectContent().Rectangle(0, 0, 1, 1)
	td.w.GetDirectContent().Fill()
	title := "Žluťoučký kůň \u2014 заголовок"
	td.doc.AddTitle(title)
	td.doc.AddAuthor("An Author")
	td.doc.AddSubject("Sujet \u00e9l\u00e9gant")
	td.doc.AddKeywords("one, two")
	td.doc.AddHeader("CustomKey", "custom value")

	rd := td.close()
	got := rd.Info()
	want := map[string]string{
		"Producer": "the producer", "Creator": "the creator (v1)", "Title": title, "Author": "An Author",
		"Subject": "Sujet \u00e9l\u00e9gant", "Keywords": "one, two", "CustomKey": "custom value",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Info[%s] = %q, want %q", k, got[k], v)
		}
	}
	if !strings.HasPrefix(got["CreationDate"], "D:") || got["ModDate"] != got["CreationDate"] {
		t.Errorf("dates: %q, %q", got["CreationDate"], got["ModDate"])
	}
	// The title is outside PDFDocEncoding, so it is UTF-16BE with a byte
	// order mark; the Latin-1 subject is PDFDocEncoding.
	if !bytes.Contains(td.buf.Bytes(), []byte("/Title(\xfe\xff\x01\x7d")) {
		t.Error("the title is not written as UTF-16BE with a byte order mark")
	}
	if !bytes.Contains(td.buf.Bytes(), []byte("/Subject(Sujet \xe9l\xe9gant)")) {
		t.Error("the subject is not written in PDFDocEncoding")
	}

	runQpdfCheck(t, td.buf.Bytes())
	if out, ok := runPdfinfo(t, td.buf.Bytes()); ok {
		for _, s := range []string{"Title:", title, "An Author", "Sujet \u00e9l\u00e9gant", "one, two", "the producer", "the creator (v1)"} {
			if !strings.Contains(out, s) {
				t.Errorf("pdfinfo output lacks %q:\n%s", s, out)
			}
		}
	}
}

func TestPdfDocEncodingRoundTrip(t *testing.T) {
	for _, s := range []string{"plain", "caf\u00e9 \u2022 \u20ac \u0141\u00f3d\u017a", "tab\tnew\nline", "\u4e2d\u6587 \U0001F600", ""} {
		if got := decodeTextString(encodeTextString(s)); got != s {
			t.Errorf("round trip of %q gave %q", s, got)
		}
	}
	if b := encodeTextString("\u20ac"); len(b) != 1 || b[0] != 0xa0 {
		t.Errorf("euro sign encoded as % x, want a0", b)
	}
}

type recordingPageEvent struct{ events []string }

func (r *recordingPageEvent) OnOpenDocument(*PdfWriter, *Document) {
	r.events = append(r.events, "open")
}
func (r *recordingPageEvent) OnStartPage(*PdfWriter, *Document) { r.events = append(r.events, "start") }
func (r *recordingPageEvent) OnEndPage(w *PdfWriter, _ *Document) {
	r.events = append(r.events, "end")
	cb := w.GetDirectContent()
	cb.Rectangle(0, 0, 5, 5)
	cb.Fill()
}
func (r *recordingPageEvent) OnCloseDocument(*PdfWriter, *Document) {
	r.events = append(r.events, "close")
}

func TestPageEvents(t *testing.T) {
	var buf bytes.Buffer
	doc := NewDocument(NewRectangle(0, 0, 100, 100), 0, 0, 0, 0)
	w, err := PdfWriterGetInstance(doc, &buf)
	if err != nil {
		t.Fatal(err)
	}
	w.SetCompressionLevel(PdfStreamNoCompression)
	events := &recordingPageEvent{}
	w.SetPageEvent(events)
	doc.Open()
	w.GetDirectContent().NewPath()
	doc.NewPage()
	w.GetDirectContent().NewPath()
	if err := doc.Close(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(events.events, " "); got != "open start end start end close" {
		t.Errorf("events = %q", got)
	}
	rd, err := ReadPDF(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	content, _ := rd.PageContent(1)
	if string(content) != "n\n0 0 5 5 re\nf\n" {
		t.Errorf("content of page 1 = %q", content)
	}
}

func TestUnsupportedFeatures(t *testing.T) {
	td := newTestDoc(t, 100, 100, false)
	w := td.w
	checks := map[string]error{
		"SetEncryption":      w.SetEncryption(nil, nil, PdfWriterAllowPrinting, PdfWriterStandardEncryption128),
		"SetPDFXConformance": w.SetPDFXConformance(PdfWriterPdfa1b),
		"SetOutputIntents":   w.SetOutputIntents("", "sRGB", "http://www.color.org", "sRGB", nil),
		"CreateXmpMetadata":  w.CreateXmpMetadata(),
		"SetPageXmpMetadata": w.SetPageXmpMetadata(nil),
		"SetTagged":          w.SetTagged(),
	}
	_, err := w.GetStructureTreeRoot()
	checks["GetStructureTreeRoot"] = err
	_, err = w.GetAcroForm()
	checks["GetAcroForm"] = err
	_, err = w.GetImportedPage(nil, 1)
	checks["GetImportedPage"] = err
	_, err = ImageGetInstanceFromSvg(nil)
	checks["ImageGetInstanceFromSvg"] = err
	_, err = CreateFont("some.afm", BaseFontCp1252, false)
	checks["CreateFont(afm)"] = err
	for name, err := range checks {
		var ue *UnsupportedFeatureError
		if !errors.As(err, &ue) {
			t.Errorf("%s returned %v, want an *UnsupportedFeatureError", name, err)
		}
	}
	if w.IsTagged() {
		t.Error("IsTagged is true")
	}
	if err := w.SetPDFXConformance(PdfWriterPdfxnone); err != nil {
		t.Errorf("SetPDFXConformance(PDFXNONE) = %v", err)
	}
}

type failingWriter struct{ n int }

func (f *failingWriter) Write(p []byte) (int, error) {
	f.n++
	if f.n > 1 {
		return 0, errors.New("disk full")
	}
	return len(p), nil
}

func TestWriteErrorIsReturnedByClose(t *testing.T) {
	doc := NewDocument(NewRectangle(0, 0, 100, 100), 0, 0, 0, 0)
	w, err := PdfWriterGetInstance(doc, &failingWriter{})
	if err != nil {
		t.Fatal(err)
	}
	doc.Open()
	w.GetDirectContent().NewPath()
	if err := doc.Close(); err == nil || err.Error() != "disk full" {
		t.Errorf("Close = %v, want the writer's error", err)
	}
}

func TestReaderRejectsOtherInput(t *testing.T) {
	if _, err := ReadPDF([]byte("hello")); err == nil {
		t.Error("ReadPDF accepted a file without a PDF header")
	}
	if _, err := ReadPDF([]byte("%PDF-1.4\nno xref here")); err == nil {
		t.Error("ReadPDF accepted a file without startxref")
	}
}
