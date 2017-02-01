//line parser/grammar.go.y:7
package parser

import __yyfmt__ "fmt"

//line parser/grammar.go.y:7
import (
	"fmt"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
	"strconv"
)

//line parser/grammar.go.y:17
type yySymType struct {
	yys     int
	node    ast.Expr
	nodes   []ast.Expr
	token   *token.Token
	funcdef *ast.FuncDef
	decls   []*ast.Symbol
}

const ILLEGAL = 57346
const COMMENT = 57347
const LPAREN = 57348
const RPAREN = 57349
const IDENT = 57350
const BOOL = 57351
const NOT = 57352
const INT = 57353
const FLOAT = 57354
const MINUS = 57355
const PLUS = 57356
const MINUS_DOT = 57357
const PLUS_DOT = 57358
const STAR_DOT = 57359
const SLASH_DOT = 57360
const EQUAL = 57361
const LESS_GREATER = 57362
const LESS_EQUAL = 57363
const LESS = 57364
const GREATER = 57365
const GREATER_EQUAL = 57366
const IF = 57367
const THEN = 57368
const ELSE = 57369
const LET = 57370
const IN = 57371
const REC = 57372
const COMMA = 57373
const ARRAY_CREATE = 57374
const DOT = 57375
const LESS_MINUS = 57376
const SEMICOLON = 57377
const prec_let = 57378
const prec_if = 57379
const prec_unary_minus = 57380
const prec_app = 57381

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"ILLEGAL",
	"COMMENT",
	"LPAREN",
	"RPAREN",
	"IDENT",
	"BOOL",
	"NOT",
	"INT",
	"FLOAT",
	"MINUS",
	"PLUS",
	"MINUS_DOT",
	"PLUS_DOT",
	"STAR_DOT",
	"SLASH_DOT",
	"EQUAL",
	"LESS_GREATER",
	"LESS_EQUAL",
	"LESS",
	"GREATER",
	"GREATER_EQUAL",
	"IF",
	"THEN",
	"ELSE",
	"LET",
	"IN",
	"REC",
	"COMMA",
	"ARRAY_CREATE",
	"DOT",
	"LESS_MINUS",
	"SEMICOLON",
	"prec_let",
	"prec_if",
	"prec_unary_minus",
	"prec_app",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line parser/grammar.go.y:217

func revSyms(syms []*ast.Symbol) []*ast.Symbol {
	l := len(syms)
	for i := 0; i < l/2; i++ {
		j := l - i - 1
		syms[i], syms[j] = syms[j], syms[i]
	}
	return syms
}

func revExprs(exprs []ast.Expr) []ast.Expr {
	l := len(exprs)
	for i := 0; i < l/2; i++ {
		j := l - i - 1
		exprs[i], exprs[j] = exprs[j], exprs[i]
	}
	return exprs
}

var genCount = 0

func genTempId() string {
	genCount += 1
	return fmt.Sprintf("$unused%d", genCount)
}

// vim: noet

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 44
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 471

