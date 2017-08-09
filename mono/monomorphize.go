// Package mono provides a function to monomorphize MIR representation.
package mono

import (
	"fmt"
	"github.com/rhysd/gocaml/mir"
	"github.com/rhysd/gocaml/types"
	"strings"
)

// Monomorphization makes polymorphic instructions into monomorphic ones.
// Polymorphic functions are duplicated for each instantiations. And polymorphic (not function) variables
// remain as-is.
//
// - Before:
//
// ```
// id$t1 = fun x$t2     ; type='a -> 'a
//   BEGIN body (id$t1)
//   $k1 = ref x$t2     ; type='a
//   END body (id$t1)
// ```
//
// - After:
//
// ```
// id$t1$int = fun x$t2$int     ; type=int -> int
//   BEGIN body (id$t1$int)
//   $k1 = ref x$t2$int     ; type=int
//   END body (id$t1$int)
//
// id$t1$bool = fun x$t2$bool     ; type=bool -> bool
//   BEGIN body (id$t1$bool)
//   $k1 = ref x$t2$bool     ; type=bool
//   END body (id$t1$bool)
//
// id$t1$unit = fun x$t2$unit     ; type=unit -> unit
//   BEGIN body (id$t1$unit)
//   $k1 = ref x$t2$unit     ; type=unit
//   END body (id$t1$unit)
// ```
//
// How polymorphic functions were instantiated is recorded as env.PolyVariants in *types.Env.
//
// Monomorphization is performed in following order:
// 1. Monomorphize the expression of entry point of program
// 2. Monomorphize non-closure functions. Each function is duplicated for each instantiations. For
//    non-polymorphic functions, they don't need to be duplicated.
// 3. Closure functions are monomorphized at makecls instruction because captures may contain polymorphic
//    type values.
//
// The reason of 3. is that captures of a closure function is passed as a hidden parameter of function.
// Monomorphization must consider the case where the hidden parameter is polymorphic.
// For example of 3., The function `f` is a polymorphic function 'a -> (). And `g` is also polymorphic
// 'b -> 'a * 'b. In this case `g` captures `x`, which is typed as 'a. So `g` need to consider the captured
// type 'a on monomorphization. `g` should be duplicated for each `x` types (int and bool). It means that
//
// ```
// let rec f x =
//   let rec g y = x, y in
//   g 3.14; g "foo"; ()
// in
// f 10; f true
// ```
//
// `g` should be duplicated into 4 functions `g$int$float`, `g$int$string`, `g$bool$float` and `g$bool$string`.
//
// ```
// let rec f$int x$int =
//   let rec g$int$float y$float = x$int, y$float in
//   let rec g$int$string y$string = x$int, y$string in
//   g$int$float 3.14; g$int$string "foo"; ()
// in
// let rec f$bool x$bool =
//   let rec g$bool$float y$float = x$bool, y$float in
//   let rec g$bool$string y$string = x$bool, y$string in
//   g$bool$float 3.14; g$bool$string "foo"; ()
// in
// f$int 10; f$bool true
// ```
//
// However, all the type variables are not instantiated in function definitions. There are some cases
// where variables are polymorphic; Type is undetermined at variable definition at `let` expression.
// This case can be split into 2 cases:
//
// 1. not function
// 2. function
//
// For 1., we leave the type variable as-is because it is not important which type it is actually
// instantiated.
//
// ```
// let o = None in Some 42 = o; Some true = o; ()
// ```
//
// This program is valid. `o` in `Some 42 = o` is instantiated as `int option` and `Some true = o`
// is instantiated as `bool option`. However, `o` in `let` expression is `'a option`. But this is
// not important because it's `None`. Although this is a case of option type, other types are the same.
//
// For 2., actual function pointer is determined at the function call. So we create a table of instantiations
// and set the pointer to the table instead of the function pointer when it is polymorphic function
// variable.
//
// ```
// let rec id x = id in
// let f = id in
// f 42;
// f ()
// ```
//
// In above example, we create a list of instantiations of `id`; `let id_ptr = [id$int; id$unit]`.
// Then the pointer to the table is embedded in closure object instead of function pointer because
// `id` is a polymorphic function. Function variable `f` can be assigned correctly.
// At invocation of `f`, we have list of instantiations statically so it's possible to get the instantiated
// function pointer from the list in the closure object by index. Finally we can call it with captures.
//
// All instructions in polymorphic functions' bodies are duplicated. We need to consider the duplications
// don't break alpha transformation. We introduce another ID counter to solve this.
// All created instructions due to monomorphization will have a new name with counter. If the counter
// value is `42`, the instruction monomorphized from `foo$t3` will be named as `foo$t3$42`.

