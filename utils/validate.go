package utils

import "regexp"

func ValidateIsEmail(s string) bool {
	// 简单邮箱正则表达式
	// 你可以根据需要使用更复杂的版本
	reg := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return reg.MatchString(s)
}
