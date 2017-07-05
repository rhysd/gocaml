package monomorphize

import (
	"fmt"
	"github.com/rhysd/gocaml/types"
)

type assignTable struct {
	parent *assignTable
	vars   map[types.VarID]types.Type
}

func (tbl *assignTable) assign(id types.VarID, t types.Type) {
	tbl.vars[id] = t
}

func (tbl *assignTable) find(id types.VarID) types.Type {
	if t, ok := tbl.vars[id]; ok {
		return t
	}
	if tbl.parent == nil {
		panic(fmt.Sprint("FATAL: Type variable not found: ", id))
	}
	return tbl.parent.find(id)
}

func (tbl *assignTable) nest() *assignTable {
	return &assignTable{
		tbl,
		map[types.VarID]types.Type{},
	}
}
