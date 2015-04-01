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

var js1 = "" //crypto/x509" //runtime very long at 30+ mins, excluded for now
var js = `
	 runtime compress/flate regexp/syntax sync/atomic  
	 database/sql/driver fmt 
	 container/ring crypto/des crypto/hmac crypto/md5 crypto/rc4 crypto/sha1 crypto/sha256 crypto/sha512 
	 debug/gosym debug/pe debug/plan9obj encoding/base64 encoding/csv encoding/hex encoding/pem 
	 go/format go/scanner hash/crc32 hash/crc64 index/suffixarray io/ioutil math/cmplx path/filepath 
	 text/scanner archive/zip os 
	 bufio bytes flag html io sort strings unicode 
	 container/heap container/list 
	 crypto/cipher crypto/rand crypto/elliptic crypto/ecdsa crypto/dsa 
	 encoding/ascii85 encoding/base32 image/color 
	 text/tabwriter unicode/utf8 unicode/utf16 
	 errors path mime net/url net/textproto net/mail 
	 go/token math archive/tar 
	 net/http/fcgi net/http/internal 
	 regexp  strconv  crypto/aes 
	 compress/bzip2 compress/gzip compress/lzw compress/zlib debug/elf 
	 hash/adler32 hash/fnv image image/draw image/gif image/jpeg 
`

var cs = `bytes strings unicode unicode/utf8 unicode/utf16 
 math/cmplx io io/ioutil archive/zip 
	 compress/bzip2   compress/flate   compress/lzw  compress/zlib 
 errors sort container/ring container/list container/heap`

var cpp = `errors sort container/ring container/list container/heap strconv 
 archive/tar  bufio 
  compress/bzip2   compress/flate  compress/gzip  compress/lzw  compress/zlib 
 math math/cmplx unicode unicode/utf8 unicode/utf16 io io/ioutil fmt bytes strings 
`

var java = `errors sort container/ring container/list container/heap
 archive/zip  
  compress/bzip2    compress/flate  compress/gzip  compress/lzw  compress/zlib 
 math/cmplx unicode unicode/utf8 unicode/utf16  io io/ioutil bytes strings 
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

func doTarget(target, pkg string) {
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
	out, lastErr = exec.Command(exe, "./testtgo.sh", target, pkg).CombinedOutput()
	layout := "%-25s %s"
	if lastErr != nil {
		//out = append(out, []byte(lastErr.Error())...)
		scores[fmt.Sprintf(layout, pkg, target)] = "Fail"
		atomic.AddUint32(&failures, 1)
	} else {
		scores[fmt.Sprintf(layout, pkg, target)] = "Pass"
		atomic.AddUint32(&passes, 1)
	}
	results <- resChan{string(out), lastErr}
}

type params struct {
	tgt, pkg string
}

var parallelism = runtime.NumCPU()

var limit = make(chan params, parallelism)

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
	numPkgs := len(jsl) + len(jsl1) + len(csl) + len(cppl) + len(javal)
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
	for pll := 0; pll < parallelism; pll++ {
		go limiter()
	}

	//very long js tests 1st
	for _, pkg := range jsl1 {
		limit <- params{"js", pkg}
	}
	for _, pkg := range cppl {
		limit <- params{"cpp", pkg}
	}
	for _, pkg := range javal {
		limit <- params{"java", pkg}
	}
	for _, pkg := range csl {
		limit <- params{"cs", pkg}
	}
	// normal length js tests
	for _, pkg := range jsl {
		limit <- params{"js", pkg}
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
