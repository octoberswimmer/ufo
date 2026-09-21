// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/css/parser/Token.java

package ufo

// TokenType ports the enum Token.Type.
type TokenType int

const (
	TokenTypeS TokenType = iota
	TokenTypeCdo
	TokenTypeCdc
	TokenTypeIncludes
	TokenTypeDashmatch
	TokenTypePrefixmatch
	TokenTypeSuffixmatch
	TokenTypeSubstringmatch
	TokenTypeLbrace
	TokenTypePlus
	TokenTypeGreater
	TokenTypeComma
	TokenTypeString
	TokenTypeInvalid
	TokenTypeIdent
	TokenTypeHash
	TokenTypeImportSym
	TokenTypePageSym
	TokenTypeMediaSym
	TokenTypeCharsetSym
	TokenTypeNamespaceSym
	TokenTypeFontFaceSym
	TokenTypeAtRule
	TokenTypeImportantSym
	TokenTypeEms
	TokenTypeExs
	TokenTypePx
	TokenTypeCm
	TokenTypeMm
	TokenTypeIn
	TokenTypePt
	TokenTypePc
	TokenTypeAngle
	TokenTypeTime
	TokenTypeFreq
	TokenTypeDimension
	TokenTypePercentage
	TokenTypeNumber
	TokenTypeUri
	TokenTypeFunction
	TokenTypeOther
	TokenTypeRbrace
	TokenTypeSemicolon
	TokenTypeVirgule
	TokenTypeColon
	TokenTypeMinus
	TokenTypeRparen
	TokenTypeLbracket
	TokenTypeRbracket
	TokenTypePeriod
	TokenTypeEquals
	TokenTypeAsterisk
	TokenTypeVerticalBar
	TokenTypeEof
)

var tokenTypeNames = [...]string{
	TokenTypeS:              "S",
	TokenTypeCdo:            "CDO",
	TokenTypeCdc:            "CDC",
	TokenTypeIncludes:       "INCLUDES",
	TokenTypeDashmatch:      "DASHMATCH",
	TokenTypePrefixmatch:    "PREFIXMATCH",
	TokenTypeSuffixmatch:    "SUFFIXMATCH",
	TokenTypeSubstringmatch: "SUBSTRINGMATCH",
	TokenTypeLbrace:         "LBRACE",
	TokenTypePlus:           "PLUS",
	TokenTypeGreater:        "GREATER",
	TokenTypeComma:          "COMMA",
	TokenTypeString:         "STRING",
	TokenTypeInvalid:        "INVALID",
	TokenTypeIdent:          "IDENT",
	TokenTypeHash:           "HASH",
	TokenTypeImportSym:      "IMPORT_SYM",
	TokenTypePageSym:        "PAGE_SYM",
	TokenTypeMediaSym:       "MEDIA_SYM",
	TokenTypeCharsetSym:     "CHARSET_SYM",
	TokenTypeNamespaceSym:   "NAMESPACE_SYM",
	TokenTypeFontFaceSym:    "FONT_FACE_SYM",
	TokenTypeAtRule:         "AT_RULE",
	TokenTypeImportantSym:   "IMPORTANT_SYM",
	TokenTypeEms:            "EMS",
	TokenTypeExs:            "EXS",
	TokenTypePx:             "PX",
	TokenTypeCm:             "CM",
	TokenTypeMm:             "MM",
	TokenTypeIn:             "IN",
	TokenTypePt:             "PT",
	TokenTypePc:             "PC",
	TokenTypeAngle:          "ANGLE",
	TokenTypeTime:           "TIME",
	TokenTypeFreq:           "FREQ",
	TokenTypeDimension:      "DIMENSION",
	TokenTypePercentage:     "PERCENTAGE",
	TokenTypeNumber:         "NUMBER",
	TokenTypeUri:            "URI",
	TokenTypeFunction:       "FUNCTION",
	TokenTypeOther:          "OTHER",
	TokenTypeRbrace:         "RBRACE",
	TokenTypeSemicolon:      "SEMICOLON",
	TokenTypeVirgule:        "VIRGULE",
	TokenTypeColon:          "COLON",
	TokenTypeMinus:          "MINUS",
	TokenTypeRparen:         "RPAREN",
	TokenTypeLbracket:       "LBRACKET",
	TokenTypeRbracket:       "RBRACKET",
	TokenTypePeriod:         "PERIOD",
	TokenTypeEquals:         "EQUALS",
	TokenTypeAsterisk:       "ASTERISK",
	TokenTypeVerticalBar:    "VERTICAL_BAR",
	TokenTypeEof:            "EOF",
}

// Name returns the name of the enum constant, e.g. "IMPORT_SYM".
func (t TokenType) Name() string {
	return tokenTypeNames[t]
}

func (t TokenType) ToString() string {
	return tokenTypeNames[t]
}

func (t TokenType) String() string {
	return tokenTypeNames[t]
}

