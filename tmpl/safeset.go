package P

import "sync"

type N struct {
	m map[T]struct{}
	l sync.Mutex
}

func New() *N {
	return &N{m: make(map[T]struct{})}
}

func NewSize(n int) *N {
	return &N{m: make(map[T]struct{}, n)}
}

func (set *N) Add(key T) {
	set.l.Lock()
	set.m[key] = struct{}{}
	set.l.Unlock()
}

func (set *N) Has(key T) (ok bool) {
	set.l.Lock()
	_, ok = set.m[key]
	set.l.Unlock()
	return
}

func (set *N) Delete(key T) {
	set.l.Lock()
	delete(set.m, key)
	set.l.Unlock()
}
