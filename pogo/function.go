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

	/* for DCE tests
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/pointer"
	*/)

// FnIsCalled tells if the function is called
func (comp *Compilation) FnIsCalled(fn *ssa.Function) bool {
	return comp.fnMap[fn]
}

// For every function, maybe emit the code...
func (comp *Compilation) emitFunctions() {
	dceList := []*ssa.Package{
		comp.mainPackage,
		comp.rootProgram.ImportedPackage(LanguageList[comp.TargetLang].Goruntime),
	}
	dceExceptions := []string{}
	if LanguageList[comp.TargetLang].TestFS != "" { // need to load file system
		dceExceptions = append(dceExceptions, "syscall") // so that we keep UnzipFS()
	}
	dceExceptions = append(dceExceptions, comp.LibListNoDCE...)
	for _, ex := range dceExceptions {
		exip := comp.rootProgram.ImportedPackage(ex)
		if exip != nil {
			dceList = append(dceList, exip)
		} else {
			//fmt.Println("DEBUG exip nil for package: ",ex)
		}
	}
	comp.fnMap, comp.grMap = tgossa.VisitedFunctions(comp.rootProgram, dceList, comp.IsOverloaded)

	/* NOTE non-working code below attempts to improve Dead Code Elimination,
	//	but is unreliable so far, in part because the target lang runtime may use "unsafe" pointers
	//  and in any case, every program of any consequence uses "reflect"

	// but pointer analysis only fully works when the "reflect" and "unsafe" packages are not used
	canPointerAnalyse := len(dceExceptions) == 0 // and with no DCE exceptions
	for _, pkg := range rootProgram.AllPackages() {
		switch pkg.Object.Name() {
		case "reflect", "unsafe":
			canPointerAnalyse = false
		}
	}
	if canPointerAnalyse {
		println("DEBUG can use pointer analysis")
		roots := []*ssa.Function{}
		for _, pkg := range rootProgram.AllPackages() {
			if pkg != nil {
				for _, mem := range pkg.Members {
					fn, ok := mem.(*ssa.Function)
					if ok { // TODO - not hard-coded values, more descrimination
						if (pkg.Object.Name() == "main" && fn.Name() == "main") ||
							strings.HasPrefix(fn.Name(), "init") {
							//(pkg.Object.Name() == LanguageList[TargetLang].Goruntime && fn.Name() == "main") ||
							//pkg.Object.Name() == "reflect" ||
							//(pkg.Object.Name() == "syscall" && fn.Name() == "UnzipFS") {
							roots = append(roots, fn)
							//fmt.Println("DEBUG root added:",fn.String())
						}
					}
				}
			}
		}
		config := &pointer.Config{
			Mains:          []*ssa.Package{mainPackage},
			BuildCallGraph: true,
			Reflection:     true,
		}
		ptrResult, err := pointer.Analyze(config)
		if err != nil {
			panic(err)
		}

		for _, pkg := range rootProgram.AllPackages() {
			funcs := []*ssa.Function{}
			for _, mem := range pkg.Members {
				fn, ok := mem.(*ssa.Function)
				if ok {
					funcs = append(funcs, fn)
				}
				typ, ok := mem.(*ssa.Type)
				if ok {
					mset := rootProgram.MethodSets.MethodSet(typ.Type())
					for i, n := 0, mset.Len(); i < n; i++ {
						mf := rootProgram.Method(mset.At(i))
						funcs = append(funcs, mf)
						//fmt.Printf("DEBUG method %v\n", mf)
					}
				}
			}
			for _, fn := range funcs {
				notRoot := true
				for _, r := range roots {
					if r == fn {
						notRoot = false
						break
					}
				}
				if notRoot {
					_, found := fnMap[fn]
					hasPath := false
					if fn != nil {
						for _, r := range roots {
							if r != nil {
								nod, ok := ptrResult.CallGraph.Nodes[r]
								if ok {
									pth := callgraph.PathSearch(nod,
										func(n *callgraph.Node) bool {
											if n == nil {
												return false
											}
											return n.Func == fn
										})
									if pth != nil {
										//fmt.Printf("DEBUG path from %v to %v = %v\n",
										//	r, fn, pth)
										hasPath = true
										break
									}

								}
							}
						}
					}
					if found != hasPath {
						if found { // we found it when we should not have
							println("DEBUG DCE function not called: ", fn.String())
							delete(fnMap, fn)
							delete(grMap, fn)
						} else {
							panic("function not found in DCE cross-check: " + fn.String())
						}
					}
				}
			}
		}
	}
	*/
	/*
		fmt.Println("DEBUG funcs not requiring goroutines:")
		for df, db := range grMap {
			if !db {
				fmt.Println(df)
			}
		}
	*/

	// Remove virtual functions
	for _, pkg := range comp.rootProgram.AllPackages() {
		for _, vf := range LanguageList[comp.TargetLang].PseudoPkgPaths {
			if pkg.Pkg.Path() == vf {
				for _, mem := range pkg.Members {
					fn, ok := mem.(*ssa.Function)
					if ok {
						//println("DEBUG DCE virtual function: ", fn.String())
						delete(comp.fnMap, fn)
						delete(comp.grMap, fn)
					}
				}
			}
		}
	}

	var dupCheck = make(map[string]*ssa.Function)
	for f := range comp.fnMap {
		p, n := comp.GetFnNameParts(f)
		first, exists := dupCheck[p+"."+n]
		if exists {
			panic(fmt.Sprintf(
				"duplicate function name: %s.%s\nparent orig %v new %v\n",
				p, n, uintptr(unsafe.Pointer(first)), uintptr(unsafe.Pointer(f))))
		}
		dupCheck[p+"."+n] = f
	}

	for _, f := range comp.fnMapSorted() {
		if !comp.IsOverloaded(f) {
			if err := tgossa.CheckNames(f); err != nil {
				panic(err)
			}
			comp.emitFunc(f)
		}
	}
}

