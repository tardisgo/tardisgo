// THIS IS NOT PRETTY, IT IS A WORK-IN-PROGRESS

// TODO separate this jumble of tests into a set of smaller ones

// This package should only test the core language functionality, all standard package tests moved elsewhere
package main

import (
	//"math"
	//"math/big" // does not currently complile - infinite loop
	//"bytes"
	"github.com/tardisgo/tardisgo/tardisgolib"
	//"runtime"
	//"strconv"
	//"strings"
	"sync" // keep these two for now...
	"sync/atomic"
	"unicode/utf8"

	// final one at end to match the constant declaration
	_ "github.com/tardisgo/tardisgo/golibruntime/sync"
	_ "github.com/tardisgo/tardisgo/golibruntime/sync/atomic"
)

const tardisgoLibRuntimePath = "github.com/tardisgo/tardisgo/golibruntime"

const tardisgoHeader = "/* TARDIS Go general header*/"

const tardisgoHaxeHeader = `class Util {
public static inline function comma(num:Interface):Interface
{
    var arr = Std.string(num.val).split(".");
    var str = "";
    while(arr[0].length > 3) 
    {
        str = "," + arr[0].substr(-3) + str;
        arr[0] = arr[0].substr(0, arr[0].length - 3);
    }
    return new Interface(TypeInfo.getId("string"),arr[0] + str) ; //cut because decimals do not always work!: + if(arr.length > 1) "." + arr[1];
}}
`

const ShowKnownErrors = false

func TEQ(l string, a, b interface{}) bool {
	if a != b {
		println("TEQ error " + l + " ")
		println(a)
		println(b)
		return false
	}
	return true
}

func TEQuint64(l string, a, b uint64) bool {
	if a != b {
		println("TEQui64 error " + l + " ")
		println("high a", uint(a>>32))
		println("low a", uint(a&0xFFFFFFFF))
		println("high b", uint(b>>32))
		println("low b", uint(b&0xFFFFFFFF))
		return false
	}
	return true
}
func TEQint64(l string, a, b int64) bool {
	if a != b {
		println("TEQi64 error " + l + " ")
		println("high a", int(a>>32))
		println("low a", int(a&0xFFFFFFFF))
		println("high b", int(b>>32))
		println("low b", int(b&0xFFFFFFFF))
		return false
	}
	return true
}
func TEQuint32(l string, a, b uint32) bool {
	if a != b {
		println("TEQui32 error " + l + " ")
		println(a)
		println(b)
		return false
	}
	return true
}
func TEQint32(l string, a, b int32) bool {
	if a != b {
		println("TEQi32 error " + l + " ")
		println(a)
		println(b)
		return false
	}
	return true
}
func TEQbyteSlice(l string, a, b []byte) bool {
	if len(a) != len(b) {
		println("TEQbyteSlice error "+l+" ", a, b)
		return false
	}
	ret := true
	for i := range a {
		if a[i] != b[i] {
			println("TEQbyteSlice error "+l+" ", a, b)
			ret = false
		}
	}
	return ret
}
func TEQruneSlice(l string, a, b []rune) bool {
	if len(a) != len(b) {
		println("TEQruneSlice error "+l+" ", a, b)
		return false
	}
	ret := true
	for i := range a {
		if a[i] != b[i] {
			println("TEQruneSlice error "+l+" ", a, b)
			ret = false
		}
	}
	return ret
}
func TEQintSlice(l string, a, b []int) bool {
	if len(a) != len(b) {
		println("TEQintSlice error "+l+" ", a, b)
		return false
	}
	ret := true
	for i := range a {
		if a[i] != b[i] {
			println("TEQintSlice error "+l+" ", a, b)
			ret = false
		}
	}
	return ret
}
func TEQfloat(l string, a, b, maxDif float64) bool {
	dif := a - b
	if dif < 0 {
		dif = -dif
	}
	if dif > maxDif {
		println("TEQfloat error " + l + " ")
		println(a)
		println(b)
		return false
	}
	return true
}

// CONSTANT TEST DATA
const Name string = "this is my name"
const ests bool = true
const Pi float64 = 3.14159265358979323846
const zero = 0.0 // untyped floating-point constant
const (
	size int = 1024
	eof      = -1 // untyped integer constant
)
const a, b, c = 3, 4, "foo" // a = 3, b = 4, c = "foo", untyped integer and string constants
const u, v float64 = 0, 3   // u = 0.0, v = 3.0
const (
	Sunday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Partyday
	numberOfDays // this constant is not exported
)
const ( // iota is reset to 0
	c0 = iota // c0 == 0
	c1 = iota // c1 == 1
	c2 = iota // c2 == 2
)
const (
	_a = 1 << iota // a == 1 (iota has been reset)
	_b = 1 << iota // b == 2
	_c = 1 << iota // c == 4
)
const (
	_u         = iota * 42 // u == 0     (untyped integer constant)
	_v float64 = iota * 42 // v == 42.0  (float64 constant)
	_w         = iota * 42 // w == 84    (untyped integer constant)
)
const _x = iota // x == 0 (iota has been reset)
const _y = iota // y == 0 (iota has been reset)
const (
	bit0, mask0 = 1 << iota, 1<<iota - 1 // bit0 == 1, mask0 == 0
	bit1, mask1                          // bit1 == 2, mask1 == 1
	_, _                                 // skips iota == 2
	bit3, mask3                          // bit3 == 8, mask3 == 7
)
const ren = '人'
const Θ float64 = 3 / 2  // Θ == 1.0   (type float64, 3/2 is integer division)
const Π float64 = 3 / 2. // Π == 1.5   (type float64, 3/2. is float division)
const d = 1 << 3.0       // d == 8     (untyped integer constant)
const e = 1.0 << 3       // e == 8     (untyped integer constant)
const h = "foo" > "bar"  // h == true  (untyped boolean constant)
const j = true           // j == true  (untyped boolean constant)
const k = 'w' + 1        // k == 'x'   (untyped rune constant)
const l = "hi"           // l == "hi"  (untyped string constant)
const m = string(k)      // m == "x"   (type string)

