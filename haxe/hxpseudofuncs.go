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

	if fnToCall == "CallbackFunc" {
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
					return "new Interface(-1," + strings.Trim(l.IndirectValue(con, errorInfo), "\"") + ");"
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
	if len(ifLogic) > 2 && ifLogic[0] == '"' { // the empty string comes back as two double quotes
		ifLogic = strings.Trim(ifLogic, `"`)
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

	if strings.HasSuffix(fnToCall, "Iface") {
		argOff = 2
		wrapStart += "new Interface(TypeInfo.getId(" + l.IndirectValue(args[1], errorInfo) + "),{"
		wrapEnd = "});" + wrapEnd
	}
	code := ""
	if strings.HasPrefix(fnToCall, "New") {
		code = "new "
	}
	code += strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`)
	if strings.HasPrefix(fnToCall, "Call") || strings.HasPrefix(fnToCall, "Meth") || strings.HasPrefix(fnToCall, "New") {
		argOff++
		if strings.HasPrefix(fnToCall, "Meth") {
			haxeType := strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`)
			if len(haxeType) > 0 {
				code = "cast(" + code + "," + haxeType + ")"
			}
			argOff++
			obj := l.IndirectValue(args[argOff], errorInfo)
			code += "." + strings.Trim(obj, `"`) + "("
			// If in need of reflection:
			//code = "#if (cpp || flash) " + code + "." + strings.Trim(obj, `"`) + "( #else Reflect.callMethod(" + code + "," +
			//	" Reflect.field(" + code + ", " + obj + "),[ #end " // prefix the code with the var to execute the method on
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
				code += fmt.Sprintf("_a.itemAddr(%d).load().val", i)
			}
		}
		if strings.HasPrefix(fnToCall, "Meth") {
			// If in need of reflection:
			//code += " #if !(cpp || flash) ] #end "
		}
		code += ");"
	}
	if strings.HasPrefix(fnToCall, "Get") {
		code += ";"
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "Set") {
		argOff++
		code = code + "=" + l.IndirectValue(args[argOff], errorInfo) + ";"
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "Fget") {
		argOff++
		if l.IndirectValue(args[argOff], errorInfo) != `""` {
			code = "cast(" + code + "," + strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`) + ")"
		}
		code += "." + strings.Trim(l.IndirectValue(args[argOff+1], errorInfo), `"`) + "; "
		//code = "#if (cpp || flash) " + code + "." + strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`) + "; #else " +
		//	"Reflect.getProperty(" + code + "," + l.IndirectValue(args[argOff], errorInfo) + "); #end "
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "Fset") {
		argOff++
		if l.IndirectValue(args[argOff], errorInfo) != `""` {
			code = "cast(" + code + "," + strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`) + ")"
		}
		code += "." + strings.Trim(l.IndirectValue(args[argOff+1], errorInfo), `"`) +
			"=" + l.IndirectValue(args[argOff+2], errorInfo) + "; "
		//code = "#if (cpp || flash) " + code + "." + strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`) +
		//	"=" + l.IndirectValue(args[argOff+1], errorInfo) + "; #else " +
		//	"Reflect.setProperty(" + code + "," +
		//	l.IndirectValue(args[argOff], errorInfo) + "," + l.IndirectValue(args[argOff+1], errorInfo) + "); #end "
		usesArgs = false
	}

	ret := "{"
	if usesArgs {
		ret += "var _a=" + l.IndirectValue(args[argOff+1], errorInfo) + "; "
	}
	return ret + wrapStart + code + wrapEnd + " }"
}
