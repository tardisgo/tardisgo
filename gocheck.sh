# OSX script to run "go vet", "golint" and "errcheck" to check completenes, style and error handling
# must be run from tardisgo package directory
go fmt *.go 
go fmt pogo/*.go 
go fmt haxe/*.go
go fmt tgossa/*.go 
go vet github.com/tardisgo/tardisgo github.com/tardisgo/tardisgo/pogo github.com/tardisgo/tardisgo/haxe  github.com/tardisgo/tardisgo/tgossa
golint *.go */*.go  
errcheck . ./pogo ./haxe ./tgossa
