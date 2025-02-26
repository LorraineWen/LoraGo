package util

import "strings"

func SubStringLast(str string, substr string) string {
	index := strings.Index(str, substr)
	if index == -1 {
		return ""
	}
	return str[index+len(substr):]
}
