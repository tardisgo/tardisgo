// Package os is not implemented for TARDIS Go, this code is only as a TEST for OSX
package tgoos

import "errors"

func init() { // stop DCE
	if false {
		Exit()
	}
}

func Exit() {
	panic("tgoos.Exit()")
}

var (
	ErrInvalid    = errors.New("invalid argument")
	ErrPermission = errors.New("permission denied")
	ErrExist      = errors.New("file already exists")
	ErrNotExist   = errors.New("file does not exist")
)

/*
var (
	Stdin  = NewFile(uintptr(syscall.Stdin), "/dev/stdin")
	Stdout = NewFile(uintptr(syscall.Stdout), "/dev/stdout")
	Stderr = NewFile(uintptr(syscall.Stderr), "/dev/stderr")
)
*/
var Args []string
