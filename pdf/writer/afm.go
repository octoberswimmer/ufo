package writer

import (
	"bufio"
	"embed"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// The Adobe Core 14 AFM files, distributed under the notice in
// afm/MustRead.html, which has to accompany them.
//
//go:embed afm/*.afm afm/MustRead.html
var afmFiles embed.FS

// afmMetrics holds what Type1Font reads from an AFM file.
type afmMetrics struct {
	fontName           string
	fullName           string
	familyName         string
	weight             string
	italicAngle        float32
	isFixedPitch       bool
	characterSet       string
	llx, lly, urx, ury int
	underlinePosition  int
	underlineThickness int
	encodingScheme     string
	capHeight          int
	xHeight            int
	ascender           int
	descender          int
	stdHW              int
	stdVW              int
	// fontSpecific is true unless the encoding scheme is
	// AdobeStandardEncoding; it is true for Symbol and ZapfDingbats.
	fontSpecific bool

	widthByName map[string]int
	bboxByName  map[string][4]int
	widthByCode map[int]int
	bboxByCode  map[int][4]int
	kernPairs   map[string]map[string]int
}

var (
	afmCacheMu sync.Mutex
	afmCache   = map[string]*afmMetrics{}
)

// loadAFM parses the embedded AFM file of a base-14 font at first use.
func loadAFM(fontName string) (*afmMetrics, error) {
	afmCacheMu.Lock()
	defer afmCacheMu.Unlock()
	if m, ok := afmCache[fontName]; ok {
		return m, nil
	}
	f, err := afmFiles.Open("afm/" + fontName + ".afm")
	if err != nil {
		return nil, newDocumentException("%s not found as resource", fontName+".afm")
	}
	defer f.Close()

	m := &afmMetrics{
		// Type1Font's defaults for the keys an AFM file may omit.
		ascender:           800,
		descender:          -200,
		capHeight:          700,
		stdVW:              80,
		underlinePosition:  -100,
		underlineThickness: 50,
		encodingScheme:     "FontSpecific",
		widthByName:        map[string]int{},
		bboxByName:         map[string][4]int{},
		widthByCode:        map[int]int{},
		bboxByCode:         map[int][4]int{},
		kernPairs:          map[string]map[string]int{},
	}
	atoi := func(s string) int {
		v, _ := strconv.ParseFloat(s, 64)
		return int(v)
	}
	section := ""
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		key, rest, _ := strings.Cut(line, " ")
		rest = strings.TrimSpace(rest)
		fields := strings.Fields(rest)
		switch section {
		case "":
			switch key {
			case "FontName":
				m.fontName = rest
			case "FullName":
				m.fullName = rest
			case "FamilyName":
				m.familyName = rest
			case "Weight":
				m.weight = rest
			case "ItalicAngle":
				v, _ := strconv.ParseFloat(rest, 32)
				m.italicAngle = float32(v)
			case "IsFixedPitch":
				m.isFixedPitch = rest == "true"
			case "CharacterSet":
				m.characterSet = rest
			case "FontBBox":
				if len(fields) >= 4 {
					m.llx, m.lly, m.urx, m.ury = atoi(fields[0]), atoi(fields[1]), atoi(fields[2]), atoi(fields[3])
				}
			case "UnderlinePosition":
				m.underlinePosition = atoi(rest)
			case "UnderlineThickness":
				m.underlineThickness = atoi(rest)
			case "EncodingScheme":
				m.encodingScheme = rest
			case "CapHeight":
				m.capHeight = atoi(rest)
			case "XHeight":
				m.xHeight = atoi(rest)
			case "Ascender":
				m.ascender = atoi(rest)
			case "Descender":
				m.descender = atoi(rest)
			case "StdHW":
				m.stdHW = atoi(rest)
			case "StdVW":
				m.stdVW = atoi(rest)
			case "StartCharMetrics":
				section = "chars"
			case "StartKernPairs":
				section = "kern"
			}
		case "chars":
			if key == "EndCharMetrics" {
				section = ""
				continue
			}
			code := -1
			width := 250
			name := ""
			var bbox [4]int
			for _, part := range strings.Split(line, ";") {
				pf := strings.Fields(part)
				if len(pf) < 2 {
					continue
				}
				switch pf[0] {
				case "C":
					code = atoi(pf[1])
				case "WX":
					width = atoi(pf[1])
				case "N":
					name = pf[1]
				case "B":
					if len(pf) >= 5 {
						bbox = [4]int{atoi(pf[1]), atoi(pf[2]), atoi(pf[3]), atoi(pf[4])}
					}
				}
			}
			if code >= 0 {
				m.widthByCode[code] = width
				m.bboxByCode[code] = bbox
			}
			if name != "" {
				m.widthByName[name] = width
				m.bboxByName[name] = bbox
			}
		case "kern":
			if key == "EndKernPairs" {
				section = ""
				continue
			}
			if key == "KPX" && len(fields) >= 3 {
				first := fields[0]
				if m.kernPairs[first] == nil {
					m.kernPairs[first] = map[string]int{}
				}
				m.kernPairs[first][fields[1]] = atoi(fields[2])
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading %s.afm: %w", fontName, err)
	}
	// Type1Font gives the non-breaking space the metrics of the space when
	// the AFM file has no glyph of that name.
	if _, ok := m.widthByName["nonbreakingspace"]; !ok {
		if w, ok := m.widthByName["space"]; ok {
			m.widthByName["nonbreakingspace"] = w
			m.bboxByName["nonbreakingspace"] = m.bboxByName["space"]
		}
	}
	m.fontSpecific = !(m.encodingScheme == "AdobeStandardEncoding" || m.encodingScheme == "StandardEncoding")
	afmCache[fontName] = m
	return m, nil
}
