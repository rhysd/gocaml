package closure

import (
	"github.com/rhysd/gocaml/mir"
)

// As post process of closure transform, CLOSURE_CALL flag is set to each 'app' instruction
type appFixer struct {
	closures       mir.Closures
	funcs          mir.Toplevel
	fixingFuncName string
	fixingFunc     *mir.Fun
}

// TODO:
// Reveal how each adhoc polymorphic types should be instantiated.
// e.g.
//   let o = None in o = (o = Some 10) || (o = Some true) in ...
// In expression `None`, it's type is 'a option and 'a should be instantiated as int and bool.
// In this process, collect the actual instantiated types for adhoc polymorphic type variables
// (in above example, they're int and bool).
// It also corrects captured values. In above case, when capturing `o`, the type of capture
// value is 'a option. However, it actually needs to capture two values typed as int option and
// bool option because we introduce code duplication to generate code for polymorphic type
// expressions.

// TODO:
// Rearrange basic blocks to represents actual DAG.
// All blocks should be flattened in a function.
//
// e.g.
//
// From:
//   block {
//       // entry block
//       if
//       then {
//           // then block
//       }
//       else {
//           // else block
//       }
//       insns...
//   }
//
// To:
//   block {
//       // entry block
//       if
//   }
//   then {
//       // then block
//   }
//   else {
//       // else block
//   }
//   precede {
//       // rest block
//       insns...
//   }

func (fix *appFixer) fixApp(insn *mir.Insn) {
	switch val := insn.Val.(type) {
	case *mir.App:
		if val.Callee == fix.fixingFuncName && fix.fixingFunc != nil {
			fix.fixingFunc.IsRecursive = true
		}
		if val.Kind == mir.EXTERNAL_CALL {
			break
		}
		if _, ok := fix.closures[val.Callee]; ok {
			val.Kind = mir.CLOSURE_CALL
			break
		}
		if _, ok := fix.funcs[val.Callee]; ok {
			// Callee register name is a name of function, but not a closure.
			// So it must be known function.
			break
		}
		// It's not an external symbol, closure nor known function. So it must be a function
		// variable. All function variables are closures. So the callee must be a closure.
		val.Kind = mir.CLOSURE_CALL
	case *mir.If:
		fix.fixAppsInBlock(val.Then)
		fix.fixAppsInBlock(val.Else)
	case *mir.Fun:
		panic("unreachable")
	}
}

func (fix *appFixer) fixAppsInBlock(block *mir.Block) {
	begin, end := block.WholeRange()
	for i := begin; i != end; i = i.Next {
		fix.fixApp(i)
	}
}

func (fix *appFixer) fixAppsInFun(n string, f *mir.Fun, b *mir.Block) {
	fix.fixingFuncName = n
	fix.fixingFunc = f
	fix.fixAppsInBlock(b)
}

func fixAppsInProg(prog *mir.Program) {
	pp := &appFixer{
		prog.Closures,
		prog.Toplevel,
		"",
		nil,
	}
	for n, f := range prog.Toplevel {
		pp.fixAppsInFun(n, f.Val, f.Val.Body)
	}
	pp.fixAppsInFun("", nil, prog.Entry)
}
