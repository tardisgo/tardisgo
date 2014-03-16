package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestCore(t *testing.T) {
	n := runtime.NumCPU()
	fmt.Printf("DEBUG NumCPU=%d\n", n)

	err := os.Chdir("tests/core")
	if err != nil {
		t.Error(err)
	}
	err = doTestable([]string{"test.go"})
	if err != nil {
		t.Error(err)
	}

	out, err := exec.Command("haxe", "-main", "tardis.Go", "--interp").CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	//fmt.Printf("The Haxe output is: %d :\n%s\n", len(out), out)

	if len(out) > 0 {
		t.Errorf(string(out))
	}
}
