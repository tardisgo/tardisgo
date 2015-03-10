// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"strings"
	"unicode"
	"unsafe"

	"github.com/tardisgo/tardisgo/tgossa"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

var fnMap, grMap map[*ssa.Function]bool // which functions are used and if the functions use goroutines/channels

// For every function, maybe emit the code...
func emitFunctions() {
	//fnMap := ssautil.AllFunctions(rootProgram)
	dceList := []*ssa.Package{mainPackage, rootProgram.ImportedPackage(LanguageList[TargetLang].Goruntime)}
	dceExceptions := []string{}
	if LanguageList[TargetLang].TestFS != "" { // need to load file system
		dceExceptions = append(dceExceptions, "syscall") // so that we keep UnzipFS()
	}
	dceExceptions = append(dceExceptions, LibListNoDCE...)
	for _, ex := range dceExceptions {
		exip := rootProgram.ImportedPackage(ex)
		if exip != nil {
			dceList = append(dceList, exip)
		}
	}
	fnMap, grMap = tgossa.VisitedFunctions(rootProgram, dceList, IsOverloaded)
	/*
		fmt.Println("DEBUG funcs not requiring goroutines:")
		for df, db := range grMap {
			if !db {
				fmt.Println(df)
			}
		}
	*/
	/*
		fmt.Println("DEBUG functions removed by Dead Code Eliminaiton:")
		for _,pkg := range rootProgram.AllPackages() {
			for _,mem := range pkg.Members {
				fn,ok := mem.(*ssa.Function)
				if ok {
					_,found := fnMap[fn]
					if !found {
						println(fn.String())
					}
				}
			}
		}
	*/

	var dupCheck = make(map[string]*ssa.Function)
	for f := range fnMap {
		p, n := GetFnNameParts(f)
		first, exists := dupCheck[p+"."+n]
		if exists {
			panic(fmt.Sprintf(
				"duplicate function name: %s.%s\nparent orig %v new %v\n",
				p, n, uintptr(unsafe.Pointer(first)), uintptr(unsafe.Pointer(f))))
		}
		dupCheck[p+"."+n] = f
	}

	for _, f := range fnMapSorted() {
		if !IsOverloaded(f) {
			if err:= tgossa.CheckNames(f); err!=nil {
				panic(err)
			}
			emitFunc(f)
		}
	}
}

func IsOverloaded(f *ssa.Function) bool {
	pn := "unknown" // Defensive, as some synthetic or other edge-case functions may not have a valid package name
	rx := f.Signature.Recv()
	if rx == nil { // ordinary function
		if f.Pkg != nil {
			if f.Pkg.Object != nil {
				pn = f.Pkg.Object.Path() //was .Name()
			}
		} else {
			if f.Object() != nil {
				if f.Object().Pkg() != nil {
					pn = f.Object().Pkg().Path() //was .Name()
				}
			}
		}
	} else { // determine the package information from the type description
		typ := rx.Type()
		ts := typ.String()
		if ts[0] == '*' {
			ts = ts[1:]
		}
		tss := strings.Split(ts, ".")
		if len(tss) >= 2 {
			ts = tss[len(tss)-2] // take the part before the final dot
		} else {
			ts = tss[0] // no dot!
		}
		pn = ts
	}
	tss := strings.Split(pn, "/") // TODO check this also works in Windows
	ts := tss[len(tss)-1]         // take the last part of the path
	pn = ts                       // TODO this is incorrect, but not currently a problem as there is no function overloading
	//println("DEBUG package name: " + pn)
	if LanguageList[TargetLang].FunctionOverloaded(pn, f.Name()) ||
		strings.HasPrefix(pn, "_") { // the package is not in the target language, signaled by a leading underscore and
		return true
	}
	return false
}

//------------------------------------------------------------------------------------------------------------
// Some target languages, notably Java and PHP, cannot handle very large functions like unicode.init(),
// and so need to be split into a number of sub-functions. As the sub-functions can use stack-based temp vars,
// this also has the advantage of reducing the number of temporary variables required on the heap.
//------------------------------------------------------------------------------------------------------------

