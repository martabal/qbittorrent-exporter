package internal

import (
	"testing"
)

// BenchmarkCompareSemVer benchmarks semantic version comparison
func BenchmarkCompareSemVer(b *testing.B) {
	v1 := "2.11.5"
	v2 := "2.11.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CompareSemVer(v1, v2)
	}
}

// BenchmarkIsValidURL benchmarks URL validation
func BenchmarkIsValidURL(b *testing.B) {
	url := "http://tracker.example.com:8080/announce"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidURL(url)
	}
}

// BenchmarkIsValidHttpsURL benchmarks HTTPS URL validation
func BenchmarkIsValidHttpsURL(b *testing.B) {
	url := "https://example.com:8080/path"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidHttpsURL(url)
	}
}
