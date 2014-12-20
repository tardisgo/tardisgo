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
	"golang.org/x/tools/go/types"

	// TARDIS Go additions
	"os/exec"

	_ "github.com/tardisgo/tardisgo/haxe" // TARDIS Go addition
	"github.com/tardisgo/tardisgo/pogo"
)

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

var testFlag = flag.Bool("test", false, "Loads test code (*_test.go) for imported packages.")

var runFlag = flag.Bool("run", false, "Invokes the SSA interpreter on the program.")

var interpFlag = flag.String("interp", "", `Options controlling the SSA test interpreter.
The value is a sequence of zero or more more of these letters:
R	disable [R]ecover() from panic; show interpreter crash instead.
T	[T]race execution of the program.  Best for single-threaded programs!
`)

// TARDIS Go addition
var allFlag = flag.Bool("testall", false, "For all targets: invokes the Haxe compiler (output ignored) and then runs the compiled program on the command line (OSX only)")
var debugFlag = flag.Bool("debug", false, "Instrument the code to enable debugging, add comments, and give more meaningful information during a stack dump (warning: increased code size)")
var traceFlag = flag.Bool("trace", false, "Output trace information for every block visited (warning: huge output)")
var buidTags = flag.String("tags", "", "build tags separated by spaces")

// TARDIS Go modification TODO review words here
const usage = `SSA builder and TARDIS Go transpiler (experimental).
Usage: tardisgo [<flag> ...] <args> ...
A shameless copy of the ssadump utility, but also writes a 'Go.hx' Haxe file into the 'tardis' sub-directory of the current location (which you must create by hand).
Example:
% tardisgo hello.go
Then to run the tardis/Go.hx file generated, type the command line: "haxe -main tardis.Go --interp", or whatever Haxe compilation options you want to use. 
(Note that to compile for PHP you currently need to add the haxe compilation option "--php-prefix tardisgo" to avoid name confilcts).
use -help to display options
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
		Build:         &build.Default,
		SourceImports: true,
	}

	// TODO(adonovan): make go/types choose its default Sizes from
	// build.Default or a specified *build.Context.
	var wordSize int64 = 8
	switch conf.Build.GOARCH {
	case "386", "arm":
		wordSize = 4
	}

	wordSize = 4              // TARDIS Go addition to force default int size to 32 bits
	conf.Build.GOARCH = "tgo" // or 386? TARDIS Go addition to ensure 32-bit int
	conf.Build.GOOS = "tgo"   // or nacl? TARDIS Go addition to ensure simplest OS-specific code to emulate

	conf.Build.BuildTags = strings.Split(*buidTags, " ")

	conf.TypeChecker.Sizes = &types.StdSizes{
		MaxAlign: 8,
		WordSize: wordSize,
	}

	var mode ssa.BuilderMode
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
		case 'G':
			conf.SourceImports = false
		case 'L':
			mode |= ssa.BuildSerially
		case 'I':
			mode |= ssa.BareInits
		default:
			return fmt.Errorf("unknown -build option: '%c'", c)
		}
	}

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

	// Really need to find a way to replace entire packages, this experiment is WIP so excluded for the latest push
	// conf.Build.GOROOT = "/Users/Elliott/Desktop/tardisgo/goroot/go1.4rc2" // TODO sort out to a sensible location
	// fmt.Println("DEBUG GOPATH", conf.Build.GOPATH) // TODO set GOROOT by concatonating with this unless TGOPATH is set

	if *testFlag {
		conf.ImportWithTests(args[0]) // assumes you give the full cannonical name of package
		args = args[1:]
	}

	// Use the initial packages from the command line.
	_, err := conf.FromArgs(args, *testFlag)
	if err != nil {
		return err
	}

	// TODO will -run still work?
	// The interpreter needs the runtime package.
	if *runFlag {
		conf.Import("runtime")
	}

	// TARDIS GO additional line to add the language specific go runtime code
	conf.Import(pogo.LanguageList[pogo.TargetLang].Goruntime) // TODO add code to set pogo.TargetLang when more than one of them

	// Load, parse and type-check the whole program.
	iprog, err := conf.Load()
	if err != nil {
		return err
	}

	// Create and build SSA-form program representation.
	prog := ssa.Create(iprog, mode)

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
		if runtime.GOARCH != build.Default.GOARCH {
			return fmt.Errorf("cross-interpretation is not yet supported (target has GOARCH %s, interpreter has %s)",
				build.Default.GOARCH, runtime.GOARCH)
		}

		interp.Interpret(main, interpMode, conf.TypeChecker.Sizes, main.Object.Path(), args)
	}

	// TARDIS Go additions: copy run interpreter code above, but call pogo class
	if true {
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
		if *allFlag {
			for _, dir := range dirs {
				err := os.RemoveAll(dir) //
				if err != nil {
					fmt.Println("Error deleting existing '" + dir + "' directory: " + err.Error())
				}
			}
			results := make(chan resChan)
			for _, cmd := range targets {
				go doTarget(cmd, results)
			}
			for _ = range targets {
				r := <-results
				fmt.Println(r.output)
				r.backChan <- true
			}
		}
	}
	return nil
}

var dirs = []string{"tardis/cpp", "tardis/java", "tardis/cs", "tardis/php"}

var targets = [][][]string{
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-cpp", "tardis/cpp"},
		[]string{"echo", `"CPP:"`},
		[]string{"time", "./tardis/cpp/Go"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-java", "tardis/java"},
		[]string{"echo", `"Java:"`},
		[]string{"time", "java", "-jar", "tardis/java/Go.jar"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-cs", "tardis/cs"},
		[]string{"echo", `"CS:"`},
		[]string{"time", "mono", "./tardis/cs/bin/Go.exe"},
	},
	// Seldom works, so removed
	//[][]string{
	//	[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-neko", "tardis/tardisgo.n"},
	//	[]string{"echo", `"Neko:"`},
	//	[]string{"time", "neko", "tardis/tardisgo.n"},
	//},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-js", "tardis/tardisgo.js"},
		[]string{"echo", `"Node/JS:"`},
		[]string{"time", "node", "tardis/tardisgo.js"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-D", "dataview", "-js", "tardis/tardisgo-dv.js"},
		[]string{"echo", `"Node/JS (using dataview):"`},
		[]string{"time", "node", "tardis/tardisgo-dv.js"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-swf", "tardis/tardisgo.swf"},
		[]string{"echo", `"Opening swf file (Chrome as a file association for swf works to test on OSX):"` + "\n"},
		[]string{"open", "tardis/tardisgo.swf"},
	},
	[][]string{
		[]string{"haxe", "-main", "tardis.Go", "-dce", "full", "-php", "tardis/php", "--php-prefix", "tgo"},
		[]string{"echo", `"PHP:"`},
		[]string{"time", "php", "tardis/php/index.php"},
	},
	[][]string{
		[]string{"echo", ``}, // Output from this line is ignored
		[]string{"echo", `"Neko (haxe --interp):"`},
		[]string{"time", "haxe", "-main", "tardis.Go", "--interp"},
	},
}

type resChan struct {
	output   string
	backChan chan bool
}

func doTarget(cl [][]string, results chan resChan) {
	res := ""
	for j, c := range cl {
		exe := c[0]
		if exe == "echo" {
			res += c[1] + "\n"
		} else {
			_, err := exec.LookPath(exe)
			if err != nil {
				switch exe {
				case "node":
					exe = "nodejs" // for Ubuntu
				default:
					res += "TARDISgo error - executable not found: " + exe + "\n"
					exe = "" // nothing to execute
				}
			}
			if exe != "" {
				out, err := exec.Command(exe, c[1:]...).CombinedOutput()
				if err != nil {
					out = append(out, []byte(err.Error())...)
				}
				if j > 0 { // ignore the output from the compile phase
					res += string(out)
				}
			}
		}
	}
	bc := make(chan bool)
	results <- resChan{res, bc}
	<-bc
}
