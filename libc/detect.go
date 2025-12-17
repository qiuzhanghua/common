package libc

import (
	"os"
	"sync"
)

var (
	once   sync.Once
	result string
)

// Detect returns "gnu", "musl", or "unknown"
func Detect() string {
	once.Do(func() {
		result = detect()
	})
	return result
}

func detect() string {

	if libc := checkFiles(); libc != "unknown" {
		return libc
	}

	// fallback for Alpine
	if _, err := os.Stat("/etc/alpine-release"); err == nil {
		return "musl"
	}

	return "unknown"
}

func checkFiles() string {
	glibcFiles := []string{
		"/lib/ld-linux-x86-64.so.2",
		"/lib64/ld-linux-x86-64.so.2",
		"/lib/libc.so.6",
		"/lib64/libc.so.6",
	}

	for _, file := range glibcFiles {
		if _, err := os.Stat(file); err == nil {
			return "gnu"
		}
	}

	muslFiles := []string{
		"/lib/ld-musl-x86_64.so.1",
		"/lib/ld-musl-aarch64.so.1",
		"/lib/ld-musl-armhf.so.1",
		"/lib/libc.musl-x86_64.so.1",
		"/usr/lib/ld-musl-x86_64.so.1",
	}

	for _, file := range muslFiles {
		if _, err := os.Stat(file); err == nil {
			return "musl"
		}
	}

	return "unknown"
}
