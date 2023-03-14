package boltrouter

import (
	"fmt"
	"sync"
)

// AtomicVar is a generic thread safe wrapper for variables of any type
type AtomicVar[T any] struct {
	sync.RWMutex
	value T
}

func (v *AtomicVar[T]) String() string {
	return fmt.Sprintf("%v", v.Get())
}

// Get uses sync.RLock to access the variable
func (v *AtomicVar[T]) Get() T {
	v.RLock()
	defer v.RUnlock()
	return v.value
}

// Set uses sync.Lock to write to the variable
func (v *AtomicVar[T]) Set(x T) {
	v.Lock()
	v.value = x
	defer v.Unlock()
}
