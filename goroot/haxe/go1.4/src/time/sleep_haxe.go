// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

package time

import ( // import is an Haxe addition
	"runtime"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

// Sleep pauses the current goroutine for at least the duration d.
// A negative or zero duration causes Sleep to return immediately.
var sleeping = false     // for testing only see interrupt()
func Sleep(d Duration) { // function body is an Haxe addition
	target := runtimeNano()
	target += d.Nanoseconds()
	sleeping = true
	haxeWait(target, &sleeping)
}

func haxeWait(target int64, whileTrue *bool) {
	// TODO(haxe): optimize to use the Timer call-back methods for the targets - flash, flash8, java, js, python
	now := runtimeNano()
	//println("DEBUG haxeWait:start now, target, *whileTrue diff = ", now, target, *whileTrue, target-now)
	for now < target && *whileTrue {
		for wait := int((target - now) / 10000000); wait > 0; wait-- { // one wait per 10 miliseconds
			runtime.Gosched() // let other code run
		}
		now = runtimeNano()
		//println("DEBUG haxeWait:loop now, target, *whileTrue diff = ", now, target, *whileTrue, target-now)
	}
}

// runtimeNano returns the current value of the runtime clock in nanoseconds.
func runtimeNano() int64 { // function body is an Haxe addition
	fv := hx.CallFloat("", "haxe.Timer.stamp", 0)
	// cs and maybe Java have stamp values too large for int64, so set a baseline
	if !runtimeNanoHaveBase {
		runtimeNanoBase = float64(uint64(fv))
		runtimeNanoHaveBase = true
	}
	fv -= runtimeNanoBase
	return int64(fv * 1000000000)
}

var runtimeNanoHaveBase bool
var runtimeNanoBase float64

// Interface to timers implemented in package runtime.
// Must be in sync with ../runtime/runtime.h:/^struct.Timer$
type runtimeTimer struct {
	i          int
	when       int64
	period     int64
	f          func(interface{}, uintptr) // NOTE: must not be closure
	arg        interface{}
	seq        uintptr
	haxeRuning bool
}

// when is a helper function for setting the 'when' field of a runtimeTimer.
// It returns what the time will be, in nanoseconds, Duration d in the future.
// If d is negative, it is ignored.  If the returned value would be less than
// zero because of an overflow, MaxInt64 is returned.
func when(d Duration) int64 {
	if d <= 0 {
		//println("DEBUG -ve duration")
		return runtimeNano()
	}
	t := runtimeNano()
	//println("DEBUG when runtimeNano()", t)
	//println("DEBUG when duration", int64(d))
	t += int64(d)
	//println("DEBUG when +duration", t)
	if t < 0 {
		t = 1<<63 - 1 // math.MaxInt64
	}
	return t
}

func haxeTimer(rt *runtimeTimer) {
again:
	haxeWait(rt.when, &rt.haxeRuning) // rt.when is in nanoseconds
	if rt.haxeRuning {
		rt.f(rt.arg, rt.seq)
		rt.seq++
		if rt.period > 0 {
			rt.when += rt.period
			goto again
		}
	}
}

func startTimer(rt *runtimeTimer) { // function body is an Haxe addition
	rt.haxeRuning = true
	go haxeTimer(rt)
}
func stopTimer(rt *runtimeTimer) bool { // function body is an Haxe addition
	if rt.haxeRuning {
		rt.haxeRuning = false
		return true
	}
	return false
}

// The Timer type represents a single event.
// When the Timer expires, the current time will be sent on C,
// unless the Timer was created by AfterFunc.
// A Timer must be created with NewTimer or AfterFunc.
type Timer struct {
	C <-chan Time
	r runtimeTimer
}

// Stop prevents the Timer from firing.
// It returns true if the call stops the timer, false if the timer has already
// expired or been stopped.
// Stop does not close the channel, to prevent a read from the channel succeeding
// incorrectly.
func (t *Timer) Stop() bool {
	if t.r.f == nil {
		panic("time: Stop called on uninitialized Timer")
	}
	return stopTimer(&t.r)
}

// NewTimer creates a new Timer that will send
// the current time on its channel after at least duration d.
func NewTimer(d Duration) *Timer {
	c := make(chan Time, 1)
	t := &Timer{
		C: c,
		r: runtimeTimer{
			when: when(d),
			f:    sendTime,
			arg:  c,
		},
	}
	startTimer(&t.r)
	return t
}

// Reset changes the timer to expire after duration d.
// It returns true if the timer had been active, false if the timer had
// expired or been stopped.
func (t *Timer) Reset(d Duration) bool {
	if t.r.f == nil {
		panic("time: Reset called on uninitialized Timer")
	}
	w := when(d)
	active := stopTimer(&t.r)
	t.r.when = w
	startTimer(&t.r)
	return active
}

func sendTime(c interface{}, seq uintptr) {
	// Non-blocking send of time on c.
	// Used in NewTimer, it cannot block anyway (buffer).
	// Used in NewTicker, dropping sends on the floor is
	// the desired behavior when the reader gets behind,
	// because the sends are periodic.
	select {
	case c.(chan Time) <- Now():
	default:
	}
}

// After waits for the duration to elapse and then sends the current time
// on the returned channel.
// It is equivalent to NewTimer(d).C.
func After(d Duration) <-chan Time {
	return NewTimer(d).C
}

// AfterFunc waits for the duration to elapse and then calls f
// in its own goroutine. It returns a Timer that can
// be used to cancel the call using its Stop method.
func AfterFunc(d Duration, f func()) *Timer {
	t := &Timer{
		r: runtimeTimer{
			when: when(d),
			f:    goFunc,
			arg:  f,
		},
	}
	startTimer(&t.r)
	return t
}

func goFunc(arg interface{}, seq uintptr) {
	go arg.(func())()
}
