package go_map

func (m *Map) missCntLock() {
	m.missCnt++
	if m.missCnt < len(m.dirty) {
		return
	}
	m.read.Store(readOnly{
		m: m.dirty,
	})
	m.dirty = nil
	m.missCnt = 0
}

func (m *Map) buildDirty() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnly)
	m.dirty = make(map[interface{}]*entry, len(read.m))
	for k, v := range read.m {
		if !v.expungedLock() {
			m.dirty[k] = v
		}
	}
}
