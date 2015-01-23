// Package hx provides pseudo-functions to interface to Haxe code, TARDIS Go re-writes these functions as Haxe.
// It follows that these functions act more like macros than functions, so some parameters must be constant strings.
//
// This package provides untyped access to Haxe, which is far from ideal.
// The next stage of development will be to provide a typed Go overlay - the gohaxelib approach (as yet incomplete).
// The final stage will be to use Haxe types directly...
//
package hx

import "unsafe"

// CallbackFunc returns the Haxe-callable form of a Go function, or
// if passed a string value, it gives the actual name of a Haxe function e.g. "Scheduler.timerEventHandler"
func CallbackFunc(function interface{}) interface{} { return nil }

// Resource loads a file resource that was added through the
// -resource host/file/path/a.dat@/target/file/path/b.dat haxe command line parameter;
// if the file resource does not exist, an empty slice is returned.
func Resource(s string) []byte { return []byte{} }

// Malloc allocates a memory Object and returns an unsafe pointer to it
func Malloc(size uintptr) unsafe.Pointer { return nil }

// IsNull returns if the haxe Dynamic variable is null
func IsNull(x uintptr) bool { return false }

// Null returns a haxe Dynamic null value
func Null() uintptr { return 0 }

// Complex provides a cast from haxe Dynamic type
func Complex(x uintptr) complex128 { return 0 + 0i }

// Int64 provides a cast from haxe Dynamic type
func Int64(x uintptr) int64 { return 0 }

// Code inserts the given constant Haxe code at this point.
// ifLogic = a constant string giving the logic for wrapping Haxe complie time condition, ignored if "": #if (ifLogic) ... #end
// resTyp = a constant string giving the Go name of the type of the data to be returned as an interface. "" if nothing is returned.
// code = must be a constant string containing a well-formed Haxe statement, probably terminated with a ";".
// args = whatever aguments are passed (as interfaces), typical haxe code to access the value of an argument is "_a[3].val".
// Try the Go code:
//   hx.Code("","trace('HAXE trace:',_a.itemAddr(0).load().val,_a.itemAddr(1).load().val);", 42,43)
func Code(ifLogic, code string, args ...interface{}) {}

// CodeIface - same as Code() but returns an interface.
func CodeIface(ifLogic, resTyp, code string, args ...interface{}) interface{} { return nil }

// CodeBool - same as Code() but returns a bool.
func CodeBool(ifLogic, code string, args ...interface{}) bool { return false }

// CodeInt - same as Code() but returns an int.
func CodeInt(ifLogic, code string, args ...interface{}) int { return 0 }

// CodeFloat - same as Code() but returns a float64.
func CodeFloat(ifLogic, code string, args ...interface{}) float64 { return 0.0 }

// CodeString - same as Code() but returns a string.
func CodeString(ifLogic, code string, args ...interface{}) string { return "" }

// CodeDynamic - same as Code() but returns a Dynamic (modeled as Haxe Dynamic in TARDIS Go, so can hold any Haxe object).
func CodeDynamic(ifLogic, code string, args ...interface{}) uintptr { return 0 }

// Call static Haxe functions, ifLogic, resTyp & target must be constant strings, nargs must be a constant number of arguments.

func Call(ifLogic, target string, nargs int, args ...interface{})                          {}
func CallIface(ifLogic, resTyp, target string, nargs int, args ...interface{}) interface{} { return nil }
func CallBool(ifLogic, target string, nargs int, args ...interface{}) bool                 { return false }
func CallInt(ifLogic, target string, nargs int, args ...interface{}) int                   { return 0 }
func CallFloat(ifLogic, target string, nargs int, args ...interface{}) float64             { return 0.0 }
func CallString(ifLogic, target string, nargs int, args ...interface{}) string             { return "" }
func CallDynamic(ifLogic, target string, nargs int, args ...interface{}) uintptr           { return 0 }

func New(ifLogic, target string, nargs int, args ...interface{}) uintptr { return 0 } // new haxe type

// Call Haxe instance functions, method must be a constant string, nargs must be a constant number of arguments.
// haxeType is required when the underlying haxe object is a simple type in compiled langs, like Date (int in cpp), otherwise ""
// ifLogic, resTyp & haxeType must be constant strings

func Meth(ifLogic string, object uintptr, haxeType string, method string, nargs int, args ...interface{}) {
}
func MethIface(ifLogic string, resTyp uintptr, object interface{}, haxeType string, nargs int, method string, args ...interface{}) interface{} {
	return nil
}
func MethBool(ifLogic string, object uintptr, haxeType string, method string, nargs int, args ...interface{}) bool {
	return false
}
func MethInt(ifLogic string, object uintptr, haxeType string, method string, nargs int, args ...interface{}) int {
	return 0
}
func MethFloat(ifLogic string, object uintptr, haxeType string, method string, nargs int, args ...interface{}) float64 {
	return 0.0
}
func MethString(ifLogic string, object uintptr, haxeType string, method string, nargs int, args ...interface{}) string {
	return ""
}
func MethDynamic(ifLogic string, object uintptr, haxeType string, method string, nargs int, args ...interface{}) uintptr {
	return 0
}

// Get a static Haxe value, ifLogic, resTyp & name must be constant strings.

func GetIface(ifLogic, resTyp, name string) interface{} { return nil }
func GetBool(ifLogic, name string) bool                 { return false }
func GetInt(ifLogic, name string) int                   { return 0 }
func GetFloat(ifLogic, name string) float64             { return 0.0 }
func GetString(ifLogic, name string) string             { return "" }
func GetDynamic(ifLogic, name string) uintptr           { return 0 }

// Set a static Haxe value, ifLogic, resTyp & name must be constant strings.

func SetIface(ifLogic, resTyp, name string, val interface{}) {} // TODO is this required?
func SetBool(ifLogic, name string, val bool)                 {}
func SetInt(ifLogic, name string, val int)                   {}
func SetFloat(ifLogic, name string, val float64)             {}
func SetString(ifLogic, name string, val string)             {}
func SetDynamic(ifLogic, name string, val uintptr)           {}

// Get a field value in a Haxe object, ifLogic, resTyp & name must be constant strings.

func FgetIface(ifLogic, resTyp string, object uintptr, haxeType string, name string) interface{} {
	return nil
}
func FgetBool(ifLogic string, object uintptr, haxeType string, name string) bool       { return false }
func FgetInt(ifLogic string, object uintptr, haxeType string, name string) int         { return 0 }
func FgetFloat(ifLogic string, object uintptr, haxeType string, name string) float64   { return 0.0 }
func FgetString(ifLogic string, object uintptr, haxeType string, name string) string   { return "" }
func FgetDynamic(ifLogic string, object uintptr, haxeType string, name string) uintptr { return 0 }

// Set a field value in a Haxe object, ifLogic, resTyp & name must be constant strings.

func FsetIface(ifLogic, resTyp string, object uintptr, haxeType string, name string, val interface{}) {
}                                                                                           // TODO is this required?
func FsetBool(ifLogic string, object uintptr, haxeType string, name string, val bool)       {}
func FsetInt(ifLogic string, object uintptr, haxeType string, name string, val int)         {}
func FsetFloat(ifLogic string, object uintptr, haxeType string, name string, val float64)   {}
func FsetString(ifLogic string, object uintptr, haxeType string, name string, val string)   {}
func FsetDynamic(ifLogic string, object uintptr, haxeType string, name string, val uintptr) {}
