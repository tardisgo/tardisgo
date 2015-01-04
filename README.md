# TARDIS Go -> Haxe transpiler

#### Haxe -> JavaScript / ActionScript / Java / C++ / C# / PHP / Neko

[![Build Status](https://travis-ci.org/tardisgo/tardisgo.png?branch=master)](https://travis-ci.org/tardisgo/tardisgo)
[![GoDoc](https://godoc.org/github.com/tardisgo/tardisgo?status.png)](https://godoc.org/github.com/tardisgo/tardisgo)
[![status](https://sourcegraph.com/api/repos/github.com/tardisgo/tardisgo/badges/status.png)](https://sourcegraph.com/github.com/tardisgo/tardisgo)

## Objectives:
The objective of this project is to enable the same [Go](http://golang.org) code to be re-deployed in  as many different execution environments as possible, thus saving development time and effort. 
The long-term vision is to provide a framework that makes it easy to target many languages as part of this project.

The first language targeted is [Haxe](http://haxe.org), because the Haxe compiler generates 7 other languages and is already well-proven for making multi-platform client-side applications, mostly games. 

Target short-term use-case is writing multi-platform client-side applications in Go using the APIs available in the Haxe ecosystem, including:
- The standard [Haxe APIs](http://api.haxe.org/) for JavaScript and Flash.
- [OpenFL](http://openfl.org) "Open Flash" to target HTML5, Windows, Mac, Linux, iOS, Android, BlackBerry, Firefox OS, Tizen and Flash, using compiled auto-generated C++ code where possible (star users are TiVo and Prezzi).

Target medium-term use-case will be to make the wider Haxe ecosystem available to Go programmers, including:
- [Flambe](https://github.com/aduros/flambe) cross-platform game engine targeting mobile via Adobe AIR (star users are Disney and Nickelodeon).
- [Kha](http://kha.ktxsoftware.com/) "World's most portable software platform" additionaly targeting Xbox and Playstation.
- Or see a list of many more projects [here](http://old.haxe.org/doc/libraries) (though that page is not exhaustive).

Target long-term use-cases (once the generated code and runtime environment is more efficient):
- For the Haxe community: provide access to the portable elements of Go's extensive libraries and open-source code base.
- For the Go community: write a library in Go and call it from  existing Haxe, JavaScript, ActionScript, Java, C# or PHP applications (in C++ you would just link as normal). 

For more background and on-line examples see the links from: http://tardisgo.github.io/

## Project status: a working proof of concept
####  DEMONSTRABLE, EXPERIMENTAL, INCOMPLETE, UN-OPTIMIZED, UNSTABLE API

All of the core [Go language specification](http://golang.org/ref/spec) is implemented, including single-threaded goroutines and channels. However the package "reflect", which is mentioned in the core specification, is not yet supported. 

Goroutines are implemented as co-operatively scheduled co-routines. Other goroutines are automatically scheduled every time there is a channel operation or goroutine creation (or call to a function which uses channels or goroutines through any called funciton). So loops without channel operations may never give up control. The function tardisgolib.Gosched() provides a convenient way to give up control (it perfoms a channel select operation).  

Some parts of the Go standard library work, as you can see in the [example TARDIS Go code](http://github.com/tardisgo/tardisgo-samples), but the bulk has not been  tested or implemented yet. If the standard package is not mentioned in the notes below, please assume it does not work. So fmt.Println("Hello world!") will not transpile, instead use the go builtin function: println("Hello world!").  

Some standard Go library packages do not call any runtime C or assembler functions and will probably work OK (though their tests still need to be rewritten and run to validate their correctness), these include:
 unsafe, errors, unicode, unicode/utf8, unicode/utf16, sort, container/heap, container/list & container/ring.

Other standard libray packages make limited use of runtime C or assembler functions without using the actual Go "runtime" or "os" packages. These limited runtime functions have been emulated for a small number of packages (math, strings, bytes, strconv) though this remains a work-in-progress. At present, standard library packages which rely on the Go "runtime", "os", "reflect" or other low-level packages are not implemented.

A start has been made on the automated integration with Haxe libraries, but this is incomplete and the API unstable, see the tardisgolib/hx directory and gohaxelib repository for the story so far. 

TARDIS Go specific runtime functions are available in [tardisgolib](https://github.com/tardisgo/tardisgo/tree/master/tardisgolib):
```
import "github.com/tardisgo/tardisgo/tardisgolib" // runtime functions for TARDIS Go, API unstable
```

The code is developed and tested on OS X 10.9.5, using Go 1.4 and Haxe 3.1.3. The CI tests run on 64-bit Ubuntu. 

No other platforms are currently regression tested, although the project has been run on Ubuntu 32-bit and Windows 7 32-bit. Compilation to the C# target fails on Win-7 and PHP is flakey (but you probably knew that).

## Installation and use:
 
Dependencies:
```
go get golang.org/x/tools
```

TARDIS Go:
```
go get -u github.com/tardisgo/tardisgo
```

If tardisgo is not installing and there is a green "build:passing" icon at the top of this page, please e-mail [Elliott](https://github.com/elliott5)!

To translate Go to Haxe, from the directory containing your .go files type the command line: 
```
tardisgo yourfilename.go 
``` 
A single Go.hx file will be created in the tardis subdirectory.

To run your transpiled code you will first need to install [Haxe](http://haxe.org).

Then to run the tardis/Go.hx file generated above, type the command line: 
```
haxe -main tardis.Go --interp
```
... or whatever [Haxe compilation options](http://haxe.org/doc/compiler) you want to use. 
See the [tgoall.sh](https://github.com/tardisgo/tardisgo-samples/blob/master/scripts/tgoall.sh) script for simple examples.

There is a TARDIS Go only Haxe compilation flag for JS to control use of the dataview method of object access (this has a smaller memory footprint and allows unsafe pointers to be modelled more accurately, but goes slower than the standard method): 
```
haxe -main tardis.Go -D dataview -js tardisgo.js
```

To add Go build tags, use -tags 'name1 name2'. Note that particular Go build tags are required when compiling for OpenFL using the [pre-built Haxe API definitions](https://github.com/tardisgo/gohaxelib). 

Use the "-debug=true" tardisgo compilation flag to instrument the code and add automated comments to the Haxe. When you experience a panic in this mode the latest Go source code line information and local variables appears in the stack dump. For the C++ & Neko (--interp) targets, a very simple debugger is also available by using the "-D godebug" Haxe flag, for example to use it in C++ type:
```
tardisgo -debug=true myprogram.go
haxe -main tardis.Go -dce full -D godebug -cpp tardis/cpp
./tardis/cpp/Go
``` 
To get a list of commands type "?" followed by carrage return, after the 1st break location is printed (there is no prompt character). 

To run cross-target command-line tests as quickly as possible, the "-testall" flag  concurrently runs the Haxe compiler and executes the resulting code for all supported targets (with compiler output suppressed and results appearing in the order they complete, with an execution time):
```
tardisgo -testall myprogram.go
```

If you can't work-out what is going on prior to a panic, you can add the "-trace" tardisgo compilation flag to instrument the code even further, printing out every part of the code visited. But be warned, the output can be huge.

PHP specific issues:
* to compile for PHP you currently need to add the haxe compilation option "--php-prefix tgo" to avoid name conflicts
* very long PHP class/file names may cause name resolution problems on some platforms

## Next steps:
Please go to http://github.com/tardisgo/tardisgo-samples for example Go code modified to work with tardisgo.

For a small technical FAQ, please see the [Wiki page](https://github.com/tardisgo/tardisgo/wiki). 

For public help or discussion please go to the [Google Group](https://groups.google.com/d/forum/tardisgo); or feel free to e-mail [Elliott](https://github.com/elliott5) direct to discuss any issues if you prefer.

The documentation is sparse at present, if there is some aspect of the system that you want to know more about, please let [Elliott](https://github.com/elliott5) know and he will prioritise that area to add to the wiki.

If you transpile your own code using TARDIS Go, please report the bugs that you find here, so that they can be fixed.

## Future plans:

The focus of short-term development is to get the Haxe implementation production ready. In particular, smooth interaction with external Haxe code is required to make the project useful for real work, [an experimental version of which is available](https://github.com/tardisgo/gohaxelib). 

In speed terms, the planned next release of Haxe (3.2) will contain cross-platform implementation of JS [typed arrays](https://github.com/HaxeFoundation/haxe/issues/3073) which, with other improvements, will allow for faster execution times by making less use of the Haxe "Dynamic" type to store values on the heap. (See the -dataview js haxe compilation flag for a partial implementation.)

Longer term development priorities:
- For all Go standard libraries, report testing and implementation status
- Improve integration with Haxe code and libraries, automating as far as possible - [in progress](https://github.com/tardisgo/gohaxelib)
- Improve currently poor execution speeds and update benchmarking results
- Research and publish the best methods to use TARDIS Go to create multi-platform client-side applications - [in progress](https://github.com/tardisgo/tardisgo-samples/tree/master/openfl)
- Improve debug and profiling capabilities
- Add command line flags to control options
- Publish more explanation and documentation
- Move more of the runtime into Go (rather than Haxe) to make it more portable 
- Implement other target languages...

If you would like to get involved in helping the project to advance, that would be wonderful. However, please contact [Elliott](https://github.com/elliott5) or discuss your plans in the [tardisgo](https://groups.google.com/d/forum/tardisgo) forum before writing any substantial amounts of code to avoid any conflicts. 

## License:

MIT license, please see the license file.