func testConst() {
	TEQ(tardisgolib.CPos(), Name, "this is my name")
	TEQ(tardisgolib.CPos(), ests, true)
	TEQfloat(tardisgolib.CPos(), Pi, 3.14159265358979323846, 0.00000000000001)
	TEQ(tardisgolib.CPos(), zero, 0.0) // untyped floating-point constant
	TEQ(tardisgolib.CPos(), size, 1024)
	TEQ(tardisgolib.CPos(), eof, -1) // untyped integer constant
	// a = 3, b = 4, c = "foo", untyped integer and string constants
	TEQ(tardisgolib.CPos(), a, 3)
	TEQ(tardisgolib.CPos(), b, 4)
	TEQ(tardisgolib.CPos(), c, "foo")
	// u = 0.0, v = 3.0
	TEQ(tardisgolib.CPos(), u, 0.0)
	TEQ(tardisgolib.CPos(), v, 3.0)
	TEQ(tardisgolib.CPos(), Sunday, 0)
	TEQ(tardisgolib.CPos(), Monday, 1)
	TEQ(tardisgolib.CPos(), Tuesday, 2)
	TEQ(tardisgolib.CPos(), Wednesday, 3)
	TEQ(tardisgolib.CPos(), Thursday, 4)
	TEQ(tardisgolib.CPos(), Friday, 5)
	TEQ(tardisgolib.CPos(), Partyday, 6)
	TEQ(tardisgolib.CPos(), numberOfDays, 7) // this constant is not exported
	TEQ(tardisgolib.CPos(), c0, 0)           // c0 == 0
	TEQ(tardisgolib.CPos(), c1, 1)           // c1 == 1
	TEQ(tardisgolib.CPos(), c2, 2)           // c2 == 2
	TEQ(tardisgolib.CPos(), _a, 1)           // a == 1 (iota has been reset)
	TEQ(tardisgolib.CPos(), _b, 2)           // b == 2
	TEQ(tardisgolib.CPos(), _c, 4)           // c == 4
	TEQ(tardisgolib.CPos(), _u, 0)           // u == 0     (untyped integer constant)
	TEQ(tardisgolib.CPos(), _v, 42.0)        // v == 42.0  (float64 constant)
	TEQ(tardisgolib.CPos(), _w, 84)          // w == 84    (untyped integer constant)
	TEQ(tardisgolib.CPos(), _x, 0)           // x == 0 (iota has been reset)
	TEQ(tardisgolib.CPos(), _y, 0)           // y == 0 (iota has been reset)
	TEQ(tardisgolib.CPos(), bit0, 1)
	TEQ(tardisgolib.CPos(), mask0, 0) // bit0 == 1, mask0 == 0
	TEQ(tardisgolib.CPos(), bit1, 2)
	TEQ(tardisgolib.CPos(), mask1, 1) // bit1 == 2, mask1 == 1
	//_, _                                 // skips iota == 2
	TEQ(tardisgolib.CPos(), bit3, 8)
	TEQ(tardisgolib.CPos(), mask3, 7) // bit3 == 8, mask3 == 7
	TEQ(tardisgolib.CPos(), ren, '人')
	TEQ(tardisgolib.CPos(), Θ, 1.0)  // Θ == 1.0   (type float64, 3/2 is integer division)
	TEQ(tardisgolib.CPos(), Π, 1.5)  // Π == 1.5   (type float64, 3/2. is float division)
	TEQ(tardisgolib.CPos(), d, 8)    // d == 8     (untyped integer constant)
	TEQ(tardisgolib.CPos(), e, 8)    // e == 8     (untyped integer constant)
	TEQ(tardisgolib.CPos(), h, true) // h == true  (untyped boolean constant)
	TEQ(tardisgolib.CPos(), j, true) // j == true  (untyped boolean constant)
	TEQ(tardisgolib.CPos(), k, 'x')  // k == 'x'   (untyped rune constant)
	TEQ(tardisgolib.CPos(), l, "hi") // l == "hi"  (untyped string constant)
	TEQ(tardisgolib.CPos(), m, "x")  // m == "x"   (type string)
}

var testUTFlength = "123456789"

func testUTF() {
	var (
		rA, rB, r  []rune
		uS, s1, s2 string
	)
	rA = []rune{0x767d, 0x9d6c, 0x7fd4}
	uS = string(rA) // "\u767d\u9d6c\u7fd4" == "白鵬翔"
	rB = []rune(uS)
	TEQruneSlice(tardisgolib.CPos(), rA, rB)

	s1 = "香港发生工厂班车砍人案12人受伤"
	r = []rune(s1)
	s2 = string(r)
	TEQ(tardisgolib.CPos(), s1, s2)

	TEQ(tardisgolib.CPos(), len(s1), 44)
	TEQ(tardisgolib.CPos(), len(testUTFlength), 9)

	hellø := "hellø"
	TEQ(tardisgolib.CPos(), string([]byte{'h', 'e', 'l', 'l', '\xc3', '\xb8'}), hellø)
	TEQbyteSlice(tardisgolib.CPos(), []byte("hellø"), []byte{'h', 'e', 'l', 'l', '\xc3', '\xb8'})

	TEQ(tardisgolib.CPos(), "ø", hellø[4:])
}

var TestInit = "init() ran OK"
var primes = [6]int{2, 3, 5, 7, 9, 2147483647}
var iFace interface{} = nil

func testInit() {
	TEQ(tardisgolib.CPos(), TestInit, "init() ran OK")
	TEQintSlice(tardisgolib.CPos(), primes[:], []int{2, 3, 5, 7, 9, 2147483647})
	TEQ(tardisgolib.CPos(), 9, primes[4]) // also testing array access with a constant index
	TEQ(tardisgolib.CPos(), nil, iFace)
}

var PublicStruct struct {
	a int
	b bool
	c string
	d float64
	e interface{}
	f [12]int
	g [6]string
	h [14]struct {
		x bool
		y [3]float64
		z [6]interface{}
	}
}

func testStruct() {
	var PrivateStruct struct {
		a int
		b bool
		c string
		d float64
		e interface{}
		f [12]int
		g [6]string
		h [14]struct {
			x bool
			y [3]float64
			z [6]interface{}
		}
	}
	// check that everything is equally initialized
	TEQ(tardisgolib.CPos(), PublicStruct.a, PrivateStruct.a)
	TEQ(tardisgolib.CPos(), PublicStruct.b, PrivateStruct.b)
	TEQ(tardisgolib.CPos(), PublicStruct.c, PrivateStruct.c)
	TEQ(tardisgolib.CPos(), PublicStruct.d, PrivateStruct.d)
	TEQ(tardisgolib.CPos(), PublicStruct.e, PrivateStruct.e)
	TEQintSlice(tardisgolib.CPos(), PublicStruct.f[:], PrivateStruct.f[:])
	PublicStruct.a = 42
	PrivateStruct.a = 42
	TEQ(tardisgolib.CPos(), PublicStruct.a, PrivateStruct.a)
	PublicStruct.c = Name
	PrivateStruct.c = Name
	TEQ(tardisgolib.CPos(), PublicStruct.c, PrivateStruct.c)
	for i := range PrivateStruct.h {
		for j := range PrivateStruct.h[i].y {
			PrivateStruct.h[i].y[j] = 42.0 * float64(i) * float64(j)
			PublicStruct.h[i].y[j] = 42.0 * float64(i) * float64(j)
			TEQfloat(tardisgolib.CPos(), PrivateStruct.h[i].y[j], PublicStruct.h[i].y[j], 1.0)
		}
	}
}
func Sqrt(x float64) float64 {
	z := x
	for i := 0; i < 1000; i++ {
		z -= (z*z - x) / (2.0 * x)
	}
	return z
}

func testFloat() { // and also slices!
	TEQfloat(tardisgolib.CPos(), Sqrt(1024), 32.0, 0.1)
	threeD := make([][][]float64, 10)
	for i := range threeD {
		threeD[i] = make([][]float64, 10)
		for j := range threeD[i] {
			threeD[i][j] = make([]float64, 10)
			for k := range threeD[i][j] {
				threeD[i][j][k] = float64(i) * float64(j) * float64(k)
				TEQfloat(tardisgolib.CPos(), threeD[i][j][k], float64(i)*float64(j)*float64(k), 0.1)
			}
		}
	}
	// TODO add more here
}

func noCaller() float64 { // this should be removed by a good target compiler...
	U_ := Sqrt(float64(64))
	return U_
}

//var aPtr *int // TODO this should generate an error

func twoRets(x int) (a int, b string) {
	return 42 * x, "forty-two"
}

func testMultiRet() {
	r1, r2 := twoRets(1)
	TEQ(tardisgolib.CPos(), r1, 42)
	TEQ(tardisgolib.CPos(), r2, "forty-two")
}

