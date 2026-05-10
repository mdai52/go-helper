package strutil

import (
	"strconv"
	"strings"
)

// ToInt 将字符串转换为 int，失败返回 0
func ToInt(str string) int {
	v, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return v
}

// ToUint 将字符串转换为 uint，失败返回 0
func ToUint(str string) uint {
	v, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return uint(v)
}

// FirstUpper 将字符串首字母大写
func FirstUpper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FirstLower 将字符串首字母小写
func FirstLower(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}