# OSX script to run "go vet", "golint" and "errcheck" to check completenes, style and error handling
# must be run from tardisgo package directory
go vet github.com/tardisgo/tardisgo github.com/tardisgo/tardisgo/pogo github.com/tardisgo/tardisgo/haxe github.com/tardisgo/tardisgo/haxe/haxegoruntime github.com/tardisgo/tardisgo/tardisgolib
golint *.go */*.go haxe/haxegoruntime/*.go 
errcheck ./...
