package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestCore(t *testing.T) {

	err := os.Chdir("tests/core")
	if err != nil {
		t.Error(err)
	}

	*debugFlag = true
	err = doTestable([]string{"test.go"})
	if err != nil {
		t.Error(err)
	}

	out, err := exec.Command("haxe", "-main", "tardis.Go", "--no-inline", "--interp").CombinedOutput()
	if err != nil {
		t.Error(err)
	}

	// any Haxe output would signal an error
	if len(out) > 0 {
		t.Errorf(string(out))
	}
}
