// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"bytes"
	"code.google.com/p/go.tools/go/ssa"
	"code.google.com/p/go.tools/go/types"
	"fmt"
	"io/ioutil"
)

// The Language interface enables multiple target languages for TARDIS Go.
type Language interface {
	DeclareTempVar(ssa.Value) string
	LanguageName() string
	FileTypeSuffix() string // e.g. ".go" ".js" ".hx"
	FileStart(packageName, headerText string) string
	FileEnd() string
	SetPosHash() string
	RunDefers() string
	GoClassStart() string
	GoClassEnd(*ssa.Package) string
	SubFnStart(int, bool) string
	SubFnEnd(int) string
	SubFnCall(int) string
	FuncName(*ssa.Function) string
	FieldAddr(register string, v interface{}, errorInfo string) string
	IndexAddr(register string, v interface{}, errorInfo string) string
	Comment(string) string
	LangName(p, o string) string
	Const(lit ssa.Const, position string) (string, string)
	NamedConst(packageName, objectName string, val ssa.Const, position string) string
	Global(packageName, objectName string, glob ssa.Global, position string, isPublic bool) string
	FuncStart(pName, mName string, fn *ssa.Function, posStr string, isPublic bool, trackPhi bool, canOptMap map[string]bool) string
	RunEnd(fn *ssa.Function) string
	FuncEnd(fn *ssa.Function) string
	BlockStart(block []*ssa.BasicBlock, num int) string
	BlockEnd(block []*ssa.BasicBlock, num int, emitPhi bool) string
	Jump(int) string
	If(v interface{}, trueNext, falseNext int, errorInfo string) string
	PhiStart(register, regTyp, regInit string) string
	PhiEntry(register string, phiVal int, v interface{}, errorInfo string) string
	PhiEnd(defaultValue string) string
	LangType(types.Type, bool, string) string
	Value(v interface{}, errorInfo string) string
	BinOp(register, op string, v1, v2 interface{}, errorInfo string) string
	UnOp(register, op string, v interface{}, commaOK bool, errorInfo string) string
	Store(v1, v2 interface{}, errorInfo string) string
	Send(v1, v2 interface{}, errorInfo string) string
	Ret0() string
	Ret1(v1 interface{}, errorInfo string) string
	RetN(values []*ssa.Value, errorInfo string) string
	RegEq(r string) string
	Call(register string, cc ssa.CallCommon, args []ssa.Value, isBuiltin, isGo, isDefer bool, fnToCall, errorInfo string) string
	Convert(register, langType string, destType types.Type, v interface{}, errorInfo string) string
	MakeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string
	ChangeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string
	ChangeType(register string, regTyp, v interface{}, errorInfo string) string
	Alloc(register string, v interface{}, errorInfo string) string
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
	Panic(v1 interface{}, errorInfo string) string
	TypeStart(*types.Named, string) string
	TypeEnd(*types.Named, string) string
	TypeAssert(Register string, X ssa.Value, AssertedType types.Type, CommaOk bool, errorInfo string) string
	EmitTypeInfo() string
	EmitInvoke(register string, isGo bool, isDefer bool, callCommon interface{}, errorInfo string) string
	PackageOverloaded(pkg string) (overloadPkgGo, overloadPkg string, isOverloaded bool)
	Select(isSelect bool, register string, v interface{}, CommaOK bool, errorInfo string) string
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
}

// LanguageList holds the languages that can be targeted. Hey, I hope we do get up to 10 target languages!!
var LanguageList = make([]LanguageEntry, 0, 10)

// TargetLang holds the language currently being targeted, default is the first on the list, initially haxe.
var TargetLang = 0

// Utility comment emitter function.
func emitComment(cmt string) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].Comment(cmt))
}

// Write out the target language file
// TODO consider writing multiple output files, if this would be better/required for some target languages.
func writeFiles() {
	l := TargetLang
	// TODO move to the correct directory based on a command line argument
	err := ioutil.WriteFile(
		"tardis/Go"+LanguageList[l].FileTypeSuffix(), // Ubuntu requires the first letter of the haxe file to be uppercase
		LanguageList[l].buffer.Bytes(), 0666)
	if err != nil {
		LogError("Unable to write output file, does the 'tardis' output directory exist in this location? (it is not created automatically in this early version of tardisgo as a safety feature)",
			"pogo", err)
	}
}

// MakeId cleans-up Go names to replace characters outside (_,0-9,a-z,A-Z) with a decimal value surrounded by underlines, with special handling of '.' and '*'.
func MakeId(s string) (r string) {
	var b []rune
	b = []rune(s)
	for i := range b {
		if b[i] == '_' || ((b[i] >= 'a') && (b[i] <= 'z')) || ((b[i] >= 'A') && (b[i] <= 'Z')) || ((b[i] >= '0') && (b[i] <= '9')) {
			r += string(b[i])
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
