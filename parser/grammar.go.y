/*
  This parser definition is based on min-caml/parser.mly
  Copyright (c) 2005-2008, Eijiro Sumii, Moe Masuko, and Kenichi Asai
*/

%{
package parser

import (
	"fmt"
	"strconv"
	"github.com/rhysd/gocaml/ast"
	"github.com/rhysd/gocaml/token"
)
%}

%union{
	node ast.Expr
	nodes []ast.Expr
	token *token.Token
	funcdef *ast.FuncDef
	decls []*ast.Symbol
}

%token<token> ILLEGAL
%token<token> COMMENT
%token<token> LPAREN
%token<token> RPAREN
%token<token> IDENT
%token<token> BOOL
%token<token> NOT
%token<token> INT
%token<token> FLOAT
%token<token> MINUS
%token<token> PLUS
%token<token> MINUS_DOT
%token<token> PLUS_DOT
%token<token> STAR_DOT
%token<token> SLASH_DOT
%token<token> EQUAL
%token<token> LESS_GREATER
%token<token> LESS_EQUAL
%token<token> LESS
%token<token> GREATER
%token<token> GREATER_EQUAL
%token<token> IF
%token<token> THEN
%token<token> ELSE
%token<token> LET
%token<token> IN
%token<token> REC
%token<token> COMMA
%token<token> ARRAY_MAKE
%token<token> DOT
%token<token> LESS_MINUS
%token<token> SEMICOLON
%token<token> STAR
%token<token> SLASH
%token<token> BAR_BAR
%token<token> AND_AND
%token<token> ARRAY_SIZE
%token<token> STRING_LITERAL

%right prec_let
%right SEMICOLON
%right prec_if
%right LESS_MINUS
%left COMMA
%left BAR_BAR
%left AND_AND
%left EQUAL LESS_GREATER LESS GREATER LESS_EQUAL GREATER_EQUAL
%left PLUS MINUS PLUS_DOT MINUS_DOT
%left STAR SLASH STAR_DOT SLASH_DOT
%right prec_unary_minus
%left prec_app
%left DOT

%type<node> exp
%type<node> parenless_exp
%type<nodes> elems
%type<nodes> args
%type<decls> params
%type<decls> pat
%type<funcdef> fundef
%type<> program

%start program

%%

program:
	exp
		{
			yylex.(*pseudoLexer).result = $1
		}

exp:
	parenless_exp
		{ $$ = $1 }
	| NOT exp
		%prec prec_app
		{ $$ = &ast.Not{$1, $2} }
	| MINUS exp
		%prec prec_unary_minus
		{ $$ = &ast.Neg{$1, $2} }
	| exp PLUS exp
		{ $$ = &ast.Add{$1, $3} }
	| exp MINUS exp
		{ $$ = &ast.Sub{$1, $3} }
	| exp STAR exp
		{ $$ = &ast.Mul{$1, $3} }
	| exp SLASH exp
		{ $$ = &ast.Div{$1, $3} }
	| exp EQUAL exp
		{ $$ = &ast.Eq{$1, $3} }
	| exp LESS_GREATER exp
		{ $$ = &ast.NotEq{$1, $3} }
	| exp LESS exp
		{ $$ = &ast.Less{$1, $3} }
	| exp GREATER exp
		{ $$ = &ast.Greater{$1, $3} }
	| exp LESS_EQUAL exp
		{ $$ = &ast.LessEq{$1, $3} }
	| exp GREATER_EQUAL exp
		{ $$ = &ast.GreaterEq{$1, $3} }
	| exp AND_AND exp
		{ $$ = &ast.And{$1, $3} }
	| exp BAR_BAR exp
		{ $$ = &ast.Or{$1, $3} }
	| IF exp THEN exp ELSE exp
		%prec prec_if
		{ $$ = &ast.If{$1, $2, $4, $6} }
	| MINUS_DOT exp
		%prec prec_unary_minus
		{ $$ = &ast.FNeg{$1, $2} }
	| exp PLUS_DOT exp
		{ $$ = &ast.FAdd{$1, $3} }
	| exp MINUS_DOT exp
		{ $$ = &ast.FSub{$1, $3} }
	| exp STAR_DOT exp
		{ $$ = &ast.FMul{$1, $3} }
	| exp SLASH_DOT exp
		{ $$ = &ast.FDiv{$1, $3} }
	| LET IDENT EQUAL exp IN exp
		%prec prec_let
		{ $$ = &ast.Let{$1, ast.NewSymbol($2.Value()), $4, $6} }
	| LET REC fundef IN exp
		%prec prec_let
		{ $$ = &ast.LetRec{$1, $3, $5} }
	| exp args
		%prec prec_app
		{ $$ = &ast.Apply{$1, revExprs($2)} }
	| elems
		{ $$ = &ast.Tuple{$1} }
	| LET LPAREN pat RPAREN EQUAL exp IN exp
		{ $$ = &ast.LetTuple{$1, $3, $6, $8} }
	| parenless_exp DOT LPAREN exp RPAREN LESS_MINUS exp
		{ $$ = &ast.Put{$1, $4, $7} }
	| exp SEMICOLON exp
		{ $$ = &ast.Let{$2, ast.NewSymbol(genTempId()), $1, $3} }
	| ARRAY_MAKE parenless_exp parenless_exp
		%prec prec_app
		{ $$ = &ast.ArrayCreate{$1, $2, $3} }
	| ARRAY_SIZE parenless_exp
		%prec prec_app
		{ $$ = &ast.ArraySize{$1, $2} }
	| ILLEGAL error
		{
			yylex.Error(fmt.Sprintf("Parsing illegal token: %s", $1.String()))
			$$ = nil
		}

