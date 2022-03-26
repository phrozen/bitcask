package bitcask

import (
	"math/rand"
	"testing"
)

func BenchmarkPut(b *testing.B) {
	cases := []struct {
		name string
		size int64
	}{
		{"128B", 128},
		{"256B", 256},
		{"1K", 1024},
		{"2K", 2048},
		{"4K", 4096},
		{"8K", 8192},
		{"16K", 16384},
		{"32K", 32768},
	}

	bc, err := Open("data", Default)
	if err != nil {
		b.Fatal(err)
	}
	defer bc.Close()

	for _, tt := range cases {
		b.Run(tt.name, func(b *testing.B) {
			b.SetBytes(tt.size + 32 + 16)
			key := make([]byte, 32)
			_, err = rand.Read(key)
			if err != nil {
				b.Fatal(err)
			}
			value := make([]byte, tt.size)
			_, err = rand.Read(value)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := bc.Put(key, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