// IsOverloaded reports if a function reference should be replaced
func (comp *Compilation) IsOverloaded(f *ssa.Function) bool {
	pn := "unknown" // Defensive, as some synthetic or other edge-case functions may not have a valid package name
	rx := f.Signature.Recv()
	if rx == nil { // ordinary function
		if f.Pkg != nil {
			if f.Pkg.Pkg != nil {
				pn = f.Pkg.Pkg.Path() //was .Name()
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
	if LanguageList[comp.TargetLang].FunctionOverloaded(pn, f.Name()) ||
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
func (comp *Compilation) emitFunc(fn *ssa.Function) {

	/* TODO research if the ssautil.Switches() function can be incorporated to provide any run-time improvement to the code
	// it would give a big adavantage, but only to a very small number of functions - so definately TODO
	sw := ssautil.Switches(fn)
		fmt.Printf("DEBUG Switches for : %s \n", fn)
	for num,swi := range sw {
		fmt.Printf("DEBUG Switches[%d]= %+v\n", num, swi)
	}
	*/

	var subFnList []subFnInstrs        // where the sub-functions are
	canOptMap := make(map[string]bool) // TODO review use of this mechanism

	//println("DEBUG processing function: ", fn.Name())
	comp.MakePosHash(fn.Pos()) // mark that we have entered a function
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
		if instrCount > LanguageList[comp.TargetLang].InstructionLimit {
			//println("DEBUG mustSplitCode => large function length:", instrCount, " in ", fn.Name())
			mustSplitCode = true
		}
		blks := fn.DomPreorder() // was fn.Blocks
		for b := range blks {    // go though the blocks looking for sub-functions
			instrsEmitted := 0
			inSubFn := false
			for i := range blks[b].Instrs {
				canPutInSubFn := true
				in := blks[b].Instrs[i]
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
						if instrsEmitted > LanguageList[comp.TargetLang].SubFnInstructionLimit {
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
				subFnList[len(subFnList)-1].end = len(blks[b].Instrs)
			}
		}
		for sf := range subFnList { // go though the sub-functions looking for optimisable temp vars
			var instrMap = make(map[ssa.Instruction]bool)
			for ii := subFnList[sf].start; ii < subFnList[sf].end; ii++ {
				instrMap[blks[subFnList[sf].block].Instrs[ii]] = true
			}

			for i := subFnList[sf].start; i < subFnList[sf].end; i++ {
				instrVal, hasVal := blks[subFnList[sf].block].Instrs[i].(ssa.Value)
				if hasVal {
					refs := *blks[subFnList[sf].block].Instrs[i].(ssa.Value).Referrers()
					switch len(refs) {
					case 0: // no other instruction uses the result of this one
					default: //multiple usage of the register
						canOpt := true
						for r := range refs {
							user := refs[r]
							if user.Block() != blks[subFnList[sf].block] {
								canOpt = false
								break
							}
							_, inRange := instrMap[user]
							if !inRange {
								canOpt = false
								break
							}
						}
						if canOpt &&
							!LanguageList[comp.TargetLang].CanInline(blks[subFnList[sf].block].Instrs[i]) {
							canOptMap[instrVal.Name()] = true
						}
					}
				}
			}
		}

		reconstruct := tgossa.Reconstruct(blks, comp.grMap[fn] || mustSplitCode)
		if reconstruct != nil {
			//fmt.Printf("DEBUG reconstruct %s %#v\n",fn.String(),reconstruct)
		}

		comp.emitFuncStart(fn, blks, trackPhi, canOptMap, mustSplitCode, reconstruct)
		thisSubFn := 0
		for b := range blks {
			emitPhi := trackPhi
			comp.emitBlockStart(blks, b, emitPhi)
			inSubFn := false
			for i := 0; i < len(blks[b].Instrs); i++ {
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
							l := comp.TargetLang
							if mustSplitCode {
								fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].SubFnCall(thisSubFn))
							} else {
								comp.emitSubFn(fn, blks, subFnList, thisSubFn, mustSplitCode, canOptMap)
							}
						}
					}
				}
				if !inSubFn {
					// optimize phi case statements
					phiList := 0
				phiLoop:
					switch blks[b].Instrs[i+phiList].(type) {
					case *ssa.Phi:
						if len(*blks[b].Instrs[i+phiList].(*ssa.Phi).Referrers()) > 0 {
							phiList++
							if (i + phiList) < len(blks[b].Instrs) {
								goto phiLoop
							}
						}
					}
					if phiList > 0 {
						comp.peephole(blks[b].Instrs[i : i+phiList])
						i += phiList - 1
					} else {
						emitPhi = comp.emitInstruction(blks[b].Instrs[i],
							blks[b].Instrs[i].Operands(make([]*ssa.Value, 0)))
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
			comp.emitBlockEnd(blks, b, emitPhi && trackPhi)
		}
		comp.emitRunEnd(fn)
		if mustSplitCode {
			for sf := range subFnList {
				comp.emitSubFn(fn, blks, subFnList, sf, mustSplitCode, canOptMap)
			}
		}
		comp.emitFuncEnd(fn)
	}
}

func (comp *Compilation) emitSubFn(fn *ssa.Function, blks []*ssa.BasicBlock, subFnList []subFnInstrs, sf int, mustSplitCode bool, canOptMap map[string]bool) {
	l := comp.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].SubFnStart(sf, mustSplitCode,
		blks[subFnList[sf].block].Instrs[subFnList[sf].start:subFnList[sf].end]))
	for i := subFnList[sf].start; i < subFnList[sf].end; i++ {
		instrVal, hasVal := blks[subFnList[sf].block].Instrs[i].(ssa.Value)
		if hasVal {
			if canOptMap[instrVal.Name()] == true {
				l := comp.TargetLang
				fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].DeclareTempVar(instrVal))
			}
		}
	}
	comp.peephole(blks[subFnList[sf].block].Instrs[subFnList[sf].start:subFnList[sf].end])
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].SubFnEnd(sf, int(comp.LatestValidPosHash), mustSplitCode))
}

