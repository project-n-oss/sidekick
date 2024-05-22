package shutdown

import (
	"sync/atomic"
)

var shutdownFuncs = make([]func(), 0)
var preShutDown = func() {}
var Terminating = &atomic.Bool{}

func OnShutdown(f func()) {
	shutdownFuncs = append(shutdownFuncs, f)
}

func WithPreShutdown(f func()) {
	preShutDown = f
}
func Shutdown() {
	Terminating.Store(true)
	preShutDown()
	for _, f := range shutdownFuncs {
		f()
	}
}