var yyAct = [...]int{

	32, 3, 77, 93, 61, 3, 3, 3, 3, 33,
	79, 42, 29, 3, 40, 81, 38, 41, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	76, 3, 3, 89, 80, 87, 64, 91, 39, 90,
	78, 68, 3, 70, 12, 60, 16, 13, 66, 14,
	15, 18, 17, 26, 25, 27, 28, 19, 20, 23,
	21, 22, 24, 3, 3, 3, 72, 62, 43, 31,
	1, 65, 67, 3, 9, 0, 0, 3, 0, 0,
	0, 88, 0, 0, 0, 3, 3, 0, 3, 0,
	3, 2, 0, 0, 3, 0, 34, 35, 36, 37,
	3, 0, 0, 0, 44, 0, 0, 0, 0, 46,
	47, 48, 49, 50, 51, 52, 53, 54, 55, 56,
	57, 0, 58, 59, 0, 0, 12, 0, 16, 13,
	0, 14, 15, 69, 12, 0, 16, 13, 0, 14,
	15, 18, 17, 26, 25, 27, 28, 19, 20, 23,
	21, 22, 24, 61, 73, 74, 75, 99, 0, 31,
	0, 0, 0, 30, 82, 0, 0, 12, 86, 16,
	13, 0, 14, 15, 0, 0, 94, 95, 0, 96,
	0, 97, 0, 0, 0, 98, 0, 12, 92, 16,
	13, 100, 14, 15, 18, 17, 26, 25, 27, 28,
	19, 20, 23, 21, 22, 24, 0, 0, 0, 0,
	0, 0, 31, 0, 0, 12, 30, 16, 13, 0,
	14, 15, 18, 17, 26, 25, 27, 28, 19, 20,
	23, 21, 22, 24, 0, 0, 0, 0, 85, 0,
	31, 0, 0, 12, 30, 16, 13, 0, 14, 15,
	18, 17, 26, 25, 27, 28, 19, 20, 23, 21,
	22, 24, 0, 0, 84, 0, 0, 0, 31, 0,
	0, 0, 30, 12, 83, 16, 13, 0, 14, 15,
	18, 17, 26, 25, 27, 28, 19, 20, 23, 21,
	22, 24, 0, 0, 0, 0, 0, 0, 31, 0,
	0, 0, 30, 12, 71, 16, 13, 0, 14, 15,
	18, 17, 26, 25, 27, 28, 19, 20, 23, 21,
	22, 24, 0, 0, 0, 0, 0, 0, 31, 0,
	0, 12, 30, 16, 13, 0, 14, 15, 18, 17,
	26, 25, 27, 28, 19, 20, 23, 21, 22, 24,
	0, 63, 0, 0, 0, 0, 31, 0, 0, 12,
	30, 16, 13, 0, 14, 15, 18, 17, 26, 25,
	27, 28, 19, 20, 23, 21, 22, 24, 0, 0,
	0, 0, 0, 0, 31, 0, 0, 11, 30, 12,
	45, 16, 13, 4, 14, 15, 5, 0, 7, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 6, 0,
	0, 8, 11, 0, 12, 10, 16, 13, 4, 14,
	15, 5, 0, 7, 0, 0, 12, 0, 16, 13,
	0, 14, 15, 6, 0, 0, 8, 27, 28, 12,
	10, 16, 13, 0, 14, 15, 18, 17, 26, 25,
	27, 28, 19, 20, 23, 21, 22, 24, 12, 0,
	16, 13, 0, 14, 15, 18, 17, 26, 25, 27,
	28,
}
var yyPact = [...]int{

	408, -1000, 353, -24, 408, 408, 408, 408, 8, -14,
	161, 66, 383, -1000, -1000, -1000, -1000, 408, 408, 408,
	408, 408, 408, 408, 408, 408, 408, 408, 408, -1000,
	408, 408, 120, 61, 161, 161, 325, 161, 17, 40,
	33, 408, 120, -1000, 297, -1000, 420, 420, 452, 452,
	452, 452, 452, 452, 420, 420, 161, 161, 353, 433,
	-1000, 60, 408, 408, 408, 1, 32, 3, -16, 433,
	-29, -1000, 408, 267, 237, 209, 408, 16, 32, 14,
	31, 29, 181, -31, 408, 408, 353, 408, -1000, 408,
	-1000, -1000, -1000, 408, 38, 353, 353, 128, 38, 408,
	353,
}
var yyPgo = [...]int{

	0, 91, 0, 74, 12, 2, 72, 71, 70,
}
var yyR1 = [...]int{

	0, 8, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 7, 5,
	5, 4, 4, 3, 3, 6, 6, 2, 2, 2,
	2, 2, 2, 2,
}
var yyR2 = [...]int{

	0, 1, 1, 2, 2, 3, 3, 3, 3, 3,
	3, 3, 3, 6, 2, 3, 3, 3, 3, 6,
	5, 2, 1, 8, 7, 3, 3, 2, 4, 2,
	1, 2, 1, 3, 3, 3, 3, 3, 2, 1,
	1, 1, 1, 5,
}
var yyChk = [...]int{

	-1000, -8, -1, -2, 10, 13, 25, 15, 28, -3,
	32, 4, 6, 9, 11, 12, 8, 14, 13, 19,
	20, 22, 23, 21, 24, 16, 15, 17, 18, -4,
	35, 31, -2, 33, -1, -1, -1, -1, 8, 30,
	6, 31, -2, 2, -1, 7, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-4, 33, 6, 26, 19, -7, 8, -6, 8, -1,
	-2, 7, 6, -1, -1, -1, 29, -5, 8, 7,
	31, 31, -1, 7, 27, 29, -1, 19, -5, 19,
	8, 8, 7, 34, -1, -1, -1, -1, -1, 29,
	-1,
}
var yyDef = [...]int{

	0, -2, 1, 2, 0, 0, 0, 0, 0, 22,
	0, 0, 0, 39, 40, 41, 42, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 21,
	0, 0, 32, 0, 3, 4, 0, 14, 0, 0,
	0, 0, 0, 27, 0, 38, 5, 6, 7, 8,
	9, 10, 11, 12, 15, 16, 17, 18, 25, 34,
	31, 0, 0, 0, 0, 0, 0, 0, 0, 33,
	26, 37, 0, 0, 0, 0, 0, 0, 30, 0,
	0, 0, 0, 43, 0, 0, 20, 0, 29, 0,
	35, 36, 43, 0, 13, 19, 28, 0, 24, 0,
	23,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:85
		{
			yylex.(*pseudoLexer).result = yyDollar[1].node
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:91
		{
			yyVAL.node = yyDollar[1].node
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:94
		{
			yyVAL.node = &ast.Not{yyDollar[1].token, yyDollar[2].node}
		}
	case 4:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:97
		{
			yyVAL.node = &ast.Neg{yyDollar[1].token, yyDollar[2].node}
		}
	case 5:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:99
		{
			yyVAL.node = &ast.Add{yyDollar[1].node, yyDollar[3].node}
		}
	case 6:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:101
		{
			yyVAL.node = &ast.Sub{yyDollar[1].node, yyDollar[3].node}
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:103
		{
			yyVAL.node = &ast.Eq{yyDollar[1].node, yyDollar[3].node}
		}
	case 8:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:105
		{
			yyVAL.node = &ast.Not{yyDollar[2].token, &ast.Eq{yyDollar[1].node, yyDollar[3].node}}
		}
	case 9:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:107
		{
			yyVAL.node = &ast.Less{yyDollar[1].node, yyDollar[3].node}
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:109
		{
			yyVAL.node = &ast.Not{yyDollar[2].token, &ast.Less{yyDollar[1].node, yyDollar[3].node}}
		}
	case 11:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:111
		{
			yyVAL.node = &ast.Not{yyDollar[2].token, &ast.Less{yyDollar[3].node, yyDollar[1].node}}
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:113
		{
			yyVAL.node = &ast.Less{yyDollar[3].node, yyDollar[1].node}
		}
	case 13:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line parser/grammar.go.y:116
		{
			yyVAL.node = &ast.If{yyDollar[1].token, yyDollar[2].node, yyDollar[4].node, yyDollar[6].node}
		}
	case 14:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:119
		{
			yyVAL.node = &ast.FNeg{yyDollar[1].token, yyDollar[2].node}
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:121
		{
			yyVAL.node = &ast.FAdd{yyDollar[1].node, yyDollar[3].node}
		}
	case 16:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:123
		{
			yyVAL.node = &ast.FSub{yyDollar[1].node, yyDollar[3].node}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:125
		{
			yyVAL.node = &ast.FMul{yyDollar[1].node, yyDollar[3].node}
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:127
		{
			yyVAL.node = &ast.FDiv{yyDollar[1].node, yyDollar[3].node}
		}
	case 19:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line parser/grammar.go.y:130
		{
			yyVAL.node = &ast.Let{yyDollar[1].token, ast.NewSymbol(yyDollar[2].token.Value()), yyDollar[4].node, yyDollar[6].node}
		}
	case 20:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line parser/grammar.go.y:133
		{
			yyVAL.node = &ast.LetRec{yyDollar[1].token, yyDollar[3].funcdef, yyDollar[5].node}
		}
	case 21:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:136
		{
			yyVAL.node = &ast.Apply{yyDollar[1].node, revExprs(yyDollar[2].nodes)}
		}
	case 22:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:138
		{
			yyVAL.node = &ast.Tuple{yyDollar[1].nodes}
		}
	case 23:
		yyDollar = yyS[yypt-8 : yypt+1]
		//line parser/grammar.go.y:140
		{
			yyVAL.node = &ast.LetTuple{yyDollar[1].token, yyDollar[3].decls, yyDollar[6].node, yyDollar[8].node}
		}
	case 24:
		yyDollar = yyS[yypt-7 : yypt+1]
		//line parser/grammar.go.y:142
		{
			yyVAL.node = &ast.Put{yyDollar[1].node, yyDollar[4].node, yyDollar[7].node}
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:144
		{
			yyVAL.node = &ast.Let{yyDollar[2].token, ast.NewSymbol(genTempId()), yyDollar[1].node, yyDollar[3].node}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:147
		{
			yyVAL.node = &ast.ArrayCreate{yyDollar[1].token, yyDollar[2].node, yyDollar[3].node}
		}
	case 27:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:149
		{
			yylex.Error(fmt.Sprintf("Parsing illegal token: %s", yyDollar[1].token.String()))
			yyVAL.node = nil
		}
	case 28:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line parser/grammar.go.y:156
		{
			yyVAL.funcdef = &ast.FuncDef{ast.NewSymbol(yyDollar[1].token.Value()), revSyms(yyDollar[2].decls), yyDollar[4].node}
		}
	case 29:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:160
		{
			yyVAL.decls = append(yyDollar[2].decls, ast.NewSymbol(yyDollar[1].token.Value()))
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:162
		{
			yyVAL.decls = []*ast.Symbol{ast.NewSymbol(yyDollar[1].token.Value())}
		}
	case 31:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:166
		{
			yyVAL.nodes = append(yyDollar[2].nodes, yyDollar[1].node)
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:168
		{
			yyVAL.nodes = []ast.Expr{yyDollar[1].node}
		}
	case 33:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:172
		{
			yyVAL.nodes = append(yyDollar[1].nodes, yyDollar[3].node)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:174
		{
			yyVAL.nodes = []ast.Expr{yyDollar[1].node, yyDollar[3].node}
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:178
		{
			yyVAL.decls = append(yyDollar[1].decls, ast.NewSymbol(yyDollar[3].token.Value()))
		}
	case 36:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:180
		{
			yyVAL.decls = []*ast.Symbol{
				ast.NewSymbol(yyDollar[1].token.Value()),
				ast.NewSymbol(yyDollar[3].token.Value()),
			}
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line parser/grammar.go.y:189
		{
			yyVAL.node = yyDollar[2].node
		}
	case 38:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line parser/grammar.go.y:191
		{
			yyVAL.node = &ast.Unit{yyDollar[1].token, yyDollar[2].token}
		}
	case 39:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:193
		{
			yyVAL.node = &ast.Bool{yyDollar[1].token, yyDollar[1].token.Value() == "true"}
		}
	case 40:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:195
		{
			i, err := strconv.Atoi(yyDollar[1].token.Value())
			if err != nil {
				yylex.Error("Parse error")
			} else {
				yyVAL.node = &ast.Int{yyDollar[1].token, i}
			}
		}
	case 41:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:204
		{
			f, err := strconv.ParseFloat(yyDollar[1].token.Value(), 64)
			if err != nil {
				yylex.Error("Parse error")
			} else {
				yyVAL.node = &ast.Float{yyDollar[1].token, f}
			}
		}
	case 42:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line parser/grammar.go.y:213
		{
			yyVAL.node = &ast.VarRef{yyDollar[1].token, ast.NewSymbol(yyDollar[1].token.Value())}
		}
	case 43:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line parser/grammar.go.y:215
		{
			yyVAL.node = &ast.Get{yyDollar[1].node, yyDollar[4].node}
		}
	}
	goto yystack /* stack new state and value */
}
