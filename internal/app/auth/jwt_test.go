package auth

import (
	"math/rand"
	"testing"
)

// функция для генерации случайной строки заданной длины
func randomStr(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = letters[rand.Intn(len(letters))]
	}
	return string(bytes)
}

func Benchmark_generateJWTToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		input := randomStr(20)
		b.StartTimer()
		generateJWTToken(input)
	}
}
