package P

import "sync"

type N struct {
	m map[T]U
	l sync.Mutex
}

func New() *N {
	return &N{m: make(map[T]U)}
}

func NewSize(n int) *N {
	return &N{m: make(map[T]U, n)}
}

func (m *N) Add(key T, val U) {
	m.l.Lock()
	m.m[key] = val
	m.l.Unlock()
}

func (m *N) Get(key T) (val U, ok bool) {
	m.l.Lock()
	val, ok = m.m[key]
	m.l.Unlock()
	return
}

func (m *N) Has(key T) (ok bool) {
	m.l.Lock()
	_, ok = m.m[key]
	m.l.Unlock()
	return
}

func (m *N) Delete(key T) {
	m.l.Lock()
	delete(m.m, key)
	m.l.Unlock()
}
