package token

import "fmt"

type Token int

type TokenPos struct {
	Filename string
	Offset   int
}

var (
	keywords map[string]Token
)

// The list of tokens
const (
	// Special Tokens
	ILLEGAL Token = iota
	EOF
	COMMENT // ;

	literal_begin
	IDENT // person
	INT   // Unisigned integer or negative integer: -9 / 9
	UINT  // Unsigned integer: 9
	NINT  // Negative integer: -9

	FLOAT   // float; one of float16, float32, float64
	FLOAT16 // A number representable as an IEEE 754 half-precision float (major type 7, additional information 25).
	FLOAT32 // A number representable as an IEEE 754 single-precision float (major type 7, additional information 26).
	FLOAT64 // A number representable as an IEEE 754 double-precision float (major type 7, additional information 27).
	TSTR    // "text"
	TEXT    // "text"

	BYTES // bytes
	BSTR  // bstr

	BOOL  // bool
	TRUE  // true
	FALSE // false
	NIL   // nil
	NULL  // null
	ANY   // any

	literal_end

	// Operators and Delimiters
	operator_begin
	MINUS  // -
	LPAREN // (
	LBRACK // [
	LBRACE // {

	RPAREN // )
	RBRACK // ]
	RBRACE // }

	COMMA  // ,
	PERIOD // .
	COLON  // :

	ASSIGN              // =
	ARROW_MAP           // =>
	OPTIONAL            // ?
	TYPE_CHOICE         // /
	GROUP_CHOICE        // //
	TYPE_CHOICE_ASSIGN  // /=
	GROUP_CHOICE_ASSIGN // /=
	AMPERSAND           // &
	HASH                // #

	LEFT_ANGLE_BRACKET  // <
	RIGHT_ANGLE_BRACKET // >

	INCLUSIVE_BOUND // ..
	EXCLUSIVE_BOUND // ...

	TYPE_SOCKET  // $
	GROUP_SOCKET // $$

	ZERO_OR_MORE // *
	ONE_OR_MORE  // +
	UNWRAP       // ~
	operator_end

	control_operators_begin
	SIZE    // .size
	BITS    // .bits
	REGEXP  // .regexp
	CBOR    // .cbor
	CBORSEQ // .cborseq
	WITHIN  // .within
	AND     // .and

	LT      // .lt
	LE      // .le
	GT      // .gt
	GE      // .ge
	EQ      // .ne
	DEFAULT // .default

	// Additional control operators RFC9165

	PLUS    // .plus
	CAT     // .cat
	DET     // .det
	ABNF    // .abnf
	ABNFB   // .abnfb
	FEATURE // .feature

	control_operators_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:   "IDENT",
	INT:     "int",
	UINT:    "uint",
	NINT:    "nint",
	FLOAT:   "float",
	FLOAT16: "float16",
	FLOAT32: "float32",
	FLOAT64: "float16",
	TSTR:    "tstr",
	TEXT:    "text",

	BYTES: "bytes",
	BSTR:  "bstr",

	BOOL:  "bool",
	TRUE:  "true",
	FALSE: "false",
	NIL:   "nil",
	NULL:  "null",
	ANY:   "any",

	MINUS: "-",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",

	RPAREN: ")",
	RBRACK: "]",
	RBRACE: "}",

	COMMA:  ",",
	PERIOD: ".",
	COLON:  ":",

	ASSIGN:              "=",
	ARROW_MAP:           "=>",
	OPTIONAL:            "?",
	TYPE_CHOICE:         "/",
	GROUP_CHOICE:        "//",
	TYPE_CHOICE_ASSIGN:  "/=",
	GROUP_CHOICE_ASSIGN: "//=",
	AMPERSAND:           "&",
	HASH:                "#",

	LEFT_ANGLE_BRACKET:  "<",
	RIGHT_ANGLE_BRACKET: ">",

	INCLUSIVE_BOUND: "..",
	EXCLUSIVE_BOUND: "...",

	TYPE_SOCKET:  "$",
	GROUP_SOCKET: "$$",

	ZERO_OR_MORE: "*",
	ONE_OR_MORE:  "+",
	UNWRAP:       "~",

	SIZE:    ".size",
	BITS:    ".bits",
	REGEXP:  ".regexp",
	CBOR:    ".cbor",
	CBORSEQ: ".cborseq",
	WITHIN:  ".within",
	AND:     ".and",

	LT:      ".lt",
	LE:      ".le",
	GT:      ".gt",
	GE:      ".ge",
	EQ:      ".ne",
	DEFAULT: ".default",

	// Additional control operators RFC9165

	PLUS:    ".plus",
	CAT:     ".cat",
	DET:     ".det",
	ABNF:    ".abnf",
	ABNFB:   ".abnfb",
	FEATURE: ".feature",
}

func (t Token) String() string {
	out := ""
	if t >= 0 && t < Token(len(tokens)) {
		out += string(tokens[t])
	}
	if out == "" {
		out = fmt.Sprintf("%d", t)
	}
	return out
}

func Lookup(str string) Token {
	if v, ok := keywords[str]; ok {
		return v
	}
	return IDENT
}

func init() {
	keywords = make(map[string]Token)
	// skip the IDENT literal
	for i := Token(INT + 1); i < literal_end; i++ {
		v := tokens[i]
		keywords[v] = i
	}
	for i := Token(control_operators_begin + 1); i < control_operators_end; i++ {
		vc := tokens[i]
		keywords[vc] = i
	}
}
