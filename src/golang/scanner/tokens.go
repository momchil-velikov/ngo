package scanner

const (
    EOF = 0

    PLUS     = '+'
    MINUS    = '-'
    MUL      = '*'
    DIV      = '/'
    REM      = '%'
    BITAND   = '&'
    BITOR    = '|'
    BITXOR   = '^'
    LT       = '<'
    GT       = '>'
    ASSIGN   = '='
    NOT      = '!'
    LPAREN   = '('
    RPAREN   = ')'
    LBRACKET = '['
    RBRACKET = ']'
    LBRACE   = '{'
    RBRACE   = '}'
    COMMA    = ','
    DOT      = '.'
    SEMI     = ';'
    COLON    = ':'

    SHL          = 256 + iota // <<
    SHR                       // >>
    ANDN                      // &^
    PLUS_ASSIGN               // +=
    MINUS_ASSIGN              // -=
    MUL_ASSIGN                // *=
    DIV_ASSIGN                // /*
    REM_ASSIGN                // %=
    AND_ASSIGN                // &=
    OR_ASSIGN                 // |=
    XOR_ASSIGN                // ^=
    SHL_ASSIGN                // <<=
    SHR_ASSIGN                // >>=
    ANDN_ASSIGN               // &^=
    AND                       // &&
    OR                        // ||
    RECV                      // <-
    INC                       // ++
    DEC                       // --
    EQ                        // ==
    NE                        // !=
    LE                        // <=
    GE                        // >=
    DEFINE                    // :=
    DOTS                      // ...

    BREAK
    CASE
    CHAN
    CONST
    CONTINUE
    DEFAULT
    DEFER
    ELSE
    FALLTHROUGH
    FOR
    FUNC
    GO
    GOTO
    IF
    IMPORT
    INTERFACE
    MAP
    PACKAGE
    RANGE
    RETURN
    SELECT
    STRUCT
    SWITCH
    TYPE
    VAR

    ID
    INTEGER
    FLOAT
    IMAGINARY
    STRING
    RUNE

    ERROR

    SPACE = '\u0020'
    CR    = '\u000d'
    NL    = '\u000a'
    TAB   = '\u0009'
)

var keywords = map[string]uint{
    "break":       BREAK,
    "case":        CASE,
    "chan":        CHAN,
    "const":       CONST,
    "continue":    CONTINUE,
    "default":     DEFAULT,
    "defer":       DEFER,
    "else":        ELSE,
    "fallthrough": FALLTHROUGH,
    "for":         FOR,
    "func":        FUNC,
    "go":          GO,
    "goto":        GOTO,
    "if":          IF,
    "import":      IMPORT,
    "interface":   INTERFACE,
    "map":         MAP,
    "package":     PACKAGE,
    "range":       RANGE,
    "return":      RETURN,
    "select":      SELECT,
    "struct":      STRUCT,
    "switch":      SWITCH,
    "type":        TYPE,
    "var":         VAR,
}

var TokenNames = map[uint]string{
    EOF:      "<eof>",
    PLUS:     "+",
    MINUS:    "-",
    MUL:      "*",
    DIV:      "/",
    REM:      "%",
    BITAND:   "&",
    BITOR:    "|",
    BITXOR:   "^",
    LT:       "<",
    GT:       ">",
    ASSIGN:   "=",
    NOT:      "!",
    LPAREN:   "(",
    RPAREN:   ")",
    LBRACKET: "[",
    RBRACKET: "]",
    LBRACE:   "{",
    RBRACE:   "}",
    COMMA:    ",",
    DOT:      ".",
    SEMI:     ";",
    COLON:    ":",

    SHL:          "<<",
    SHR:          ">>",
    ANDN:         "&^",
    PLUS_ASSIGN:  "+=",
    MINUS_ASSIGN: "-=",
    MUL_ASSIGN:   "*=",
    DIV_ASSIGN:   "/=",
    REM_ASSIGN:   "%=",
    AND_ASSIGN:   "&=",
    OR_ASSIGN:    "|=",
    XOR_ASSIGN:   "^=",
    SHL_ASSIGN:   "<<=",
    SHR_ASSIGN:   ">>=",
    ANDN_ASSIGN:  "&^=",
    AND:          "&&",
    OR:           "||",
    RECV:         "<-",
    INC:          "++",
    DEC:          "--",
    EQ:           "==",
    NE:           "!=",
    LE:           "<=",
    GE:           ">=",
    DEFINE:       ":=",
    DOTS:         "...",

    BREAK:       "break",
    CASE:        "case",
    CHAN:        "chan",
    CONST:       "const",
    CONTINUE:    "continue",
    DEFAULT:     "default",
    DEFER:       "defer",
    ELSE:        "else",
    FALLTHROUGH: "fallthrough",
    FOR:         "for",
    FUNC:        "func",
    GO:          "go",
    GOTO:        "goto",
    IF:          "if",
    IMPORT:      "import",
    INTERFACE:   "interface",
    MAP:         "map",
    PACKAGE:     "package",
    RANGE:       "range",
    RETURN:      "return",
    SELECT:      "select",
    STRUCT:      "struct",
    SWITCH:      "switch",
    TYPE:        "type",
    VAR:         "var",

    ID:        "<id>",
    INTEGER:   "<integer>",
    FLOAT:     "<float>",
    IMAGINARY: "<imaginary>",
    STRING:    "<string>",
    RUNE:      "<rune>",

    ERROR: "<error>",
}
