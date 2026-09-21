# ufo

ufo is a Go port of [Flying Saucer](https://github.com/flyingsaucerproject/flyingsaucer),
the CSS 2.1 renderer that lays out XHTML and writes it as PDF. It is the
renderer Salesforce uses for Visualforce pages with `renderAs="pdf"`, which is
why [aer](https://github.com/octoberswimmer/aer) uses ufo for the same pages,
running it as a separate program (`ufo pipe`).

The port follows the Java source class by class (see `PORTING.md`), so that
Flying Saucer's own test suite, ported alongside the code, validates it.

## Status

Every class of `flying-saucer-core` and `flying-saucer-pdf` is ported except
the Swing, Java2D and SWT renderers, Swing form controls, AcroForm fields, SVG,
encryption, PDF/A, tagged PDF and iText's built-in CJK fonts; those entry
points exist and return an error naming the missing feature. Flying Saucer's
JUnit suite is ported alongside the code. `CHECKLIST.md` lists every Java file
and test.

The port is checked against Flying Saucer itself:

- `layout_java_comparison_test.go` lays out 200 generated documents and
  compares every box's position and size with Flying Saucer's.
- Parts of the CSS lexer, parser, line breaking, table layout, float
  placement and border painting were checked against the Java classes on
  generated inputs; the tests keep the resulting expected values.
- `pdf/writer` formats numbers as OpenPDF 3.0.5 does, so content streams
  match Flying Saucer's.

## Comparing with Flying Saucer

`reference/build.sh` builds Flying Saucer (JDK 21 or later and Maven) and
`reference/render.sh in.xhtml out.pdf out.boxes` renders a document with it,
writing the box tree too. The Java build reads only XML;
`go run ./cmd/html2xhtml page.html > page.xhtml` converts an HTML document with
the parser ufo uses, so both implementations read the same tree.

## Command line

`ufo input.html output.pdf` renders a document; input that is not well-formed
XML is parsed as HTML.

`ufo pipe` renders one document for another program over standard input and
output: the program sends the HTML and answers ufo's requests for the
stylesheets, images and fonts it refers to, and ufo sends back the PDF. This
lets a program use ufo without linking it. The protocol is documented in
package `github.com/octoberswimmer/ufo/pipe`.

## Building and releasing

- `make` builds `./ufo`; `make install` installs it. `ufo version` prints the
  version it was built as (the latest tag).
- `make test` checks formatting, runs `go vet`, and runs the tests with and
  without the race detector.
- `make tag` proposes the next minor version tag with a changelog and creates
  it on confirmation.
- `make release`, with HEAD at that tag, runs the tests, builds zips for
  Linux, macOS and Windows on amd64 and arm64, signs and notarizes the macOS
  binaries, writes `SHA256SUMS-<version>`, pushes the tag to the
  `octoberswimmer` remote and creates the GitHub release with `gh`. Each zip
  holds the binary, `LICENSE` and this README; the LGPL requires the licence
  text to accompany the binary, and the release's source is the tag it was
  built from.

## Layout

- `github.com/octoberswimmer/ufo/dom` — the document tree the renderer reads
  (the role `org.w3c.dom` plays in Java), with parsers for XHTML and HTML.
- `github.com/octoberswimmer/ufo/geom` — the geometry types the Java source
  takes from `java.awt` (`Rectangle`, `Point`, `Dimension`, `Shape`, `Area`,
  `AffineTransform`).
- `github.com/octoberswimmer/ufo` — `flying-saucer-core`: CSS parser, cascade,
  box tree, layout, tables, painting.
- `github.com/octoberswimmer/ufo/pdf` — `flying-saucer-pdf`: the PDF output
  device and the `ITextRenderer` entry point.
- `github.com/octoberswimmer/ufo/pipe` — the `ufo pipe` protocol.
- `github.com/octoberswimmer/ufo/pdf/writer` — the PDF file writer the output
  device draws through, with the content-stream level API the Java source
  uses from OpenPDF (`PdfContentByte`, `BaseFont`, `Image`, outlines,
  annotations), and a reader the tests use to check what was written.

## License

Flying Saucer is licensed under the GNU Lesser General Public License, version
2.1 or later. ufo is a derivative work and carries the same license; see
`LICENSE`.
