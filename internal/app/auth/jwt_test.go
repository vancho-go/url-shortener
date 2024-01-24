package auth

import "testing"

func Benchmark_generateJWTToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateJWTToken("test string bla bla bla")
	}
}
