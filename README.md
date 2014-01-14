tardisgo
========

The TARDIS Go->Haxe transpiler 

BEWARE! THIS PROJECT IS EXPERIMENTAL

### Project status as at 14-Jan-14:
- Summary: the transpiler is demonstrable but incomplete, it requires a huge amount of additional testing and further development to make it usable in the real-world.
- All of the core Go language specification (http://golang.org/ref/spec) should be implemented (except for the “System considerations” section of the specification regarding “Package unsafe” and “Size and alignment guarantees”). 
- Some small elements of the Go standard library work, but the bulk has not yet even been tested.
- A start has been made on integrating with Haxe libaries, but this is currently incomplete see: https://github.com/elliott5/gohaxelib
- The only platforms tested are OSX 10.9.1, Ubuntu 13.10 32-bit, Ubuntu 12.04 64-bit and Windows 7 32-bit. 
- The Haxe targets tested are JavaScript, Java, Flash, C++, C#, PHP and Neko VM. 
- Some core elements of the design may still change, so please do not rely on the current Haxe or Go APIs.

TARDIS Go can be installed by typing:
```
go get github.com/tardisgo/tardisgo
```

To try the project out, please go to http://github.com/tardisgo/tardisgo-samples and follow the instructions there. 

If you write your own code using TARDIS Go, please report the bugs that you find so that they can be fixed.

Thank you for your interest in TARDIS Go. If you would like to get involved in helping the project to advance, please contact me (elliott.stoneham@gmail.com) before writing any code so that we are not both working on the same thing.  
