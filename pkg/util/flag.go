package util

import (
	"strings"
)

type StringsFlags []string

// impl flag.Value
func (i *StringsFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// impl flag.Value
func (i *StringsFlags) String() string {
	return strings.Join(*i, ",")
}