// GetFnNameParts gets the elements of the function's name
func (comp *Compilation) GetFnNameParts(fn *ssa.Function) (pack, nam string) {
	mName := fn.Name()
	pName, _ := comp.FuncPathName(fn) //fmt.Sprintf("fn%d", fn.Pos()) //uintptr(unsafe.Pointer(fn)))
	if fn.Pkg != nil {
		if fn.Pkg.Pkg != nil {
			pName = fn.Pkg.Pkg.Path() // was .Name()
		}
	}
	if fn.Signature.Recv() != nil { // we have a method
		pName = fn.Signature.Recv().Pkg().Name() + ":" + fn.Signature.Recv().Type().String() // note no underlying()
		//pName = LanguageList[l].PackageOverloadReplace(pName)
	}
	return pName, mName
}

// Emit the start of a function.
func (comp *Compilation) emitFuncStart(fn *ssa.Function, blks []*ssa.BasicBlock, trackPhi bool, canOptMap map[string]bool, mustSplitCode bool, reconstruct []tgossa.BlockFormat) {
	l := comp.TargetLang
	posStr := comp.CodePosition(fn.Pos())
	pName, mName := comp.GetFnNameParts(fn)
	isPublic := unicode.IsUpper(rune(mName[0])) // TODO check rules for non-ASCII 1st characters and fix
	fmt.Fprintln(&LanguageList[l].buffer,
		LanguageList[l].FuncStart(pName, mName, fn, blks, posStr, isPublic, trackPhi, comp.grMap[fn] || mustSplitCode, canOptMap, reconstruct))
}