func testAppend() {
	s0 := []int{0, 0}
	s1 := append(s0, 2) // append a single element     s1 == []int{0, 0, 2}
	TEQintSlice(tardisgolib.CPos(), []int{0, 0, 2}, s1)
	s2 := append(s1, 3, 5, 7) // append multiple elements    s2 == []int{0, 0, 2, 3, 5, 7}
	TEQintSlice(tardisgolib.CPos(), []int{0, 0, 2, 3, 5, 7}, s2)
	s3 := append(s2, s0...) // append a slice              s3 == []int{0, 0, 2, 3, 5, 7, 0, 0}
	TEQintSlice(tardisgolib.CPos(), []int{0, 0, 2, 3, 5, 7, 0, 0}, s3)
	var t []interface{}
	t = append(t, 42, 3.1415, "foo", nil) //       				 t == []interface{}{42, 3.1415, "foo"}
	TEQ(tardisgolib.CPos(), t[0], 42)
	TEQ(tardisgolib.CPos(), t[1], 3.1415)
	TEQ(tardisgolib.CPos(), t[2], "foo")
	TEQ(tardisgolib.CPos(), t[3], nil)

	var b []byte
	b = append(b, "bar"...)
	TEQbyteSlice(tardisgolib.CPos(), b, []byte{'b', 'a', 'r'})
}

func testHeader() {
	if tardisgolib.Host() == "Haxe" { // test of "pogoHeaderHaxe"
		//TODO implement a way to do this properly, probably using _package
		//TEQ(tardisgolib.CPos(), pg.C1("Util.comma", 75840032), string("75,840,032"))
		//TEQ(tardisgolib.CPos(), pg.C1("Util.comma", 300012301), string("300,012,301"))
	}
}

func testCopy() {
	var a = [...]int{0, 1, 2, 3, 4, 5, 6, 7}
	var s = make([]int, 6)
	n1 := copy(s, a[0:]) // n1 == 6, s == []int{0, 1, 2, 3, 4, 5}
	TEQ(tardisgolib.CPos(), n1, 6)
	TEQintSlice(tardisgolib.CPos(), s, []int{0, 1, 2, 3, 4, 5})
	n2 := copy(s, s[2:]) // n2 == 4, s == []int{2, 3, 4, 5, 4, 5}
	TEQ(tardisgolib.CPos(), n2, 4)
	TEQintSlice(tardisgolib.CPos(), s, []int{2, 3, 4, 5, 4, 5})
	var b = make([]byte, 5)
	n3 := copy(b, "Hello, World!") // n3 == 5, b == []byte("Hello")
	TEQ(tardisgolib.CPos(), n3, 5)
	TEQbyteSlice(tardisgolib.CPos(), b, []byte("Hello"))
}

func testInFuncPtr() { // there is no way to stop this use of pointers...
	var ss = 12
	var ssa = &ss
	TEQ(tardisgolib.CPos(), *ssa, 12)
}

func testCallByValue(a struct{ b int }, x [10]int, y []int, z int) {
	a.b = 42
	x[0] = 43
	y[0] = 44
	z = 45
}

func testCallByReference(a *struct{ b int }, x *[10]int, y []int, z *int) {
	a.b = 46
	x[0] = 47
	y[0] = 48
	*z = 49
}
func testTweakFloatByReference(i *float64) {
	if *i == 0 {
		*i = 0
	} else {
		*i = Sqrt(*i)
	}
}
func testCallBy() {
	var a struct {
		b int
	}
	var x [10]int
	var y []int = make([]int, 1)
	var z int
	testCallByValue(a, x, y, z)
	TEQ(tardisgolib.CPos(), a.b, 0)
	TEQ(tardisgolib.CPos(), x[0], 0)
	TEQ(tardisgolib.CPos(), y[0], 44)
	TEQ(tardisgolib.CPos(), z, 0)

	testCallByReference(&a, &x, y, &z)
	TEQ(tardisgolib.CPos(), a.b, 46)
	TEQ(tardisgolib.CPos(), x[0], 47)
	TEQ(tardisgolib.CPos(), y[0], 48)
	TEQ(tardisgolib.CPos(), z, 49)

	var xx [10]float64
	for i := range x {
		xx[i] = float64(i * i)
		testTweakFloatByReference(&xx[i])
		TEQfloat(tardisgolib.CPos(), xx[i], float64(i), 0.1)
	}
}

func testMap() { // and map-like constucts
	// vowels[ch] is true if ch is a vowel
	vowels := [128]bool{'a': true, 'e': true, 'i': true, 'o': true, 'u': true, 'y': true}
	for k, v := range vowels {
		switch k {
		case 'a', 'e', 'i', 'o', 'u', 'y':
			TEQ(tardisgolib.CPos(), true, v)
		default:
			TEQ(tardisgolib.CPos(), false, v)
		}
	}

	filter := [10]float64{-1, 4: -0.1, -0.1, 9: -1}
	TEQfloat(tardisgolib.CPos(), filter[5], -0.1, 0.01)

	// frequencies in Hz for equal-tempered scale (A4 = 440Hz)
	noteFrequency := map[string]float64{
		"C0": 16.35, "D0": 18.35, "E0": 20.60, "F0": 21.83,
		"G0": 24.50, "A0": 27.50, "B0": 30.87,
	}
	noteFrequency["Test"] = 42.42
	TEQ(tardisgolib.CPos(), len(noteFrequency), 8)
	for k, v := range noteFrequency {
		r := 0.0
		switch k {
		case "C0":
			r = 16.35
		case "D0":
			r = 18.35
		case "E0":
			r = 20.60
		case "F0":
			r = 21.83
		case "G0":
			r = 24.50
		case "A0":
			r = 27.50
		case "B0":
			r = 30.87
		case "Test":
			r = 42.42
		default:
			r = -1
		}
		if !TEQfloat(tardisgolib.CPos()+" Value itterator in map", v, r, 0.01) {
			break
		}
	}
	x, isok := noteFrequency["Test"]
	TEQfloat(tardisgolib.CPos(), 42.42, x, 0.01)
	TEQ(tardisgolib.CPos(), true, isok)
	_, notok := noteFrequency["notHere"]
	TEQ(tardisgolib.CPos(), false, notok)
	delete(noteFrequency, "Test")
	_, isok = noteFrequency["Test"]
	TEQ(tardisgolib.CPos(), false, isok)
}

type MyFloat float64
type MyFloat2 MyFloat

var namedGlobal MyFloat

type IntArray [8]int

type (
	Point struct {
		x, y float64
	}
	Polar Point
)

var myPolar Polar

func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

func (mf *MyFloat) set42() {
	*mf = 42
}

func (f MyFloat2) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}
func (ia *IntArray) set42() {
	for i := range ia {
		ia[i] = 42
	}
}
func (p Polar) BearVal() bool {
	return p.x == p.y
}

// from the language spec section Method Values
type T struct {
	a int
}

func (tv T) Mv(a int) int          { return a } // value receiver
func (tp *T) Mp(f float32) float32 { return f } // pointer receiver

var t T
var pt *T

func testNamed() {
	var ia IntArray
	for i := range ia {
		ia[i] = i
	}
	TEQintSlice(tardisgolib.CPos(), ia[:], []int{0, 1, 2, 3, 4, 5, 6, 7})
	var namedLocal MyFloat = 41.42
	namedGlobal = 42.42
	namedLocal += 1.0
	TEQfloat(tardisgolib.CPos(), float64(namedGlobal), float64(namedLocal), 0.0002)
	myPolar.x = 11.11
	myPolar.y = 10.11
	myPolar.y++
	TEQfloat(tardisgolib.CPos(), float64(myPolar.x), float64(myPolar.y), 0.0002)
	// method expression tests...
	TEQ(tardisgolib.CPos(), myPolar.BearVal(), true)
	f := MyFloat(-555)
	g := MyFloat2(-555)
	TEQfloat(tardisgolib.CPos(), f.Abs(), g.Abs(), 0.0002)
	ia.set42()
	f.set42()
	TEQfloat(tardisgolib.CPos(), float64(ia[3]), float64(f), 0.0002)

	// from the language spec section on method values (requires ssa.MakeClosure instruction)
	f1 := t.Mv
	TEQ(tardisgolib.CPos(), f1(7), t.Mv(7))
	pt = &t
	f2 := pt.Mp
	TEQ(tardisgolib.CPos(), f2(7), pt.Mp(7))
	f3 := pt.Mv
	TEQ(tardisgolib.CPos(), f3(7), (*pt).Mv(7))
	f4 := t.Mp
	TEQ(tardisgolib.CPos(), f4(7), (&t).Mp(7))

	// more from the language spec on Method expressions

	TEQ(tardisgolib.CPos(), t.Mv(7), T.Mv(t, 7))
	TEQ(tardisgolib.CPos(), t.Mv(7), (T).Mv(t, 7))

	f1a := T.Mv
	TEQ(tardisgolib.CPos(), t.Mv(7), f1a(t, 7))
	f2a := (T).Mv
	TEQ(tardisgolib.CPos(), t.Mv(7), f2a(t, 7))

}

