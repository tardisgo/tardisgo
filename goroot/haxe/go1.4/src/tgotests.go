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

// time tardisgo -haxe all -test bytes container/heap container/list container/ring crypto/aes crypto/cipher crypto/des crypto/dsa crypto/ecdsa crypto/elliptic crypto/hmac crypto/md5 crypto/rand crypto/rc4 crypto/sha1 crypto/sha256 crypto/sha512 database/sql/driver encoding/ascii85 encoding/base32 encoding/csv encoding/hex errors flag
var allList = []string{
	"bytes container/heap container/list container/ring ",
	"crypto/aes crypto/cipher crypto/des crypto/dsa crypto/ecdsa crypto/elliptic crypto/hmac " +
		"crypto/md5 crypto/rand crypto/rc4 crypto/sha1 crypto/sha256 crypto/sha512 ",
	"database/sql/driver encoding/ascii85 encoding/base32 encoding/csv encoding/hex errors flag",
	"go/scanner go/token hash/adler32 hash/crc32 hash/crc64 hash/fnv html image/color ",
	"index/suffixarray io math/cmplx net/http/internal net/mail net/textproto net/url path ",
	"regexp/syntax sort strings sync/atomic text/scanner text/tabwriter text/template/parse ",
	"unicode unicode/utf16 unicode/utf8 ",
}

var js1 = "crypto/x509" //runtime very long at 30+ mins
var js = ` archive/tar archive/zip bufio 	
	 compress/bzip2 compress/flate compress/gzip compress/lzw compress/zlib	 
	 debug/elf debug/gosym debug/pe debug/plan9obj encoding/base64  encoding/pem fmt
	 go/format go/parser go/printer  image image/draw image/gif image/jpeg io/ioutil  	
	 math mime net/http/fcgi  os path/filepath regexp  strconv  
`

// TODO reorder list below
var cs = ` debug/elf debug/gosym debug/pe debug/plan9obj
  image image/draw image/gif image/jpeg
 io/ioutil archive/zip  os 	 compress/bzip2   compress/flate   compress/lzw  compress/zlib mime
 crypto/x509	 path/filepath go/format go/scanner regexp  go/parser
 `

// TODO reorder list below
var cpp = ` strconv debug/elf debug/gosym
  image image/draw image/gif image/jpeg
 archive/tar  bufio  debug/pe debug/plan9obj mime
  compress/bzip2   compress/flate  compress/gzip  compress/lzw  compress/zlib 
    regexp  go/format path/filepath  encoding/base64 encoding/pem
 math  io/ioutil fmt  os  go/printer go/parser archive/zip
`

// TODO reorder list below
var java = ` mime os  image image/draw image/gif image/jpeg
 archive/zip  debug/gosym debug/pe debug/plan9obj
  compress/bzip2    compress/flate compress/lzw  compress/zlib flag
   regexp path/filepath encoding/base64 io/ioutil go/format go/printer go/parser
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

var parallelism = runtime.NumCPU()

var limit = make(chan params)

var results = make(chan resChan)

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

	go limiter() // need this in case only 1 cpu
	for pll := 1; pll < parallelism/2; pll++ {
		go limiter()
	}
	for _, ap := range allList {
		limit <- params{"all", pkgList(ap)}
	}
	for pll := parallelism / 2; pll < parallelism; pll++ {
		go limiter()
	}
	for _, pkg := range jsl1 { //very long js tests 1st
		limit <- params{"js", []string{pkg}}
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
