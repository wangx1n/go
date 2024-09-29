package go_map

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var expunged = unsafe.Pointer(new(interface{}))

type Map struct {
	mu      sync.Mutex
	read    atomic.Value
	dirty   map[interface{}]*entry
	missCnt int
}

type readOnly struct {
	m       map[interface{}]*entry
	amended bool
}

func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		defer m.mu.Unlock()

		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			m.missCntLock()
		}
	}

	if !ok {
		return nil, false
	}
	return e.load()
}

func (m *Map) Store(key, value interface{}) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if ok && e.tryStore(value) {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	read, _ = m.read.Load().(readOnly)
	if e, ok = read.m[key]; ok {
		if e.unExpungedLock() {
			m.dirty[key] = e
		}
		e.storeLock(value)
	} else if e, ok = m.dirty[key]; ok {
		e.storeLock(value)
	} else {
		if !read.amended {
			m.buildDirty()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
	}
}

func (m *Map) Delete(key interface{}) (value interface{}, ok bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		defer m.mu.Unlock()

		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			delete(m.dirty, key)
			m.missCntLock()
			return value, ok
		}
	}
	if ok {
		return e.delete()
	}
	return nil, false
}
