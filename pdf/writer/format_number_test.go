package writer

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestFormatNumberMatchesOpenPDF compares formatNumber with the output of
// OpenPDF 3.0.5's ByteBuffer.formatDouble(double, ByteBuffer) for 20,032
// values (testdata/openpdf_format_double.txt: the raw bits of each double,
// a tab, and what OpenPDF appended). OpenPDF appends nothing for a value above
// 32767; formatNumber writes the rounded integer there instead, so those rows
// are checked against that rule.
func TestFormatNumberMatchesOpenPDF(t *testing.T) {
	f, err := os.Open("testdata/openpdf_format_double.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	rows := 0
	for scanner.Scan() {
		bitsText, want, _ := strings.Cut(scanner.Text(), "\t")
		bits, err := strconv.ParseInt(bitsText, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		d := math.Float64frombits(uint64(bits))
		if want == "" {
			if math.Abs(d) <= 32767 {
				t.Fatalf("OpenPDF wrote nothing for %v, which is within range", d)
			}
			want = strconv.FormatInt(int64(math.Abs(d)+0.5), 10)
			if d < 0 {
				want = "-" + want
			}
		}
		if got := formatNumber(d); got != want {
			t.Errorf("formatNumber(%v) = %q, OpenPDF writes %q", d, got, want)
		}
		rows++
	}
	if rows < 20000 {
		t.Fatalf("read %d rows", rows)
	}
}
