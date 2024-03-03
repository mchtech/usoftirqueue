package util

type ScatterRoundRobinIterator[T any] struct {
	data     []T
	length   int
	offset   int
	cursor   int
	interval int
	fn       func(num int) (r []T)
}

func NewScatterRoundRobinIterator[T any](data []T) (b *ScatterRoundRobinIterator[T]) {
	return &ScatterRoundRobinIterator[T]{
		data:     data,
		length:   len(data),
		offset:   -1,
		cursor:   0,
		interval: 1,
		fn: func(count int) (r []T) {
			for i := 0; i < count; i++ {
				j := b.offset + b.cursor*b.interval
				if j >= b.length {
					j = (j - b.length) % b.length
				}
				r = append(r, b.data[j])
				b.cursor++
			}
			return
		},
	}
}

func (b *ScatterRoundRobinIterator[T]) NewInterval(interval int) *ScatterRoundRobinIterator[T] {
	b.interval = interval
	if b.interval <= 0 {
		b.interval = 1
	}
	b.offset++
	if b.offset == b.length {
		b.offset = 0
	}
	b.cursor = 0
	return b
}

func (b *ScatterRoundRobinIterator[T]) NextN(num int) (r []T) {
	if b.offset == -1 {
		b.offset = 0
	}
	return b.fn(num)
}
