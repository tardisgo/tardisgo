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

	out, err := exec.Command("haxe", "-main", "tardis.Go", "-cp", "tardis", "--interp").CombinedOutput()
	if err != nil {
		t.Error(err)
	}

	// any Haxe output would signal an error
	if len(out) > 0 {
		t.Errorf(string(out))
	}

	err = os.Chdir("../..")
	if err != nil {
		t.Error(err)
	}
}

// NOTE: main Travis CI standard library tests are in a shell script in goroot/...
