# Porting conventions

ufo is a class-by-class port of Flying Saucer
(https://github.com/flyingsaucerproject/flyingsaucer), as of commit
14f4747b65ca4abba0bd75edb06617af7da46036 (2026-09-21). Every Go file names the
Java file it ports, relative to that repository, in its first comment. These rules are mechanical on purpose: two people porting
two classes that refer to each other must arrive at the same Go names without
seeing each other's code.

## Packages

Flying Saucer's core packages import each other in cycles (`css.style` ↔
`render` ↔ `layout` ↔ `context` ↔ `css.*`), which Go does not allow, so all of
`flying-saucer-core` is ONE Go package, `ufo`, at the module root. One Go file
per Java file, named after the class in snake case, prefixed with the Java
package's last segment: `render/BlockBox.java` → `render_block_box.go`,
`css/parser/CSSParser.java` → `parser_css_parser.go`,
`css/parser/property/PrimitivePropertyBuilders.java` →
`property_primitive_property_builders.go`, `css/style/derived/LengthValue.java`
→ `derived_length_value.go`, `layout/breaker/Breaker.java` →
`breaker_breaker.go`.

- `ufo/dom` replaces `org.w3c.dom`; `ufo/geom` replaces `java.awt` geometry.
  Neither imports `ufo`.
- `ufo/pdf` ports `flying-saucer-pdf` and imports `ufo`.
- Not ported: `swing/`, `simple/extend/form/`, `debug/`, `test/`, `event/`,
  the Java2D and SWT renderers, OSGi, FOP, log4j. Where core code refers to a
  Swing class only for a default (e.g. `Java2DTextRenderer`), the reference is
  dropped and the dependency is supplied by the caller.

## Names

- Class or interface `Foo` → Go type `Foo`. Classes are structs used through
  pointers (`*Foo`); Java interfaces are Go interfaces.
- Constructor → `NewFoo(...)`; overloaded constructors follow "Overloads".
- Instance method `fooBar` → method `FooBar` (exported, same words). Getters
  keep their `Get`/`Is`/`Has` prefix: `getStyle()` → `GetStyle()`. Record
  accessors `foo()` → `Foo()`.
- Static method `Foo.barBaz` → function `FooBarBaz`.
- Static final constant or enum constant `Foo.BAR_BAZ` → `FooBarBaz`
  (`CSSName.FONT_SIZE` → `CSSNameFontSize`, `IdentValue.AUTO` →
  `IdentValueAuto`). Java enums → a Go type with the constants as package
  variables (pointers when the enum has fields or methods, otherwise ints).
- Fields → unexported camelCase with the Java name (`_foo` → `foo`). Public
  Java fields that other classes read → exported.
- Nested class `Outer.Inner` → `OuterInner`.
- Private methods → unexported, same words. If an unexported name collides with
  another class's (one package!), prefix it with the class name in camelCase:
  `boxBuilderCreateChildren`.
- Overloads: the overload with the fewest parameters keeps the plain name.
  Each other overload appends `With` + the capitalized names of the parameters
  it adds, or, when overloads differ only by type, the type:
  `layout(c)` / `layout(c, contentStart)` → `Layout(c)` /
  `LayoutWithContentStart(c, contentStart)`;
  `asFloat(String)` / `asFloat(PropertyValue)` → `AsFloatString`,
  `AsFloatPropertyValue`.

## Inheritance and virtual dispatch

Go embedding does not dispatch a base-class call to a subclass override, and
Flying Saucer relies on that (`Box.layout` calling `this.calcDimensions()`
overridden in `TableBox`). The rule for every class that is subclassed:

- The struct embeds its parent's struct by value:
  `type BlockBox struct { Box; ... }`.
