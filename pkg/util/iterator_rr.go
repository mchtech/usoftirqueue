package util

type RoundRobinIterator[T any] struct {
	fn func() T
}

func NewRoundRobinIterator[T any](data []T) *RoundRobinIterator[T] {
	var i int
	return &RoundRobinIterator[T]{
		fn: func() T {
			v := data[i]
			i++
			if i == len(data) {
				i = 0
			}
			return v
		},
	}
}

func (x *RoundRobinIterator[T]) NextN(num int) (n []T) {
	for i := 0; i < num; i++ {
		n = append(n, x.fn())
	}
	return
}
