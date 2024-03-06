package policy

type IRQPolicy string

type Policy interface {
	Name() string
	Apply() error
}
