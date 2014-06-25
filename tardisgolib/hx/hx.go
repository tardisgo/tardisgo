// Package hx provides pseudo-functions to interface to Haxe code, TARDIS Go re-writes these functions as Haxe.
// It follows that these functions act more like macros than functions, so some parameters must be constant strings.
//
// This package provides untyped access to Haxe, which is far from ideal.
// The next stage of development will be to provide a typed Go overlay - the gohaxelib approach (as yet incomplete).
// The final stage will be to use Haxe types directly...
//
package hx

// Code inserts the given constant Haxe code at this point.
// resTyp = a string giving the Go name of the type of the data to be returned as an interface. "" if nothing is returned.
// code = must be a constant string containing a well-formed Haxe statement, probably terminated with a ";".
// args = whatever aguments are passed (as interfaces), typical haxe code to access the value of an argument is "_a[3].val".
// Try the Go code:
//   hx.Code("trace('HAXE trace:',_a[0].val,_a[1].val,_a[2].val,_a[3].val);", 11, 22, 33, 44)
func Code(code string, args ...interface{}) {}

// CodeIface - same as Code() but returns an interface.
func CodeIface(resTyp, code string, args ...interface{}) interface{} { return nil }

// CodeBool - same as Code() but returns a bool.
func CodeBool(code string, args ...interface{}) bool { return false }

// CodeInt - same as Code() but returns an int.
func CodeInt(code string, args ...interface{}) int { return 0 }

// CodeFloat - same as Code() but returns a float64.
func CodeFloat(code string, args ...interface{}) float64 { return 0.0 }

// CodeString - same as Code() but returns a string.
func CodeString(code string, args ...interface{}) string { return "" }

// CodeDynamic - same as Code() but returns a uintptr (modeled as Haxe Dynamic in TARDIS Go, so can hold any Haxe object).
func CodeDynamic(code string, args ...interface{}) uintptr { return 0 }

// Call static Haxe functions, target must be a constant string, nargs must be a constant number of arguments.

func Call(target string, nargs int, args ...interface{})                          {}
func CallIface(resTyp, target string, nargs int, args ...interface{}) interface{} { return nil }
func CallBool(target string, nargs int, args ...interface{}) bool                 { return false }
func CallInt(target string, nargs int, args ...interface{}) int                   { return 0 }
func CallFloat(target string, nargs int, args ...interface{}) float64             { return 0.0 }
func CallString(target string, nargs int, args ...interface{}) string             { return "" }
func CallDynamic(target string, nargs int, args ...interface{}) uintptr           { return 0 }

// Call instance Haxe functions, method must be a constant string, nargs must be a constant number of arguments.

func Meth(object uintptr, method string, nargs int, args ...interface{}) {}
func MethIface(resTyp, object uintptr, nargs int, method string, args ...interface{}) interface{} {
	return nil
}
func MethBool(object uintptr, method string, nargs int, args ...interface{}) bool       { return false }
func MethInt(object uintptr, method string, nargs int, args ...interface{}) int         { return 0 }
func MethFloat(object uintptr, method string, nargs int, args ...interface{}) float64   { return 0.0 }
func MethString(object uintptr, method string, nargs int, args ...interface{}) string   { return "" }
func MethDynamic(object uintptr, method string, nargs int, args ...interface{}) uintptr { return 0 }

// Get a static Haxe value, name must be a constant string.

func GetIface(resTyp, name string) interface{} { return nil }
func GetBool(name string) bool                 { return false }
func GetInt(name string) int                   { return 0 }
func GetFloat(name string) float64             { return 0.0 }
func GetString(name string) string             { return "" }
func GetDynamic(name string) uintptr           { return 0 }

// Set a static Haxe value, name must be a constant string.

func SetIface(resTyp, name string, val interface{}) {}
func SetBool(name string, val bool)                 {}
func SetInt(name string, val int)                   {}
func SetFloat(name string, val float64)             {}
func SetString(name string, val string)             {}
func SetDynamic(name string, val uintptr)           {}

// Get a field value in a Haxe object, name must be a constant string.

func FgetIface(resTyp string, object uintptr, name string) interface{} { return nil }
func FgetBool(object uintptr, name string) bool                        { return false }
func FgetInt(object uintptr, name string) int                          { return 0 }
func FgetFloat(object uintptr, name string) float64                    { return 0.0 }
func FgetString(object uintptr, name string) string                    { return "" }
func FgetDynamic(object uintptr, name string) uintptr                  { return 0 }

// Set a field value in a Haxe object, name must be a constant string.

func FsetIface(resTyp string, object uintptr, name string, val interface{}) {}
func FsetBool(object uintptr, name string, val bool)                        {}
func FsetInt(object uintptr, name string, val int)                          {}
func FsetFloat(object uintptr, name string, val float64)                    {}
func FsetString(object uintptr, name string, val string)                    {}
func FsetDynamic(object uintptr, name string, val uintptr)                  {}