var hypot1 = func(x, y float64) float64 {
	return Sqrt(x*x + y*y)
}

func testFuncPtr() {
	var hypot2 = func(x, y float64) float64 {
		return Sqrt(x*x + y*y)
	}
	TEQfloat(tardisgolib.CPos(), hypot1(3, 4), hypot2(3, 4), 0.2)
}

var int64_max int64 = 0x7FFFFFFFFFFFFFFF
var int32_max int32 = 0x7FFFFFFF
var int16_max int16 = 0x7FFF
var int8_max int8 = 0x7F
var uint64_max uint64 = 0xFFFFFFFFFFFFFFFF
var uint32_max uint32 = 0xFFFFFFFF // This value too big and too ambiguous for cpp when held as an Int...
var uint16_max uint16 = 0xFFFF
var uint8_max uint8 = 0xFF
var int8_mostNeg int8 = -128
var int16_mostNeg int16 = -32768
var int32_mostNeg int32 = -2147483648
var int64_mostNeg int64 = -9223372036854775808

var five int = 5
var three int = 3

var uint64Global uint64
var uint64GlobalArray [4]uint64

func testIntOverflow() { //TODO add int64
	TEQ(tardisgolib.CPos()+" int16 overflow test 1", int16_max+1, int16_mostNeg)
	TEQ(tardisgolib.CPos()+" int8 overflow test 1", int8_max+1, int8_mostNeg)
	TEQ(tardisgolib.CPos()+" uint16 overflow test 2", uint16(uint16_max+1), uint16(0))
	TEQ(tardisgolib.CPos()+" uint8 overflow test 2", uint8(uint8_max+1), uint8(0))
	TEQ(tardisgolib.CPos()+" int8 overflow test 3", int8(int8_mostNeg-1), int8_max)
	TEQ(tardisgolib.CPos()+" int16 overflow test 3", int16(int16_mostNeg-1), int16_max)

	TEQint64(tardisgolib.CPos()+" int64 overflow test 1 ", int64_max+1, int64_mostNeg)
	TEQint32(tardisgolib.CPos()+" int32 overflow test 1 ", int32_max+1, int32_mostNeg)
	TEQuint64(tardisgolib.CPos()+" uint64 overflow test 2 ", uint64(uint64_max+1), uint64(0))
	TEQuint32(tardisgolib.CPos()+" uint32 overflow test 2 ", uint32(uint32_max+1), uint32(0))

	TEQint32(tardisgolib.CPos()+" int32 overflow test 3 ", int32(int32_mostNeg-int32(1)), int32_max)
	TEQint64(tardisgolib.CPos()+" int64 overflow test 3 ", int64(int64_mostNeg-int64(1)), int64_max)

	//Math.imul test case at https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Math/imul
	TEQuint32(tardisgolib.CPos(), ((uint32_max << 1) * 5), uint32_max+1-10)
	TEQuint32(tardisgolib.CPos(), (((uint32_max << 1) + 1) * 5), uint32_max+1-5)

	/* from Go spec:
	   For two integer values x and y, the integer quotient q = x / y and remainder r = x % y satisfy the following relationships:

	   x = q*y + r  and  |r| < |y|
	   with x / y truncated towards zero ("truncated division").

	    x     y     x / y     x % y
	    5     3       1         2
	   -5     3      -1        -2
	    5    -3      -1         2
	   -5    -3       1        -2

	*/

	TEQ(tardisgolib.CPos(), five/three, 1)
	TEQ(tardisgolib.CPos(), five%three, 2)
	TEQ(tardisgolib.CPos(), (-five)/three, -1)
	TEQ(tardisgolib.CPos(), (-five)%three, -2)
	TEQ(tardisgolib.CPos(), five/(-three), -1)
	TEQ(tardisgolib.CPos(), five%(-three), 2)
	TEQ(tardisgolib.CPos(), (-five)/(-three), 1)
	TEQ(tardisgolib.CPos(), (-five)%(-three), -2)

	TEQint64(tardisgolib.CPos(), int64(five)/int64(three), int64(1))
	TEQint64(tardisgolib.CPos(), int64(five)%int64(three), int64(2))
	TEQint64(tardisgolib.CPos(), int64(-five)/int64(three), int64(-1))
	TEQint64(tardisgolib.CPos(), int64(-five)%int64(three), int64(-2))
	TEQint64(tardisgolib.CPos(), int64(five)/int64(-three), int64(-1))
	TEQint64(tardisgolib.CPos(), int64(five)%int64(-three), int64(2))
	TEQint64(tardisgolib.CPos(), int64(-five)/int64(-three), int64(1))
	TEQint64(tardisgolib.CPos(), int64(-five)%int64(-three), int64(-2))

	/*
		As an exception to this rule, if the dividend x is the most negative value for the int type of x,
		the quotient q = x / -1 is equal to x (and r = 0).
	*/
	TEQint64(tardisgolib.CPos()+" int64 div special case", int64_mostNeg/int64(-1), int64_mostNeg)
	TEQint64(tardisgolib.CPos()+" int64 mod special case", int64_mostNeg%int64(-1), 0)
	TEQint32(tardisgolib.CPos()+" int32 div special case", int32(int32_mostNeg/int32(-1)), int32_mostNeg)
	TEQint32(tardisgolib.CPos()+" int32 mod special case", int32_mostNeg%int32(-1), 0)
	if int16(int16_mostNeg/int16(-1)) != int16_mostNeg {
		println(tardisgolib.CPos() + " int16 div special case")
	}
	if int16(int16_mostNeg%int16(-1)) != int16(0) {
		println(tardisgolib.CPos() + " int16 mod special case")
	}
	if int8(int8_mostNeg/int8(-1)) != int8_mostNeg {
		println(tardisgolib.CPos() + " int8 div special case")
	}
	if int8(int8_mostNeg%int8(-1)) != int8(0) {
		println(tardisgolib.CPos() + " int8 mod special case")
	}

	/*THESE VALUES ARE NOT IN THE SPEC, SO UNTESTED
	if uint64(int64_mostNeg)/0xFFFFFFFFFFFFFFFF == uint64(int64_mostNeg) {
		println(tardisgolib.CPos() + " uint64 div special case")
	}
	if uint64(int64_mostNeg)%0xFFFFFFFFFFFFFFFF == uint64(0) {
		println(tardisgolib.CPos() + " uint64 mod special case")
	}
	if uint32(int32_mostNeg)/0xFFFFFFFF == uint32(int32_mostNeg) {
		println(tardisgolib.CPos() + " uint32 div special case")
	}
	if uint32(int32_mostNeg)%0xFFFFFFFF == uint32(0) {
		println(tardisgolib.CPos() + " uint32 mod special case")
	}
	if uint16(int16_mostNeg)/0xFFFF == uint16(int16_mostNeg) {
		println(tardisgolib.CPos() + " uint16 div special case")
	}
	if uint16(int16_mostNeg)%0xFFFF == uint16(0) {
		println(tardisgolib.CPos() + " uint16 mod special case")
	}
	if uint8(int8_mostNeg)/0xFF == uint8(int8_mostNeg) {
		println(tardisgolib.CPos() + " uint8 div special case")
	}
	if uint8(int8_mostNeg)%0xFF == uint8(0) {
		println(tardisgolib.CPos() + " uint8 mod special case")
	}
	*/

	//TODO more tests for unsigned comparisons, need to check all possibilities are covered
	TEQ(tardisgolib.CPos(), uint8(int8_mostNeg) > uint8(0), true)
	TEQ(tardisgolib.CPos(), uint8(int8_mostNeg) < uint8(0), false)
	TEQ(tardisgolib.CPos(), uint16(int16_mostNeg) > uint16(0), true)
	TEQ(tardisgolib.CPos(), uint16(int16_mostNeg) < uint16(0), false)

	TEQ(tardisgolib.CPos(), uint32(int32_mostNeg) > uint32(0), true)
	TEQ(tardisgolib.CPos(), uint32(int32_mostNeg) < uint32(0), false)
	TEQ(tardisgolib.CPos()+" uint64(int64_mostNeg) > uint64(0) ", uint64(int64_mostNeg) > uint64(0), true)
	TEQ(tardisgolib.CPos(), uint64(int64_mostNeg) < uint64(0), false)

	//TEQint64(tardisgolib.CPos(), int64(int64_mostNeg), int64(uint64(0x8000000000000000)))
	//println(float64(int64_mostNeg))
	//println(int64_mostNeg)
	uint64Global = uint64(int64_mostNeg)
	TEQuint64(tardisgolib.CPos(), uint64Global, uint64(0x8000000000000000))

	for i := range uint64GlobalArray {
		uint64GlobalArray[i] = uint64(int64_mostNeg)
		TEQ(tardisgolib.CPos()+" uint64(int64_mostNeg) > uint64(0) [Array] ", uint64(uint64GlobalArray[i]) > uint64(0), true)
	}

	// TODO test for equality too & check these constants are not being resolved by the compiler, rather than genereating tests!
	TEQ(tardisgolib.CPos(), uint8(int8_mostNeg)-uint8(42) > uint8(0), true)
	TEQ(tardisgolib.CPos(), uint8(int8_mostNeg)-uint8(42) < uint8(three), false)
	TEQ(tardisgolib.CPos(), uint16(int16_mostNeg)-uint16(42) > uint16(0), true)
	TEQ(tardisgolib.CPos(), uint16(0xffff)-uint16(five) < uint16(three), false)
	TEQ(tardisgolib.CPos(), uint32(0xffffffff)-uint32(five) > uint32(0), true)
	TEQ(tardisgolib.CPos(), uint32(0xffffffff)-uint32(five) < uint32(three), false)
	TEQ(tardisgolib.CPos(), uint64(0xffffffffffffffff)-uint64(five) > uint64(0), true)
	TEQ(tardisgolib.CPos(), uint64(0xffffffffffffffff)-uint64(five) < uint64(three), false)
	TEQ(tardisgolib.CPos(), uint8(0xff) > uint8(0xfe)-uint8(five), true)
	TEQ(tardisgolib.CPos(), uint8(five) < uint8(three), false)
	TEQ(tardisgolib.CPos(), uint16(0xffff) > uint16(0xfffe)-uint16(five), true)
	TEQ(tardisgolib.CPos(), uint16(10000)-uint16(five) < uint16(1000), false)
	TEQ(tardisgolib.CPos(), uint32(0xffffffff) > uint32(0xfffffffe)-uint32(five), true)
	TEQ(tardisgolib.CPos(), uint32(12)-uint32(five) < uint32(three), false)
	TEQ(tardisgolib.CPos(), uint64(0xffffffffffffffff) > uint64(0xfffffffffffffffe)-uint64(five), true)
	TEQ(tardisgolib.CPos(), uint64(12)-uint64(five) < uint64(three), false)

	// test Float / Int64 conversions
	fiveI64 := int64(five)
	TEQfloat(tardisgolib.CPos(), float64(fiveI64), 5.0, 0.1)
	TEQfloat(tardisgolib.CPos(), float64(int32_mostNeg), float64(-2147483648.0), 0.1)

	TEQfloat(tardisgolib.CPos()+" PHP error",
		float64(int64_mostNeg/int64(100000)), float64(int64(-9223372036854775808)/int64(100000)), float64(1.0))
	TEQfloat(tardisgolib.CPos()+" PHP error ",
		float64(int64_max/200), float64(int64(0x7fffffffffffffff)/200), float64(10.0))
	TEQfloat(tardisgolib.CPos(), float64(int64_mostNeg+1), float64(int64(-9223372036854775808+1)), float64(2000.0))
	TEQfloat(tardisgolib.CPos(), float64(int64_mostNeg), float64(int64(-9223372036854775808)), float64(2000.0))
	TEQfloat(tardisgolib.CPos(), float64(uint64Global), float64(int64(0x7fffffffffffffff)), float64(2000.0))
	uint64Global = 0xFFFFFFFFFFFFFFFF
	TEQfloat(tardisgolib.CPos(), float64(uint64Global), float64(uint64(0xffffffffffffffff)), float64(2000.0))

	// tests below removed to avoid also loading the math package
	//TEQint64(tardisgolib.CPos()+" NaN ->int64 conversion", int64(math.NaN()), -9223372036854775808)
	//TEQuint64(tardisgolib.CPos()+" NaN ->uint64 conversion (error on php)", uint64(math.NaN()), 9223372036854775808)

	myPi := float64(7)
	myPi64 := int64(myPi)
	myPu64 := uint64(myPi)
	limit := float64(1 << 52)
	for myPi < limit {
		a := TEQint64(tardisgolib.CPos()+" +ve float->int64 conversion  ", int64(myPi), myPi64)
		b := TEQint64(tardisgolib.CPos()+" -ve float->int64 conversion  ", int64(-myPi), -myPi64)
		c := TEQuint64(tardisgolib.CPos()+" float->uint64 conversion  ", uint64(myPi), myPu64)
		if a == false || b == false || c == false {
			break
		}
		myPi *= myPi
		myPi64 *= myPi64
		myPu64 *= myPu64
	}
}

