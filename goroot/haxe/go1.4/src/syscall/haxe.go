package syscall

func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno) {
	//println("DEBUG: syscall.Syscall()" + hx.CallString("", "Std.string", 1, trap))
	return 0, 0, ENOSYS
}
