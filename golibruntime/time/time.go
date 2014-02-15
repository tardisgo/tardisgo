// Package time is not implemented in TARDIS Go, this is not-fully-working TEST code, for OSX only
package time

import "github.com/tardisgo/tardisgo/tardisgolib"

// Provided by package runtime.
func now() (sec int64, nsec int32) {
	return int64(tardisgolib.HAXE("GOint64.ofFloat(Date.now().getTime()/1000.0);")),
		int32(tardisgolib.HAXE("cast(Date.now().getTime()%1000.0,Int)*1000000;"))
}

// Interface to timers implemented in package runtime.
// Must be in sync with ../runtime/runtime.h:/^struct.Timer$
type runtimeTimer struct {
	i      int32
	when   int64
	period int64
	f      func(int64, interface{}) // NOTE: must not be closure
	arg    interface{}
}

func startTimer(*runtimeTimer) {
	panic("time.startTimer() NOT IMPLEMENTED")
}
func stopTimer(*runtimeTimer) bool {
	panic("time.stopTimer() NOT IMPLEMENTED")
	return false
}
