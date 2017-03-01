package codegen

import (
	"github.com/rhysd/gocaml/token"
	"github.com/rhysd/gocaml/typing"
	"llvm.org/llvm/bindings/go/llvm"
	"path/filepath"
)

type sizeEntry struct {
	allocInBits uint64
	alignInBits uint32
}

type sizeTable struct {
	table       map[typing.Type]sizeEntry
	data        llvm.TargetData
	typeBuilder *typeBuilder
	ptrSize     sizeEntry
}

func newSizeTable(types *typeBuilder, data llvm.TargetData) *sizeTable {
	ptrSize := sizeEntry{
		data.TypeSizeInBits(types.voidPtrT),
		uint32(data.ABITypeAlignment(types.voidPtrT) * 8),
	}
	return &sizeTable{
		map[typing.Type]sizeEntry{},
		data,
		types,
		ptrSize,
	}
}

func (sizes *sizeTable) calcSize(t typing.Type) sizeEntry {
	ty := sizes.typeBuilder.convertGCIL(t)
	if _, ok := t.(*typing.Tuple); ok {
		// Tuple is managed by GC with pointer. What we want is size of actual allocated type, not a pointer.
		ty = ty.ElementType()
	}
	bits := sizes.data.TypeSizeInBits(ty)
	align := sizes.data.ABITypeAlignment(ty)
	s := sizeEntry{uint64(bits), uint32(align * 8)}
	sizes.table[t] = s
	return s
}

func (sizes *sizeTable) sizeOf(ty typing.Type) sizeEntry {
	if s, ok := sizes.table[ty]; ok {
		return s
	}
	return sizes.calcSize(ty)
}

func (sizes *sizeTable) allocInBitsOf(ty typing.Type) uint64 {
	return sizes.sizeOf(ty).allocInBits
}

type debugInfoBuilder struct {
	builder     *llvm.DIBuilder
	file        llvm.Metadata
	scope       llvm.Metadata
	compileUnit llvm.Metadata
	typeBuilder *typeBuilder
	sizes       *sizeTable
	voidPtrInfo llvm.Metadata
}

func newDebugInfoBuilder(module llvm.Module, file *token.Source, tb *typeBuilder, target llvm.TargetData) (*debugInfoBuilder, error) {
	d := &debugInfoBuilder{}
	d.typeBuilder = tb
	d.sizes = newSizeTable(tb, target)
	d.builder = llvm.NewDIBuilder(module)

	filename := file.Name
	directory := ""
	if file.Exists {
		p, err := filepath.Abs(file.Name)
		if err != nil {
			return nil, err
		}
		filename = filepath.Base(p)
		directory = filepath.Dir(p)
	}
	d.file = d.builder.CreateFile(filename, directory)

	d.compileUnit = d.builder.CreateCompileUnit(llvm.DICompileUnit{
		Language: llvm.DwarfLang(0xdead), // DW_LANG_USER (0x8000~0xFFFF)
		File:     filename,
		Dir:      directory,
		Producer: "gocaml",
	})

	d.voidPtrInfo = d.builder.CreatePointerType(llvm.DIPointerType{
		Pointee:     d.builder.CreateBasicType(llvm.DIBasicType{}),
		SizeInBits:  d.sizes.ptrSize.allocInBits,
		AlignInBits: d.sizes.ptrSize.alignInBits,
		Name:        "captures",
	})

	return d, nil
}

func (d *debugInfoBuilder) basicTypeInfo(ty typing.Type, enc llvm.DwarfTypeEncoding) llvm.Metadata {
	return d.builder.CreateBasicType(llvm.DIBasicType{
		Name:       ty.String(),
		SizeInBits: d.sizes.allocInBitsOf(ty),
		Encoding:   enc,
	})
}

func (d *debugInfoBuilder) funcTypeInfo(ty *typing.Fun, isClosure bool) llvm.Metadata {
	length := len(ty.Params) + 1
	if isClosure {
		length++
	}
	params := make([]llvm.Metadata, 0, length)

	// Return type is registered as 0th element of params
	params = append(params, d.typeInfo(ty.Ret))

	if isClosure {
		params = append(params, d.voidPtrInfo)
	}

	for _, p := range ty.Params {
		params = append(params, d.typeInfo(p))
	}

	return d.builder.CreateSubroutineType(llvm.DISubroutineType{d.file, params})
}

