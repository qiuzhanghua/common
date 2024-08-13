package util

import (
	"os"
	"testing"
)

func TestReplaceString(t *testing.T) {
	content := "Hello, ${JAVA_HOME}"
	_ = os.Setenv("JAVA_HOME", "/usr/bin")
	actual := ReplaceString(content)
	expected := "Hello, /usr/bin"
	if actual != expected {
		t.Errorf("Expected : %v, actual is '%v'", expected, actual)
	}
}

func TestReplaceString2(t *testing.T) {
	content := "Hello, %JAVA_HOME%"
	_ = os.Setenv("JAVA_HOME", "/usr/bin")
	actual := ReplaceString(content)
	expected := "Hello, /usr/bin"
	if actual != expected {
		t.Errorf("Expected : %v, actual is '%v'", expected, actual)
	}
}

func TestReplaceString3(t *testing.T) {
	content := "setx /m JAVA_HOME ${TDP_HOME}/${TDP_LIB}/JAVA/${TDP_CURRENT}"
	_ = os.Setenv("JAVA_HOME", "/usr/bin")
	_ = os.Setenv("TDP_HOME", "tdp")
	_ = os.Setenv("TDP_LIB", "lib")
	_ = os.Setenv("TDP_CURRENT", "current")
	actual := ReplaceString(content)
	expected := "setx /m JAVA_HOME tdp/lib/JAVA/current"
	if actual != expected {
		t.Errorf("Expected : %v, actual is '%v'", expected, actual)
	}
}

func TestReplaceWithMode(t *testing.T) {
	content := "setx ${TDP_HOME}"
	_ = os.Setenv("TDP_HOME", "C:\\Users\\q\\tdp")
	actual := ReplaceStringWithMode(content, Slash)
	expected := "setx C:/Users/q/tdp"
	if actual != expected {
		t.Errorf("Expected : %v, actual is '%v'", expected, actual)
	}

	_ = os.Setenv("TDP_HOME", "C:/Users/q/tdp")
	actual = ReplaceStringWithMode(content, BackSlash)
	expected = "setx C:\\Users\\q\\tdp"
	if actual != expected {
		t.Errorf("Expected : %v, actual is '%v'", expected, actual)
	}

	_ = os.Setenv("TDP_HOME", "C:\\Users\\q\\tdp")
	actual = ReplaceStringWithMode(content, DoubleBackSlash)
	expected = "setx C:\\\\Users\\\\q\\\\tdp"
	if actual != expected {
		t.Errorf("Expected : %v, actual is '%v'", expected, actual)
	}
}
