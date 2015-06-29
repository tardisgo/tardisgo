// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modifications:
// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// TARDIS Go is a Go->Haxe transpiler.
// However the tool is written with a "language" interface type separating the generic from the language specific parts of the code, which will allow other languages to be targeted in future.
// To see example code working in your browser please visit http://tardisgo.github.io .
// For simplicity, the current command line tool is simply a modified version of ssadump: a tool for displaying and interpreting the SSA form of Go programs.
package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/interp"
	"golang.org/x/tools/go/ssa/ssautil"
	"golang.org/x/tools/go/types"

	// TARDIS Go additions
	"os/exec"

	_ "github.com/tardisgo/tardisgo/haxe" // TARDIS Go addition
	"github.com/tardisgo/tardisgo/pogo"
)

/*
var buildFlag = flag.String("build", "", `Options controlling the SSA builder.
The value is a sequence of zero or more of these letters:
C	perform sanity [C]hecking of the SSA form.
D	include [D]ebug info for every function.
P	log [P]ackage inventory.
F	log [F]unction SSA code.
S	log [S]ource locations as SSA builder progresses.
G	use binary object files from gc to provide imports (no code).
L	build distinct packages seria[L]ly instead of in parallel.
N	build [N]aive SSA form: don't replace local loads/stores with registers.
`)
*/

var testFlag = flag.Bool("test", false, "Loads test code (*_test.go) for imported packages.")
var LoadTestZipFS = false

const TestFS = "tgotestfs.zip"

var runFlag = flag.Bool("run", false, "Invokes the SSA interpreter on the program.")

var interpFlag = flag.String("interp", "", `Options controlling the SSA test interpreter.
The value is a sequence of zero or more more of these letters:
R	disable [R]ecover() from panic; show interpreter crash instead.
T	[T]race execution of the program.  Best for single-threaded programs!
`)

// TARDIS Go addition
var allFlag = flag.String("haxe", "", "invokes the Haxe compiler (output ignored) and then runs the compiled program on the command line (OSX only): all=all targets, math=math-safe targets (cpp & js -D fullunsafe), interp=haxe interpreter")
var debugFlag = flag.Bool("debug", false, "Instrument the code to enable debugging, add comments, and give more meaningful information during a stack dump (warning: increased code size)")
var traceFlag = flag.Bool("trace", false, "Output trace information for every block visited (warning: huge output)")
var buidTags = flag.String("tags", "", "build tags separated by spaces")
var tgoroot = flag.String("tgoroot", "", "set goroot to the given value")

var modeFlag = ssa.BuilderModeFlag(flag.CommandLine, "build", 0)

// TODO
//var traceFlag = flag.Bool("v", false, "Verbose compiler mode (including files written)")
//var hxPackFlag = flag.String("hxpack", "tardis", "Sets the Haxe package name to use")
//var hxDirFlag = flag.String("hxdir", "tardis", "Sets the directory in which to output generated Haxe code")
//var hxLibFlag = flag.Bool("hxlib", false, "Generates code suitable for use as a Haxe library (no Dead Code Elimination)")

// TARDIS Go modification TODO review words here
const usage = `SSA builder and TARDIS Go transpiler (experimental).
Usage: tardisgo [<flag> ...] <args> ...
A shameless copy of the ssadump utility, but also writes a 'Go.hx' Haxe file into the 'tardis' sub-directory of the current location (which you must create by hand).
Example:
% tardisgo hello.go
Then to compile the tardis/Go.hx file generated, type the command line: "haxe -main tardis.Go -cp tardis -js tardis/go.js", or whatever Haxe compilation options you want to use. 

Use -help to display other options.
`
const ignore = `
Use -help flag to display options.

Examples:
% ssadump -build=FPG hello.go         # quickly dump SSA form of a single package
% ssadump -run -interp=T hello.go     # interpret a program, with tracing
% ssadump -run unicode -- -test.v     # interpret the unicode package's tests, verbosely
` + loader.FromArgsUsage +
	`
When -run is specified, ssadump will find the first package that
defines a main function and run it in the interpreter.
If none is found, the tests of each package will be run instead.
`

// end TARDIS Go modification

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func init() {
	// If $GOMAXPROCS isn't set, use the full capacity of the machine.
	// For small machines, use at least 4 threads.
	if os.Getenv("GOMAXPROCS") == "" {
		n := runtime.NumCPU()
		if n < 4 {
			n = 4
		}
		runtime.GOMAXPROCS(n)
	}
}

