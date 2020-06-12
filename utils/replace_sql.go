package utils

import (
	"strconv"
	"strings"
)

func ReplaceSQL(old, pattern string, startsWith uint64) string {
	count := strings.Count(old, pattern)
	for i := 1; i <= count; i++ {
		old = strings.Replace(old, pattern, "$"+strconv.Itoa(i+int(startsWith)-1), 1)
	}
	return old
}
