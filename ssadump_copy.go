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

	"code.google.com/p/go.tools/go/loader"
	"code.google.com/p/go.tools/go/ssa"
	"code.google.com/p/go.tools/go/ssa/interp"
	"code.google.com/p/go.tools/go/types"

	_ "github.com/tardisgo/tardisgo/haxe" // TARDIS Go addition
	"github.com/tardisgo/tardisgo/pogo"   // TARDIS Go addition
	//"go/parser"                           // TARDIS Go addition
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

// TARDIS Go modification TODO review words here
const usage = `SSA builder and TARDIS Go transpiler (version 0.0.1-experimental).
Usage: tardisgo [<flag> ...] <args> ...
A shameless copy of the ssadump utility, but also writes a 'Go.hx' Haxe file into the 'tardis' sub-directory of the current location (which you must create by hand).
Example:
% tardisgo hello.go
Then to run the tardis/Go.hx file generated, type the command line: "haxe -main tardis.Go --interp", or whatever Haxe compilation options you want to use. 
(Note that to compile for PHP you currently need to add the haxe compilation option "--php-prefix tardisgo" to avoid name confilcts).
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
		fmt.Fprintf(os.Stderr, "TARDISgo: %s.\n", err) // TARDISgo alteration
		os.Exit(1)
	}
}

func doMain() error {
	flag.Parse()
	args := flag.Args()

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

	wordSize = 4 // TARDIS Go addition to force default int size to 32 bits
	//conf.Build.GOARCH = "tardisgo" // TARDIS Go addition to ensure no architecure-specific code will compile
	//conf.Build.GOOS = "tardisgo"   // TARDIS Go addition to ensure no OS-specific code will compile

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
			mode |= ssa.LogPackages | ssa.BuildSerially
		case 'F':
			mode |= ssa.LogFunctions | ssa.BuildSerially
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
		default:
			log.Fatalf("Unknown -build option: '%c'.", c)
		}
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
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	// Profiling support.
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			return err
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Use the initial packages from the command line.
	args, err := conf.FromArgs(args, *testFlag)
	if err != nil {
		return err
	}

	// The interpreter needs the runtime package.
	if *runFlag {
		conf.Import("runtime")
		conf.Import("github.com/tardisgo/tardisgo/golibruntime/runtime") // This required for TARDIS go to run runtime
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

		// TARDIS Go removal as we alter the GOARCH to stop architecture-specific code
		/*
			if runtime.GOARCH != build.Default.GOARCH {
				return fmt.Errorf("cross-interpretation is not yet supported (target has GOARCH %s, interpreter has %s)",
					build.Default.GOARCH, runtime.GOARCH)
			}
		*/
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
		pogo.EntryPoint(main) // TARDIS Go entry point, no return, does os.Exit at end
	}
	return nil
}
