// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/tardisgo/tardisgo/pogo"
	"golang.org/x/tools/go/exact"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

// Start the main Go class in haxe
func (langType) GoClassStart() string {
	// the code below makes the Go class globally visible in JS as window.Go in the browser or exports.Go in nodejs
	// TODO consider how to make Go/Haxe libs available across all platforms
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
    #elseif python
    	#error "SORRY: the python target is not yet ready for general use"
    	return "python";
    #elseif php
    	return "php";
    #elseif neko
    	return "neko";
    #else 
        #error "Only the js, flash, cpp (C++), java, cs (C#), php, python and neko Haxe targets are supported as a Go platform" 
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

	main += "var _sfgr=new Go_haxegoruntime_init(gr,[]).run();\n" //haxegoruntime.init() NOTE can't use .hx() to call from Haxe as that would call this fn
	main += `Go.haxegoruntime_ZZiLLen.store_uint32('字'.length);`  // value required by haxegoruntime to know what type of strings we have
	main += "while(_sfgr._incomplete) Scheduler.runAll();\n"
	main += "var _sf=new Go_" + l.LangName(pkg.Object.Path(), "init") + `(gr,[]).run();` + "\n" //NOTE can't use .hx() to call from Haxe as that would call this fn
	main += "while(_sf._incomplete) Scheduler.runAll();\n"
	main += ""
	main += "Scheduler.doneInit=true;\n"
	main += "}\n"
	// Haxe main function, only called in a go-only environment
	main += "\npublic static function main() : Void {\n"
	main += "Go_" + l.LangName(pkg.Object.Path(), "main") + `.hx();` + "\n"
	main += "}\n"

	pos := "public static function CPos(pos:Int):String {\nvar prefix:String=\"\";\n"
	pos += fmt.Sprintf(`if (pos==%d) return "(pogo.NoPosHash)";`, pogo.NoPosHash) + "\n"
	pos += "if (pos<0) { pos = -pos; prefix= \"near \";}\n"
	for p := len(pogo.PosHashFileList) - 1; p >= 0; p-- {
		pos += fmt.Sprintf(`if(pos>%d) return prefix+"%s:"+Std.string(pos-%d);`,
			pogo.PosHashFileList[p].BasePosHash,
			strings.Replace(pogo.PosHashFileList[p].FileName, "\\", "\\\\", -1),
			pogo.PosHashFileList[p].BasePosHash) + "\n"
	}
	pos += "return \"(invalid pogo.PosHash:\"+Std.string(pos)+\")\";\n}\n"

	if pogo.DebugFlag {
		pos += "\npublic static function getStartCPos(s:String):Int {\n"
		for p := len(pogo.PosHashFileList) - 1; p >= 0; p-- {
			pos += "\t" + fmt.Sprintf(`if("%s".indexOf(s)!=-1) return %d;`,
				strings.Replace(pogo.PosHashFileList[p].FileName, "\\", "\\\\", -1),
				pogo.PosHashFileList[p].BasePosHash) + "\n"
		}
		pos += "\treturn -1;\n}\n"

		pos += "\npublic static function getGlobal(s:String):String {\n"
		globs := pogo.GlobalList()
		for _, g := range globs {
			goName := strings.Replace(g.Package+"."+g.Member, "\\", "\\\\", -1)
			pos += "\t" + fmt.Sprintf(`if("%s".indexOf(s)!=-1) return "%s = "+%s.toString();`,
				goName, goName, l.LangName(g.Package, g.Member)) + "\n"
		}
		pos += "\treturn \"Couldn't find global: \"+s;\n}\n"

	}

	return main + pos + "} // end Go class"
}

func haxeStringConst(sconst string, position string) string {
	s, err := strconv.Unquote(sconst)
	if err != nil {
		pogo.LogError(position, "Haxe", errors.New(err.Error()+" : "+sconst))
		return ""
	}
	ret0 := ""
	hadEsc := false
	for i := 0; i < len(s); i++ {
		c := rune(s[i])
		if unicode.IsPrint(c) && c < unicode.MaxASCII && c != '"' && c != '`' && c != '\\' && !hadEsc {
			ret0 += string(c)
		} else {
			ret0 += fmt.Sprintf("\\x%02X", c)
			hadEsc = true
		}
	}
	ret0 = `"` + ret0 + `"`

	ret := ``
	compound := ""
	hadStr := false
	for i := 0; i < len(s); i++ {
		c := rune(s[i])
		if unicode.IsPrint(c) && c < unicode.MaxASCII && c != '"' && c != '`' && c != '\\' {
			compound += string(c)
		} else {
			if hadStr {
				ret += "+"
			}
			if compound != "" {
				compound = `"` + compound + `"+`
			}
			ret += fmt.Sprintf("%sString.fromCharCode(%d)", compound, c)
			compound = ""
			hadStr = true
		}
	}
	if hadStr {
		if compound != "" {
			ret += fmt.Sprintf("+\"%s\"", compound)
		}
	} else {
		ret += fmt.Sprintf("\"%s\"", compound)
	}

	if ret0 == ret {
		return ret
	}
	return ` #if (cpp || neko || php) ` + ret0 + ` #else ` + ret + " #end "
}