var (
	TokenTkS              = newToken(TokenTypeS, "S", "whitespace")
	TokenTkCdo            = newToken(TokenTypeCdo, "CDO", "<!--")
	TokenTkCdc            = newToken(TokenTypeCdc, "CDC", "-->")
	TokenTkIncludes       = newToken(TokenTypeIncludes, "INCLUDES", "an attribute word match")
	TokenTkDashmatch      = newToken(TokenTypeDashmatch, "DASHMATCH", "an attribute hyphen match")
	TokenTkPrefixmatch    = newToken(TokenTypePrefixmatch, "PREFIXMATCH", "an attribute prefix match")
	TokenTkSuffixmatch    = newToken(TokenTypeSuffixmatch, "SUFFIXMATCH", "an attribute suffix match")
	TokenTkSubstringmatch = newToken(TokenTypeSubstringmatch, "SUBSTRINGMATCH", "an attribute substring match")
	TokenTkLbrace         = newToken(TokenTypeLbrace, "LBRACE", "a {")
	TokenTkPlus           = newToken(TokenTypePlus, "PLUS", "a +")
	TokenTkGreater        = newToken(TokenTypeGreater, "GREATER", "a >")
	TokenTkComma          = newToken(TokenTypeComma, "COMMA", "a comma")
	TokenTkString         = newToken(TokenTypeString, "STRING", "a string")
	TokenTkInvalid        = newToken(TokenTypeInvalid, "INVALID", "an unclosed string")
	TokenTkIdent          = newToken(TokenTypeIdent, "IDENT", "an identifier")
	TokenTkHash           = newToken(TokenTypeHash, "HASH", "a hex color")
	TokenTkImportSym      = newToken(TokenTypeImportSym, "IMPORT_SYM", "@import")
	TokenTkPageSym        = newToken(TokenTypePageSym, "PAGE_SYM", "@page")
	TokenTkMediaSym       = newToken(TokenTypeMediaSym, "MEDIA_SYM", "@media")
	TokenTkCharsetSym     = newToken(TokenTypeCharsetSym, "CHARSET_SYM", "@charset")
	TokenTkNamespaceSym   = newToken(TokenTypeNamespaceSym, "NAMESPACE_SYM", "@namespace,")
	TokenTkFontFaceSym    = newToken(TokenTypeFontFaceSym, "FONT_FACE_SYM", "@font-face")
	TokenTkAtRule         = newToken(TokenTypeAtRule, "AT_RULE", "at rule")
	TokenTkImportantSym   = newToken(TokenTypeImportantSym, "IMPORTANT_SYM", "!important")
	TokenTkEms            = newToken(TokenTypeEms, "EMS", "an em value")
	TokenTkExs            = newToken(TokenTypeExs, "EXS", "an ex value")
	TokenTkPx             = newToken(TokenTypePx, "PX", "a pixel value")
	TokenTkCm             = newToken(TokenTypeCm, "CM", "a centimeter value")
	TokenTkMm             = newToken(TokenTypeMm, "MM", "a millimeter value")
	TokenTkIn             = newToken(TokenTypeIn, "IN", "an inch value")
	TokenTkPt             = newToken(TokenTypePt, "PT", "a point value")
	TokenTkPc             = newToken(TokenTypePc, "PC", "a pica value")
	TokenTkAngle          = newToken(TokenTypeAngle, "ANGLE", "an angle value")
	TokenTkTime           = newToken(TokenTypeTime, "TIME", "a time value")
	TokenTkFreq           = newToken(TokenTypeFreq, "FREQ", "a freq value")
	TokenTkDimension      = newToken(TokenTypeDimension, "DIMENSION", "a dimension")
	TokenTkPercentage     = newToken(TokenTypePercentage, "PERCENTAGE", "a percentage")
	TokenTkNumber         = newToken(TokenTypeNumber, "NUMBER", "a number")
	TokenTkUri            = newToken(TokenTypeUri, "URI", "a URI")
	TokenTkFunction       = newToken(TokenTypeFunction, "FUNCTION", "function")
	TokenTkOther          = newToken(TokenTypeOther, "OTHER", "other")
	TokenTkRbrace         = newToken(TokenTypeRbrace, "RBRACE", "}")
	TokenTkSemicolon      = newToken(TokenTypeSemicolon, "SEMICOLON", ";")
	TokenTkVirgule        = newToken(TokenTypeVirgule, "VIRGULE", "/")
	TokenTkColon          = newToken(TokenTypeColon, "COLON", ":")
	TokenTkMinus          = newToken(TokenTypeMinus, "MINUS", "-")
	TokenTkRparen         = newToken(TokenTypeRparen, "RPAREN", ")")
	TokenTkLbracket       = newToken(TokenTypeLbracket, "LBRACKET", "[")
	TokenTkRbracket       = newToken(TokenTypeRbracket, "RBRACKET", "]")
	TokenTkPeriod         = newToken(TokenTypePeriod, "PERIOD", ".")
	TokenTkEquals         = newToken(TokenTypeEquals, "EQUALS", "=")
	TokenTkAsterisk       = newToken(TokenTypeAsterisk, "ASTERISK", "*")
	TokenTkVerticalBar    = newToken(TokenTypeVerticalBar, "VERTICAL_BAR", "|")
	TokenTkEof            = newToken(TokenTypeEof, "EOF", "end of file")
)

type Token struct {
	tokenType    TokenType
	name         string
	externalName string
}

func newToken(tokenType TokenType, name string, externalName string) *Token {
	return &Token{tokenType: tokenType, name: name, externalName: externalName}
}

func (t *Token) GetType() TokenType {
	return t.tokenType
}

func (t *Token) GetName() string {
	return t.name
}

func (t *Token) GetExternalName() string {
	return t.externalName
}

func (t *Token) ToString() string {
	return t.name
}

func (t *Token) String() string {
	return t.name
}

func TokenCreateOtherToken(value string) *Token {
	return newToken(TokenTypeOther, "OTHER", value+" (other)")
}
