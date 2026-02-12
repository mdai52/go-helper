package strutil

import (
	"math/rand"
	"strconv"
	"strings"
)

// 随机字符串

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
