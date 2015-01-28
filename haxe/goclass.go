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

	main += "var _sfgr=new Go_haxegoruntime_init(gr,[]).run();\n" //haxegoruntime.init() NOTE can't use callFromHaxe() as that would call this fn
	main += "while(_sfgr._incomplete) Scheduler.runAll();\n"
	main += "var _sf=new Go_" + pkg.Object.Name() + `_init(gr,[]).run();` + "\n" //NOTE can't use callFromHaxe() as that would call this fn
	main += "while(_sf._incomplete) Scheduler.runAll();\n"
	main += "Scheduler.doneInit=true;\n"
	main += `Go.haxegoruntime_ZZiLLen.store_uint32('字'.length);` // value required by haxegoruntime to know what type of strings we have
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
	pos += "else return \"(invalid pogo.PosHash:\"+Std.string(pos)+\")\";\n}\n"

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
		if unicode.IsPrint(c) && c < unicode.MaxASCII && c != '"' && c != '\\' && !hadEsc {
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
		if unicode.IsPrint(c) && c < unicode.MaxASCII && c != '"' && c != '\\' {
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
	//sigBits := uint(53)
	//if bits == 32 {
	//	sigBits = 24
	//}
	f, _ /*f64ok*/ = exact.Float64Val(lit.Value)
	f32, _ /*f32ok*/ = exact.Float32Val(lit.Value)
	if bits == 32 {
		f = float64(f32)
	}
	haxeVal := pogo.FloatVal(lit.Value, bits, position)
	if math.IsInf(f, +1) {
		haxeVal = "Math.POSITIVE_INFINITY"
	} else {
		if math.IsInf(f, -1) {
			haxeVal = "Math.NEGATIVE_INFINITY"
		} else {
			if math.IsNaN(f) { // must come after infinity checks
				haxeVal = "Math.NaN"
			} else {
				// there is a problem with haxe constant processing for some floats
				// try to be as exact as the host can be ... but also concise
				//if float64(int64(f)) != f { // not a simple integer
				/*
					frac, exp := math.Frexp(f)
					intPart := int64(frac * float64(uint64(1)<<sigBits))
					expPart := exp - int(sigBits)
					if float64(intPart) == frac*float64(uint64(1)<<sigBits) &&
						expPart >= -1022 && expPart <= 1023 {
						//it is an integer in the correct range
						haxeVal = fmt.Sprintf("(%d*Math.pow(2,%d))", intPart, expPart) // NOTE: need the Math.pow to avoid haxe constant folding
					}
				*/
				/*
					val := exact.MakeFloat64(frac)
					num := exact.Num(val)
					den := exact.Denom(val)
					n64i, nok := exact.Int64Val(num)
					d64i, dok := exact.Int64Val(den)
					res := float64(n64i) * math.Pow(2, float64(exp)) / float64(d64i)
					if !math.IsNaN(res) && !math.IsInf(res, +1) && !math.IsInf(res, -1) { //drop through
						if nok && dok {
							nh, nl := pogo.IntVal(num, position)
							dh, dl := pogo.IntVal(den, position)
							n := fmt.Sprintf("%d", nl)
							if n64i < 0 {
								n = "(" + n + ")"
							}
							if nh != 0 && nh != -1 {
								n = fmt.Sprintf("GOint64.toFloat(Force.toInt64(GOint64.make(0x%x,0x%x)))", uint32(nh), uint32(nl))
							}
							if float64(d64i) == math.Pow(2, float64(exp)) {
								haxeVal = n // divisor and multiplier the same
							} else {
								d := fmt.Sprintf("%d", dl)
								if dh != 0 && dh != -1 {
									d = fmt.Sprintf("GOint64.toFloat(Force.toInt64(GOint64.make(0x%x,0x%x)))", uint32(dh), uint32(dl))
								}
								if n64i == 1 {
									n = "" // no point multiplying by 1
								} else {
									n = n + "*"
								}
								if d64i == 1 {
									d = "" // no point in dividing by 1
								} else {
									d = "/" + d
								}
								haxeVal = fmt.Sprintf("(%sMath.pow(2,%d)%s)", n, exp, d) // NOTE: need the Math.pow to avoid haxe constant folding
							}
						}
					}
				*/
				//}
			}
		}
	}
	return haxeVal
	/*
		bits64 := *(*uint64)(unsafe.Pointer(&f))
		bitVal := exact.MakeUint64(bits64)
		h, l := pogo.IntVal(bitVal, position)
		bitStr := fmt.Sprintf("GOint64.make(0x%x,0x%x)", uint32(h), uint32(l))
		return "Force.float64const(" + bitStr + "," + haxeVal + ")"
	*/
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
	pub := "public "                                                      // all globals have to be public in Haxe terms
	gTyp := glob.Type().Underlying().(*types.Pointer).Elem().Underlying() // globals are always pointers to an underlying element
	/*
		ptrTyp := "Pointer"
		//ltDesc := "Dynamic" // these values suitable for *types.Struct
		ltInit := "null"
		switch gTyp.(type) {
		case *types.Basic, *types.Pointer, *types.Interface, *types.Chan, *types.Map, *types.Signature:
			ptrTyp = "Pointer"
			//ltDesc = l.LangType(gTyp, false, position)
			ltInit = l.LangType(gTyp, true, position)
		case *types.Array:
			ptrTyp = "Pointer"
			//ltDesc = "Array<" + l.LangType(gTyp.(*types.Array).Elem().Underlying(), false, position) + ">"
			ltInit = l.LangType(gTyp, true, position)
		case *types.Slice:
			ptrTyp = "Pointer"
			//ltDesc = "Slice" // was: l.LangType(gTyp.(*types.Slice).Elem().Underlying(), false, position)
			ltInit = l.LangType(gTyp, true, position)
		case *types.Struct:
			ptrTyp = "Pointer"
			//ltDesc = "Dynamic" // TODO improve!
			ltInit = l.LangType(gTyp, true, position)
		}
		init := "new " + ptrTyp + "(" + ltInit + ")" // initialize basic types only
	*/
	//return fmt.Sprintf("%sstatic %s %s",
	//	pub, haxeVar(l.LangName(packageName, objectName), ptrTyp, init, position, "Global()"),
	//	l.Comment(position))
	return fmt.Sprintf("%sstatic var %s:Pointer=new Pointer(new Object(%d)); %s",
		pub, l.LangName(packageName, objectName), haxeStdSizes.Sizeof(gTyp),
		l.Comment(position))
}
