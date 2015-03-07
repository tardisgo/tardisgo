// +build haxe

package ioutil

import "os"

func init() {
	err := os.Mkdir("/ioutil", 0777)
	if err != nil {
		panic(err)
	}
}
