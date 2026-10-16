// Package env đọc biến môi trường, trả về giá trị mặc định khi biến chưa được đặt
// hoặc không đúng kiểu.
package env

import (
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return valInt
}