func testSlices() {
	// from the Go tour...
	p := []int{2, 3, 5, 7, 11, 13}
	TEQintSlice(tardisgolib.CPos(), p[1:4], []int{3, 5, 7})
	TEQintSlice(tardisgolib.CPos(), p[:3], []int{2, 3, 5})
	TEQintSlice(tardisgolib.CPos(), p[4:], []int{11, 13})

	a := make([]int, 5)
	TEQintSlice(tardisgolib.CPos(), a, []int{0, 0, 0, 0, 0})
	TEQ(tardisgolib.CPos(), len(a), 5)
	TEQ(tardisgolib.CPos(), cap(a), 5)
	b := make([]int, 0, 5)
	TEQintSlice(tardisgolib.CPos(), b, []int{})
	TEQ(tardisgolib.CPos(), len(b), 0)
	TEQ(tardisgolib.CPos(), cap(b), 5)
	c := b[:2]
	TEQintSlice(tardisgolib.CPos(), c, []int{0, 0})
	TEQ(tardisgolib.CPos(), len(c), 2)
	TEQ(tardisgolib.CPos(), cap(c), 5)
	d := c[2:5]
	TEQintSlice(tardisgolib.CPos(), d, []int{0, 0, 0})
	TEQ(tardisgolib.CPos(), len(d), 3)
	TEQ(tardisgolib.CPos(), cap(d), 3)

	var z []int
	TEQ(tardisgolib.CPos(), len(z), 0)
	TEQ(tardisgolib.CPos(), cap(z), 0)
	TEQ(tardisgolib.CPos(), z == nil, true)

}

