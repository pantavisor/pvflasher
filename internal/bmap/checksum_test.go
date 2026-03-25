package bmap

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"testing"
)

func TestGetHasher_SHA1(t *testing.T) {
	h, err := GetHasher("sha1")
	if err != nil {
		t.Fatalf("GetHasher error: %v", err)
	}

	// Check we got the right type
	var _ hash.Hash = h

	// Test that it produces correct output
	h.Write([]byte("test"))
	sum := fmt.Sprintf("%x", h.Sum(nil))

	// SHA1 of "test" is known
	expected := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	if sum != expected {
		t.Errorf("SHA1 sum = %q, want %q", sum, expected)
	}
}

func TestGetHasher_SHA256(t *testing.T) {
	h, err := GetHasher("sha256")
	if err != nil {
		t.Fatalf("GetHasher error: %v", err)
	}

	var _ hash.Hash = h

	h.Write([]byte("test"))
	sum := fmt.Sprintf("%x", h.Sum(nil))

	// SHA256 of "test" is known
	expected := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if sum != expected {
		t.Errorf("SHA256 sum = %q, want %q", sum, expected)
	}
}

func TestGetHasher_SHA512(t *testing.T) {
	h, err := GetHasher("sha512")
	if err != nil {
		t.Fatalf("GetHasher error: %v", err)
	}

	var _ hash.Hash = h

	h.Write([]byte("test"))
	sum := fmt.Sprintf("%x", h.Sum(nil))

	// SHA512 of "test" is known
	expected := "ee26b0dd4af7e749aa1a8ee3c10ae9923f618980772e473f8819a5d4940e0db27ac185f8a0e1d5f84f88bc887fd67b143732c304cc5fa9ad8e6f57f50028a8ff"
	if sum != expected {
		t.Errorf("SHA512 sum = %q, want %q", sum, expected)
	}
}

func TestGetHasher_Unsupported(t *testing.T) {
	tests := []string{
		"md5",
		"crc32",
		"adler32",
		"invalid",
		"",
	}

	for _, algo := range tests {
		t.Run(algo, func(t *testing.T) {
			_, err := GetHasher(algo)
			if err == nil {
				t.Errorf("GetHasher(%q) should return error", algo)
			}
		})
	}
}

func TestGetHasher_Reset(t *testing.T) {
	h, _ := GetHasher("sha256")

	// Write first data
	h.Write([]byte("first"))
	firstSum := h.Sum(nil)

	// Reset and write same data
	h.Reset()
	h.Write([]byte("first"))
	secondSum := h.Sum(nil)

	if string(firstSum) != string(secondSum) {
		t.Error("Reset did not clear hash state properly")
	}
}

func TestGetHasher_MultipleWrites(t *testing.T) {
	h, _ := GetHasher("sha256")

	// Write in chunks
	h.Write([]byte("Hello, "))
	h.Write([]byte("World!"))
	chunkedSum := fmt.Sprintf("%x", h.Sum(nil))

	// Write all at once
	h.Reset()
	h.Write([]byte("Hello, World!"))
	fullSum := fmt.Sprintf("%x", h.Sum(nil))

	if chunkedSum != fullSum {
		t.Error("Multiple writes produced different hash than single write")
	}
}

func TestGetHasher_Consistency(t *testing.T) {
	algorithms := []string{"sha1", "sha256", "sha512"}

	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			h1, _ := GetHasher(algo)
			h2, _ := GetHasher(algo)

			testData := []byte("The quick brown fox jumps over the lazy dog")

			h1.Write(testData)
			h2.Write(testData)

			sum1 := fmt.Sprintf("%x", h1.Sum(nil))
			sum2 := fmt.Sprintf("%x", h2.Sum(nil))

			if sum1 != sum2 {
				t.Errorf("Inconsistent %s hashes", algo)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetHasher_SHA256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHasher("sha256")
	}
}

func BenchmarkSHA256_Write(b *testing.B) {
	h := sha256.New()
	data := []byte("benchmark data for hashing performance testing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Write(data)
		h.Sum(nil)
		h.Reset()
	}
}
