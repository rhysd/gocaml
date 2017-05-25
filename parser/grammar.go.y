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
	decl *ast.Symbol
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
%token<token> ARRAY_LENGTH
%token<token> STRING_LITERAL
%token<token> PERCENT
%token<token> MATCH
%token<token> WITH
%token<token> BAR
%token<token> SOME
%token<token> NONE
%token<token> MINUS_GREATER

%right prec_let
%right SEMICOLON
%right prec_if
%right prec_match
%right LESS_MINUS
%left COMMA
%left BAR_BAR
%left AND_AND
%left EQUAL LESS_GREATER LESS GREATER LESS_EQUAL GREATER_EQUAL
%left PLUS MINUS PLUS_DOT MINUS_DOT
%left STAR SLASH STAR_DOT SLASH_DOT PERCENT
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
%type<token> match_arm_start
%type<decl> match_ident
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
	| exp PERCENT exp
		{ $$ = &ast.Mod{$1, $3} }
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
	| MATCH exp match_arm_start SOME match_ident MINUS_GREATER exp BAR NONE MINUS_GREATER exp
		%prec prec_match
		{
			none := $11
			$$ = &ast.Match{$1, $2, $7, none, $5, none.Pos()}
		}
	| MATCH exp match_arm_start NONE MINUS_GREATER exp BAR SOME match_ident MINUS_GREATER exp
		%prec prec_match
		{
			some := $11
			$$ = &ast.Match{$1, $2, some, $6, $9, some.Pos()}
		}
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
		{ $$ = &ast.Apply{$1, $2} }
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
	| ARRAY_LENGTH parenless_exp
		%prec prec_app
		{ $$ = &ast.ArraySize{$1, $2} }
	| SOME parenless_exp
		{ $$ = &ast.Some{$1, $2} }
	| ILLEGAL error
		{
			yylex.Error(fmt.Sprintf("Parsing illegal token: %s", $1.String()))
			$$ = nil
		}

fundef:
	IDENT params EQUAL exp
		{ $$ = &ast.FuncDef{ast.NewSymbol($1.Value()), $2, $4} }

params:
	params IDENT
		{ $$ = append($1, ast.NewSymbol($2.Value())) }
	| IDENT
		{ $$ = []*ast.Symbol{ast.NewSymbol($1.Value())} }

args:
	args parenless_exp
		{ $$ = append($1, $2) }
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
			i, err := strconv.ParseInt($1.Value(), 10, 64)
			if err != nil {
				yylex.Error("Parse error at int literal: " + err.Error())
			} else {
				$$ = &ast.Int{$1, i}
			}
		}
	| FLOAT
		{
			f, err := strconv.ParseFloat($1.Value(), 64)
			if err != nil {
				yylex.Error("Parse error at float literal: " + err.Error())
			} else {
				$$ = &ast.Float{$1, f}
			}
		}
	| STRING_LITERAL
		{
			from := $1.Value()
			s, err := strconv.Unquote(from)
			if err != nil {
				yylex.Error(fmt.Sprintf("Parse error at string literal %s: %s", from, err.Error()))
			} else {
				$$ = &ast.String{$1, s}
			}
		}
	| NONE
		{ $$ = &ast.None{$1} }
	| IDENT
		{ $$ = &ast.VarRef{$1, ast.NewSymbol($1.Value())} }
	| parenless_exp DOT LPAREN exp RPAREN
		{ $$ = &ast.Get{$1, $4} }

match_arm_start:
	WITH BAR | WITH

match_ident:
	LPAREN IDENT RPAREN
		{ $$ = ast.NewSymbol($2.Value()) }
	| IDENT
		{ $$ = ast.NewSymbol($1.Value()) }

%%

var genCount = 0
func genTempId() string {
	genCount += 1
	return fmt.Sprintf("$unused%d", genCount)
}

// vim: noet