func testUTF8() {
	b := []byte("Hello, 世界")
	r, size := utf8.DecodeLastRune(b)
	TEQ(tardisgolib.CPos(), '界', r)
	TEQ(tardisgolib.CPos(), size, 3)
	b = b[:len(b)-size]
	r, size = utf8.DecodeLastRune(b)
	TEQ(tardisgolib.CPos(), '世', r)
	TEQ(tardisgolib.CPos(), size, 3)
	b = b[:len(b)-size]
	r, size = utf8.DecodeLastRune(b)
	TEQ(tardisgolib.CPos(), ' ', r)
	TEQ(tardisgolib.CPos(), size, 1)

	str := "Hello, 世界"
	r, size = utf8.DecodeLastRuneInString(str)
	TEQ(tardisgolib.CPos(), '界', r)
	TEQ(tardisgolib.CPos(), size, 3)
	str = str[:len(str)-size]
	r, size = utf8.DecodeLastRuneInString(str)
	TEQ(tardisgolib.CPos(), '世', r)
	TEQ(tardisgolib.CPos(), size, 3)
	str = str[:len(str)-size]
	r, size = utf8.DecodeLastRuneInString(str)
	TEQ(tardisgolib.CPos(), ' ', r)
	TEQ(tardisgolib.CPos(), size, 1)

	ru := '世'
	buf := make([]byte, 3)
	n := utf8.EncodeRune(buf, ru)
	TEQ(tardisgolib.CPos(), n, 3)
	TEQbyteSlice(tardisgolib.CPos(), buf, []byte{228, 184, 150})

	buf = []byte{228, 184, 150} // 世
	TEQ(tardisgolib.CPos(), true, utf8.FullRune(buf))
	TEQ(tardisgolib.CPos(), false, utf8.FullRune(buf[:2]))

	str = "世"
	TEQ(tardisgolib.CPos(), true, utf8.FullRuneInString(str))
	if ShowKnownErrors || tardisgolib.Zilen() == 3 {
		TEQ(tardisgolib.CPos()+" NOTE: known error handling incorrect strings on UTF16 platforms", false, utf8.FullRuneInString(str[:2]))
	}
	buf = []byte("Hello, 世界")
	TEQ(tardisgolib.CPos(), 13, len(buf))
	TEQ(tardisgolib.CPos(), 9, utf8.RuneCount(buf))

	str = "Hello, 世界"
	TEQ(tardisgolib.CPos(), 13, len(str))
	TEQ(tardisgolib.CPos(), 9, utf8.RuneCountInString(str))

	TEQ(tardisgolib.CPos(), 1, utf8.RuneLen('a'))
	TEQ(tardisgolib.CPos(), 3, utf8.RuneLen('界'))

	buf = []byte("a界")
	TEQ(tardisgolib.CPos(), true, utf8.RuneStart(buf[0]))
	TEQ(tardisgolib.CPos(), true, utf8.RuneStart(buf[1]))
	TEQ(tardisgolib.CPos(), false, utf8.RuneStart(buf[2]))

	valid := []byte("Hello, 世界")
	invalid := []byte{0xff, 0xfe, 0xfd}
	TEQ(tardisgolib.CPos(), true, utf8.Valid(valid))
	TEQ(tardisgolib.CPos(), false, utf8.Valid(invalid))

	valid_rune := 'a'
	invalid_rune := rune(0xfffffff)
	TEQ(tardisgolib.CPos(), true, utf8.ValidRune(valid_rune))
	TEQ(tardisgolib.CPos(), false, utf8.ValidRune(invalid_rune))

	valid_string := "Hello, 世界"
	invalid_string := string([]byte{0xff, 0xfe, 0xfd})
	TEQ(tardisgolib.CPos(), true, utf8.ValidString(valid_string))
	if ShowKnownErrors || tardisgolib.Zilen() == 3 {
		TEQ(tardisgolib.CPos()+" NOTE: known error handling incorrect strings on UTF16 platforms", false, utf8.ValidString(invalid_string))
	}
}

func testChan() {
	c := make(chan int, 2)
	c <- 1
	c <- 2
	close(c)
	TEQ(tardisgolib.CPos(), <-c, 1)
	TEQ(tardisgolib.CPos(), <-c, 2)
	v, ok := <-c
	TEQ(tardisgolib.CPos(), v, 0)
	TEQ(tardisgolib.CPos(), ok, false)

	ch := make(chan bool, 2)
	ch <- true
	ch <- true
	close(ch)
	rangeCount := 0
	for v := range ch {
		TEQ(tardisgolib.CPos(), v, true)
		rangeCount++
	}
	TEQ(tardisgolib.CPos(), rangeCount, 2)

	//TODO much more to come here...
}

func testComplex() {

	var x, y, z complex64
	var ss complex128

	x = 1 + 2i
	TEQfloat(tardisgolib.CPos(), float64(real(x)), 1, 0.1)
	TEQfloat(tardisgolib.CPos(), float64(imag(x)), 2, 0.1)

	y = complex(3, 4)
	TEQfloat(tardisgolib.CPos(), float64(real(y)), 3, 0.1)
	TEQfloat(tardisgolib.CPos(), float64(imag(y)), 4, 0.1)

	//this previously failed in the SSA interpreter
	z = -x
	TEQfloat(tardisgolib.CPos(), float64(real(z)), -1, 0.1)
	TEQfloat(tardisgolib.CPos(), float64(imag(z)), -2, 0.1)

	z = x + y
	TEQfloat(tardisgolib.CPos(), float64(real(z)), 4, 0.1)
	TEQfloat(tardisgolib.CPos(), float64(imag(z)), 6, 0.1)

	z = x - y
	TEQfloat(tardisgolib.CPos(), float64(real(z)), -2, 0.1)
	TEQfloat(tardisgolib.CPos(), float64(imag(z)), -2, 0.1)

	z = x + y - y
	TEQfloat(tardisgolib.CPos(), float64(real(z)), float64(real(x)), 0.1)
	TEQfloat(tardisgolib.CPos(), float64(imag(z)), float64(imag(x)), 0.1)
	/*
		z = x * y
		printf64("real(x*y)", float64(real(z)))
		printf64("imag(x*y)", float64(imag(z)))

		z = x / y
		printf64("real(x/y)", float64(real(z)))
		printf64("imag(x/y)", float64(imag(z)))
	*/
	z = x * y / y
	TEQfloat(tardisgolib.CPos(), float64(real(z)), float64(real(x)), 0.1)
	TEQfloat(tardisgolib.CPos(), float64(imag(z)), float64(imag(x)), 0.1)

	TEQ(tardisgolib.CPos(), x == y, false)

	TEQ(tardisgolib.CPos(), x != y, true)

	ss = complex128(x)
	tt := complex128(y)
	TEQ(tardisgolib.CPos(), ss != tt, true)
}

var aString = "A"
var aaString = "AA"
var bbString = "BB"

func testString() {
	TEQ(tardisgolib.CPos(), aString <= "A", true)
	TEQ(tardisgolib.CPos(), aString <= aaString, true)
	TEQ(tardisgolib.CPos(), aString > aaString, false)
	TEQ(tardisgolib.CPos(), aString == aaString, false)
	TEQ(tardisgolib.CPos(), aString+aString == aaString, true)
	TEQ(tardisgolib.CPos(), bbString < aaString, false)
}

