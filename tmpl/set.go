package P

type N map[T]struct{}

func New() N {
	return N{}
}

func NewSize(n int) N {
	return make(N, n)
}

func (set N) Add(key T) {
	set[key] = struct{}{}
}

func (set N) Has(key T) (ok bool) {
	_, ok = set[key]
	return
}

func (set N) Delete(key T) {
	delete(set, key)
}
