// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"code.google.com/p/go.tools/go/exact"
	"code.google.com/p/go.tools/go/ssa"
	"code.google.com/p/go.tools/go/types"
	"fmt"
	"github.com/tardisgo/tardisgo/pogo"
	"runtime"
	"strings"
	"unsafe"
)

// Start the main Go class in haxe
func (langType) GoClassStart() string {
	// the code below makes the Go class globally visible in JS as window.Go in the browser or exports.Go in nodejs
	//TODO consider how to make Go/Haxe libs available across all platforms
	return `
#if js
@:expose("Go")
#end
class Go
{

	public static function Platform():String { // codes returned the same as used by Haxe 
    #if flash
    	return "flash";
    #elseif js
    	return "js";
    #elseif cpp
    	return "cpp";
    #elseif java
    	return "java";
    #elseif cs
    	return "cs";
    #elseif php
    	return "php";
    #elseif neko
    	return "neko";
    #else 
        #error "Only the js, flash, cpp (C++), java, cs (C#), php and neko Haxe targets are supported as a Go platform" 
    #end
	}
`
}

// end the main Go class
func (l langType) GoClassEnd(pkg *ssa.Package) string {
	// init function
	main := "public static var doneInit:Bool=false;\n"                                                          // flag to run this routine only once
	main += "\npublic static function init() : Void {\ndoneInit=true;\nvar gr:Int=Scheduler.makeGoroutine();\n" // first goroutine number is always 0
	main += `if(gr!=0) throw "non-zero goroutine number in init";` + "\n"                                       // first goroutine number is always 0, NOTE using throw as panic not setup

	//NOTE HACK start
	ap := pkg.Prog.AllPackages()
	for p := range ap {
		// fmt.Println("DEBUG: ", ap[p].Object.Name())
		if ap[p].Object.Name() == "runtime" {
			var memStats runtime.MemStats // see magic variable setting required below
			// this magic number required to init the runtime module, may change in future versions
			// see go/tip/go/src/pkg/runtime/mem.go:68
			main += fmt.Sprintf("Go.runtime_sizeof_C_MStats.store(%d);\n", unsafe.Sizeof(memStats))
			break
		}
	}
	//NOTE HACK end

	main += "var _sfgr=new Go_haxegoruntime_init(gr,[]).run();\n" //haxegoruntime.init() NOTE can't use callFromHaxe() as that would call this fn
	main += "while(_sfgr._incomplete) Scheduler.runAll();\n"
	main += "var _sf=new Go_" + pkg.Object.Name() + `_init(gr,[]).run();` + "\n" //NOTE can't use callFromHaxe() as that would call this fn
	main += "while(_sf._incomplete) Scheduler.runAll();\n"
	main += "Scheduler.doneInit=true;\n"
	main += "Go.haxegoruntime_ZiLen.store('字'.length);\n" // value required by haxegoruntime to know what type of strings we have
	main += "}\n"
	// Haxe main function, only called in a go-only environment
	main += "\npublic static function main() : Void {\n"
	main += "Go_" + pkg.Object.Name() + `_main.callFromHaxe();` + "\n"
	main += "}\n"

	pos := "public static function CPos(pos:Int):String {\nvar prefix:String=\"\";\n"
	pos += fmt.Sprintf(`if (pos==%d) return "(pogo.NoPosHash)";`, pogo.NoPosHash) + "\n"
	pos += "if (pos<0) { pos = -pos; prefix= \"near \";}\n"
	for p := len(pogo.PosHashFileList) - 1; p >= 0; p-- {
		if p != len(pogo.PosHashFileList)-1 {
			pos += "else "
		}
		pos += fmt.Sprintf(`if(pos>%d) return prefix+"%s:"+Std.string(pos-%d);`,
			pogo.PosHashFileList[p].BasePosHash,
			strings.Replace(pogo.PosHashFileList[p].FileName, "\\", "\\\\", -1),
			pogo.PosHashFileList[p].BasePosHash) + "\n"
	}
	pos += "else return \"(invalid pogo.PosHash:\"+Std.string(pos)+\")\";}\n"

	return main + pos + "} // end Go class"
}