// Type to track the details of each sub-function.
type subFnInstrs struct {
	block int
	start int
	end   int
}

// Emit a particular function.
func emitFunc(fn *ssa.Function) {

	/* TODO research if the ssautil.Switches() function can be incorporated to provide any run-time improvement to the code
	sw := ssautil.Switches(fn)
	if len(sw) > 0 {
		fmt.Printf("DEBUG Switches: %s = %+v\n", fn, sw)
	}
	*/

	var subFnList []subFnInstrs        // where the sub-functions are
	canOptMap := make(map[string]bool) // TODO review use of this mechanism

	//println("DEBUG processing function: ", fn.Name())
	MakePosHash(fn.Pos()) // mark that we have entered a function
	trackPhi := true
	switch len(fn.Blocks) {
	case 0: // NoOp - only output a function if it has a body... so ignore pure definitions (target language may generate an error, if truely undef)
		//fmt.Printf("DEBUG function has no body, ignored: %v %v \n", fn.Name(), fn.String())
	case 1: // Only one block, so no Phi tracking required
		trackPhi = false
		fallthrough
	default:
		if trackPhi {
			// check that there actually are Phi instructions to track
			trackPhi = false
		phiSearch:
			for b := range fn.Blocks {
				for i := range fn.Blocks[b].Instrs {
					_, trackPhi = fn.Blocks[b].Instrs[i].(*ssa.Phi)
					if trackPhi {
						break phiSearch
					}
				}
			}
		}
		instrCount := 0
		for b := range fn.Blocks {
			instrCount += len(fn.Blocks[b].Instrs)
		}
		mustSplitCode := false
		if instrCount > LanguageList[TargetLang].InstructionLimit {
			//println("DEBUG mustSplitCode => large function length:", instrCount, " in ", fn.Name())
			mustSplitCode = true
		}
		for b := range fn.Blocks { // go though the blocks looking for sub-functions
			instrsEmitted := 0
			inSubFn := false
			for i := range fn.Blocks[b].Instrs {
				canPutInSubFn := true
				in := fn.Blocks[b].Instrs[i]
				switch in.(type) {
				case *ssa.Phi: // phi uses self-referential temp vars that must be pre-initialised
					canPutInSubFn = false
				case *ssa.Return:
					canPutInSubFn = false
				case *ssa.Call:
					switch in.(*ssa.Call).Call.Value.(type) {
					case *ssa.Builtin:
						//NoOp
					default:
						canPutInSubFn = false
					}
				case *ssa.Select, *ssa.Send, *ssa.Defer, *ssa.RunDefers, *ssa.Panic:
					canPutInSubFn = false
				case *ssa.UnOp:
					if in.(*ssa.UnOp).Op == token.ARROW {
						canPutInSubFn = false
					}
				}
				if canPutInSubFn {
					if inSubFn {
						if instrsEmitted > LanguageList[TargetLang].SubFnInstructionLimit {
							subFnList[len(subFnList)-1].end = i
							subFnList = append(subFnList, subFnInstrs{b, i, 0})
							instrsEmitted = 0
						}
					} else {
						subFnList = append(subFnList, subFnInstrs{b, i, 0})
						inSubFn = true
					}
				} else {
					if inSubFn {
						subFnList[len(subFnList)-1].end = i
						inSubFn = false
					}
				}
				instrsEmitted++
			}
			if inSubFn {
				subFnList[len(subFnList)-1].end = len(fn.Blocks[b].Instrs)
			}
		}
		for sf := range subFnList { // go though the sub-functions looking for optimisable temp vars
			var instrMap = make(map[ssa.Instruction]bool)
			for ii := subFnList[sf].start; ii < subFnList[sf].end; ii++ {
				instrMap[fn.Blocks[subFnList[sf].block].Instrs[ii]] = true
			}

			for i := subFnList[sf].start; i < subFnList[sf].end; i++ {
				instrVal, hasVal := fn.Blocks[subFnList[sf].block].Instrs[i].(ssa.Value)
				if hasVal {
					refs := *fn.Blocks[subFnList[sf].block].Instrs[i].(ssa.Value).Referrers()
					switch len(refs) {
					case 0: // no other instruction uses the result of this one
					default: //multiple usage of the register
						canOpt := true
						for r := range refs {
							user := refs[r]
							if user.Block() != fn.Blocks[subFnList[sf].block] {
								canOpt = false
								break
							}
							_, inRange := instrMap[user]
							if !inRange {
								canOpt = false
								break
							}
						}
						if canOpt {
							canOptMap[instrVal.Name()] = true
						}
					}
				}
			}
		}

		emitFuncStart(fn, trackPhi, canOptMap, mustSplitCode)
		thisSubFn := 0
		for b := range fn.Blocks {
			emitPhi := trackPhi
			emitBlockStart(fn.Blocks, b, emitPhi)
			inSubFn := false
			for i := 0; i < len(fn.Blocks[b].Instrs); i++ {
				if thisSubFn >= 0 && thisSubFn < len(subFnList) { // not at the end of the list
					if b == subFnList[thisSubFn].block {
						if i >= subFnList[thisSubFn].end && inSubFn {
							inSubFn = false
							thisSubFn++
							if thisSubFn >= len(subFnList) {
								thisSubFn = -1 // we have come to the end of the list
							}
						}
					}
				}
				if thisSubFn >= 0 && thisSubFn < len(subFnList) { // not at the end of the list
					if b == subFnList[thisSubFn].block {
						if i == subFnList[thisSubFn].start {
							inSubFn = true
							l := TargetLang
							if mustSplitCode {
								fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].SubFnCall(thisSubFn))
							} else {
								emitSubFn(fn, subFnList, thisSubFn, mustSplitCode, canOptMap)
							}
						}
					}
				}
				if !inSubFn {
					//TODO optimize phi case statements
					phiList := 0
				phiLoop:
					switch fn.Blocks[b].Instrs[i+phiList].(type) {
					case *ssa.Phi:
						if len(*fn.Blocks[b].Instrs[i+phiList].(*ssa.Phi).Referrers()) > 0 {
							phiList++
							if (i + phiList) < len(fn.Blocks[b].Instrs) {
								goto phiLoop
							}
						}
					}
					if phiList > 1 {
						peephole(fn.Blocks[b].Instrs[i : i+phiList])
						i += phiList - 1
					} else {
						emitPhi = emitInstruction(fn.Blocks[b].Instrs[i],
							fn.Blocks[b].Instrs[i].Operands(make([]*ssa.Value, 0)))
					}
				}
			}
			if thisSubFn >= 0 && thisSubFn < len(subFnList) { // not at the end of the list
				if b == subFnList[thisSubFn].block {
					if inSubFn {
						thisSubFn++
						if thisSubFn >= len(subFnList) {
							thisSubFn = -1 // we have come to the end of the list
						}
					}
				}
			}
			emitBlockEnd(fn.Blocks, b, emitPhi && trackPhi)
		}
		emitRunEnd(fn)
		if mustSplitCode {
			for sf := range subFnList {
				emitSubFn(fn, subFnList, sf, mustSplitCode, canOptMap)
			}
		}
		emitFuncEnd(fn)
	}
}