type typeVarAssignment map[types.VarID]types.Type

func (assign typeVarAssignment) assignToTypes(ts []types.Type) ([]types.Type, bool) {
	assigned := make([]types.Type, 0, len(ts))
	changed := false
	for _, t := range ts {
		t, c := assign.assign(t)
		changed = changed || c
		assigned = append(assigned, t)
	}
	return assigned, changed
}

func (assign typeVarAssignment) assignToVar(v *types.Var) (types.Type, bool) {
	if v.Ref != nil {
		t, _ := assign.assign(v.Ref)
		return t, false
	}
	t, ok := assign[v.ID]
	if !ok {
		// Unknown type variable means that it's instantiated later. It can be ignored.
		// In following code, type of o is 'a option and the 'a is the case.
		// e.g.
		//   let o = None in Some 10 = o
		return v, false
	}
	ret, _ := assign.assign(t)
	return ret, true
}

func (assign typeVarAssignment) assign(t types.Type) (types.Type, bool) {
	switch t := t.(type) {
	case *types.Fun:
		ret, changed := assign.assign(t.Ret)
		params, changed2 := assign.assignToTypes(t.Params)
		if changed || changed2 {
			return &types.Fun{ret, params}, true
		}
	case *types.Tuple:
		elems, changed := assign.assignToTypes(t.Elems)
		if changed {
			return &types.Tuple{elems}, true
		}
	case *types.Array:
		elem, changed := assign.assign(t.Elem)
		if changed {
			return &types.Array{elem}, true
		}
	case *types.Option:
		elem, changed := assign.assign(t.Elem)
		if changed {
			return &types.Option{elem}, true
		}
	case *types.Var:
		return assign.assignToVar(t)
	}
	return t, false
}

func (assign typeVarAssignment) applyTo(t types.Type) (copied types.Type) {
	copied, _ = assign.assign(t)
	return
}

func (assign typeVarAssignment) dump() {
	fmt.Println("Type assignments")
	for id, t := range assign {
		fmt.Printf("  '%d => %s\n", id, types.Debug(t))
	}
}

func mangleType(t types.Type) string {
	// TODO: More efficient type mangling
	return t.String()
}

type codeDup struct {
	*monomorphizer
	replacedIdents map[string]string
	typeVarAssign  typeVarAssignment
}

func (dup *codeDup) mangleFun(name string, inst *types.Instantiation) string {
	ss := append(make([]string, 0, len(inst.Mapping)+1), name)
	for _, m := range inst.Mapping {
		ss = append(ss, mangleType(dup.typeVarAssign.applyTo(m.Type)))
	}
	return strings.Join(ss, "$")
}

// Make new ID from existing ID to identify duplicated instructions from original ones.
// This is needed to avoid breaking alpha-transformed identifiers.
func (dup *codeDup) newIdent(from string) string {
	ident := fmt.Sprintf("%s$%d", from, dup.genID())
	dup.replacedIdents[from] = ident
	return ident
}

func (dup *codeDup) resolveIdent(ident string) string {
	// Consider the ident may be replaced by monomorphization recursively.
	for {
		replaced, ok := dup.replacedIdents[ident]
		if !ok {
			break
		}
		ident = replaced
	}
	return ident
}

