package haxe

import (
	"fmt"
	"strconv"
	"strings"

	"code.google.com/p/go.tools/go/ssa"
)

func (l langType) hxPseudoFuncs(fnToCall string, args []ssa.Value, errorInfo string) string {
	if fnToCall == "hx_init" {
		return ""
	}
	fnToCall = strings.TrimPrefix(fnToCall, "hx_")
	argOff := 0
	wrapStart := ""
	wrapEnd := ""
	usesArgs := true
	if strings.HasSuffix(fnToCall, "Iface") {
		argOff = 1
		wrapStart = "new Interface(TypeInfo.getId(" + l.IndirectValue(args[0], errorInfo) + "),{"
		wrapEnd = "});"
	}
	code := strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`)
	if strings.HasPrefix(fnToCall, "Call") || strings.HasPrefix(fnToCall, "Meth") {
		argOff++
		if strings.HasPrefix(fnToCall, "Meth") {
			obj := l.IndirectValue(args[argOff], errorInfo)
			code = "#if (cpp || flash) " + code + "." + strings.Trim(obj, `"`) + "( #else Reflect.callMethod(" + code + "," +
				" Reflect.field(" + code + ", " + obj + "),[ #end " // prefix the code with the var to execute the method on
			//code = "Reflect.callMethod(" + code + ", #if cpp " + obj +
			//	" #else Reflect.field(" + code + ", " + obj + ") #end ,[" // prefix the code with the var to execute the method on
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
				code += fmt.Sprintf("_a[%d].val", i)
			}
		}
		if strings.HasPrefix(fnToCall, "Meth") {
			code += " #if !(cpp || flash) ] #end "
		}
		code += ");"
	}
	if strings.HasPrefix(fnToCall, "Get") {
		code += ";"
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "Set") {
		argOff++
		code = l.IndirectValue(args[argOff], errorInfo) + "=" + code + ";"
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "Fget") {
		argOff++

		code = "#if (cpp || flash) " + code + "." + strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`) + "; #else " +
			"Reflect.getProperty(" + code + "," + l.IndirectValue(args[argOff], errorInfo) + "); #end "
		usesArgs = false
	}
	if strings.HasPrefix(fnToCall, "Fset") {
		argOff++
		code = "#if (cpp || flash) " + code + "." + strings.Trim(l.IndirectValue(args[argOff], errorInfo), `"`) +
			"=" + l.IndirectValue(args[argOff+1], errorInfo) + "; #else " +
			"Reflect.setProperty(" + code + "," +
			l.IndirectValue(args[argOff], errorInfo) + "," + l.IndirectValue(args[argOff+1], errorInfo) + "); #end "
		usesArgs = false
	}

	ret := "{"
	if usesArgs {
		ret += "var _a=" + l.IndirectValue(args[argOff+1], errorInfo) + "; "
	}
	return ret + wrapStart + code + wrapEnd + " }"
}