func adder() func(int) int {
	sum := 0
	return func(x int) int {
		sum += x
		return sum
	}
}

// fib returns a function that returns
// successive Fibonacci numbers.
func fib() func() int {
	a, b := 0, 1
	return func() int {
		a, b = b, a+b
		return a
	}
}

func testClosure() {
	// example from the go tour
	pos, neg := adder(), adder()
	for i := 0; i < 10; i++ {
		pos(i)
		neg(-2 * i)
	}
	TEQ(tardisgolib.CPos(), pos(0), 45)
	TEQ(tardisgolib.CPos(), neg(0), -90)

	// example from http://jordanorelli.tumblr.com/post/42369331748/function-types-in-go-golang
	x := 5
	fn := func(y int) {
		TEQ(tardisgolib.CPos(), x, y)
	}
	fn(5)
	x++
	fn(6)

	f := fib()
	TEQ(tardisgolib.CPos(), f(), 1)
	TEQ(tardisgolib.CPos(), f(), 1)
	TEQ(tardisgolib.CPos(), f(), 2)
	TEQ(tardisgolib.CPos(), f(), 3)
	TEQ(tardisgolib.CPos(), f(), 5)
}

func testVariadic(values ...int) {
	total := 0
	for i := range values {
		total += values[i]
	}
	TEQ(tardisgolib.CPos(), total, 42)
}

func testMath() {
	// comment out for quicker testing
	/*
		if int(math.Sqrt(16.0)) != 4 {
			println(tardisgolib.CPos() + ": Incorrect square root of 16")
		}
	*/
}

func testInterface() {
	var i interface{}

	i = "test"
	if i.(string) != "test" {
		println("testInterface string not equal 'test':")
		println(i)
	}

	i = int(42)
	if i.(int) != 42 {
		println("testInterface int not equal 42:")
		println(i)
	}

	j, ok := i.(rune)
	if ok {
		println("error rune!=int")
	}
	TEQ(tardisgolib.CPos(), j, rune(0))
}

// from the go tour
type Vertex struct {
	X, Y float64
}

func (v *Vertex) Abs() float64 {
	return Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v *Vertex) Scale(f float64) float64 {
	v.X = v.X * f
	v.Y = v.Y * f
	return v.Abs()
}

func (v MyFloat) Scale(f float64) float64 {
	return float64(v) * f
}

type Abser interface {
	Abs() float64
	Scale(x float64) float64
}

//from the go.tools/go/types documentation
type T0 struct {
	X float64
}

var p0 *T0

type Value struct {
	typ *rtype
	flag
}

type flag uintptr

func (f flag) mv(x flag) flag { f = x; return f }

type rtype struct {
	hash      uint32
	_         uint8  // test unused/padding
	ptrToThis *rtype // test self-ref
	flag
}

var rt *rtype

func (v Value) testFieldFn() {
	f0 := flag.mv(0, 42)
	f1 := flag.mv
	TEQ(tardisgolib.CPos(), f1(f0, 42), f0)
	rt = new(rtype)
	v.typ = rt
	v.typ.hash = 42
	x := rt.hash
	TEQ(tardisgolib.CPos(), v.typ.hash, x)
	y := &rt.hash
	z := (*y)
	TEQ(tardisgolib.CPos(), z, x)
	var ff interface{} = &v.flag
	*(ff.(*flag)) = 42
	var ffh interface{} = v.typ
	ffh.(*rtype).flag = f1(0, 42)
	TEQ(tardisgolib.CPos(), *(ff.(*flag)), ffh.(*rtype).flag)
}

func testInterfaceMethods() {
	var v0 Value
	v0.testFieldFn()

	v := &Vertex{3, 4}
	TEQfloat(tardisgolib.CPos(), v.Abs(), 5, 0.00001)
	TEQfloat(tardisgolib.CPos(), v.Scale(5), 25, 0.001)
	TEQfloat(tardisgolib.CPos(), v.X, 15, 0.0000001)
	TEQfloat(tardisgolib.CPos(), v.Y, 20, 0.0000001)

	var a Abser
	f := MyFloat(-42)
	vt := Vertex{3, 4}

	a = f // a MyFloat implements Abser
	x, ok := a.(Abser)
	//println(reflect.TypeOf(x).String()) => main.MyFloat
	if !ok {
		println("Error in testInterfaceMethods(): MyFloat should be in Abser interface")
	}
	TEQ(tardisgolib.CPos()+"testInterfaceMethods():MyFloat in Abser", a, f)
	TEQ(tardisgolib.CPos()+"testInterfaceMethods():MyFloat.Abs()", a.Abs(), float64(42))
	TEQ(tardisgolib.CPos()+"testInterfaceMethods():x.Abs()", x.Abs(), float64(42))
	TEQ(tardisgolib.CPos()+"testInterfaceMethods():MyFloat.Scale(10)", a.Scale(10), float64(-420))
	TEQ(tardisgolib.CPos()+"testInterfaceMethods():x.Scale(10)", x.Scale(10), float64(-420))

	a = &vt // a *Vertex implements Abser
	y, ok := a.(Abser)
	//println(reflect.TypeOf(y).String()) => *main.Vertex
	if !ok {
		println("Error in testInterfaceMethods(): Vertex should be in Abser interface")
	}
	TEQ(tardisgolib.CPos()+"testInterfaceMethods():*Vertex in Abser", a, &vt)
	TEQfloat(tardisgolib.CPos()+"testInterfaceMethods():*Vertex.Abs()", a.Abs(), float64(5), 0.000001)
	TEQfloat(tardisgolib.CPos()+"testInterfaceMethods():y.Abs()", y.Abs(), float64(5), 0.000001)
	TEQfloat(tardisgolib.CPos()+"testInterfaceMethods():*Vertex.Scale(10)", a.Scale(10), float64(50), 0.000001)
	TEQfloat(tardisgolib.CPos()+"testInterfaceMethods():y.Scale(10)", y.Scale(10), float64(653.35), 0.01)

	// a=vt // a Vertex, does NOT

	//from the go.tools/go/types documentation
	p0 = new(T0) // TODO should fail with this line missing, but does not (globals pre-initialised when they should not be)
	p0.X = 42
	TEQfloat(tardisgolib.CPos(), p0.X, 42.0, 0.01)

}

func testStrconv() {
	/*
		TEQ(tardisgolib.CPos()+"testStrconv():Itoa", "424242", strconv.Itoa(424242))

		TEQ(tardisgolib.CPos(), strings.HasPrefix("say what", "say"), true)
		TEQ(tardisgolib.CPos()+" string.Contains (error on js)", strings.Contains("say what", "ay"), true)
		TEQ(tardisgolib.CPos()+" string.Contains (error on js)", strings.Contains("seafood", "foo"), true)
		TEQ(tardisgolib.CPos(), strings.Contains("seafood", "bar"), false)
		TEQ(tardisgolib.CPos(), strings.Contains("seafood", ""), true)
		TEQ(tardisgolib.CPos(), strings.Contains("", ""), true)
		TEQ(tardisgolib.CPos(), strings.Contains("equal?", "equal?"), true)

		TEQ(tardisgolib.CPos(), bytes.HasPrefix([]byte("say what"), []byte("say")), true)
		TEQ(tardisgolib.CPos(), bytes.Contains([]byte("say what"), []byte("ay")), true)
	*/
}

func sum(a []int, c chan int) {
	sum := 0
	for _, v := range a {
		sum += v
	}
	c <- sum // send sum to c
}

func testTour64() {
	a := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(a[:len(a)/2], c)
	go sum(a[len(a)/2:], c)
	x, y := <-c, <-c // receive from c

	TEQ(tardisgolib.CPos(), x+y, 12) // x & y could arrive in any order...
}