func (dup *codeDup) resolveRef(orig string, insnIdent string) string {
	resolved := dup.resolveIdent(orig)

	inst, ok := dup.env.RefInsts[insnIdent]
	if !ok {
		return resolved
	}

	// Instantiations at variable references were recorded with identifiers before being replaced.
	// So we need to get them with original identifiers.
	resolved = dup.mangleFun(resolved, inst)

	return resolved
}

func (dup *codeDup) resolveIdents(is []string) []string {
	ret := make([]string, 0, len(is))
	for _, i := range is {
		ret = append(ret, dup.resolveIdent(i))
	}
	return ret
}

func (dup *codeDup) dupInsn(from *mir.Insn) *mir.Insn {
	// Monomorphize type of the instruction
	ident := from.Ident
	ty, _ := dup.env.DeclTable[from.Ident]
	if ty, assigned := dup.typeVarAssign.assign(ty); assigned {
		// Only when the type is generic type and some type variable was assigned by
		// monomorphization, the type should be updated with new identifier name.
		ident = dup.newIdent(from.Ident)
		dup.env.DeclTable[ident] = ty
	}

	// Give new ident not to break alpha transformation.
	to := &mir.Insn{
		Ident: ident,
		Pos:   from.Pos,
	}

	switch val := from.Val.(type) {
	case *mir.Unit, *mir.Bool, *mir.Int, *mir.Float, *mir.String, *mir.None, *mir.XRef:
		// Don't need to duplicate instruction because they don't refer any idents
		to.Val = val
	case *mir.Unary:
		to.Val = &mir.Unary{val.Op, dup.resolveIdent(val.Child)}
	case *mir.Binary:
		to.Val = &mir.Binary{val.Op, dup.resolveIdent(val.LHS), dup.resolveIdent(val.RHS)}
	case *mir.Ref:
		// Instantiation occurs at variable reference
		to.Val = &mir.Ref{dup.resolveRef(val.Ident, from.Ident)}
	case *mir.If:
		to.Val = &mir.If{
			dup.resolveIdent(val.Cond),
			dup.dupBlock(val.Then),
			dup.dupBlock(val.Else),
		}
	case *mir.App:
		// Callee of 'app' instruction is function name itself instead of register name
		// when it is known function call. In the case, we need to resolve instantiation of
		// callee (if it's polymorphic function).
		var callee string
		switch val.Kind {
		case mir.DIRECT_CALL:
			callee = dup.resolveRef(val.Callee, from.Ident)
		case mir.CLOSURE_CALL:
			callee = dup.resolveIdent(val.Callee)
		case mir.EXTERNAL_CALL:
			callee = val.Callee
		}

		to.Val = &mir.App{
			Kind:   val.Kind,
			Callee: callee,
			Args:   dup.resolveIdents(val.Args),
		}
	case *mir.Tuple:
		to.Val = &mir.Tuple{dup.resolveIdents(val.Elems)}
	case *mir.TplLoad:
		to.Val = &mir.TplLoad{dup.resolveIdent(val.From), val.Index}
	case *mir.Array:
		to.Val = &mir.Array{
			dup.resolveIdent(val.Size),
			dup.resolveIdent(val.Elem),
		}
	case *mir.ArrLit:
		to.Val = &mir.ArrLit{dup.resolveIdents(val.Elems)}
	case *mir.ArrLoad:
		to.Val = &mir.ArrLoad{dup.resolveIdent(val.From), dup.resolveIdent(val.Index)}
	case *mir.ArrStore:
		to.Val = &mir.ArrStore{
			dup.resolveIdent(val.To),
			dup.resolveIdent(val.Index),
			dup.resolveIdent(val.RHS),
		}
	case *mir.ArrLen:
		to.Val = &mir.ArrLen{dup.resolveIdent(val.Array)}
	case *mir.Some:
		to.Val = &mir.Some{dup.resolveIdent(val.Elem)}
	case *mir.IsSome:
		to.Val = &mir.IsSome{dup.resolveIdent(val.OptVal)}
	case *mir.DerefSome:
		to.Val = &mir.DerefSome{dup.resolveIdent(val.SomeVal)}
	case *mir.MakeCls:
		fun := dup.dupClosure(val.Fun, val.Vars)
		caps, _ := dup.toProg.Closures[fun.Name]
		to.Val = &mir.MakeCls{caps, fun.Name}
	}

	return to
}

