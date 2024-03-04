package policy

// var (
// 	Policies []Policy
// )

type Policy interface {
	Name() string
	Apply() error
}