func emitSubFn(fn *ssa.Function, subFnList []subFnInstrs, sf int, mustSplitCode bool, canOptMap map[string]bool) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].SubFnStart(sf, mustSplitCode))
	for i := subFnList[sf].start; i < subFnList[sf].end; i++ {
		instrVal, hasVal := fn.Blocks[subFnList[sf].block].Instrs[i].(ssa.Value)
		if hasVal {
			if canOptMap[instrVal.Name()] == true {
				l := TargetLang
				fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].DeclareTempVar(instrVal))
			}
		}
	}
	peephole(fn.Blocks[subFnList[sf].block].Instrs[subFnList[sf].start:subFnList[sf].end])
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].SubFnEnd(sf, int(LatestValidPosHash), mustSplitCode))
}

func GetFnNameParts(fn *ssa.Function) (pack, nam string) {
	mName := fn.Name()
	pName, _ := FuncPathName(fn) //fmt.Sprintf("fn%d", fn.Pos()) //uintptr(unsafe.Pointer(fn)))
	if fn.Pkg != nil {
		if fn.Pkg.Object != nil {
			pName = fn.Pkg.Object.Path() // was .Name()
		}
	}
	if fn.Signature.Recv() != nil { // we have a method
		pName = fn.Signature.Recv().Pkg().Name() + ":" + fn.Signature.Recv().Type().String() // note no underlying()
		//pName = LanguageList[l].PackageOverloadReplace(pName)
	}
	return pName, mName
}

