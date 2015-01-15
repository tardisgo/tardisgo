package haxe

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tardisgo/tardisgo/pogo"
	"golang.org/x/tools/go/ssa"
)

func (l langType) hxPseudoFuncs(fnToCall string, args []ssa.Value, errorInfo string) string {
	//fmt.Println("DEBUG l.hxPseudoFuncs()", fnToCall, args, errorInfo)
	fnToCall = strings.TrimPrefix(fnToCall, "hx_")

	if fnToCall == "init" {
		return "" // no need to generate code for the go init function
	}

	if fnToCall == "RResource" {
		return "Slice.fromResource(" + l.IndirectValue(args[0], errorInfo) + ");"
	}

	if fnToCall == "IIsNNull" {
		return l.IndirectValue(args[0], errorInfo) + "==null;"
	}
	if fnToCall == "NNull" {
		return "null;"
	}

	if fnToCall == "CCallbackFFunc" {
		// NOTE there will be a preceeding MakeInterface call that is made redundant by this code
		if len(args) == 1 {
			goMI, ok := args[0].(*ssa.MakeInterface)
			if ok {
				goFn, ok := (*(goMI.Operands(nil)[0])).(*ssa.Function)
				if ok {
					return "new Interface(-1," + l.IndirectValue(args[0], errorInfo) + ".val.buildCallbackFn()); // Go_" + l.FuncName(goFn)
				}
				_, ok = (*(goMI.Operands(nil)[0])).(*ssa.MakeClosure)
				if ok {
					return "new Interface(-1," + l.IndirectValue(args[0], errorInfo) + ".val.buildCallbackFn());"
				}
				con, ok := (*(goMI.Operands(nil)[0])).(*ssa.Const)
				if ok {
					return "new Interface(-1," + tgoString(l.IndirectValue(con, errorInfo), errorInfo) + ");"
				}
			}
		}
		pogo.LogError(errorInfo, "Haxe", fmt.Errorf("hx.Func() argument is not a function constant"))
		return ""
	}

	argOff := 1 // because of the ifLogic
	wrapStart := ""
	wrapEnd := ""
	usesArgs := true

	ifLogic := l.IndirectValue(args[0], errorInfo)
	//fmt.Println("DEBUG:ifLogic=", ifLogic, "AT", errorInfo)
	ifLogic = tgoString(ifLogic, errorInfo)
	if len(ifLogic) > 0 {
		wrapStart = " #if (" + ifLogic + ") "
		defVal := "null"
		if strings.HasSuffix(fnToCall, "Bool") {
			defVal = "false"
		}
		if strings.HasSuffix(fnToCall, "Int") {
			defVal = "0"
		}
		if strings.HasSuffix(fnToCall, "Float") {
			defVal = "0.0"
		}
		if strings.HasSuffix(fnToCall, "String") {
			defVal = `""`
		}
		wrapEnd = " #else " + defVal + "; #end "
	}

	if strings.HasSuffix(fnToCall, "SString") && !strings.HasPrefix(fnToCall, "Fset") && !strings.HasPrefix(fnToCall, "Set") {
		wrapStart += " Force.fromHaxeString({"
		wrapEnd = "});" + wrapEnd
	}

	if strings.HasSuffix(fnToCall, "IIface") {
		argOff = 2
		wrapStart += "new Interface(TypeInfo.getId(\"" + tgoString(l.IndirectValue(args[1], errorInfo), errorInfo) + "\"),{"
		wrapEnd = "});" + wrapEnd
	}
	code := ""
	if strings.HasPrefix(fnToCall, "NNew") {
		code = "new "
	}
	code += strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`) // trim quotes if it has any
	if strings.HasPrefix(fnToCall, "CCall") ||
		strings.HasPrefix(fnToCall, "MMeth") || strings.HasPrefix(fnToCall, "NNew") {
		argOff++
		if strings.HasPrefix(fnToCall, "MMeth") {
			haxeType := tgoString(l.IndirectValue(args[argOff], errorInfo), errorInfo)
			if len(haxeType) > 0 {
				code = "cast(" + code + "," + haxeType + ")"
			}
			argOff++
			obj := l.IndirectValue(args[argOff], errorInfo)
			code += "." + strings.Trim(obj, `"`) + "("
			argOff++
		} else {
			code += "("
		}
		aLen, err := strconv.ParseUint(l.IndirectValue(args[argOff], errorInfo), 0, 64)
		if err != nil {
			code += " ERROR Go ParseUint on number of arguments to hx.Meth() or hx.Call() - " + err.Error() + "! "
		} else {
			if aLen == 0 {
				usesArgs = false
			}
			for i := uint64(0); i < aLen; i++ {
				if i > 0 {
					code += ","
				}
				code += fmt.Sprintf("Force.toHaxeParam({var b=_a.itemAddr(%d).load();b==null?null:b.val;})", i)
			}
		}
		code += ");"
	}
	if strings.HasPrefix(fnToCall, "GGet") {
		code += ";"
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "SSet") {
		argOff++
		code = code + "=" + l.IndirectValue(args[argOff], errorInfo) + ";"
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "FFget") {
		argOff++
		if l.IndirectValue(args[argOff], errorInfo) != `""` {
			code = "cast(" + code + "," + tgoString(l.IndirectValue(args[argOff], errorInfo), errorInfo) + ")"
		}
		code += "." + tgoString(l.IndirectValue(args[argOff+1], errorInfo), errorInfo) + "; "
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "FFset") {
		argOff++
		if l.IndirectValue(args[argOff], errorInfo) != `""` {
			code = "cast(" + code + "," + tgoString(l.IndirectValue(args[argOff], errorInfo), errorInfo) + ")"
		}
		code += "." + tgoString(l.IndirectValue(args[argOff+1], errorInfo), errorInfo) +
			"=Force.toHaxeParam(" + l.IndirectValue(args[argOff+2], errorInfo) + "); "
		usesArgs = false
	}

	ret := "{"
	if usesArgs {
		ret += "var _a=" + l.IndirectValue(args[argOff+1], errorInfo) + "; "
	}
	return ret + wrapStart + code + wrapEnd + " }"
}

func tgoString(s, errorInfo string) string {
	bits := strings.Split(s, `"`)
	if len(bits) < 2 {
		pogo.LogError(errorInfo, "Haxe", fmt.Errorf("hx.() argument is not a usable string constant"))
		return ""
	}
	return bits[1]
}