- A subclassed class `Foo` gets an interface `FooI` listing every non-private
  method of `Foo` (including inherited ones via embedding the parent's `...I`),
  plus `AsFoo() *Foo`. `BoxI`, `BlockBoxI`. Classes that are never subclassed
  (`LineBox`, `InlineLayoutBox`, `TableBox`, `TableCellBox`, ...) get no
  interface of their own and are referred to as `*LineBox`.
- The root struct holds `self BoxI` (named `self`), set by every constructor of
  every subclass to the outermost object (`NewTableBox` sets
  `t.self = t`). Code inside `Box`/`BlockBox` methods that calls an
  overridable method calls it through `b.self`. A subclass method calling
  `super.foo()` calls `t.BlockBox.Foo()`.
- A Java variable, field, parameter or return typed `Box` is `BoxI` in Go;
  typed `BlockBox` is `BlockBoxI`; typed as a leaf class is the pointer.
  `x instanceof TableBox` → `_, ok := x.(*TableBox)`; `x instanceof BlockBox`
  → `_, ok := x.(BlockBoxI)`; casts likewise.
- Never store a nil `*T` in an interface value: return a literal `nil`
  interface. Compare interfaces to `nil` only under that rule.
- Small abstract bases without self-calls (`DerivedValue`,
  `AbstractPropertyBuilder`) are a struct to embed plus the Java interface;
  no `self` needed. Where a base method does call an abstract method, the base
  takes the needed function or interface as a field set by the constructor.

## Exceptions

- An exception that Java code catches for control flow (`CSSParseException`
  inside `CSSParser`, which recovers and skips the bad declaration) is a Go
  `panic` with the typed value (`*CSSParseException`), recovered at the place
  the Java `catch` is. Code that only propagates does nothing special.
- `XRRuntimeException`, `IllegalArgumentException`, `NullPointerException`
  and the like → `panic(NewXRRuntimeException(msg))`. Exported entry points
  (`pdf.ITextRenderer.CreatePDF`, `StylesheetFactory.ParseStylesheet`) recover
  and return an `error`.
- `IOException` and other checked exceptions on I/O paths → a returned `error`.

## JDK types

- `String` → `string`; nullable strings stay `string` with `""` for null
  unless the code distinguishes null from empty, then `*string`.
- `List<T>` → `[]T`; `Map<K,V>` → `map[K]V` (`LinkedHashMap`/`TreeMap` whose
  order is observed → keep a separate ordered key slice or sort on read; say
  which in a comment); `Set<T>` → `map[T]struct{}`; `Optional<T>` → pointer or
  `(T, bool)`.
- `float` → `float32`, `double` → `float64`, `int` → `int`, `long` → `int64`,
  `short` → `int16`. Keep Java's arithmetic: integer division, `(int)` casts
  truncating toward zero, `Math.round` → `int(math.Floor(x + 0.5))`.
- `org.w3c.dom.{Node,Element,Document,Text,NodeList}` → `dom.Node`,
  `*dom.Element`, `*dom.Document`, `*dom.Text`, `[]dom.Node`.
- `java.awt.{Rectangle,Point,Dimension,Shape,Area,geom.AffineTransform}` →
  `geom.Rectangle` etc. (pointers where Java mutates or returns null).
  `java.awt.Color` does not appear in core outside Swing; colors are `FSColor`.
- `java.text.BreakIterator` → `BreakIterator` in `util_break_iterator.go`
  (line-break opportunities per UAX #14 as the JDK implements them for the
  root locale).
- Logging: `XRLog.cssParse(Level.WARNING, msg)` → `XRLogCssParse(LevelWarning,
  msg)`; the default sink discards below warning and writes to the standard
  logger otherwise.
- `Configuration.valueFor("xr.foo", default)` → `ConfigurationValueFor(...)`,
  reading the same keys from the same default properties file, embedded.

## Tests

Flying Saucer's JUnit tests are ported as Go tests in the matching file
(`FooTest.java` → `<prefix>_foo_test.go`), one Go test (or subtest) per Java
test method with the method's name in the Go name. Test resources are copied
under `testdata/` with their Java paths. PDF assertions that the Java suite
makes through `com.codeborne:pdf-test` (text content, page count) are made by
reading the produced PDF back. Documents that load `classpath:` URLs find
their resources because the test's `TestMain` adds `testdata` to the class
path with `GeneralUtilAddClasspathEntry`, as Maven puts
`src/test/resources` on the test class path.

## Style

Tabs. `gofmt`. Comments only where they tell a future reader something the
code does not; carry over Java comments that do. No TODO stubs that silently
do nothing: a piece that is not ported panics with
`NewXRRuntimeException("not ported: <Java class>.<method>")` and is listed in
`CHECKLIST.md`.

## Decisions settled during the port

- `org.w3c.dom.css.CSSPrimitiveValue` as a Java type is `*PropertyValue`
  (its only implementation); its constants are `CSSPrimitiveValueCssPx` etc.
- `java.util.BitSet` → the type `BitSet` in
  `property_abstract_property_builder.go` (`Get`, `Set`, `Or`), keyed by
  `IdentValue.FS_ID`.
- A Java constant whose derived Go name collides with a nested class's type
  name takes a suffix saying what it is: the builder constants are
  `PrimitivePropertyBuildersColorBuilder`, `...BorderStyleBuilder`,
  `...BorderWidthBuilder`, `...BorderRadiusBuilder`, `...MarginBuilder`,
  `...PaddingBuilder`.
- `AbstractPropertyBuilder` does hold a `self PropertyBuilder`: its
  four-argument `BuildDeclarations` calls the abstract five-argument
  `BuildDeclarationsWithInheritAllowed`.
- Before guessing a name, grep the Go tree: if the class is already ported,
  its actual exported names win over what these rules would derive.
- `CalculatedStyle` is subclassed by `EmptyStyle`, so Java's `CalculatedStyle`
  type is `CalculatedStyleI` (`AsCalculatedStyle()` gives the struct); a root
  style's parent is a nil interface. `RectPropertySet` is subclassed by
  `BorderPropertySet`: Java's `RectPropertySet` type is `RectPropertySetI`.
- Selector matching works on `dom.Node` (as the current Java does), not `any`.
  A namespace URI that Java may pass as null is `*string` (nil = any
  namespace); `AttributeResolver.GetAttributeValue` returns `*string`.
  Class, id, lang, page name, pseudo-element: `string`, `""` for null.
- `StylesheetInfo.Origin` → `StylesheetInfoOrigin` with
  `StylesheetInfoOriginUserAgent/User/Author`.
- Every class with a Java `toString()` has both `String()` and `ToString()`.
- Java float formatting in messages: `propertyBuilderFloatToString`
  (`Float.toString`), `propertyBuilderRound` (`Math.round`);
  `calculatedStyleFloatToInt` is Java's `(int)` cast of a float.
- Text positions are BYTE offsets into Go strings, everywhere (`BreakIterator`,
  `InlineText` start/end, `LineBreakContext`, `Breaker`, `WhitespaceStripper`):
  `s.substring(a, b)` → `s[a:b]`, `s.length()` → `len(s)`, `indexOf` →
  `strings.Index`. Code that walks characters (`charAt(i)` in a loop) decodes
  runes (`utf8.DecodeRuneInString(s[i:])`) and advances by the rune's width, so
  an offset never lands inside a character. A Java `char` variable is a `rune`.
  Counts of characters that reach output (letter-spacing, justification
  `CharCounts`) count runes, which equals Java's count for BMP text.
- `java.util.logging.Level` → `*Level` (`LevelWarning`...). `XRLogGeneral(args
  ...any)` and its siblings accept `(msg)`, `(level, msg)`, `(level, msg, err)`.
- `org.xml.sax.InputSource` → `InputSource` in `util_input_sources.go`.
- JDK and OpenPDF classes replaced by ufo packages: `geom` (see its doc
  comments for names such as `NewRectangle`, `Rectangle.Add`, `AddXY`,
  `AffineTransformGetTranslateInstance`, `Area.Shapes`), `pdf/writer` (see its
  `doc.go` mapping table).
