// Package writer writes PDF files through an API shaped like the part of
// OpenPDF (com.lowagie.text / org.openpdf.text) that Flying Saucer's PDF
// module calls, so that module can be ported class by class. It also has a
// minimal PDF reader (ReadPDF) for tests, in place of com.codeborne:pdf-test.
//
// The package imports the standard library and golang.org/x/image/font/sfnt
// only. It does not import the ufo package or ufo/geom: geometry is passed as
// float32 numbers, as OpenPDF takes Java floats.
//
// # Conventions
//
// Coordinates are PDF user space: origin at the bottom left of the page, y
// up, in points. Nothing is flipped.
//
// Numbers are written with up to 4 decimals, trailing zeros removed.
//
// Names follow PORTING.md: a class keeps its name (PdfContentByte), an
// instance method is capitalized (saveState -> SaveState), a constant
// Foo.BAR_BAZ is FooBarBaz (BaseFont.CP1252 -> BaseFontCp1252,
// PdfWriter.VERSION_1_7 -> PdfWriterVersion17), a static method Foo.bar is
// FooBar (PdfWriter.getInstance -> PdfWriterGetInstance, Image.getInstance ->
// ImageGetInstance), with the exceptions CreateFont, CreateFontWithBytes and
// EnumerateTTCNames, which are BaseFont's static methods without the prefix.
// Where Java overloads differ by parameter types the Go names are listed
// below.
//
// Content stream operators that OpenPDF declares without a checked exception
// panic with a *DocumentException on misuse (RestoreState without SaveState,
// ShowText before SetFontAndSize, an unbalanced page at NewPage). Methods
// that do I/O or parse data return an error. Features outside the scope of
// this package return an *UnsupportedFeatureError; see "Not supported".
//
// # Document, PdfWriter
//
//	new Document(rect, l, r, t, b)          NewDocument(rect, l, r, t, b)
//	new Rectangle(llx, lly, urx, ury)       NewRectangle(llx, lly, urx, ury)
//	new Rectangle(urx, ury)                 NewRectangleWithUrxUry(urx, ury)
//	rect.getLeft/getBottom/getRight/getTop  rect.GetLeft/GetBottom/GetRight/GetTop
//	rect.getWidth/getHeight                 rect.GetWidth/GetHeight
//	rect.setBorder / setBorderWidth         rect.SetBorder / SetBorderWidth (recorded only)
//	PdfWriter.getInstance(doc, os)          PdfWriterGetInstance(doc, w) (*PdfWriter, error)
//	doc.open()                              doc.Open()
//	doc.setPageSize(rect)                   doc.SetPageSize(rect)
//	doc.newPage()                           doc.NewPage()
//	doc.close()                             doc.Close() error
//	doc.addTitle/addAuthor/addSubject/      doc.AddTitle/AddAuthor/AddSubject/
//	  addKeywords/addCreator/addProducer      AddKeywords/AddCreator/AddProducer
//	doc.addHeader(name, content)            doc.AddHeader(name, content)
//	writer.setPdfVersion(VERSION_1_7)       writer.SetPdfVersion(PdfWriterVersion17)
//	writer.getInfo()                        writer.GetInfo() *PdfDictionary
//	writer.getExtraCatalog()                writer.GetExtraCatalog() *PdfDictionary
//	writer.setCompressionLevel(n)           writer.SetCompressionLevel(n); PdfStreamNoCompression
//	                                          writes streams without a filter
//	writer.setFullCompression()             writer.SetFullCompression() (no effect: objects and the
//	                                          cross-reference table are always written plainly)
//	writer.setViewerPreferences(bits)       writer.SetViewerPreferences(PdfWriterDisplayDocTitle | ...)
//	writer.setPageEvent(event)              writer.SetPageEvent(PdfPageEvent)
//	writer.getDirectContent()               writer.GetDirectContent() *PdfContentByte
//	writer.getDirectContentUnder()          writer.GetDirectContentUnder()
//	writer.getPageNumber()                  writer.GetPageNumber() (1-based number of the current page)
//	writer.getCurrentPageNumber()           writer.GetCurrentPageNumber()
//	writer.getPageReference(n)              writer.GetPageReference(n) *PdfIndirectReference
//	writer.getRootOutline()                 writer.GetRootOutline() *PdfOutline
//	writer.addAnnotation(annot)             writer.AddAnnotation(annot)
//	writer.addToBody(obj)                   writer.AddToBody(obj) (*PdfIndirectObject, error)
//	  .getIndirectReference()                 .GetIndirectReference()
//	writer.getPdfIndirectReference()        writer.GetPdfIndirectReference()
//
// A page takes the size the document has when the page starts: call
// SetPageSize, then NewPage (or Open for the first page). As in OpenPDF,
// NewPage on a page nothing was written to adds no page; the pending page
// takes the new size. Close returns a *DocumentException "The document has
// no pages." when nothing was written at all.
//
// The file is written to the io.Writer progressively: the header at Open,
// each page and its content stream when the page ends, images when first
// drawn, and templates, fonts, outlines, the catalog, the information
// dictionary, the cross-reference table and the trailer at Close. The first
// error of the io.Writer is returned by Close.
//
// The information dictionary starts with /Producer, /CreationDate and
// /ModDate; any key can be set or replaced through GetInfo().Put.
//
// # Objects
//
//	new PdfName("X"), PdfName.TITLE         NewPdfName("X") or PdfName("X"), PdfNameTitle
//	new PdfString(s)                        NewPdfString(s)
//	new PdfString(s, TEXT_UNICODE)          NewPdfStringWithEncoding(s, PdfObjectTextUnicode)
//	new PdfNumber(n)                        NewPdfNumber(n) or PdfNumber(n)
//	new PdfArray(); add; isEmpty            NewPdfArray(items...); Add; IsEmpty; Size
//	new PdfDictionary(); put; get           NewPdfDictionary(); Put; Get; Contains; Remove
//	new PdfRectangle(llx, lly, urx, ury)    NewPdfRectangle(llx, lly, urx, ury) *PdfArray
//	PdfBoolean, PdfNull, PdfLiteral         PdfBoolean(b), PdfNull{}, PdfLiteral(s)
//	PdfStream                               NewPdfStream(data); FlateCompress(level)
//
// A text string is written in PDFDocEncoding when every character has a code
// there and as UTF-16BE with a byte order mark otherwise, for both encodings.
//
// # PdfContentByte
//
//	saveState / restoreState                SaveState / RestoreState          q / Q
//	concatCTM(a,b,c,d,e,f)                  ConcatCTM(a,b,c,d,e,f)            cm
//	moveTo / lineTo                         MoveTo / LineTo                   m / l
//	curveTo(x1,y1,x2,y2,x3,y3)              CurveTo(x1,y1,x2,y2,x3,y3)        c
//	curveTo(x2,y2,x3,y3)                    CurveToWithX2Y2X3Y3(x2,y2,x3,y3)  v
//	curveFromTo(x1,y1,x3,y3)                CurveFromTo(x1,y1,x3,y3)          y
//	rectangle(x,y,w,h)                      Rectangle(x,y,w,h)                re
//	closePath / newPath                     ClosePath / NewPath               h / n
//	fill / eoFill / stroke                  Fill / EoFill / Stroke            f / f* / S
//	fillStroke / eoFillStroke               FillStroke / EoFillStroke         B / B*
//	closePathStroke / closePathFillStroke   ClosePathStroke / ClosePathFillStroke  s / b
//	clip / eoClip                           Clip / EoClip                     W / W*
//	setLineWidth / setLineCap / setLineJoin SetLineWidth / SetLineCap / SetLineJoin  w / J / j
//	setMiterLimit / setFlatness             SetMiterLimit / SetFlatness       M / i
//	setLineDash(array, phase)               SetLineDash(array, phase)         d
//	setLineDash(phase)                      SetLineDashPhase(phase)
//	setLineDash(on, phase)                  SetLineDashOn(on, phase)
//	setLineDash(on, off, phase)             SetLineDashOnOff(on, off, phase)
//	setLiteral(String)                      SetLiteral(s)
//	setLiteral(char)                        SetLiteralChar(c)
//	setLiteral(float)                       SetLiteralFloat(n)
//	setColorFill(Color)                     SetColorFill(Color)               rg / k / g
//	setColorStroke(Color)                   SetColorStroke(Color)             RG / K / G
//	  new java.awt.Color(r,g,b[,a])           NewRGBColor(r,g,b), NewRGBColorWithAlpha(r,g,b,a)
//	  new CMYKColor(c,m,y,k)                  NewCMYKColor(c,m,y,k)
//	  new GrayColor(g)                        NewGrayColor(g)
//	setRGBColorFill(r,g,b) / ...Stroke      SetRGBColorFill / SetRGBColorStroke
//	setRGBColorFillF / ...StrokeF           SetRGBColorFillF / SetRGBColorStrokeF
//	setCMYKColorFillF / ...StrokeF          SetCMYKColorFillF / SetCMYKColorStrokeF
//	setGrayFill / setGrayStroke             SetGrayFill / SetGrayStroke
//	setGState(gs)                           SetGState(gs)                     gs
//	  new PdfGState()                         NewPdfGState()
//	  gs.setFillOpacity / setStrokeOpacity    gs.SetFillOpacity / SetStrokeOpacity   /ca /CA
//	  gs.setBlendMode(PdfGState.BM_NORMAL)    gs.SetBlendMode(PdfGStateBmNormal)     /BM
//	beginText / endText                     BeginText / EndText               BT / ET
//	setFontAndSize(bf, size)                SetFontAndSize(bf, size)          Tf
//	setTextMatrix(a,b,c,d,x,y)              SetTextMatrix(a,b,c,d,x,y)        Tm
//	setTextMatrix(x,y)                      SetTextMatrixXY(x,y)
//	moveText / setLeading / newlineText     MoveText / SetLeading / NewlineText  Td / TL / T*
//	showText(String)                        ShowText(s)                       Tj
//	showText(PdfTextArray)                  ShowTextArray(array)              TJ
//	  new PdfTextArray(); add(String);        NewPdfTextArray(); AddString(s);
//	  add(float)                              AddNumber(n)
//	showTextKerned(String)                  ShowTextKerned(s)                 TJ
//	setTextRenderingMode(mode)              SetTextRenderingMode(PdfContentByteTextRenderModeFillStroke)  Tr
//	setCharacterSpacing / setWordSpacing    SetCharacterSpacing / SetWordSpacing  Tc / Tw
//	setHorizontalScaling / setTextRise      SetHorizontalScaling / SetTextRise    Tz / Ts
//	addImage(image, a,b,c,d,e,f)            AddImage(image, a,b,c,d,e,f) error    q cm Do Q
//	addImage(image)                         AddImageAbsolute(image) error
//	createTemplate(w, h)                    CreateTemplate(w, h) *PdfTemplate
//	addTemplate(t, a,b,c,d,e,f)             AddTemplate(t, a,b,c,d,e,f)       q cm Do Q
//	addTemplate(t, x, y)                    AddTemplateAt(t, x, y)
//	beginMarkedContentSequence(PdfName)     BeginMarkedContentSequence(tag)   BMC
//	beginMarkedContentSequence(tag, dict,   BeginMarkedContentSequenceWithProperties(tag, dict)  BDC
//	  true)
//	endMarkedContentSequence()              EndMarkedContentSequence()        EMC
//	size() / reset()                        Size() / Reset()
//
// A color with an alpha below 255 sets an ExtGState with /ca (fill) or /CA
// (stroke) before the color operator, as current OpenPDF does. The opacity
// in effect is tracked with SaveState and RestoreState.
//
// PdfTemplate has every PdfContentByte method plus GetWidth, GetHeight,
// SetWidth, SetHeight, GetBoundingBox, SetBoundingBox, SetMatrix and
// GetIndirectReference. A template is written at Close, so its content may
// be added after it was drawn.
//
// # BaseFont
//
//	BaseFont.createFont(name, enc, emb)     CreateFont(name, enc, emb) (*BaseFont, error)
//	BaseFont.createFont(name, enc, emb,     CreateFontWithBytes(name, enc, emb, cached, ttfAfm, pfb)
//	  cached, ttfAfm, pfb)
//	BaseFont.enumerateTTCNames(path)        EnumerateTTCNames(path), EnumerateTTCNamesBytes(data)
//	BaseFont.HELVETICA ...                  BaseFontHelvetica, BaseFontTimesRoman, BaseFontCourierBoldoblique, ...
//	BaseFont.CP1252 / IDENTITY_H            BaseFontCp1252 / BaseFontIdentityH
//	BaseFont.EMBEDDED / NOT_EMBEDDED        BaseFontEmbedded / BaseFontNotEmbedded
//	BaseFont.ASCENT ... AWT_MAXADVANCE,     BaseFontAscent, BaseFontCapheight, BaseFontDescent,
//	  UNDERLINE_*, STRIKETHROUGH_*            BaseFontItalicangle, BaseFontBboxllx ... BaseFontBboxury,
//	                                          BaseFontAwtAscent, BaseFontAwtDescent, BaseFontAwtLeading,
//	                                          BaseFontAwtMaxadvance, BaseFontUnderlinePosition,
//	                                          BaseFontUnderlineThickness, BaseFontStrikethroughPosition,
//	                                          BaseFontStrikethroughThickness
//	bf.getWidth(int)                        bf.GetWidth(rune) int
//	bf.getWidth(String)                     bf.GetWidthString(s) int
//	bf.getWidthPoint(String, size)          bf.GetWidthPoint(s, size) float32
//	bf.getWidthPoint(int, size)             bf.GetWidthPointChar(c, size)
//	bf.getWidthPointKerned(String, size)    bf.GetWidthPointKerned(s, size)
//	bf.getKerning(c1, c2)                   bf.GetKerning(c1, c2) int
//	bf.hasKernPairs()                       bf.HasKernPairs()
//	bf.charExists(c)                        bf.CharExists(c)
//	bf.getFontDescriptor(key, size)         bf.GetFontDescriptor(key, size) float32
//	bf.getPostscriptFontName()              bf.GetPostscriptFontName()
//	bf.getFamilyFontName()                  bf.GetFamilyFontName() [][]string  ({platform, encoding, language, name})
//	bf.getFullFontName()                    bf.GetFullFontName() [][]string
//	bf.getAllNameEntries()                  bf.GetAllNameEntries() [][]string
//	bf.getEncoding / isEmbedded /           bf.GetEncoding / IsEmbedded / GetFontType / IsFontSpecific
//	  getFontType / isFontSpecific
//
// The 14 standard fonts use WinAnsi (Cp1252) encoding and the metrics of the
// Adobe Core 14 AFM files embedded under afm/ (with Adobe's notice,
// afm/MustRead.html), parsed at first use. Widths and kerning are looked up
// as Type1Font does: a character maps to the glyph name OpenPDF's glyph list
// gives it; the non-breaking space has the metrics of the space. Symbol and
// ZapfDingbats keep their built-in encoding: the Cp1252 code of a character
// selects the glyph. Characters outside Cp1252 are left out of shown text
// and have width 0. Further AFM values: GetXHeight, GetStdVW, GetWeight,
// GetCharBBox, IsFixedPitch.
//
// A TrueType font (.ttf, glyf-based .otf, or "file.ttc,N") is always written
// as a Type 0 font with Identity-H encoding over a CIDFontType2 descendant,
// with the whole font file as FontFile2, a /W array and a ToUnicode CMap for
// the glyphs shown; text is written as two-byte glyph ids. The encoding and
// embedded arguments of CreateFont do not change that. Characters without a
// glyph are left out. Metrics follow TrueTypeFont (see GetFontDescriptor).
// What TrueTypeUtil and ITextFontResolver read through reflection or a
// RandomAccessFileOrArray is available as GetOS2WeightClass,
// GetOS2FsSelection, GetHeadMacStyle, IsItalic, IsBold, IsFixedPitch,
// GetOS2StrikeoutSize, GetOS2StrikeoutPosition, GetPostUnderlinePosition,
// GetPostUnderlineThickness (raw font units), GetUnitsPerEm, and as the raw
// tables: GetTables (offset and length by tag) over GetFontData.
//
// # Image
//
//	Image.getInstance(byte[])               ImageGetInstance(data) (*Image, error)
//	Image.getInstance(Image)                ImageGetInstanceFromImage(img)
//	(java.awt.Image)                        ImageGetInstanceFromGoImage(image.Image)
//	getWidth / getHeight                    GetWidth / GetHeight (pixels)
//	getPlainWidth / getPlainHeight          GetPlainWidth / GetPlainHeight
//	getScaledWidth / getScaledHeight        GetScaledWidth / GetScaledHeight
//	scaleAbsolute(w, h)                     ScaleAbsolute(w, h)
//	scaleAbsoluteWidth / ...Height          ScaleAbsoluteWidth / ScaleAbsoluteHeight
//	scalePercent(p) / scalePercent(px, py)  ScalePercent(p) / ScalePercentXY(px, py)
//	scaleToFit(w, h)                        ScaleToFit(w, h)
//	setAbsolutePosition(x, y)               SetAbsolutePosition(x, y)
//	getDpiX / getDpiY / setDpi              GetDpiX / GetDpiY / SetDpi
//
// A JPEG is embedded unchanged with DCTDecode. PNG and GIF are decoded and
// written as 8-bit DeviceRGB or DeviceGray with an SMask when any pixel is
// not opaque. Copies made with ImageGetInstanceFromImage share one XObject.
//
// # Outlines, destinations, actions, annotations
//
//	new PdfDestination(FIT)                 NewPdfDestination(PdfDestinationFit)
//	new PdfDestination(FITH, top)           NewPdfDestinationWithParameter(PdfDestinationFith, top)
//	new PdfDestination(XYZ, l, t, zoom)     NewPdfDestinationWithLeftTopZoom(PdfDestinationXyz, l, t, zoom)
//	new PdfDestination(FITR, l, b, r, t)    NewPdfDestinationWithLeftBottomRightTop(PdfDestinationFitr, l, b, r, t)
//	dest.addPage(ref)                       dest.AddPage(ref)
//	new PdfOutline(parent, dest, title)     NewPdfOutline(parent, dest, title)
//	new PdfOutline(parent, dest, title,     NewPdfOutlineWithOpen(parent, dest, title, open)
//	  open)
//	new PdfOutline(parent, action, title)   NewPdfOutlineWithAction(parent, action, title)
//	new PdfAction()                         NewPdfAction(); Put(PdfNameS, PdfNameGoto); Put(PdfNameD, dest)
//	new PdfAction(url)                      NewPdfActionWithUrl(url)
//	PdfAction.gotoLocalPage(p, dest, w)     PdfActionGotoLocalPage(p, dest, w)
//	PdfAction.javaScript(code, w)           PdfActionJavaScript(code, w)
//	new PdfAnnotation(w, llx, lly, urx,     NewPdfAnnotation(w, llx, lly, urx, ury, action)
//	  ury, action)
//	annot.put / setBorderStyle / setBorder  annot.Put / SetBorderStyle / SetBorder / SetFlags
//	new PdfBorderDictionary(width, style)   NewPdfBorderDictionary(width, style)
//	new PdfBorderArray(h, v, width)         NewPdfBorderArray(h, v, width)
//	PdfAnnotation.FLAGS_PRINT               PdfAnnotationFlagsPrint
//
// Named destinations are written as ITextOutputDevice writes them: objects
// added with AddToBody and a /Names entry put into GetExtraCatalog.
//
// # Not supported
//
// These exist so ported code can call them; each returns an
// *UnsupportedFeatureError: SetEncryption, SetPDFXConformance (other than
// PdfWriterPdfxnone), SetOutputIntents, CreateXmpMetadata, SetXmpMetadata,
// SetPageXmpMetadata, SetTagged (IsTagged is always false),
// GetStructureTreeRoot, NewPdfStructureElement,
// BeginMarkedContentSequenceStruct, GetAcroForm, AddFormField,
// CreateAppearance, NewPdfReader, GetImportedPage, ImageGetInstanceFromSvg.
// CreateFont returns one for a Type 1 font from an AFM or PFM file and for
// an OpenType font with CFF outlines.
//
// # Reader
//
// ReadPDF reads a file with a classic cross-reference table (what this
// package writes) and no encryption. Page numbers are 0-based. See
// ReadDocument.PageText for how text is extracted and where spaces and
// newlines are inserted.
// The OpenPDF source the port follows is the 3.0.5 release (the version
// Flying Saucer 10 depends on): https://github.com/LibrePDF/OpenPDF, tag
// 3.0.5.
package writer
