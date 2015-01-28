package fmt

import (
	"errors"
	"io"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

func init() {
	/*
		if false {
			Println("tgofmt.init()")
			Printf("tgofmt.init()")
			Print("tgofmt.init()")
			Errorf("format string, a ...interface{}", 42)
			Fprint(nil, "a ...interface{}")
			Fprintf(nil, "format string, a ...interface{}")
			Fprintln(nil, "blah")
		}
	*/
}

func Errorf(format string, a ...interface{}) error {
	return errors.New(Sprintf(format, a...))
}
func Fprint(w io.Writer, a ...interface{}) (n int, err error) {
	return Print(a...) // DUMMY
}
func Fprintf(w io.Writer, format string, a ...interface{}) (n int, err error) {
	return Printf(format, a...) // DUMMY
}
func Fprintln(w io.Writer, a ...interface{}) (n int, err error) {
	return Println(a...) // DUMMY
}

func Print(a ...interface{}) (n int, err error) {
	print(Sprint(a...))
	return 0, nil
}

func Println(a ...interface{}) (n int, err error) {
	println(Sprint(a...))
	return 0, nil
}

func Printf(format string, a ...interface{}) (n int, err error) {
	print(Sprintf(format, a...))
	return 0, nil
}

func Sprintf(format string, a ...interface{}) string {
	return " { " + Sprint(a...) + " } " + Sprint(format)
}

func Sprintln(a ...interface{}) string {
	return Sprint(a...) + "\n"
}
func Sprint(a ...interface{}) string {
	ret := ""
	for i := range a {
		ret += hx.CallString("", "Std.string", 1, a[i])
		ret += " "
	}
	return ret
}
