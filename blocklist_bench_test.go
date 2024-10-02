package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeSyntheticBlocklist(tb testing.TB, n int) string {
	tb.Helper()
	p := filepath.Join(tb.TempDir(), "hosts")
	f, err := os.Create(p)
	if err != nil {
		tb.Fatal(err)
	}
	defer f.Close()
	for i := 0; i < n; i++ {
		fmt.Fprintf(f, "0.0.0.0 bench-%d.example.com\n", i)
	}
	return p
}

func BenchmarkContainsHit(b *testing.B) {
	bl, err := LoadBlocklist(writeSyntheticBlocklist(b, 100_000))
	if err != nil {
		b.Fatal(err)
	}
	target := "bench-54321.example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !bl.Contains(target) {
			b.Fatal("expected hit")
		}
	}
}

func BenchmarkContainsMiss(b *testing.B) {
	bl, err := LoadBlocklist(writeSyntheticBlocklist(b, 100_000))
	if err != nil {
		b.Fatal(err)
	}
	target := "not-in-list.example.net"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if bl.Contains(target) {
			b.Fatal("expected miss")
		}
	}
}
