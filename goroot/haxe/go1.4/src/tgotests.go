// Copyright Elliott Stoneham 2015 see licence file

// Usage: go run tgotests.go

package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// space required before and after package names

// the allList only contains package tests that pass for all 4 targets
var allList = []string{
	// these tests do not read any files
	// for speed of compilation, they are grouped together into as large sets as will work
	"bufio bytes container/heap container/list container/ring ",
	"crypto/aes crypto/cipher crypto/des crypto/dsa crypto/ecdsa crypto/elliptic crypto/hmac " +
		"crypto/md5 crypto/rand crypto/rc4 crypto/sha1 crypto/sha256 crypto/sha512 ",
	"debug/gosym database/sql/driver encoding/ascii85 encoding/base32 encoding/base64 " +
		"encoding/csv encoding/hex errors flag ",
	"go/scanner go/token hash/adler32 hash/crc32 hash/crc64 hash/fnv html image/color ",
	"index/suffixarray io math/cmplx net/http/internal net/mail net/textproto net/url path ",
	"regexp/syntax runtime sort strings sync/atomic text/scanner text/tabwriter text/template/parse ",
	"unicode unicode/utf16 unicode/utf8 ",
	// below are those packages that require their own testdata zip file, and so must be run individually
	"archive/zip ",
	"compress/bzip2", "compress/flate", "compress/gzip", "compress/lzw", "compress/zlib",
	"debug/pe", "debug/plan9obj",
	"go/format", "go/parser", "go/printer",
	"image", "image/draw", "image/gif", "image/jpeg",
	"io/ioutil",
	"mime",
	"os",
	"path/filepath",
	"regexp",
}

var js1 = "crypto/x509" //runtime very long at 30+ mins
var js = ` archive/tar 
 debug/elf   encoding/pem fmt
 math net/http/fcgi  strconv
`

var cs = ` 
 crypto/x509
 debug/elf    
`

var cpp = ` 
  archive/tar 
  debug/elf  encoding/pem fmt   
  math  strconv
`

var java = ` 
`

func pkgList(jumble string) []string {
	pkgs := strings.Split(jumble, " ")
	edited := []string{}
	for _, pkg := range pkgs {
		pkg = strings.TrimSpace(pkg)
		if pkg != "" {
			edited = append(edited, pkg)
		}
	}
	sort.Strings(edited)
	//println("DEBUG sorted list: ", strings.Join(edited, " "))
	return edited
}

type resChan struct {
	output string
	err    error
}

var scores = make(map[string]string)
var passes, failures uint32

func doTarget(target string, pkg []string) {
	//println("DEBUG ", target, pkg)
	var lastErr error
	exe := "bash"
	_, err := exec.LookPath(exe)
	if err != nil {
		switch exe {
		default:
			panic(" error - executable not found: " + exe)
		}
	}
	out := []byte{}
	if target == "all" {
		prms := append([]string{"./testtgoall.sh"}, pkg...)
		out, lastErr = exec.Command(exe, prms...).CombinedOutput()
	} else {
		out, lastErr = exec.Command(exe, "./testtgo.sh", target, pkg[0]).CombinedOutput()
	}
	layout := "%-25s %s"
	for n := range pkg {
		if lastErr != nil {
			//out = append(out, []byte(lastErr.Error())...)
			scores[fmt.Sprintf(layout, pkg[n], target)] = "Fail"
			atomic.AddUint32(&failures, 1)
		} else {
			scores[fmt.Sprintf(layout, pkg[n], target)] = "Pass"
			atomic.AddUint32(&passes, 1)
		}
	}
	results <- resChan{string(out), lastErr}
}

type params struct {
	tgt string
	pkg []string
}

var parallelism = 1 + runtime.NumCPU()/2 // control resource usage here

var limit = make(chan params)

var results = make(chan resChan, parallelism)

func limiter() {
	for {
		p := <-limit
		doTarget(p.tgt, p.pkg)
	}
}

func main() {
	jsl := pkgList(js)
	jsl1 := pkgList(js1)
	csl := pkgList(cs)
	cppl := pkgList(cpp)
	javal := pkgList(java)
	numPkgs := len(jsl) + len(jsl1) + len(csl) + len(cppl) + len(javal) + len(allList)
	var wg sync.WaitGroup
	wg.Add(numPkgs)
	go func() {
		for count := 0; count < numPkgs; count++ {
			r := <-results
			fmt.Println(r.output)
			fmt.Printf("\n%d passes, %d failures.\n", passes, failures)
			wg.Done()
		}
	}()

	go limiter()                               // need this in case parallism == 1
	for pll := 0; pll < parallelism/2; pll++ { // the "all" option runs 4 target tests in parallel
		go limiter()
	}
	for _, ap := range allList {
		limit <- params{"all", pkgList(ap)}
	}
	for _, pkg := range jsl1 { //very long js tests 1st, with lower contention
		limit <- params{"js", []string{pkg}}
	}
	for pll := parallelism / 2; pll < parallelism; pll++ { // other options run 1 test each
		go limiter()
	}
	for _, pkg := range cppl {
		limit <- params{"cpp", []string{pkg}}
	}
	for _, pkg := range javal {
		limit <- params{"java", []string{pkg}}
	}
	for _, pkg := range csl {
		limit <- params{"cs", []string{pkg}}
	}
	// normal length js tests
	for _, pkg := range jsl {
		limit <- params{"js", []string{pkg}}
	}

	wg.Wait()
	joint := []string{}
	for k := range scores {
		joint = append(joint, k)
	}
	sort.Strings(joint)
	fmt.Println("\nResults\n=======")
	for _, k := range joint {
		fmt.Printf("%-35s %s\n", k, scores[k])
	}
	fmt.Printf("\n%d passes, %d failures.\n", passes, failures)
}