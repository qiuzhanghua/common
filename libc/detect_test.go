package libc

import (
	"fmt"
	"testing"
)

func TestDetect(t *testing.T) {
	libc := Detect()
	fmt.Printf("Detected libc: %s\n", libc)
	if libc != "gnu" && libc != "musl" && libc != "unknown" {
		t.Errorf("Unexpected libc detected: %s", libc)
	}
}