func main() {
	if err := doMain(); err != nil {
		fmt.Fprintf(os.Stderr, "TARDISgo: %s\n", err) // TARDISgo alteration
		os.Exit(1)
	}
	os.Exit(0)
}

func doMain() error {
	flag.Parse()
	args := flag.Args()
	return doTestable(args)
}

func doTestable(args []string) error {

	conf := loader.Config{
		Build: &build.Default,
	}

	// TODO(adonovan): make go/types choose its default Sizes from
	// build.Default or a specified *build.Context.
	var wordSize int64 = 8
	switch conf.Build.GOARCH {
	case "386", "arm":
		wordSize = 4
	}

	//if !(*runFlag) {
	wordSize = 4               // TARDIS Go addition to force default int size to 32 bits
	conf.Build.GOARCH = "haxe" // TARDIS Go addition - 32-bit int
	conf.Build.GOOS = "nacl"   // or haxe? TARDIS Go addition - simplest OS-specific code to emulate?
	//}

	conf.Build.BuildTags = strings.Split(*buidTags, " ")

	conf.TypeChecker.Sizes = &types.StdSizes{ // must equal haxe.haxeStdSizes when (!*runFlag)
		MaxAlign: 8,
		WordSize: wordSize,
	}

	var mode ssa.BuilderMode
	/*
		for _, c := range *buildFlag {
			switch c {
			case 'D':
				mode |= ssa.GlobalDebug
			case 'P':
				mode |= ssa.PrintPackages
			case 'F':
				mode |= ssa.PrintFunctions
			case 'S':
				mode |= ssa.LogSource | ssa.BuildSerially
			case 'C':
				mode |= ssa.SanityCheckFunctions
			case 'N':
				mode |= ssa.NaiveForm
			case 'L':
				mode |= ssa.BuildSerially
			case 'I':
				mode |= ssa.BareInits
			default:
				return fmt.Errorf("unknown -build option: '%c'", c)
			}
		}
	*/

	// TARDIS go addition
	if *debugFlag {
		mode |= ssa.GlobalDebug
	}

	var interpMode interp.Mode
	for _, c := range *interpFlag {
		switch c {
		case 'T':
			interpMode |= interp.EnableTracing
		case 'R':
			interpMode |= interp.DisableRecover
		default:
			log.Fatalf("Unknown -interp option: '%c'.", c)
		}
	}

	if len(args) == 0 {
		//fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("%v", usage)
	}

	// Profiling support.
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			return err
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}

	// TODO Eventually this might be better as an environment variable
	//if !(*runFlag) {
	if *tgoroot == "" {
		if conf.Build.GOPATH == "" {
			return fmt.Errorf("GOPATH must be set")
		}
		conf.Build.GOROOT = strings.Split(conf.Build.GOPATH, ":")[0] + "/src/github.com/tardisgo/tardisgo/goroot/haxe/go1.4"
	} else {
		conf.Build.GOROOT = *tgoroot
	}
	//}
	//fmt.Println("DEBUG GOPATH", conf.Build.GOPATH)
	//fmt.Println("DEBUG GOROOT", conf.Build.GOROOT)

	if *testFlag {
		conf.ImportWithTests(args[0]) // assumes you give the full cannonical name of the package to test
		args = args[1:]
	}

	// Use the initial packages from the command line.
	_, err := conf.FromArgs(args, *testFlag)
	if err != nil {
		return err
	}

	// The interpreter needs the runtime package.
	//if *runFlag {
	conf.Import("runtime")
	//} else {
	// TARDIS GO additional line to add the language specific go runtime code
	conf.Import(pogo.LanguageList[pogo.TargetLang].Goruntime) // TODO add code to set pogo.TargetLang when more than one of them
	//}

	// Load, parse and type-check the whole program, including the type definitions.
	iprog, err := conf.Load()
	if err != nil {
		return err
	}

	// Create and build SSA-form program representation.
	*modeFlag |= mode | ssa.SanityCheckFunctions
	prog := ssautil.CreateProgram(iprog, *modeFlag)

	prog.BuildAll()

	// Run the interpreter.
	if *runFlag {
		var main *ssa.Package
		pkgs := prog.AllPackages()
		if *testFlag {
			// If -test, run all packages' tests.
			if len(pkgs) > 0 {
				main = prog.CreateTestMainPackage(pkgs...)
			}
			if main == nil {
				return fmt.Errorf("no tests")
			}
		} else {
			// Otherwise, run main.main.
			for _, pkg := range pkgs {
				if pkg.Object.Name() == "main" {
					main = pkg
					if main.Func("main") == nil {
						return fmt.Errorf("no func main() in main package")
					}
					break
				}
			}
			if main == nil {
				return fmt.Errorf("no main package")
			}
		}

		// NOTE TARDIS Go removal of this test required if we alter the GOARCH to stop architecture-specific code
		//if runtime.GOARCH != build.Default.GOARCH {
		//	return fmt.Errorf("cross-interpretation is not yet supported (target has GOARCH %s, interpreter has %s)",
		//		build.Default.GOARCH, runtime.GOARCH)
		//}

		interp.Interpret(main, interpMode, conf.TypeChecker.Sizes, main.Object.Path(), args)
	} else {
		// if not interpreting...
		// TARDIS Go additions: copy run interpreter code above, but call pogo class
		var main *ssa.Package
		pkgs := prog.AllPackages()
		//fmt.Println("DEBUG pkgs:", pkgs)

		if *testFlag {
			// If -test, run all packages' tests.
			if len(pkgs) > 0 {
				main = prog.CreateTestMainPackage(pkgs...)
			}
			if main == nil {
				return fmt.Errorf("no tests")
			}
			fd, err := os.Open(TestFS)
			fd.Close()
			if err == nil {
				LoadTestZipFS = true
				for l := range pogo.LanguageList {
					pogo.LanguageList[l].TestFS = TestFS
				}
			}
		} else {
			// Otherwise, run main.main.
			for _, pkg := range pkgs {
				if pkg.Object.Name() == "main" {
					main = pkg
					if main.Func("main") == nil {
						return fmt.Errorf("no func main() in main package")
					}
					break
				}
			}
			if main == nil {
				return fmt.Errorf("no main package")
			}
		}
		/*
			if runtime.GOARCH != build.Default.GOARCH {
				return fmt.Errorf("cross-interpretation is not yet supported (target has GOARCH %s, interpreter has %s)",
					build.Default.GOARCH, runtime.GOARCH)
			}

			interp.Interpret(main, interpMode, conf.TypeChecker.Sizes, main.Object.Path(), args)
		*/
		pogo.DebugFlag = *debugFlag
		pogo.TraceFlag = *traceFlag
		err = pogo.EntryPoint(main) // TARDIS Go entry point, returns an error
		if err != nil {
			return err
		}
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
				go doTarget(cmd, results)
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
				go doTarget(cmd, results)
			}
			for _ = range mathCmds {
				r := <-results
				fmt.Println(r.output)
				if r.err != nil {
					os.Exit(1) // exit with an error if the test fails
				}
				r.backChan <- true
			}

		case "interp", "cpp", "cs", "js", "jsfu", "java": // for running tests
			switch *allFlag {
			case "interp":
				go doTarget([][]string{
					[]string{"echo", ``}, // Output from this line is ignored
					[]string{"echo", `"Neko (haxe --interp):"`},
					[]string{"time", "haxe", "-main", "tardis.Go", "-cp", "tardis", "--interp"},
				}, results)
			case "cpp":
				go doTarget([][]string{
					[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-cpp", "tardis/cpp"},
					[]string{"echo", `"CPP:"`},
					[]string{"time", "./tardis/cpp/Go"},
				}, results)
			case "cs":
				go doTarget([][]string{
					[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-cs", "tardis/cs"},
					[]string{"echo", `"CS:"`},
					[]string{"time", "mono", "./tardis/cs/bin/Go.exe"},
				}, results)
			case "js":
				go doTarget([][]string{
					[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-D", "uselocalfunctions", "-js", "tardis/go.js"},
					[]string{"echo", `"Node/JS:"`},
					[]string{"time", "node", "tardis/go.js"},
				}, results)
			case "jsfu":
				go doTarget([][]string{
					[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-D", "uselocalfunctions", "-D", "fullunsafe", "-js", "tardis/go-fu.js"},
					[]string{"echo", `"Node/JS using fullunsafe memory mode (js dataview):"`},
					[]string{"time", "node", "tardis/go-fu.js"},
				}, results)
			case "java":
				go doTarget([][]string{
					[]string{"haxe", "-main", "tardis.Go", "-cp", "tardis", "-dce", "full", "-D", "inlinepointers", "-java", "tardis/java"},
					[]string{"echo", `"Java:"`},
					[]string{"time", "java", "-jar", "tardis/java/Go.jar"},
				}, results)
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
	return nil
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

func doTarget(cl [][]string, results chan resChan) {
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
