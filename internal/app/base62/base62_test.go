package base62

import "testing"

func BenchmarkBase62Encode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Base62Encode(20)
	}
}

func TestBase62Encode(t *testing.T) {
	tests := []struct {
		number   uint64
		expected string
	}{
		{0, ""},
		{1, "b"},
		{10, "k"},
		{61, "9"},
		{62, "ab"},
		{123, "9b"},
		{3844, "aab"},
	}

	for _, test := range tests {
		encoded := Base62Encode(test.number)
		if encoded != test.expected {
			t.Errorf("Base62Encode(%d): expected %s, got %s", test.number, test.expected, encoded)
		}
	}
}
