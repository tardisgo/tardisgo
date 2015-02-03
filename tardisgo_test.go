package main

import (
	"os"
	"os/exec"
	"runtime"
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

	out, err := exec.Command("haxe", "-main", "tardis.Go", "--interp").CombinedOutput()
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

/*
func stdLibTest(t *testing.T, lib string) {
	if !testing.Short() && runtime.GOOS != "darwin" {
		out, err := exec.Command("bash", "stdlibtest.sh", lib).CombinedOutput()
		t.Log(string(out))
		if err != nil {
			t.Error(err)
		}
	}
}
func TestContainerHeap(t *testing.T)   { stdLibTest(t, "container/heap") }
func TestContainerList(t *testing.T)   { stdLibTest(t, "container/list") }
func TestContainerRing(t *testing.T)   { stdLibTest(t, "container/ring") }
func TestEncodingAscii85(t *testing.T) { stdLibTest(t, "encoding/ascii85") }
func TestEncodingBase32(t *testing.T)  { stdLibTest(t, "encoding/base32") }
func TestEncodingBase64(t *testing.T)  { stdLibTest(t, "encoding/base64") }
func TestEncodingHex(t *testing.T)     { stdLibTest(t, "encoding/hex") }
func TestErrors(t *testing.T)          { stdLibTest(t, "errors") }
func TestPath(t *testing.T)            { stdLibTest(t, "path") }
func TestStrings(t *testing.T)         { stdLibTest(t, "strings") }
func TestTextTabwriter(t *testing.T)   { stdLibTest(t, "text/tabwriter") }
func TestUnicode(t *testing.T)         { stdLibTest(t, "unicode") }
func TestUnicodeUTF8(t *testing.T)     { stdLibTest(t, "unicode/utf8") }
func TestUnicodeUTF16(t *testing.T)    { stdLibTest(t, "unicode/utf16") }

// Long so last
func TestSort(t *testing.T)  { stdLibTest(t, "sort") }
func TestBytes(t *testing.T) { stdLibTest(t, "bytes") }
*/

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type stdLibResT struct {
	out []byte
	err error
}

var stdLibRes = make(chan stdLibResT)

func stdLibTestGo(lib string) {
	var r stdLibResT
	r.out, r.err = exec.Command("bash", "stdlibtest.sh", lib).CombinedOutput()
	stdLibRes <- r
}

var libs = []string{
	"bytes", "errors", "path", "strings", "sort", "unicode", "container/heap", "container/list",
	"container/ring", "encoding/ascii85", "encoding/base32", "encoding/base64", "encoding/hex", "text/tabwriter",
	"unicode/utf8", "unicode/utf16",
}

func TestStdLib(t *testing.T) {
	if !testing.Short() {
		for _, lib := range libs {
			go stdLibTestGo(lib)
		}
		for _ = range libs {
			r := <-stdLibRes
			t.Log(string(r.out))
			if r.err != nil {
				t.Error(r.err)
			}
		}
	}
}
