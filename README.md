tardisgo
========

The TARDIS Go -> Haxe -> JavaScript/Flash/Java/C++/C#/PHP/Neko transpiler 

The objective of this project is to enable the same Go code to be re-deployed in  as many different execution environments as possible, thus saving you time and effort. 

It currently targets Haxe (because the Haxe compiler generates 7 other languages), but is written with the intention of targeting other languages too. The long-term vision is to provide a framework that makes it easy to add new transpilers from Go to new target langauages as part of this project.

The Go->Haxe transpiler is demonstrable, but currently generates large, slow and  occasionally incorrect code. The reason for making the project open source is to get <u>your</u> help with the considerable amount of additional testing, optimizing and further development required to make it usable in a production environment.

Planned current use cases: 
- Write a library in Go and call it from  existing Haxe, JavaScript, ActionScript, Java, C++, C# or PHP applications.
- Write a multi-platform client-side application mostly in Go, using [OpenFL](http://openfl.org), [Lime](https://github.com/openfl/lime) or [Kha] (http://kha.ktxsoftware.com/) to target a sub-set of: 
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

Project status: WORKING, EXPERIMENTAL, UNTESTED:
- Almost all of the core [Go language specification] (http://golang.org/ref/spec) is implemented, including single-threaded goroutines.
- The parts of the specification that have not been implemented are the “System considerations” section regarding “Package unsafe” and “Size and alignment guarantees”. 
- Some parts of the Go standard library work, but the bulk has not been implemented or even tested yet. Indeed elements of the standard library may not even be appropriate for transpilation into Haxe. 
- If in doubt, assume the standard package does not work. So fmt.Println("Hello world!") will not transpile, instead use the go builtin function: println("Hello world!")
- A start has been made on the automated integration with Haxe libaries, but this is currently incomplete see: https://github.com/elliott5/gohaxelib
- The only platforms tested are OSX 10.9.1, Ubuntu 13.10 32-bit, Ubuntu 12.04 64-bit and Windows 7 32-bit. 
- The "magnificent seven" Haxe targets tested are JavaScript, Java, Flash, C++, C#, PHP and Neko VM. 
- Some core elements of the design are very likely to change, making the current Haxe and Go APIs experimental.

For more background see the links from: http://tardisgo.github.io/

TARDIS Go can be installed very easily:
```
go get github.com/tardisgo/tardisgo
```

After installation, please go to http://github.com/tardisgo/tardisgo-samples and follow the instructions there for how to run the transpiler. 
For help or general discussions please go to the [Google Group](https://groups.google.com/d/forum/tardisgo). 

If you transpile your own code using TARDIS Go, please report the bugs that you find here, so that they can be fixed.

Thank you for your interest in TARDIS Go. If you would like to get involved in helping the project to advance, I welcome pull requests. However, please contact me (elliott.stoneham@gmail.com) or discuss your plans in the [tardisgo](https://groups.google.com/d/forum/tardisgo) forum before writing any substantial amounts of code so that we can avoid any conflicts.  It would also be great if you wanted to add another target language, or help with building the framework to make adding new target languages easier.

I plan to write substantially more documentation and project notes here, as time allows...
