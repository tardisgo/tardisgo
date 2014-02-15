// Package syscall is not implemented for TARDIS Go, this code is a non-functioning TEST, only for OSX
package syscall

import host "syscall"

// Syscall unimplemented
func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err host.Errno) {
	return 0, 0, host.EBADARCH
}

// Syscall6 unimplemented
func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err host.Errno) {
	return 0, 0, host.EBADARCH
}

// RawSyscall unimplemented
func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err host.Errno) {
	return 0, 0, host.EBADARCH
}

// RawSyscall6 unimplemented
func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err host.Errno) {
	return 0, 0, host.EBADARCH
}

// setenv_c is provided by the runtime, but is a no-op if cgo isn't
// loaded.
func setenv_c(k, v string) {}

// should be implemented in runtime package
func runtime_BeforeFork() {
	panic("syscall.runtime_BeforeFork() NOT IMPLEMENTED")
}
func runtime_AfterFork() {
	panic("syscall.runtime_AfterFork() NOT IMPLEMENTED")
}
