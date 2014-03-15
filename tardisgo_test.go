package main

import (
	"os"
	//"os/exec"
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
	/*
		out, err := exec.Command("haxe", "-main=tardis.Go", "--interp").Output()
		if err != nil {
			t.Error(err)
		}
		t.Log("The Haxe output is:\n%s\n", out)
	*/
}
