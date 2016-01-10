# OSX script to run "gometalinter" to check completenes, style and error handling
# must be run from tardisgo package directory
go fmt *.go 
go fmt pogo/*.go 
go fmt haxe/*.go
go fmt tgossa/*.go 
gometalinter --skip=goroot --skip=tests --skip=haxe --deadline=30s ./...
