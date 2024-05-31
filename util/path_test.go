package util

import (
	"os"
	"runtime"
	"testing"
)

func TestEnsurePathSeparator(t *testing.T) {
	if runtime.GOOS == "windows" {
		expected := "C:\\Users\\go"
		actual := EnsurePathSeparator("C:\\Users\\go")
		if expected != actual {
			t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
		}
		return
	}
	expected := "C:/Users/go"
	actual := EnsurePathSeparator("C:\\Users\\go")
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

func TestExpandHomePath(t *testing.T) {
	expected, _ := os.UserHomeDir()
	expected += string(os.PathSeparator) + "go"
	actual, err := ExpandHome("~/\\go")
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}