// Emit the start of a function.
func emitFuncStart(fn *ssa.Function, trackPhi bool, canOptMap map[string]bool, mustSplitCode bool) {
	l := TargetLang
	posStr := CodePosition(fn.Pos())
	pName, mName := GetFnNameParts(fn)
	isPublic := unicode.IsUpper(rune(mName[0])) // TODO check rules for non-ASCII 1st characters and fix
	fmt.Fprintln(&LanguageList[l].buffer,
		LanguageList[l].FuncStart(pName, mName, fn, posStr, isPublic, trackPhi, grMap[fn] || mustSplitCode, canOptMap))
}

// Emit the end of a function.
func emitFuncEnd(fn *ssa.Function) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].FuncEnd(fn))
}

// Emit code for after the end of all the case statements for a functions _Next phi switch, but before the sub-functions.
func emitRunEnd(fn *ssa.Function) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].RunEnd(fn))
}

// Emit the start of the code to handle a particular SSA code block,
// for Haxe this handles a particular _Next value (in phi or -ve if synthetic because of call or channel Rx/Tx).
func emitBlockStart(block []*ssa.BasicBlock, num int, emitPhi bool) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].BlockStart(block, num, emitPhi))
}

// Emit the end of the SSA code block
func emitBlockEnd(block []*ssa.BasicBlock, num int, emitPhi bool) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].BlockEnd(block, num, emitPhi))
}

// Emit the code for a call to a function or builtin, which could be deferred.
func emitCall(isBuiltin, isGo, isDefer, usesGr bool, register string, callInfo ssa.CallCommon, errorInfo, comment string) {
	// usesGr gives the default position
	l := TargetLang
	fnToCall := ""
	if isBuiltin {
		fnToCall = callInfo.Value.(*ssa.Builtin).Name()
		usesGr = false
	} else if callInfo.StaticCallee() != nil {
		pName, _ := FuncPathName(callInfo.StaticCallee()) //fmt.Sprintf("fn%d", callInfo.StaticCallee().Pos())
		if callInfo.Signature().Recv() != nil {
			pName = callInfo.Signature().Recv().Pkg().Name() + ":" + callInfo.Signature().Recv().Type().String() // no use of Underlying() here
		} else {
			pkg := callInfo.StaticCallee().Package()
			if pkg != nil {
				pName = pkg.Object.Path() // was .Name()
			}
		}
		fnToCall = LanguageList[l].LangName(pName, callInfo.StaticCallee().Name())
		usesGr = grMap[callInfo.StaticCallee()]
	} else { // Dynamic call (take the default on usesGr)
		fnToCall = LanguageList[l].Value(callInfo.Value, errorInfo)
	}

	if isBuiltin {
		switch fnToCall {
		case "len", "cap", "append", "real", "imag", "complex": //  "copy" may have the results unused
			if register == "" {
				LogError(errorInfo, "pogo", fmt.Errorf("the result from a built-in function is not used"))
			}
		default:
		}
	} else {
		if callInfo.Signature().Results().Len() > 0 {
			if register == "" {
				LogWarning(errorInfo, "pogo", fmt.Errorf("the result from a function call is not used")) //TODO is this needed?
			}
		}
	}
	// target language code must do builtin emulation
	text := LanguageList[l].Call(register, callInfo, callInfo.Args, isBuiltin, isGo, isDefer, usesGr, fnToCall, errorInfo)
	fmt.Fprintln(&LanguageList[l].buffer, text+LanguageList[l].Comment(comment))
}

// FuncValue is a utility function to avoid publishing rootProgram from this package.
func FuncValue(obj *types.Func) ssa.Value {
	return rootProgram.FuncValue(obj)
}
