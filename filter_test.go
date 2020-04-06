package cuckoo

import (
	"bufio"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"testing/quick"
)

func samples(t testing.TB) [][32]byte {
	t.Helper()

	f, err := os.Open("/usr/share/dict/words")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	scanner := bufio.NewScanner(f)

	var samples [][32]byte

	for scanner.Scan() {
		samples = append(samples, sha256.Sum256([]byte(scanner.Text())))
	}

	return samples
}

/**
go test -bench=. -benchtime=10s -benchmem

goos: linux
goarch: amd64
pkg: github.com/lithdew/cuckoo
BenchmarkNewFilter-8                       39187            290737 ns/op         2105346 B/op          1 allocs/op
BenchmarkInsert-8                       239464290               51.8 ns/op             0 B/op          0 allocs/op
BenchmarkLookup-8                       238444526               49.7 ns/op             0 B/op          0 allocs/op
BenchmarkMarshalBinary-8                   24799            468098 ns/op         2097161 B/op          1 allocs/op
BenchmarkUnsafeUnmarshalBinary-8            6715           2097175 ns/op         2105777 B/op          5 allocs/op
BenchmarkUnmarshalBinary-8                  5738           2042495 ns/op         2105777 B/op          5 allocs/op
*/

func BenchmarkNewFilter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFilter()
	}
}

func BenchmarkInsert(b *testing.B) {
	filter := NewFilter()

	samples := samples(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Insert(samples[i%len(samples)])
	}
}

func BenchmarkLookup(b *testing.B) {
	filter := NewFilter()

	samples := samples(b)

	for _, sample := range samples {
		filter.Insert(sample)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Lookup(samples[i%len(samples)])
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	filter := NewFilter()

	samples := samples(b)

	for _, sample := range samples {
		filter.Insert(sample)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.MarshalBinary()
	}
}

func BenchmarkUnsafeUnmarshalBinary(b *testing.B) {
	filter := NewFilter()

	samples := samples(b)

	for _, sample := range samples {
		filter.Insert(sample)
	}

	data := filter.MarshalBinary()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := UnsafeUnmarshalBinary(data)
		require.NoError(b, err)
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	filter := NewFilter()

	samples := samples(b)

	for _, sample := range samples {
		filter.Insert(sample)
	}

	data := filter.MarshalBinary()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := UnmarshalBinary(data)
		require.NoError(b, err)
	}
}

func TestFilter(t *testing.T) {
	filter := NewFilter()

	samples := samples(t)

	if len(samples) > 1000 {
		samples = samples[:1000]
	}

	for _, sample := range samples {
		require.True(t, filter.Insert(sample))
	}

	require.EqualValues(t, filter.Count, len(samples))

	for _, sample := range samples {
		require.False(t, filter.Insert(sample))
	}

	require.EqualValues(t, filter.Count, len(samples))

	for _, sample := range samples {
		require.True(t, filter.Lookup(sample))
	}

	for _, sample := range samples {
		require.True(t, filter.Delete(sample))
	}

	require.EqualValues(t, filter.Count, 0)

	for _, sample := range samples {
		require.False(t, filter.Delete(sample))
	}

	for _, sample := range samples {
		require.False(t, filter.Lookup(sample))
	}

	empty := NewFilter()
	filter.Reset()

	require.Equal(t, filter, empty)
}

func TestEncoding(t *testing.T) {
	f := func(entries [][]byte) bool {
		a := NewFilter()

		for _, entry := range entries {
			if len(entry) > 100 {
				entry = entry[:100]
			}

			a.Insert(sha256.Sum256(entry))
		}

		b, err := UnmarshalBinary(a.MarshalBinary())
		if !assert.NoError(t, err) {
			return false
		}

		if !assert.Equal(t, a, b) {
			return false
		}

		return true
	}

	require.NoError(t, quick.Check(f, &quick.Config{MaxCount: 10}))
}
