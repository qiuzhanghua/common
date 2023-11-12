package util

import (
	"os"
	"testing"
)

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
