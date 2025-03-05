package util

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unsafe"
)

func SubStringLast(str string, substr string) string {
	index := strings.Index(str, substr)
	if index == -1 {
		return ""
	}
	return str[index+len(substr):]
}
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
func StringToByte(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
func ByteToString(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}
func JoinStrings(str ...any) string {
	var sb strings.Builder
	for _, v := range str {
		sb.WriteString(check(v))
	}
	return sb.String()
}

func check(v any) string {
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.String:
		return v.(string)
	default:
		return fmt.Sprintf("%v", v)
	}
}
