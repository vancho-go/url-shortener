// Модуль base62 позволяет сгенерировать строку, которая
// будет использоваться в качестве shortenURL.
package base62

import "strings"

// Разрешенный алфавит для генерирования строки.
const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Base62Encode генерирует строку, на основе полученного на вход числа.
func Base62Encode(number uint64) string {
	length := len(alphabet)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(alphabet[(number % uint64(length))])
	}

	return encodedBuilder.String()
}
