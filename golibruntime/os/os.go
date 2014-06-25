// Package os is not implemented for TARDIS Go, this code is only as a TEST for OSX
package os

func init() { // stop DCE
	if false {
		//sigpipe()
	}
}

// dummy
func sigpipe() {
	panic("os.sigpipe() NOT IMPLEMENTED")
}
