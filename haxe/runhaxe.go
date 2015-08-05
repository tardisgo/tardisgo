package haxe

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func RunHaxe(allFlag *string, LoadTestZipFS bool, TestFS string) {
	results := make(chan resChan)
	switch *allFlag {
	case "": // NoOp
	case "all", "bench":
		//for _, dir := range dirs {
		//	err := os.RemoveAll(dir)
		//	if err != nil {
		//		fmt.Println("Error deleting existing '" + dir + "' directory: " + err.Error())
		//	}
		//}

		var targets [][][]string
		if *allFlag == "bench" {
			targets = allBenchmark // fast execution time
		} else {
			targets = allCompile // fast compile time
		}
		for _, cmd := range targets {
			go doTarget(cmd, results, LoadTestZipFS, TestFS)
		}
		for _ = range targets {
			r := <-results
			fmt.Println(r.output)
			if (r.err != nil || len(strings.TrimSpace(r.output)) == 0) && *allFlag != "bench" {
				os.Exit(1) // exit with an error if the test fails, but not for benchmarking
			}
			r.backChan <- true
		}

	case "math": // which is faster for the test with correct math processing, cpp or js?
		//err := os.RemoveAll("tardis/cpp")
		//if err != nil {
		//	fmt.Println("Error deleting existing '" + "tardis/cpp" + "' directory: " + err.Error())
		//}
		mathCmds := [][][]string{
			[][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-cpp", "tardis/cpp"},
				[]string{"echo", `"CPP:"`},
				[]string{"time", "./tardis/cpp/Go"},
			},
			[][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-D", "fullunsafe", "-js", "tardis/go-fu.js"},
				[]string{"echo", `"Node/JS using fullunsafe memory mode (js dataview):"`},
				[]string{"time", "node", "tardis/go-fu.js"},
			},
		}
		for _, cmd := range mathCmds {
			go doTarget(cmd, results, LoadTestZipFS, TestFS)
		}
		for _ = range mathCmds {
			r := <-results
			fmt.Println(r.output)
			if r.err != nil {
				os.Exit(1) // exit with an error if the test fails
			}
			r.backChan <- true
		}

	case "interp", "cpp", "cs", "js", "jsfu", "java", "flash": // for running tests
		switch *allFlag {
		case "interp":
			go doTarget([][]string{
				[]string{"echo", ``}, // Output from this line is ignored
				[]string{"echo", `"Neko (haxe --interp):"`},
				[]string{"time", "haxe", "-main", "tardis.Go", "-cp", "tardis", "--interp"},
			}, results, LoadTestZipFS, TestFS)
		case "cpp":
			go doTarget([][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-cpp", "tardis/cpp"},
				[]string{"echo", `"CPP:"`},
				[]string{"time", "./tardis/cpp/Go"},
			}, results, LoadTestZipFS, TestFS)
		case "cs":
			go doTarget([][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-cs", "tardis/cs"},
				[]string{"echo", `"CS:"`},
				[]string{"time", "mono", "./tardis/cs/bin/Go.exe"},
			}, results, LoadTestZipFS, TestFS)
		case "js":
			go doTarget([][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-D", "uselocalfunctions", "-js", "tardis/go.js"},
				[]string{"echo", `"Node/JS:"`},
				[]string{"time", "node", "tardis/go.js"},
			}, results, LoadTestZipFS, TestFS)
		case "jsfu":
			go doTarget([][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-D", "uselocalfunctions", "-D", "fullunsafe", "-js", "tardis/go-fu.js"},
				[]string{"echo", `"Node/JS using fullunsafe memory mode (js dataview):"`},
				[]string{"time", "node", "tardis/go-fu.js"},
			}, results, LoadTestZipFS, TestFS)
		case "java":
			go doTarget([][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-java", "tardis/java"},
				[]string{"echo", `"Java:"`},
				[]string{"time", "java", "-jar", "tardis/java/Go.jar"},
			}, results, LoadTestZipFS, TestFS)
		case "flash":
			go doTarget([][]string{
				[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-swf", "tardis/go.swf"},
				[]string{"echo", `"Flash:"`},
				[]string{"time", "open", "tardis/go.swf"},
			}, results, LoadTestZipFS, TestFS)
		}
		r := <-results
		fmt.Println(r.output)
		if r.err != nil {
			os.Exit(1) // exit with an error if the test fails
		}
		r.backChan <- true

	default:
		panic("invalid value for -haxe flag: " + *allFlag)
	}
}

//var dirs = []string{"tardis/cpp", "tardis/java", "tardis/cs" /*, "tardis/php"*/}

var allCompile = [][][]string{
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-cpp", "tardis/cpp"},
		[]string{"echo", `"CPP:"`},
		[]string{"time", "./tardis/cpp/Go"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-java", "tardis/java"},
		[]string{"echo", `"Java:"`},
		[]string{"time", "java", "-jar", "tardis/java/Go.jar"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-cs", "tardis/cs"},
		[]string{"echo", `"CS:"`},
		[]string{"time", "mono", "./tardis/cs/bin/Go.exe"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-D", "uselocalfunctions", "-js", "tardis/go.js"},
		[]string{"echo", `"Node/JS:"`},
		[]string{"time", "node", "tardis/go.js"},
	},
}
var allBenchmark = [][][]string{
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full" /*, "-D", "nulltempvars"*/, "-D", "inlinepointers" /*, "-D", "abstractobjects"*/, "-cpp", "tardis/cpp-bench"},
		[]string{"echo", `"CPP (bench):"`},
		[]string{"time", "./tardis/cpp-bench/Go"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full" /*, "-D", "nulltempvars"*/, "-D", "inlinepointers" /*, "-D", "abstractobjects"*/, "-java", "tardis/java-bench"},
		[]string{"echo", `"Java (bench):"`},
		[]string{"time", "java", "-jar", "tardis/java-bench/Go.jar"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full" /*, "-D", "nulltempvars"*/, "-D", "inlinepointers" /*, "-D", "abstractobjects"*/, "-cs", "tardis/cs-bench"},
		[]string{"echo", `"CS (bench):"`},
		[]string{"time", "mono", "./tardis/cs-bench/bin/Go.exe"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full" /*, "-D", "nulltempvars"*/, "-D", "inlinepointers" /*, "-D", "abstractobjects" */, "-D", "jsinit", "-D", "uselocalfunctions", "-js", "tardis/go-bench.js"},
		[]string{"echo", `"Node/JS (bench):"`},
		[]string{"time", "node", "tardis/go-bench.js"},
	},
	// as this mode is no longer used for testing, remove it from the "all" tests
	//[][]string{
	//	[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "fullunsafe", "-js", "tardis/go-fu.js"},
	//	[]string{"echo", `"Node/JS using fullunsafe memory mode (js dataview):"`},
	//	[]string{"time", "node", "tardis/go-fu.js"},
	//},
	// Cannot automate testing for SWF so removed
	//[][]string{
	//	[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-swf", "tardis/go.swf"},
	//	[]string{"echo", `"Opening swf file (Chrome as a file association for swf works to test on OSX):"` + "\n"},
	//	[]string{"open", "tardis/go.swf"},
	//},
	// PHP will never be a reliable target, so removed
	//[][]string{
	//	[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-php", "tardis/php", "--php-prefix", "tgo"},
	//	[]string{"echo", `"PHP:"`},
	//	[]string{"time", "php", "tardis/php/index.php"},
	//},
	// Seldom works, so removed
	//[][]string{
	//	[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-neko", "tardis/go.n"},
	//	[]string{"echo", `"Neko (does not work for large code):"`},
	//	[]string{"time", "neko", "tardis/go.n"},
	//},
	// only really useful for testing, so can be run from the command line
	//[][]string{
	//	[]string{"echo", ``}, // Output from this line is ignored
	//	[]string{"echo", `"Neko (haxe --interp):"`},
	//	[]string{"time", "haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "--interp"},
	//},
}

type resChan struct {
	output   string
	err      error
	backChan chan bool
}

func doTarget(cl [][]string, results chan resChan, LoadTestZipFS bool, TestFS string) {
	res := ""
	var lastErr error
	for j, c := range cl {
		if lastErr != nil {
			break
		}
		exe := c[0]
		if exe == "echo" {
			res += c[1] + "\n"
		} else {
			_, err := exec.LookPath(exe)
			if err != nil {
				res += "TARDISgo error - executable not found: " + exe + "\n"
				exe = "" // nothing to execute
			}
			if exe == "time" && c[1] == "node" && runtime.GOOS == "linux" {
				c[1] = "nodejs" // for Ubuntu
			}
			if (exe == "haxe" || (exe == "time" && c[1] == "haxe")) && LoadTestZipFS {
				c = append(c, "-resource")
				c = append(c, TestFS)
			}
			if exe != "" {
				out := []byte{}
				out, lastErr = exec.Command(exe, c[1:]...).CombinedOutput()
				if lastErr != nil {
					out = append(out, []byte(lastErr.Error())...)
				}
				if j > 0 { // ignore the output from the compile phase
					res += string(out)
				}
			}
		}
	}
	bc := make(chan bool)
	results <- resChan{res, lastErr, bc}
	<-bc
}