func (dup *codeDup) dupBlock(from *mir.Block) *mir.Block {
	to := mir.NewEmptyBlock(from.Name)
	for insn, end := from.WholeRange(); insn != end; insn = insn.Next {
		to.Append(dup.dupInsn(insn))
	}
	return to
}

func (dup *codeDup) dupFun(fun mir.FunInsn, inst *types.Instantiation) mir.FunInsn {
	// If the instantiated function is already monomorphized, just return it
	if insts, ok := dup.funInsts[fun.Name]; ok {
		for _, i := range insts {
			if i.inst == inst {
				return i.insn
			}
		}
	}

	funName := dup.mangleFun(fun.Name, inst)
	for _, m := range inst.Mapping {
		dup.typeVarAssign[m.ID] = dup.typeVarAssign.applyTo(m.Type)
	}

	val := &mir.Fun{
		Params:      make([]string, 0, len(fun.Val.Params)),
		IsRecursive: fun.Val.IsRecursive,
	}

	for _, param := range fun.Val.Params {
		p := dup.newIdent(param)
		dup.env.DeclTable[p] = dup.typeVarAssign.applyTo(dup.env.DeclTable[param])
		val.Params = append(val.Params, p)
	}

	t := dup.typeVarAssign.applyTo(dup.env.DeclTable[fun.Name])
	dup.env.DeclTable[funName] = t

	val.Body = dup.dupBlock(fun.Val.Body)

	insn := mir.FunInsn{
		Name: funName,
		Val:  val,
		Pos:  fun.Pos,
	}

	dup.funInsts[fun.Name] = append(dup.funInsts[fun.Name], funInst{inst, insn})
	dup.toProg.Toplevel[funName] = insn
	return insn
}

func (dup *codeDup) dupClosure(name string, captures []string) mir.FunInsn {
	fty, _ := dup.env.DeclTable[name]
	insts, isPolyFun := dup.env.PolyTypes[fty]

	hasPolyCapture := false
	monoCaps := make([]string, 0, len(captures))
	for _, cap := range captures {
		replaced := dup.resolveIdent(cap)
		if replaced != cap {
			// Some capture for the closure instance was updated by replacement of
			// identifier in monomorphization.
			hasPolyCapture = true
		}
		monoCaps = append(monoCaps, replaced)
	}

	if !isPolyFun && !hasPolyCapture {
		// The closure function is monomorphic. It doesn't need to be monomorphized.
		// Simply visit it's body and do shallow copy like as visitFun().
		if f, ok := dup.toProg.Toplevel[name]; ok {
			return f
		}
		f, _ := dup.toplevel[name]
		dup.visitBlock(f.Val.Body)
		dup.toProg.Toplevel[name] = f
		dup.toProg.Closures[name] = captures
		return f
	}

	ident := dup.newIdent(name)
	from, _ := dup.toplevel[name]

	insn := mir.FunInsn{
		Name: ident,
		Val:  from.Val,
		Pos:  from.Pos,
	}

	if !isPolyFun {
		// When some capture has new identifier name by monomorphization, it means that the capture
		// is polymorphic. To resolve polymorphic captures, we need to duplicate the closure even if
		// the closure itself is not polymorphic. This is because captures are passed as hidden parameter
		// to the closure function.

		val := &mir.Fun{
			Params:      make([]string, 0, len(from.Val.Params)),
			IsRecursive: from.Val.IsRecursive,
			Body:        from.Val.Body,
		}

		for _, param := range from.Val.Params {
			p := dup.newIdent(param)
			dup.env.DeclTable[p] = dup.typeVarAssign.applyTo(dup.env.DeclTable[param])
			val.Params = append(val.Params, p)
		}

		dup.env.DeclTable[ident] = fty
		val.Body = dup.dupBlock(val.Body)

		insn.Val = val
		dup.toProg.Toplevel[ident] = insn
		dup.toProg.Closures[ident] = monoCaps
		return insn
	}

	// Some capture is polymorphic and closure function is also polymorphic.
	mappings := make(map[string][]*types.VarMapping, len(insts))
	for _, inst := range insts {
		mapping := make([]*types.VarMapping, 0, len(inst.Mapping))
		for _, m := range inst.Mapping {
			t := dup.typeVarAssign.applyTo(m.Type)
			dup.typeVarAssign[m.ID] = t
			mapping = append(mapping, &types.VarMapping{m.ID, t})
		}
		monoInst := &types.Instantiation{
			inst.From,
			dup.typeVarAssign.applyTo(inst.To),
			mapping,
		}
		monoFun := dup.dupFun(insn, monoInst)
		dup.toProg.Closures[monoFun.Name] = monoCaps
		mappings[monoFun.Name] = mapping
	}

	// TODO: Store mappings

	// TODO: Create polymorphic closure table? (e.g. having 'a -> 'a, int -> int, bool -> bool, ...)

	return insn
}