func constFloat64(lit ssa.Const, bits int, position string) string {
	var f float64
	var f32 float32
	f, _ /*f64ok*/ = exact.Float64Val(lit.Value)
	f32, _ /*f32ok*/ = exact.Float32Val(lit.Value)
	if bits == 32 {
		f = float64(f32)
	}
	haxeVal := pogo.FloatVal(lit.Value, bits, position)
	switch {
	case math.IsInf(f, +1):
		haxeVal = "Math.POSITIVE_INFINITY"
	case math.IsInf(f, -1):
		haxeVal = "Math.NEGATIVE_INFINITY"
	case math.IsNaN(f): // must come after infinity checks
		haxeVal = "Math.NaN"
	}
	return haxeVal
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
			return "String", haxeStringConst(lit.Value.String(), position)
		case *types.Slice:
			return "Slice", "Force.toUTF8slice(this._goroutine," + haxeStringConst(lit.Value.String(), position) + ")"
		default:
			pogo.LogError(position, "Haxe", fmt.Errorf("haxe.Const() internal error, unknown string type"))
		}
	case exact.Float:
		switch lit.Type().Underlying().(*types.Basic).Kind() {
		case types.Float32:
			return "Float", constFloat64(lit, 32, position)
		case types.Float64, types.UntypedFloat:
			return "Float", constFloat64(lit, 64, position)
		case types.Complex64:
			return "Complex", fmt.Sprintf("new Complex(%s,0)", pogo.FloatVal(lit.Value, 32, position))
		case types.Complex128:
			return "Complex", fmt.Sprintf("new Complex(%s,0)", pogo.FloatVal(lit.Value, 64, position))
		}
	case exact.Int:
		h, l := pogo.IntVal(lit.Value, position)
		switch lit.Type().Underlying().(*types.Basic).Kind() {
		case types.Int64:
			return "GOint64", fmt.Sprintf("Force.toInt64(GOint64.make(0x%x,0x%x))", uint32(h), uint32(l))
		case types.Uint64:
			return "GOint64", fmt.Sprintf("Force.toUint64(GOint64.make(0x%x,0x%x))", uint32(h), uint32(l))
		case types.Float32:
			return "Float", constFloat64(lit, 32, position)
		case types.Float64, types.UntypedFloat:
			return "Float", constFloat64(lit, 64, position)
		case types.Complex64:
			return "Complex", fmt.Sprintf("new Complex(%s,0)", pogo.FloatVal(lit.Value, 32, position))
		case types.Complex128:
			return "Complex", fmt.Sprintf("new Complex(%s,0)", pogo.FloatVal(lit.Value, 64, position))
		default:
			if h != 0 && h != -1 {
				pogo.LogWarning(position, "Haxe", fmt.Errorf("integer constant value > 32 bits : %v", lit.Value))
			}
			ret := ""
			switch lit.Type().Underlying().(*types.Basic).Kind() {
			case types.Uint, types.Uint32, types.Uintptr:
				q := uint32(l)
				ret = fmt.Sprintf(
					" #if js untyped __js__(\"0x%x\") #elseif php untyped __php__(\"0x%x\") #else 0x%x #end ",
					q, q, q)
			case types.Uint16:
				q := uint16(l)
				ret = fmt.Sprintf(" 0x%x ", q)
			case types.Uint8: // types.Byte
				q := uint8(l)
				ret = fmt.Sprintf(" 0x%x ", q)
			case types.Int, types.Int32, types.UntypedRune, types.UntypedInt: // types.Rune
				if l < 0 {
					ret = fmt.Sprintf("(%d)", int32(l))
				} else {
					ret = fmt.Sprintf("%d", int32(l))
				}
			case types.Int16:
				if l < 0 {
					ret = fmt.Sprintf("(%d)", int16(l))
				} else {
					ret = fmt.Sprintf("%d", int16(l))
				}
			case types.Int8:
				if l < 0 {
					ret = fmt.Sprintf("(%d)", int8(l))
				} else {
					ret = fmt.Sprintf("%d", int8(l))
				}
			case types.UnsafePointer:
				if l == 0 {
					return "Pointer", "null"
				}
				pogo.LogError(position, "Haxe", fmt.Errorf("unsafe pointers cannot be initialized in TARDISgo/Haxe to a non-zero value: %v", l))
			default:
				panic("haxe.Const() unhandled integer constant for: " +
					lit.Type().Underlying().(*types.Basic).String())
			}
			return "Int", ret
		}
	case exact.Unknown: // not sure we should ever get here!
		return "Dynamic", "null"
	case exact.Complex:
		realV, _ := exact.Float64Val(exact.Real(lit.Value))
		imagV, _ := exact.Float64Val(exact.Imag(lit.Value))
		switch lit.Type().Underlying().(*types.Basic).Kind() {
		case types.Complex64:
			return "Complex", fmt.Sprintf("new Complex(%g,%g)", float32(realV), float32(imagV))
		default:
			return "Complex", fmt.Sprintf("new Complex(%g,%g)", realV, imagV)
		}
	}
	pogo.LogError(position, "Haxe", fmt.Errorf("haxe.Const() internal error, unknown constant type: %v", lit.Value.Kind()))
	return "", ""
}

// only public Literals are created here, so that they can be used by Haxe callers of the Go code
func (l langType) NamedConst(packageName, objectName string, lit ssa.Const, position string) string {
	typ, rhs := l.Const(lit, position+":"+packageName+"."+objectName)
	return fmt.Sprintf("public static var %s:%s = %s;%s",
		l.LangName(packageName, objectName), typ, rhs, l.Comment(position))
}

func (l langType) Global(packageName, objectName string, glob ssa.Global, position string, isPublic bool) string {
	pub := "public " // all globals have to be public in Haxe terms
	obj := allocNewObject(glob.Type().Underlying().(*types.Pointer))
	return fmt.Sprintf("%sstatic var %s:Pointer=new Pointer(%s); %s",
		pub, l.LangName(packageName, objectName), obj, l.Comment(position))
}
