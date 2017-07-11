/*
  This parser definition is based on min-caml/parser.mly
  Copyright (c) 2005-2008, Eijiro Sumii, Moe Masuko, and Kenichi Asai
*/

%{
package syntax

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
	params []ast.Param
	program *ast.AST
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
%token<token> FUN
%token<token> COLON
%token<token> TYPE
%token<token> LBRACKET_BAR
%token<token> BAR_RBRACKET
%token<token> LBRACKET
%token<token> RBRACKET
%token<token> EXTERNAL

%nonassoc IN
%right prec_let
%right prec_seq
%right SEMICOLON
%nonassoc WITH
%right prec_if
%right prec_match
%right prec_fun
%right LESS_MINUS
%nonassoc BAR
%left prec_tuple
%left COMMA
%left BAR_BAR
%left AND_AND
%left EQUAL LESS_GREATER LESS GREATER LESS_EQUAL GREATER_EQUAL
%left PLUS MINUS PLUS_DOT MINUS_DOT
%left STAR SLASH STAR_DOT SLASH_DOT PERCENT
%right prec_unary_minus
%left prec_app
%left DOT
%nonassoc prec_below_ident
%nonassoc IDENT

%type<node> exp
%type<node> simple_exp
%type<node> seq_exp
%type<nodes> elems
%type<nodes> args
%type<params> params
%type<decls> pat
%type<funcdef> fundef
%type<token> match_arm_start
%type<decl> match_ident
%type<nodes> semi_elems
%type<node> type_annotation
%type<node> simple_type_annotation
%type<node> type
%type<node> simple_type
%type<node> simple_type_or_tuple
%type<nodes> arrow_types
%type<nodes> simple_type_star_list
%type<nodes> type_comma_list
%type<program> toplevels
%type<> opt_semi
%type<> program

%start program

%%

program:
	toplevels seq_exp
		{
			tree := $1
			tree.Root = $2
			yylex.(*pseudoLexer).result = tree
		}

toplevels:
	/* empty */
		{ $$ = &ast.AST{} }
	| toplevels TYPE IDENT EQUAL type SEMICOLON
		{
			decl := &ast.TypeDecl{$2, ast.NewSymbol($3.Value()), $5}
			tree := $1
			tree.TypeDecls = append(tree.TypeDecls, decl)
			$$ = tree
		}
	| toplevels EXTERNAL IDENT COLON type EQUAL STRING_LITERAL SEMICOLON
		{
			from := $7.Value()
			lit, err := strconv.Unquote(from)
			if err != nil {
				yylex.Error(fmt.Sprintf("Parse error at string literal in 'external' decl: %s: %s", from, err.Error()))
			} else {
				tree := $1
				ext := &ast.External{$2, $7, sym($3), $5, lit}
				tree.Externals = append(tree.Externals, ext)
				$$ = tree
			}
		}

seq_exp:
	exp %prec prec_seq
		{ $$ = $1 }
	| seq_exp SEMICOLON exp
		{ $$ = &ast.Let{$2, ast.IgnoredSymbol(), $1, $3, nil} }

exp:
	simple_exp
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
	| IF seq_exp THEN seq_exp ELSE exp
		%prec prec_if
		{ $$ = &ast.If{$1, $2, $4, $6} }
	| MATCH seq_exp match_arm_start SOME match_ident MINUS_GREATER seq_exp BAR NONE MINUS_GREATER exp
		%prec prec_match
		{
			none := $11
			$$ = &ast.Match{$1, $2, $7, none, $5, none.Pos()}
		}
	| MATCH seq_exp match_arm_start NONE MINUS_GREATER seq_exp BAR SOME match_ident MINUS_GREATER exp
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
	| LET IDENT type_annotation EQUAL seq_exp IN seq_exp
		%prec prec_let
		{ $$ = &ast.Let{$1, sym($2), $5, $7, $3} }
	| LET REC fundef IN seq_exp
		%prec prec_let
		{ $$ = &ast.LetRec{$1, $3, $5} }
	| simple_exp args
		%prec prec_app
		{ $$ = &ast.Apply{$1, $2} }
	| elems
		%prec prec_tuple
		{ $$ = &ast.Tuple{$1} }
	| LET LPAREN pat RPAREN type_annotation EQUAL seq_exp IN seq_exp
		{ $$ = &ast.LetTuple{$1, $3, $7, $9, $5} }
	| simple_exp DOT LPAREN exp RPAREN LESS_MINUS exp
		{ $$ = &ast.ArrayPut{$1, $4, $7} }
	| ARRAY_MAKE simple_exp simple_exp
		%prec prec_app
		{ $$ = &ast.ArrayMake{$1, $2, $3} }
	| ARRAY_LENGTH simple_exp
		%prec prec_app
		{ $$ = &ast.ArraySize{$1, $2} }
	| SOME simple_exp
		{ $$ = &ast.Some{$1, $2} }
	| FUN params simple_type_annotation MINUS_GREATER seq_exp
		%prec prec_fun
		{
			t := $1
			ident := ast.NewSymbol(fmt.Sprintf("lambda.line%d.col%d", t.Start.Line, t.Start.Column))
			def := &ast.FuncDef{ident, $2, $5, $3}
			ref := &ast.VarRef{$1, ident}
			$$ = &ast.LetRec{$1, def, ref}
		}
	| ILLEGAL error
		{
			yylex.Error("Parsing illegal token: " + $1.String())
			$$ = nil
		}

fundef:
	IDENT params type_annotation EQUAL seq_exp
		{ $$ = &ast.FuncDef{ast.NewSymbol($1.Value()), $2, $5, $3} }

params:
	IDENT
		{ $$ = []ast.Param{{sym($1), nil}} }
	| LPAREN IDENT COLON type RPAREN
		{ $$ = []ast.Param{{sym($2), $4}} }
	| params IDENT
		{ $$ = append($1, ast.Param{sym($2), nil}) }
	| params LPAREN IDENT COLON type RPAREN
		{ $$ = append($1, ast.Param{sym($3), $5}) }

args:
	args simple_exp
		{ $$ = append($1, $2) }
	| simple_exp
		{ $$ = []ast.Expr{$1} }

elems:
	elems COMMA exp
		{ $$ = append($1, $3) }
	| exp COMMA exp
		{ $$ = []ast.Expr{$1, $3} }

pat:
	pat COMMA IDENT
		{ $$ = append($1, sym($3)) }
	| IDENT COMMA IDENT
		{ $$ = []*ast.Symbol{sym($1), sym($3)} }

simple_exp:
	LPAREN seq_exp type_annotation RPAREN
		{
			t := $3
			if t == nil {
				$$ = $2
			} else {
				$$ = &ast.Typed{$2, $3}
			}
		}
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
	| LBRACKET_BAR BAR_RBRACKET
		{ $$ = &ast.ArrayLit{$1, $2, nil} }
	| LBRACKET_BAR semi_elems opt_semi BAR_RBRACKET
		{ $$ = &ast.ArrayLit{$1, $4, $2} }
	| LBRACKET RBRACKET error
		{ yylex.Error("List literal is not implemented yet. Please use array literal [| e1; e2; ... |] instead") }
	| LBRACKET semi_elems opt_semi RBRACKET error
		{ yylex.Error("List literal is not implemented yet. Please use array literal [| e1; e2; ... |] instead") }
	| NONE
		{ $$ = &ast.None{$1} }
	| IDENT
		{ $$ = &ast.VarRef{$1, ast.NewSymbol($1.Value())} }
	| simple_exp DOT LPAREN exp RPAREN
		{ $$ = &ast.ArrayGet{$1, $4} }

match_arm_start:
	WITH BAR | WITH

match_ident:
	LPAREN IDENT RPAREN
		{ $$ = ast.NewSymbol($2.Value()) }
	| IDENT
		{ $$ = ast.NewSymbol($1.Value()) }

semi_elems:
	exp %prec prec_seq
		{ $$ = []ast.Expr{$1} }
	| semi_elems SEMICOLON exp
		{ $$ = append($1, $3) }

opt_semi:
	/* empty */ {} | SEMICOLON {}
type_annotation:
		{ $$ = nil }
	| COLON type
		{ $$ = $2 }

simple_type_annotation:
		{ $$ = nil }
	| COLON simple_type
		{ $$ = $2 }

type:
	simple_type_or_tuple
		{ $$ = $1 }
	| simple_type_or_tuple MINUS_GREATER arrow_types
		{
			ts := $3
			i := len(ts)-1
			ret := ts[i]
			params := append([]ast.Expr{$1}, ts[:i]...)
			$$ = &ast.FuncType{params, ret}
		}

arrow_types:
	simple_type_or_tuple
		{ $$ = []ast.Expr{$1} }
	| arrow_types MINUS_GREATER simple_type_or_tuple
		{ $$ = append($1, $3) }

simple_type_or_tuple:
	simple_type
		{ $$ = $1 }
	| simple_type STAR simple_type_star_list
		{ $$ = &ast.TupleType{append([]ast.Expr{$1}, $3...)} }

simple_type_star_list:
	simple_type
		{ $$ = []ast.Expr{$1} }
	| simple_type_star_list STAR simple_type
		{ $$ = append($1, $3) }

simple_type:
	IDENT
		{
			t := $1
			$$ = &ast.CtorType{nil, t, nil, ast.NewSymbol(t.Value())}
		}
	| simple_type IDENT
		{
			t := $2
			$$ = &ast.CtorType{nil, t, []ast.Expr{$1}, ast.NewSymbol(t.Value())}
		}
	| LPAREN type_comma_list RPAREN IDENT
		{
			t := $4
			$$ = &ast.CtorType{$1, t, $2, ast.NewSymbol(t.Value())}
		}
	| LPAREN type_comma_list RPAREN
		%prec prec_below_ident
		{
			ts := $2
			if len(ts) > 1 {
				yylex.Error("(t1, t2, ...) is not a type. For tuple, use t1 * t2 * ... * tn")
			} else {
				$$ = $2[0]
			}
		}

type_comma_list:
	type
		{ $$ = []ast.Expr{$1} }
	| type_comma_list COMMA type
		{ $$ = append($1, $3) }

%%

func sym(tok *token.Token) *ast.Symbol {
	s := tok.Value()
	if s == "_" {
		return ast.IgnoredSymbol()
	} else {
		return ast.NewSymbol(s)
	}
}

// vim: noet
