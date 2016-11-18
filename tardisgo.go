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

	"go/types"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/interp"
	"golang.org/x/tools/go/ssa/ssautil"

	// TARDIS Go additions

	_ "github.com/tardisgo/tardisgo/asmgo" // TARDIS Go addition
	"github.com/tardisgo/tardisgo/haxe"    // TARDIS Go addition
	// TARDIS Go addition
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
var loadTestZipFS = false

const testFS = "tgotestfs.zip"

var runFlag = flag.Bool("run", false, "Invokes the SSA interpreter on the program.")

var interpFlag = flag.String("interp", "", `Options controlling the SSA test interpreter.
The value is a sequence of zero or more more of these letters:
R	disable [R]ecover() from panic; show interpreter crash instead.
T	[T]race execution of the program.  Best for single-threaded programs!
`)

// TARDIS Go addition
var targetFlag = flag.String("target", "haxe", "language to target (default is haxe)")
var allFlag = flag.String("haxe", "", "invokes the Haxe compiler (output ignored) and then runs the compiled program on the command line (OSX only): all=all targets, math=math-safe targets (cpp & js -D fullunsafe), interp=haxe interpreter")
var debugFlag = flag.Bool("debug", false, "Instrument the code to enable debugging, add comments, and give more meaningful information during a stack dump (warning: increased code size)")
var traceFlag = flag.Bool("trace", false, "Output trace information for every block visited (warning: huge output)")
var buidTags = flag.String("tags", "", "build tags separated by spaces")
var tgoroot = flag.String("tgoroot", "", "set goroot to the given value")

//var modeFlag = ssa.BuilderModeFlag(flag.CommandLine, "build", 0)
var modeFlag = ssa.BuilderMode(0)

func init() { flag.Var(&modeFlag, "build", ssa.BuilderModeDoc) }

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

/*
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
*/
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

	// TARDISgo addition
	langName := *targetFlag
	langEntry, e := pogo.FindTargetLang(langName)
	if e != nil {
		return e
	}

	// TODO(adonovan): make go/types choose its default Sizes from
	// build.Default or a specified *build.Context.
	var wordSize int64 = 8
	switch conf.Build.GOARCH {
	case "386", "arm":
		wordSize = 4
	}

	if *runFlag {
		// nothing here at the moment
	} else {
		wordSize = 4                 // TARDIS Go addition to force default int size to 32 bits
		conf.Build.GOOS = "nacl"     // TARDIS Go addition - simplest OS-specific code to emulate?
		conf.Build.GOARCH = langName // TARDIS Go addition
	}

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

	if !(*runFlag) {
		if *tgoroot == "" {
			if conf.Build.GOPATH == "" {
				return fmt.Errorf("GOPATH must be set")
			}
			langGOROOT := pogo.LanguageList[langEntry].GOROOT
			if langGOROOT != "" {
				conf.Build.GOROOT = strings.Split(conf.Build.GOPATH, ":")[0] + langGOROOT
			} else {
				if conf.Build.GOROOT == "" {
					return fmt.Errorf("GOROOT must be set (hint: use -tgoroot flag)")
				}
			}
		} else {
			conf.Build.GOROOT = *tgoroot
		}
	}
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
	if *runFlag {
		conf.Import("runtime")
	} else {
		// TARDIS GO additional line to add the language specific go runtime code
		rt := pogo.LanguageList[langEntry].Goruntime
		if rt != "" {
			conf.Import(rt)
		}
	}

	// Load, parse and type-check the whole program.
	iprog, err := conf.Load()
	if err != nil {
		return err
	}

	// Create and build SSA-form program representation.
	modeFlag |= mode | ssa.SanityCheckFunctions
	prog := ssautil.CreateProgram(iprog, modeFlag)

	prog.Build()

	var main *ssa.Package
	pkgs := prog.AllPackages()
	//fmt.Println("DEBUG pkgs:", pkgs)

	testFSname := ""
	if *testFlag {
		// If -test, run all packages' tests.
		if len(pkgs) > 0 {
			main = prog.CreateTestMainPackage(pkgs[0])
		}
		if main == nil {
			return fmt.Errorf("no tests")
		}
		fd, openErr := os.Open(testFS)
		closeErr := fd.Close()
		if openErr == nil && closeErr == nil {
			loadTestZipFS = true
			testFSname = testFS
		}
	} else {
		// Otherwise, run main.main.
		for _, pkg := range pkgs {
			if pkg.Pkg.Name() == "main" {
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

	if *runFlag { // Run the golang.org/x/tools/go/ssa/interp interpreter.
		interp.Interpret(main, interpMode, conf.TypeChecker.Sizes, main.Pkg.Path(), args)
	} else {
		comp, err := pogo.Compile(main, *debugFlag, *traceFlag, langName, testFSname) // TARDIS Go entry point, returns an error
		if err != nil {
			return err
		}
		comp.Recycle()

		switch langName {
		case "haxe":
			haxe.RunHaxe(allFlag, loadTestZipFS, testFSname)
		}
	}
	return nil
}
