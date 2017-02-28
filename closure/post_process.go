package closure

import (
	"github.com/rhysd/gocaml/gcil"
)

// In post process:
//   - CLOSURE_CALL flag is set to each 'app' instruction
type postProcess struct {
	closures           gcil.Closures
	funcs              gcil.Toplevel
	processingFuncName string
	processingFunc     *gcil.Fun
}

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

func (pp *postProcess) processInsn(insn *gcil.Insn) {
	switch val := insn.Val.(type) {
	case *gcil.App:
		if val.Callee == pp.processingFuncName && pp.processingFunc != nil {
			pp.processingFunc.IsRecursive = true
		}
		if val.Kind == gcil.EXTERNAL_CALL {
			break
		}
		if _, ok := pp.closures[val.Callee]; ok {
			val.Kind = gcil.CLOSURE_CALL
			break
		}
		if _, ok := pp.funcs[val.Callee]; ok {
			// Callee register name is a name of function, but not a closure.
			// So it must be known function.
			break
		}
		// It's not an external symbol, closure nor known function. So it must be a function
		// variable. All function variables are closures. So the callee must be a closure.
		val.Kind = gcil.CLOSURE_CALL
	case *gcil.If:
		pp.processBlock(val.Then)
		pp.processBlock(val.Else)
	case *gcil.Fun:
		panic("unreachable")
	}
}

func (pp *postProcess) processBlock(block *gcil.Block) {
	begin, end := block.WholeRange()
	for i := begin; i != end; i = i.Next {
		pp.processInsn(i)
	}
}

func (pp *postProcess) process(n string, f *gcil.Fun, b *gcil.Block) {
	pp.processingFuncName = n
	pp.processingFunc = f
	pp.processBlock(b)
}

func doPostProcess(prog *gcil.Program) {
	pp := &postProcess{
		prog.Closures,
		prog.Toplevel,
		"",
		nil,
	}
	for n, f := range prog.Toplevel {
		pp.process(n, f.Val, f.Val.Body)
	}
	pp.process("", nil, prog.Entry)
}
