package formatter

import "gopkg.in/yaml.v2"

type YAML struct {
	data any
}

func NewYAML(data any) *YAML {
	return &YAML{data: data}
}

func (f *YAML) String() string {
	s, _ := yaml.Marshal(f.data)
	return string(s)
}