type funInst struct {
	inst *types.Instantiation
	insn mir.FunInsn
}

type monomorphizer struct {
	env      *types.Env
	closures mir.Closures
	toplevel map[string]mir.FunInsn
	toProg   *mir.Program
	ID       uint
	funInsts map[string][]funInst
}

func newMonomorphizer(from *mir.Program, env *types.Env) *monomorphizer {
	return &monomorphizer{
		env,
		from.Closures,
		from.Toplevel,
		&mir.Program{mir.Toplevel{}, mir.Closures{}, nil},
		0,
		make(map[string][]funInst, 3),
	}
}

func (mono *monomorphizer) newCodeDup() *codeDup {
	return &codeDup{mono, make(map[string]string, 10), typeVarAssignment{}}
}

func (mono *monomorphizer) genID() uint {
	mono.ID++
	return mono.ID
}

func (mono *monomorphizer) visitInsn(from *mir.Insn) {
	switch val := from.Val.(type) {
	case *mir.MakeCls:
		mono.newCodeDup().dupClosure(val.Fun, val.Vars)
	case *mir.App:
		panic("TODO:: Consider callee is generic and monomorphized")
	case *mir.If:
		mono.visitBlock(val.Then)
		mono.visitBlock(val.Else)
	}
}

func (mono *monomorphizer) visitBlock(from *mir.Block) {
	begin, end := from.WholeRange()
	for i := begin; i != end; i = i.Next {
		mono.visitInsn(i)
	}
}

func (mono *monomorphizer) visitFun(fun mir.FunInsn) {
	ty, _ := mono.env.DeclTable[fun.Name]
	insts, isPoly := mono.env.PolyTypes[ty]

	// When monomorphic functioin, simply visit functioin
	if !isPoly {
		// Note: Don't need to check the function was already visited because this function
		// is called once per each function.
		mono.visitBlock(fun.Val.Body)
		mono.toProg.Toplevel[fun.Name] = fun
		return
	}

	// When polymorphic function, monomorphize it with each instantiation
	for _, inst := range insts {
		mono.newCodeDup().dupFun(fun, inst)
	}
}

func Monomorphize(prog *mir.Program, env *types.Env) *mir.Program {
	mono := newMonomorphizer(prog, env)

	for _, fun := range prog.Toplevel {
		if _, isClosure := prog.Closures[fun.Name]; !isClosure {
			mono.visitFun(fun)
		}
	}

	mono.visitBlock(prog.Entry)
	mono.toProg.Entry = prog.Entry

	return mono.toProg
}
