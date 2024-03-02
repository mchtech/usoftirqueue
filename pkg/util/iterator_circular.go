package util

type CircularIterator[T any] struct {
	fn func() T
}

func NewCircularIterator[T any](data []T) *CircularIterator[T] {
	var i int
	return &CircularIterator[T]{
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

func (x *CircularIterator[T]) NextN(num int) (n []T) {
	for i := 0; i < num; i++ {
		n = append(n, x.fn())
	}
	return
}
