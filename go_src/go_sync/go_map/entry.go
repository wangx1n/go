package go_map

import (
	"sync/atomic"
	"unsafe"
)

type entry struct {
	p unsafe.Pointer
}

func newEntry(i interface{}) *entry {
	return &entry{
		p: unsafe.Pointer(&i),
	}
}

func (e *entry) load() (value interface{}, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expunged {
		return nil, false
	}
	return *(*interface{})(p), true
}

func (e *entry) tryStore(value interface{}) bool {
	for {
		p := atomic.LoadPointer(&e.p)

		if p == expunged {
			return false
		}

		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(&value)) {
			return true
		}
	}
}

func (e *entry) unExpungedLock() bool {
	return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}

func (e *entry) storeLock(value interface{}) {
	atomic.StorePointer(&e.p, unsafe.Pointer(&value))
}

func (e *entry) expungedLock() bool {
	return atomic.CompareAndSwapPointer(&e.p, nil, expunged)
}

func (e *entry) delete() (value interface{}, ok bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == expunged || p == nil {
			return nil, false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return *(*interface{})(p), true
		}
	}
}
