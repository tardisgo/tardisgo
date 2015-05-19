package haxegoruntime

import (
	"runtime"
	"unsafe"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

// TODO(haxe): optimize to use the Timer call-back methods for the targets - flash, flash8, java, js, python
func HaxeWait(target *int64, whileTrue *bool) {
	fNow := hx.CallFloat("", "haxe.Timer.stamp", 0)
	firstTarget := *target
	fTarget := reverseNano(*target)
	//println("DEBUG haxeWait:start now, target, *whileTrue diff = ", fNow, *target, *whileTrue, fTarget-fNow)
	for fNow < fTarget && *target == firstTarget && *whileTrue {
		runtime.Gosched() // let other code run
		fNow = hx.CallFloat("", "haxe.Timer.stamp", 0)
		//println("DEBUG haxeWait:loop now, target, *whileTrue diff = ", fNow, *target, *whileTrue, fTarget-fNow)
	}
}

// RuntimeNano returns the current value of the runtime clock in nanoseconds.
func RuntimeNano() int64 { // function body is an Haxe addition
	fv := hx.CallFloat("", "haxe.Timer.stamp", 0)
	// cs and maybe Java have stamp values too large for int64, so set a baseline
	if runtimeNanoBase == 0 {
		//println("DEBUG set runtimeNanoBase")
		runtimeNanoBase = fv
	}
	fv -= runtimeNanoBase
	return int64(fv * 1000000000) // haxe.Timer.stamp is in seconds
}

var runtimeNanoBase float64

func reverseNano(i int64) float64 { // reverse of the above
	return runtimeNanoBase + float64(i)/1000000000
}

// Interface to timers implemented in package runtime.
// Must be in sync with ../runtime/runtime.h:/^struct.Timer$
type runtimeTimer struct { // NOTE a copy of this datastructure is in both time and syscall packages
	i          int
	when       int64
	period     int64
	f          func(interface{}, uintptr) // NOTE: must not be closure
	arg        interface{}
	seq        uintptr
	haxeRuning bool
}

func HaxeTimer(up unsafe.Pointer) {
	rt := (*runtimeTimer)(up)
	defer func() {
		rt.haxeRuning = false
		rt.seq = hx.Null()
	}()
	rt.seq = 0
	rt.haxeRuning = true
again:
	HaxeWait(&rt.when, &rt.haxeRuning) // rt.when is in nanoseconds
	if rt.haxeRuning {
		rt.f(rt.arg, rt.seq)
		rt.seq++
		if rt.period > 0 {
			rt.when += rt.period
			goto again
		}
	}
}

func StartTimer(up unsafe.Pointer) { // function body is an Haxe addition
	StopTimer(up) // just in case it is still running
	rt := (*runtimeTimer)(up)
	for !hx.IsNull(rt.seq) { // wait for the timer to stop -- NOTE potential for deadlock?
		//println("DEBUG Wait for timer to stop")
		runtime.Gosched()
	}
	go HaxeTimer(up)
}

func StopTimer(up unsafe.Pointer) bool { // function body is an Haxe addition
	rt := (*runtimeTimer)(up)
	if rt.haxeRuning {
		rt.haxeRuning = false
		rt.when = 0
		return true
	}
	return false
}
