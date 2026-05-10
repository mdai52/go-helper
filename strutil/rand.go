package strutil

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Rand 生成随机字符串
func Rand(length uint) string {
	rs := make([]string, length)

	for range length {
		t := rand.Intn(3)
		switch t {
		case 0:
			rs = append(rs, strconv.Itoa(rand.Intn(10)))
		case 1:
			rs = append(rs, string(rune(rand.Intn(26)+65)))
		default:
			rs = append(rs, string(rune(rand.Intn(26)+97)))
		}
	}

	return strings.Join(rs, "")
}

// NewString 生成基于 UUID v7 的唯一字符串
func NewString() string {
	tid, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return tid.String()
}