fundef:
	IDENT params EQUAL exp
		{ $$ = &ast.FuncDef{ast.NewSymbol($1.Value()), revSyms($2), $4} }

params:
	IDENT params
		{ $$ = append($2, ast.NewSymbol($1.Value())) }
	| IDENT
		{ $$ = []*ast.Symbol{ast.NewSymbol($1.Value())} }

args:
	parenless_exp args
		{ $$ = append($2, $1) }
	| parenless_exp
		{ $$ = []ast.Expr{$1} }

elems:
	elems COMMA exp
		{ $$ = append($1, $3) }
	| exp COMMA exp
		{ $$ = []ast.Expr{$1, $3} }

pat:
	pat COMMA IDENT
		{ $$ = append($1, ast.NewSymbol($3.Value())) }
	| IDENT COMMA IDENT
		{
			$$ = []*ast.Symbol{
				ast.NewSymbol($1.Value()),
				ast.NewSymbol($3.Value()), 
			}
		}

parenless_exp:
	LPAREN exp RPAREN
		{ $$ = $2 }
	| LPAREN RPAREN
		{ $$ = &ast.Unit{$1, $2} }
	| BOOL
		{ $$ = &ast.Bool{$1, $1.Value() == "true"} }
	| INT
		{
			i, err := strconv.Atoi($1.Value())
			if err != nil {
				yylex.Error("Parse error")
			} else {
				$$ = &ast.Int{$1, i}
			}
		}
	| FLOAT
		{
			f, err := strconv.ParseFloat($1.Value(), 64)
			if err != nil {
				yylex.Error("Parse error")
			} else {
				$$ = &ast.Float{$1, f}
			}
		}
	| STRING_LITERAL
		{
			from := $1.Value()
			s, err := strconv.Unquote(from)
			if err != nil {
				yylex.Error("Parse error on string literal: " + from)
			} else {
				$$ = &ast.String{$1, s}
			}
		}
	| IDENT
		{ $$ = &ast.VarRef{$1, ast.NewSymbol($1.Value())} }
	| parenless_exp DOT LPAREN exp RPAREN
		{ $$ = &ast.Get{$1, $4} }

%%

func revSyms(syms []*ast.Symbol) []*ast.Symbol {
	l := len(syms)
	for i := 0; i < l / 2; i++ {
		j := l - i - 1
		syms[i], syms[j] = syms[j], syms[i]
	}
	return syms
}

func revExprs(exprs []ast.Expr) []ast.Expr {
	l := len(exprs)
	for i := 0; i < l / 2; i++ {
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
