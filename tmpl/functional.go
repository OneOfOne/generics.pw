package P

type EmitFunc func(key T) U
type ReduceFunc func(initial U, value T) U
type FilterFunc func(value T) bool

func Map(f EmitFunc, in []T) (res []U) {
	res = make([]U, len(in))

	for _, it := range in {
		res = append(res, f(it))
	}

	return
}

func Reduce(f ReduceFunc, initial U, in []T) U {
	for _, it := range in {
		initial = f(initial, it)
	}
	return initial
}

func Filter(f FilterFunc, in []T) (res []T) {
	for _, it := range in {
		if f(it) {
			res = append(res, it)
		}
	}
	return res[:len(res):len(res)]
}

func MakeMap(ina []T, inb []U) (m map[T]U) {
	if len(ina) != len(inb) {
		return
	}
	m = make(map[T]U, len(ina))
	for i := range ina {
		m[ina[i]] = inb[i]
	}
	return
}
