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
	decls []ast.Decl
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
%token<token> ARRAY
%token<token> CREATE
%token<token> DOT
%token<token> LESS_MINUS
%token<token> SEMICOLON

%right prec_let
%right SEMICOLON
%right prec_if
%right LESS_MINUS
%left COMMA
%left EQUAL LESS_GREATER LESS GREATER LESS_EQUAL GREATER_EQUAL
%left PLUS MINUS PLUS_DOT MINUS_DOT
%left AST_DOT SLASH_DOT
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
	| exp EQUAL exp
		{ $$ =&ast.Eq{$1, $3} }
	| exp LESS_GREATER exp
		{ $$ = &ast.Not{$2, &ast.Eq{$1, $3}} } /* XXX: Position is not accurate */
	| exp LESS exp
		{ $$ = &ast.Not{$2, &ast.Less{$3, $1}} } /* XXX: Position is not accurate */
	| exp GREATER exp
		{ $$ = &ast.Not{$2, &ast.Less{$1, $3}} } /* XXX: Position is not accurate */
	| exp LESS_EQUAL exp
		{ $$ = &ast.Less{$1, $3} }
	| exp GREATER_EQUAL exp
		{ $$ = &ast.Less{$3, $1} } /* XXX: Position is not accurate */
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
	| exp AST_DOT exp
		{ $$ = &ast.FMul{$1, $3} }
	| exp SLASH_DOT exp
		{ $$ = &ast.FDiv{$1, $3} }
	| LET IDENT EQUAL exp IN exp
		%prec prec_let
		{ $$ = &ast.Let{$1, ast.Decl{$2.Value()}, $4, $6} }
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
		{ $$ = &ast.Let{$2, ast.Decl{genTempId()}, $1, $3} }
	| ARRAY DOT CREATE parenless_exp parenless_exp
		%prec prec_app
		{ $$ = &ast.Array{$1, $4, $5} }
	| ILLEGAL error
		{
			yylex.Error(fmt.Sprintf("Parsing illegal token: %s", $1.String()))
			$$ = nil
		}


fundef:
	IDENT params EQUAL exp
		{ $$ = &ast.FuncDef{ast.Decl{$1.Value()}, $2, $4} }

params:
	IDENT params
		{ $$ = append($2, ast.Decl{$1.Value()}) }
	| IDENT
		{ $$ = []ast.Decl{{$1.Value()}} }

args:
	args parenless_exp
		%prec prec_app
		{ $$ = append($1, $2) }
	| parenless_exp
		%prec prec_app
		{ $$ = []ast.Expr{$1} }

elems:
	elems COMMA exp
		{ $$ = append($1, $3) }
	| exp COMMA exp
		{ $$ = []ast.Expr{$1, $3} }

pat:
	pat COMMA IDENT
		{ $$ = append($1, ast.Decl{$3.Value()}) }
	| IDENT COMMA IDENT
		{ $$ = []ast.Decl{{$1.Value()}, {$3.Value()}} }

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
	| IDENT
		{ $$ = &ast.Var{$1, $1.Value()} }
	| parenless_exp DOT LPAREN exp RPAREN
		{ $$ = &ast.Get{$1, $4} }

%%

var genCount = 0
func genTempId() string {
	genCount += 1
	return fmt.Sprintf("$tmp%d", genCount)
}

// vim: noet
