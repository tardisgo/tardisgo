package haxegoruntime

import (
	"errors"
	"runtime"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

var JScallbackOK = false

// Main allows code that might be targeted to broser-based JS code to schedule TARDISgo code periodically, separate from call-backs.
// For non-JS targets it simply calls the given mainFN.
// This should be the last function called in main.main()
func BrowserMain(mainFN func(), msInvocationInterval, runLimit int) {
	if runtime.GOARCH == "js" { // probably only want to do this in a browser
		JScallbackOK = true
		ht := hx.New("js", "haxe.Timer", 1, msInvocationInterval)
		go func() { // put the main program into a goroutine
			mainFN()
			hx.Meth("js", ht, "haxe.Timer", "stop", 0)
		}()
		hx.SetInt("js", "Scheduler.runLimit", runLimit)
		hx.Code("js", "_a.param(0).val.run=Scheduler.timerEventHandler;", ht)
	} else {
		mainFN()
	}
}

type urlReply struct {
	err error
	dat string
}

func GetURL(url string) (string, error) {
	var urc = make(chan urlReply)
	defer close(urc)
	r := getUrl(url, urc)
	return r.dat, r.err
}

func getUrl(url string, urc chan urlReply) urlReply {
	httpHandler := hx.New("", "haxe.Http", 1, url)
	hx.Code("", "_a.param(0).val.onData=_a.param(1).val;", httpHandler, hx.CallbackFunc(
		func(data string) {
			urc <- urlReply{err: nil, dat: data}
		}))
	hx.Code("", "_a.param(0).val.onError=_a.param(1).val;", httpHandler, hx.CallbackFunc(
		func(msg string) {
			urc <- urlReply{err: errors.New(msg), dat: ""}
		}))
	hx.Meth("", httpHandler, "haxe.Http", "request", 0)
	return <-urc
}
