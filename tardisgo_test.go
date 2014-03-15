package main

import (
	"os"
	"os/exec"
	"testing"
)

func Test1(t *testing.T) {
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