func testDefer_a() {
	i := 0
	defer TEQ(tardisgolib.CPos(), i, 0)
	i++
	return
}
func testDefer_b(ch chan int) {
	for i := 0; i < 4; i++ {
		defer func(j int) { ch <- j }(i)
	}
}
func testDefer_c() (i int) {
	defer func() { i++ }()
	return 1
}
func protect(g func(int)) {
	defer func() {
		TEQ(tardisgolib.CPos(), recover(), "test panic")
	}()
	g(0)
}

func g(i int) {
	if i > 3 {
		panic("test panic")
	}
	for j := 0; j < i; j++ {
		defer testDefer_d()
	}
	g(i + 1)
}

var tddCount = 0

func testDefer_d() {
	tddCount++ // just to give the routine something to do
}

func testDefer() {
	// examples from http://blog.golang.org/defer-panic-and-recover
	testDefer_a()
	b := make(chan int, 4)
	testDefer_b(b)
	TEQ(tardisgolib.CPos(), <-b, 3)
	TEQ(tardisgolib.CPos(), <-b, 2)
	TEQ(tardisgolib.CPos(), <-b, 1)
	TEQ(tardisgolib.CPos(), <-b, 0)
	TEQ(tardisgolib.CPos(), testDefer_c(), 2)
	protect(g)
	TEQ(tardisgolib.CPos(), tddCount, 6)
}

// these two names were failing in java as being duplicates, now failing in PHP...
func Ilogb(x float64) int {
	return int(Sqrt(x))
}
func ilogb(x float64) int {
	return int(Sqrt(x))
}
func testCaseSensitivity() {
	//moved to a separate test file
	//TEQ(tardisgolib.CPos(), ilogb(64), Ilogb(64))
}

var (
	aGrCtr    int32
	aGrCtrMux sync.Mutex
	aGrWG     sync.WaitGroup
)

func aGoroutine(a int) {
	if a == 4 {
		//panic("test panic in goroutine 4")
	}
	for i := 0; i < a; i++ {
		tardisgolib.Gosched()
	}
	(&aGrCtrMux).Lock()
	atomic.AddInt32(&aGrCtr, -1)
	(&aGrCtrMux).Unlock()

	aGrWG.Done()
}

const numGR = 5

func testManyGoroutines() {
	var n = numGR
	aGrCtr = numGR * 2 // set up the goroutine counter
	for i := 0; i < n; i++ {
		aGrWG.Add(1)
		go aGoroutine(i)
	}
	for i := n; i > 0; i-- {
		aGrWG.Add(1)
		go aGoroutine(i)
	}
}

//
// Code from http://golangtutorials.blogspot.co.uk/2011/06/channels-in-go-range-and-select.html
//
func makeCakeAndSend(cs chan string, flavor string, count int) {
	for i := 1; i <= count; i++ {
		TEQ("Delay", i, i)
		cakeName := flavor + " Cake " + string('0'+i)
		cs <- cakeName //send a strawberry cake
	}
	close(cs)
}

func receiveCakeAndPack(strbry_cs chan string, choco_cs chan string) {
	strbry_closed, choco_closed := false, false

	for {
		//if both channels are closed then we can stop
		if strbry_closed && choco_closed {
			return
		}
		//println("Waiting for a new cake ...")
		select {
		case cakeName, strbry_ok := <-strbry_cs:
			if !strbry_ok {
				strbry_closed = true
				//println(" ... Strawberry channel closed!")
			} else {
				//println("Received from Strawberry channel.  Now packing", cakeName)
				_ = cakeName
			}
		case cakeName, choco_ok := <-choco_cs:
			if !choco_ok {
				choco_closed = true
				//println(" ... Chocolate channel closed!")
			} else {
				//println("Received from Chocolate channel.  Now packing", cakeName)
				_ = cakeName
			}
		default:
			//println("no cake!")
		}
	}
}

func testChanSelect() {
	strbry_cs := make(chan string)
	choco_cs := make(chan string)

	//two cake makers
	go makeCakeAndSend(choco_cs, "Chocolate", 3)   //make 3 chocolate cakes and send
	go makeCakeAndSend(strbry_cs, "Strawberry", 3) //make 3 strawberry cakes and send

	//one cake receiver and packer
	receiveCakeAndPack(strbry_cs, choco_cs) //pack all cakes received on these cake channels

	//sleep for a while so that the program doesn’t exit immediately
	//time.Sleep(2 * 1e9)
}

//end code from http://golangtutorials.blogspot.co.uk/2011/06/channels-in-go-range-and-select.html

//From the go tour http://tour.golang.org/#69
func fibonacci(c, quit chan int) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			//println("quit")
			return
		}
	}
}

func tourfib() {
	c := make(chan int)
	quit := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			/*println(*/ <-c /*)*/
		}
		quit <- 0
	}()
	fibonacci(c, quit)
}

// end tour

func testUintDiv32() {
	var uifs, pwr2 uint32
	uifs = uint32(0xffffffff) // test for  "-1" number constant
	pwr2 = uint32(1)
	for i := uint32(0); i < 32; i++ {
		if !TEQuint32("testUintDiv32() T1 ", (uifs)>>i, (uifs)/pwr2) {
			println("ProblemT1 i=", int(i))
		}
		pwr2 *= 2
	}
	uifs2 := uint32(0xfffffff0) //  test another "-ve" number constant
	uifs = uifs2
	pwr2 = uint32(1)
	for i := uint32(0); i < 32; i++ {
		if !TEQuint32("testUintDiv32() T2 ", (uifs)>>i, (uifs)/pwr2) {
			println("ProblemT2 i=", int(i))
		}
		pwr2 *= 2
	}
}
func testUintDiv64() {
	var uifs, pwr2 uint64
	uifs = uint64(0xfffffffffffffff0)
	pwr2 = uint64(1)
	for i := uint64(0); i < 64; i++ {
		if !TEQuint64("testUintDiv64() ", uifs>>i, uifs/pwr2) {
			println("Problem i=", int(i))
		}
		pwr2 *= 2
	}
}

func main() {
	println("Start test running in: " + tardisgolib.Platform())
	testManyGoroutines()
	testChanSelect()
	tourfib()
	testCaseSensitivity()
	testInit()
	testConst()
	testUTF()
	testFloat()
	testMultiRet()
	testAppend()
	testStruct()
	testHeader()
	testCopy()
	testInFuncPtr()
	testCallBy()
	testMap()
	testNamed()
	testFuncPtr()
	testIntOverflow()
	testSlices()
	testChan()
	testComplex()
	testUTF8()
	testString()
	testClosure()
	testVariadic(42)
	testVariadic(40, 2)
	testVariadic(42, -5, 3, 2)
	testInterface()
	testInterfaceMethods()
	testStrconv()
	testTour64()
	testUintDiv32()
	testUintDiv64()
	testDefer()
	aGrWG.Wait()
	TEQint32(tardisgolib.CPos()+" testManyGoroutines() sync/atomic counter:", aGrCtr, 0)
	if tardisgolib.Host() == "haxe" {
		TEQ(tardisgolib.CPos(), int(tardisgolib.HAXE("42;")), int(42))
		TEQ(tardisgolib.CPos(), string(tardisgolib.HAXE("'test';")), "test")
		TEQ(tardisgolib.CPos()+"Num Haxe GR post-wait", tardisgolib.NumGoroutine(), 1)
	} else {
		TEQ(tardisgolib.CPos()+"Num Haxe GR post-wait", tardisgolib.NumGoroutine(), 2)
	}
	println("End test running in: " + tardisgolib.Platform())
	println("再见！Previous two chinese characters should say goodbye! (testing unicode output)")
	println()
}
