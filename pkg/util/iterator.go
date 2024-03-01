package util

type BiasIterator[T any] struct {
	data []T
	bias int
	idx  int
}

type g[T any] struct {
	fn func(count int) (r []T)
}

func (g *g[T]) Generate(count int) (r []T) {
	return g.fn(count)
}

func NewBiasIterator[T any](data []T) *BiasIterator[T] {
	return &BiasIterator[T]{
		data: data,
		bias: -1,
		idx:  0,
	}
}

func (b *BiasIterator[T]) Interval(intv int) *g[T] {
	if intv <= 0 {
		intv = 1
	}
	b.idx = 0
	b.bias++
	if b.bias == len(b.data) {
		b.bias = 0
	}
	return &g[T]{
		fn: func(count int) (r []T) {
			for i := 0; i < count; i++ {
				j := b.bias + b.idx*intv
				if over := j - len(b.data) + 1; over > 0 {
					j = (over - 1) % len(b.data)
				}
				b.idx++
				r = append(r, b.data[j])
			}
			return
		},
	}
}
