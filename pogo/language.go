// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"bytes"
	"fmt"
	"unicode"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

// The Language interface enables multiple target languages for TARDIS Go.
type Language interface {
	RegisterName(val ssa.Value) string
	DeclareTempVar(ssa.Value) string
	LanguageName() string
	FileTypeSuffix() string // e.g. ".go" ".js" ".hx"
	FileStart(packageName, headerText string) string
	FileEnd() string
	SetPosHash() string
	RunDefers(usesGr bool) string
	GoClassStart() string
	GoClassEnd(*ssa.Package) string
	SubFnStart(int, bool) string
	SubFnEnd(id int, pos int, mustSplit bool) string
	SubFnCall(int) string
	FuncName(*ssa.Function) string
	FieldAddr(register string, v interface{}, errorInfo string) string
	IndexAddr(register string, v interface{}, errorInfo string) string
	Comment(string) string
	LangName(p, o string) string
	Const(lit ssa.Const, position string) (string, string)
	NamedConst(packageName, objectName string, val ssa.Const, position string) string
	Global(packageName, objectName string, glob ssa.Global, position string, isPublic bool) string
	FuncStart(pName, mName string, fn *ssa.Function, posStr string, isPublic, trackPhi, usesGr bool, canOptMap map[string]bool) string
	RunEnd(fn *ssa.Function) string
	FuncEnd(fn *ssa.Function) string
	BlockStart(block []*ssa.BasicBlock, num int, emitPhi bool) string
	BlockEnd(block []*ssa.BasicBlock, num int, emitPhi bool) string
	Jump(int) string
	If(v interface{}, trueNext, falseNext int, errorInfo string) string
	Phi(register string, phiEntries []int, valEntries []interface{}, defaultValue, errorInfo string) string
	LangType(types.Type, bool, string) string
	Value(v interface{}, errorInfo string) string
	BinOp(register string, regTyp types.Type, op string, v1, v2 interface{}, errorInfo string) string
	UnOp(register string, regTyp types.Type, op string, v interface{}, commaOK bool, errorInfo string) string
	Store(v1, v2 interface{}, errorInfo string) string
	Send(v1, v2 interface{}, errorInfo string) string
	Ret(values []*ssa.Value, errorInfo string) string
	RegEq(r string) string
	Call(register string, cc ssa.CallCommon, args []ssa.Value, isBuiltin, isGo, isDefer, usesGr bool, fnToCall, errorInfo string) string
	Convert(register, langType string, destType types.Type, v interface{}, errorInfo string) string
	MakeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string
	ChangeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string
	ChangeType(register string, regTyp, v interface{}, errorInfo string) string
	Alloc(register string, heap bool, v interface{}, errorInfo string) string
	MakeClosure(register string, v interface{}, errorInfo string) string
	MakeSlice(register string, v interface{}, errorInfo string) string
	MakeChan(register string, v interface{}, errorInfo string) string
	MakeMap(register string, v interface{}, errorInfo string) string
	Slice(register string, x, low, high interface{}, errorInfo string) string
	Index(register string, v1, v2 interface{}, errorInfo string) string
	RangeCheck(x, i interface{}, length int, errorInfo string) string
	Field(register string, v interface{}, fNum int, name, errorInfo string, isFunctionName bool) string
	MapUpdate(Map, Key, Value interface{}, errorInfo string) string
	Lookup(register string, Map, Key interface{}, commaOk bool, errorInfo string) string
	Extract(register string, tuple interface{}, index int, errorInfo string) string
	Range(register string, v interface{}, errorInfo string) string
	Next(register string, v interface{}, isString bool, errorInfo string) string
	Panic(v1 interface{}, errorInfo string, usesGr bool) string
	TypeStart(*types.Named, string) string
	//TypeEnd(*types.Named, string) string
	TypeAssert(Register string, X ssa.Value, AssertedType types.Type, CommaOk bool, errorInfo string) string
	EmitTypeInfo() string
	EmitInvoke(register, path string, isGo, isDefer, usesGr bool, callCommon interface{}, errorInfo string) string
	FunctionOverloaded(pkg, fun string) bool
	Select(isSelect bool, register string, v interface{}, CommaOK bool, errorInfo string) string
	PeepholeOpt(opt, register string, code []ssa.Instruction, errorInfo string) string
	DebugRef(userName string, v interface{}, errorInfo string) string
	CanInline(v interface{}) bool
}

// LanguageEntry holds the static infomation about each of the languages, expect this list to extend as more languages are added.
type LanguageEntry struct {
	Language                           // All of the interface functions.
	buffer                bytes.Buffer // Where the output is collected.
	InstructionLimit      int          // How many instructions in a function before we need to split it up.
	SubFnInstructionLimit int          // When we split up a function, how large can each sub-function be?
	PackageConstVarName   string       // The special constant name to specify a Package/Module name in the target language.
	HeaderConstVarName    string       // The special constant name for a target-specific header.
	Goruntime             string       // The location of the core implementation go runtime code for this target language.
	TestFS                string       // the location of the test zipped file system, if present
	files                 []FileOutput // files to write if no errors in compilation
}

type FileOutput struct {
	filename string
	data     []byte
}

// LanguageList holds the languages that can be targeted.
var LanguageList = make([]LanguageEntry, 0, 1)

// TargetLang holds the language currently being targeted, default is the first on the list, initially haxe.
var TargetLang = 0

// Utility comment emitter function.
func emitComment(cmt string) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].Comment(cmt))
}

// MakeID cleans-up Go names to replace characters outside (_,0-9,a-z,A-Z) with a decimal value surrounded by underlines, with special handling of '.' and '*'.
// It also doubles-up uppercase letters, because file names are made from these names and OSX is case insensitive.
func MakeID(s string) (r string) {
	var b []rune
	b = []rune(s)
	for i := range b {
		if b[i] == '_' || ((b[i] >= 'a') && (b[i] <= 'z')) || ((b[i] >= 'A') && (b[i] <= 'Z')) || ((b[i] >= '0') && (b[i] <= '9')) {
			r += string(b[i])
			if unicode.IsUpper(b[i]) {
				r += string(b[i])
			}
		} else {
			switch b[i] {
			case '.':
				r += "_dot_"
			case '*':
				r += "_star_"
			default:
				r += fmt.Sprintf("_%d_", b[i])
			}
		}
	}
	return r
}

// is there more than one package with this name?
// TODO consider using this function in pogo.emitFunctions()
func isDupPkg(pn string) bool {
	pnCount := 0
	ap := rootProgram.AllPackages()
	for p := range ap {
		if pn == ap[p].Object.Name() {
			pnCount++
		}
	}
	return pnCount > 1
}

// FunctionName returns a unique function path and name.
// TODO refactor this code and everywhere it is called to remove duplication.
func FuncPathName(fn *ssa.Function) (path, name string) {
	rx := fn.Signature.Recv()
	pf := MakeID(rootProgram.Fset.Position(fn.Pos()).String()) //fmt.Sprintf("fn%d", fn.Pos())
	if rx != nil {                                             // it is not the name of a normal function, but that of a method, so append the method description
		pf = rx.Type().String() // NOTE no underlying()
	} else {
		if fn.Pkg != nil {
			pf = fn.Pkg.Object.Path() // was .Name(), but not unique
		}
	}
	return pf, fn.Name()
}
