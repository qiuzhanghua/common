package clib

import (
	"fmt"
	"testing"
)

func TestDetect(t *testing.T) {
	libc := Detect()
	fmt.Printf("Detected libc: %s\n", libc)
	if libc != "glibc" && libc != "musl" && libc != "unknown" {
		t.Errorf("Unexpected libc detected: %s", libc)
	}
}
