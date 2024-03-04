package config

import (
	"os"
	"strings"
)

func Read(file string) string {
	bf, _ := os.ReadFile(file)
	return strings.TrimSpace(string(bf))
}