func (d *debugInfoBuilder) pointerOf(pointee llvm.Metadata, name string) llvm.Metadata {
	size := d.sizes.ptrSize
	return d.builder.CreatePointerType(llvm.DIPointerType{
		Pointee:     pointee,
		SizeInBits:  size.allocInBits,
		AlignInBits: size.alignInBits,
		Name:        name,
	})
}

func (d *debugInfoBuilder) typeInfo(ty typing.Type) llvm.Metadata {
	switch ty := ty.(type) {
	case *typing.Int:
		return d.basicTypeInfo(ty, llvm.DW_ATE_signed)
	case *typing.Bool:
		return d.basicTypeInfo(ty, llvm.DW_ATE_boolean)
	case *typing.Float:
		return d.basicTypeInfo(ty, llvm.DW_ATE_float)
	case *typing.Unit:
		size := d.sizes.sizeOf(ty)
		return d.builder.CreateStructType(d.compileUnit, llvm.DIStructType{
			Name:        "()",
			File:        d.file,
			SizeInBits:  size.allocInBits,
			AlignInBits: size.alignInBits,
			Elements:    []llvm.Metadata{},
		})
	case *typing.Fun:
		return d.funcTypeInfo(ty, false /*XXX: Is this correct?*/)
	case *typing.Array:
		size := d.sizes.sizeOf(ty)
		elems := []llvm.Metadata{d.pointerOf(d.typeInfo(ty.Elem), ""), d.basicTypeInfo(typing.IntType, llvm.DW_ATE_signed)}
		return d.builder.CreateStructType(d.scope, llvm.DIStructType{
			Name:        ty.String(),
			File:        d.file,
			SizeInBits:  size.allocInBits,
			AlignInBits: size.alignInBits,
			Elements:    elems,
		})
	case *typing.Tuple:
		size := d.sizes.sizeOf(ty)
		elems := make([]llvm.Metadata, 0, len(ty.Elems))
		for _, e := range ty.Elems {
			elems = append(elems, d.typeInfo(e))
		}
		name := ty.String()
		allocated := d.builder.CreateStructType(d.scope, llvm.DIStructType{
			Name:        name,
			File:        d.file,
			SizeInBits:  size.allocInBits,
			AlignInBits: size.alignInBits,
			Elements:    elems,
		})
		return d.pointerOf(allocated, name)
	default:
		panic("cannot handle debug info for type " + ty.String())
	}
}

func (d *debugInfoBuilder) setMainFuncInfo(mainfun llvm.Value, line int) {
	voidInfo := llvm.Metadata{}
	info := d.builder.CreateSubroutineType(llvm.DISubroutineType{d.file, []llvm.Metadata{voidInfo}})
	meta := d.builder.CreateFunction(d.compileUnit, llvm.DIFunction{
		Name:         "main",
		LinkageName:  "main",
		Line:         line,
		Type:         info,
		IsDefinition: true,
	})
	mainfun.SetSubprogram(meta)
	d.scope = meta
}

func (d *debugInfoBuilder) setFuncInfo(funptr llvm.Value, ty *typing.Fun, line int, isClosure bool, isDecl bool) {
	// Note:
	// All functions are at toplevel, so any function will be never nested in others.
	name := funptr.Name()
	meta := d.builder.CreateFunction(d.compileUnit, llvm.DIFunction{
		Name:         name, // XXX: Should show unmangled name (e.g. f$t1 -> f)?
		LinkageName:  name,
		Line:         line,
		Type:         d.funcTypeInfo(ty, isClosure),
		IsDefinition: !isDecl,
	})
	if !isDecl {
		funptr.SetSubprogram(meta)
		d.scope = meta
	}
}

func (d *debugInfoBuilder) setLocation(b llvm.Builder, pos token.Position) {
	scope := d.scope
	if scope.C == nil {
		scope = d.compileUnit
	}
	b.SetCurrentDebugLocation(uint(pos.Line), uint(pos.Column), scope, llvm.Metadata{})
}

func (d *debugInfoBuilder) finalize() {
	d.builder.Finalize()
}

func (d *debugInfoBuilder) dispose() {
	d.builder.Destroy()
}
