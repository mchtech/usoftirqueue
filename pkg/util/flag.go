package util

import (
	"strings"
)

type ArrayFlags []string

// impl flag.Value
func (i *ArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// impl flag.Value
func (i *ArrayFlags) String() string {
	return strings.Join(*i, ",")
}
