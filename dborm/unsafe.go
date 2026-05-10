package dborm

import (
	"errors"
	"regexp"
	"strings"
)

// OrderSafe 校验排序字段是否安全
func OrderSafe(data string) error {
	var expr = regexp.MustCompile(`^(\w+)( DESC)?$`)

	for _, v := range strings.Split(data, ",") {
		if !expr.MatchString(v) {
			return errors.New("unsafe sorting field")
		}
	}

	return nil
}
