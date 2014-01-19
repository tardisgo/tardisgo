# TARDIS Go transpiler

#### Go -> Haxe -> JavaScript / ActionScript / Java / C++ / C# / PHP / Neko

## Objectives:
The objective of this project is to enable the same Go code to be re-deployed in  as many different execution environments as possible, thus saving development time and effort. 
The long-term vision is to provide a framework that makes it easy to target new languages as part of this project.

The first language targeted is [Haxe](http://haxe.org), because the Haxe compiler generates 7 other languages and is already well-proven for making multi-platform client-side applications, mostly games. 
Planned current use cases: 
- Write a library in Go and call it from  existing Haxe, JavaScript, ActionScript, Java, C++, C# or PHP applications.
- Write a multi-platform client-side application in Go and Haxe, using [OpenFL](http://openfl.org) / [Lime](https://github.com/openfl/lime) or [Kha] (http://kha.ktxsoftware.com/) to target a sub-set of: 
Windows,
Mac,
Linux,
iOS,
Android,
BlackBerry,
Tizen,
Emscripten,
HTML5,
webOS,
Flash,
Xbox and PlayStation.

For more background see the links from: http://tardisgo.github.io/

## Project status: 
#### WORKING, EXPERIMENTAL, INCOMPLETE, IN-TESTING, UN-OPTIMIZED

> "Premature optimization is the root of all evil (or at least most of it) in programming." - Donald Knuth

Almost all of the core [Go language specification] (http://golang.org/ref/spec) is implemented, including single-threaded goroutines and channels. 

Some parts of the Go standard library work, but the bulk has not been implemented or even tested yet. Indeed elements of the standard library may not even be appropriate for transpilation into Haxe. If in doubt, assume the standard package does not work. So fmt.Println("Hello world!") will not transpile, instead use the go builtin function: println("Hello world!"). Unsafe pointers are not currently supported and the reflection package is not yet implemented. 

A start has been made on the automated integration with Haxe libraries, but this is currently incomplete see: https://github.com/elliott5/gohaxelib

The only development platforms tested are OSX 10.9.1, Ubuntu 13.10 32-bit, Ubuntu 12.04 64-bit and Windows 7 32-bit. 

(TODO a development road-map and much more documentation)

## Installation and use:

TARDIS Go can be installed very easily:
```
go get github.com/tardisgo/tardisgo
```
From the directory containing your .go files, first create a "tardis" sub-directory (TODO review this requirement):
```
mkdir tardis
```
Then to translate Go to Haxe, go to the directory containing your .go files (TODO review) and type the command line: 
```
tardisgo filename.go filename2.go
``` 
A single Go.hx file will be created in the tardis subdirectory.

To run your transpiled code you will first need to install [Haxe](http://haxe.org).

Then to run the tardis/Go.hx file generated above, type the command line: 
```
haxe -main tardis.Go --interp
```
... or whatever Haxe compilation options you want to use. Note that to compile for PHP you currently need to add the haxe compilation option "--php-prefix tardisgo" to avoid name conflicts.

## Next steps:
Please go to http://github.com/tardisgo/tardisgo-samples for example Go code modified to work with tardisgo.

For help or general discussions please go to the [Google Group](https://groups.google.com/d/forum/tardisgo). 

The documentation is sparse at present, if there is some aspect of the system that you want to know more about, please let [me](https://github.com/elliott5) know and I will prioritise that area.

If you transpile your own code using TARDIS Go, please report the bugs that you find here, so that they can be fixed.

If you would like to get involved in helping the project to advance, I welcome pull requests. However, please contact [me](https://github.com/elliott5) or discuss your plans in the [tardisgo](https://groups.google.com/d/forum/tardisgo) forum before writing any substantial amounts of code so that we can avoid any conflicts. 

## License:
MIT license, please see the license file.