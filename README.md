tardisgo
========

TARDIS Go->Haxe transpiler 

BEWARE! THIS PROJECT IS EXPERIMENTAL

Project status:
- All of the code language specification (http://golang.org/ref/spec) should be implemented, except for the “System considerations” section of the specification regarding “Package unsafe” and “Size and alignment guarantees”. However there are some edge-conditions that are not yet correct and need fixing.
- Some small elements of the Go standard library work, but the bulk has not yet even been tested.
- A start has been made on integrating with Haxe libaries, but this is currently incomplete see: https://github.com/elliott5/gohaxelib
- The project has only been tested on OSX 10.9.1, Ubuntu 13.10 32-bit, Ubuntu 12.04 64-bit and Windows 7 32-bit. 
- The Haxe targets tested are JavaScript, Java, Flash, C++, C#, PHP and Neko VM. 

TARDIS Go can be installed by typing:
```
go get github.com/tardisgo/tardisgo
```

To try the project out, please go to http://github.com/tardigo/tardisgo-samples and follow the instructions there. 

Thank you for your interest in TARDIS Go. If you would like to get involved in helping the project to advance, please contact me (elliott.stoneham@gmail.com) before writing any code so that we are not both working on the same thing.  