func (langType) Const(lit ssa.Const, position string) (typ, val string) {
	if lit.Value == nil {
		return "Dynamic", "null"
	}
	lit.Name()
	switch lit.Value.Kind() {
	case exact.Bool:
		return "Bool", lit.Value.String()
	case exact.String:
		// TODO check if conversion of some string constant declarations are required
		switch lit.Type().Underlying().(type) {
		case *types.Basic:
			return "String", lit.Value.String()
		case *types.Slice:
			return "Slice", "Force.toUTF8slice(this._goroutine," + lit.Value.String() + ")"
		default:
			pogo.LogError(position, "Haxe", fmt.Errorf("Const() internal error, unknown string type"))
		}
	case exact.Float:
		return "Float", pogo.Float64Val(lit.Value, position)
	case exact.Int:
		h, l := pogo.IntVal(lit.Value, position)
		switch lit.Type().Underlying().(*types.Basic).Kind() {
		case types.Int64, types.Uint64:
			return "GOint64", fmt.Sprintf("GOint64.make(0x%x,0x%x)", uint32(h), uint32(l))
		case types.Float32, types.Float64, types.UntypedFloat:
			return "Float", pogo.Float64Val(lit.Value, position)
		case types.Complex64, types.Complex128:
			return "Complex", fmt.Sprintf("new Complex(%s,0)", pogo.Float64Val(lit.Value, position))
		default:
			if h != 0 && h != -1 {
				pogo.LogWarning(position, "Haxe", fmt.Errorf("Integer constant value > 32 bits, rendered as 64-bit : %v", lit.Value))
				return "GOint64", fmt.Sprintf("GOint64.make(0x%x,0x%x)", uint32(h), uint32(l))
			}
			switch lit.Type().Underlying().(*types.Basic).Kind() {
			case types.Uint, types.Uint32, types.Uint16, types.Uint8:
				if l == -1 {
					return "Int",
						" #if js untyped __js__(\"0xffffffff\") #elseif php untyped __php__(\"0xffffffff\") #else (-1) #end "
				} else {
					return "Int", fmt.Sprintf(" (%d) ", l)
				}
			default:
				if l < 0 {
					return "Int", fmt.Sprintf("(%d)", l)
				} else {
					return "Int", fmt.Sprintf("%d", l)
				}
			}
		}
	case exact.Unknown: // not sure we should ever get here!
		return "Dynamic", "null"
	case exact.Complex:
		realV, _ := exact.Float64Val(exact.Real(lit.Value))
		imagV, _ := exact.Float64Val(exact.Imag(lit.Value))
		return "Complex", fmt.Sprintf("new Complex(%g,%g)", realV, imagV)
	default:
		pogo.LogError(position, "Haxe", fmt.Errorf("Const() internal error, unknown constant type: %v", lit.Value.Kind()))
	}
	return "", ""
}

// only public Literals are created here, so that they can be used by Haxe callers of the Go code
func (l langType) NamedConst(packageName, objectName string, lit ssa.Const, position string) string {
	typ, rhs := l.Const(lit, position+":"+packageName+"."+objectName)
	return fmt.Sprintf("public static var %s:%s = %s;%s",
		l.LangName(packageName, objectName), typ, rhs, l.Comment(position))
}

func (l langType) Global(packageName, objectName string, glob ssa.Global, position string, isPublic bool) string {
	pub := "public "                                                      // all globals have to be public in Haxe terms
	gTyp := glob.Type().Underlying().(*types.Pointer).Elem().Underlying() // globals are always pointers to an underlying element
	init := "new Pointer(null)"                                           // default is nothing in the global variable
	switch gTyp.(type) {
	case *types.Basic, *types.Struct, *types.Array, *types.Pointer:
		init = "new Pointer(" + l.LangType(gTyp, true, position) + ")" // initialize basic types only
	}
	return fmt.Sprintf("%sstatic %s %s",
		pub, haxeVar(l.LangName(packageName, objectName), "Pointer" /*typ*/, init, position, "Global()"),
		l.Comment(position))
}