// Emit the end of a function.
func (comp *Compilation) emitFuncEnd(fn *ssa.Function) {
	l := comp.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].FuncEnd(fn))
}

// Emit code for after the end of all the case statements for a functions _Next phi switch, but before the sub-functions.
func (comp *Compilation) emitRunEnd(fn *ssa.Function) {
	l := comp.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].RunEnd(fn))
}

// Emit the start of the code to handle a particular SSA code block
func (comp *Compilation) emitBlockStart(block []*ssa.BasicBlock, num int, emitPhi bool) {
	l := comp.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].BlockStart(block, num, emitPhi))
}

// Emit the end of the SSA code block
func (comp *Compilation) emitBlockEnd(block []*ssa.BasicBlock, num int, emitPhi bool) {
	l := comp.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].BlockEnd(block, num, emitPhi))
}

// Emit the code for a call to a function or builtin, which could be deferred.
func (comp *Compilation) emitCall(isBuiltin, isGo, isDefer, usesGr bool, register string, callInfo ssa.CallCommon, errorInfo, comment string) {
	// usesGr gives the default position
	l := comp.TargetLang
	fnToCall := ""
	if isBuiltin {
		fnToCall = callInfo.Value.(*ssa.Builtin).Name()
		usesGr = false
	} else if callInfo.StaticCallee() != nil {
		pName, _ := comp.FuncPathName(callInfo.StaticCallee()) //fmt.Sprintf("fn%d", callInfo.StaticCallee().Pos())
		if callInfo.Signature().Recv() != nil {
			pName = callInfo.Signature().Recv().Pkg().Name() + ":" + callInfo.Signature().Recv().Type().String() // no use of Underlying() here
		} else {
			pkg := callInfo.StaticCallee().Package()
			if pkg != nil {
				pName = pkg.Pkg.Path() // was .Name()
			}
		}
		fnToCall = LanguageList[l].LangName(pName, callInfo.StaticCallee().Name())
		usesGr = comp.grMap[callInfo.StaticCallee()]
	} else { // Dynamic call (take the default on usesGr)
		fnToCall = LanguageList[l].Value(callInfo.Value, errorInfo)
	}

	if isBuiltin {
		switch fnToCall {
		case "len", "cap", "append", "real", "imag", "complex": //  "copy" may have the results unused
			if register == "" {
				comp.LogError(errorInfo, "pogo", fmt.Errorf("the result from a built-in function is not used"))
			}
		default:
		}
	} else {
		if callInfo.Signature().Results().Len() > 0 {
			if register == "" {
				comp.LogWarning(errorInfo, "pogo", fmt.Errorf("the result from a function call is not used")) //TODO is this needed?
			}
		}
	}
	// target language code must do builtin emulation
	text := LanguageList[l].Call(register, callInfo, callInfo.Args, isBuiltin, isGo, isDefer, usesGr, fnToCall, errorInfo)
	fmt.Fprintln(&LanguageList[l].buffer, text+LanguageList[l].Comment(comment))
}

// FuncValue is a utility function to avoid publishing rootProgram from this package.
func (comp *Compilation) FuncValue(obj *types.Func) ssa.Value {
	return comp.rootProgram.FuncValue(obj)
}
