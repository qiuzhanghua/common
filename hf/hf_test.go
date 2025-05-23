package hf

import (
	"os"
	"testing"

	util "github.com/qiuzhanghua/common/util"
)

// TDP should be activated before running this test
func TestHfHome(t *testing.T) {
	tdpHome, ok := os.LookupEnv("TDP_HOME")
	if !ok {
		t.Errorf("TDP_HOME environment variable is not set")
		return
	}
	expected, _ := util.ExpandHome(tdpHome + "/cache/huggingface")
	actual, _ := HfHome()
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

// TDP should be activated before running this test, and the model should be downloaded
func TestHfModelPath(t *testing.T) {
	tdpHome, ok := os.LookupEnv("TDP_HOME")
	if !ok {
		t.Errorf("TDP_HOME environment variable is not set")
		return
	}
	expected, _ := util.ExpandHome(tdpHome + "/cache/huggingface/hub/models--baai--bge-small-zh/snapshots/1d2363c5de6ce9ba9c890c8e23a4c72dce540ca8")
	actual, _ := HfModelPath("baai/bge-small-zh")
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

// TDP should be activated before running this test
func TestHuggingfaceHubCache(t *testing.T) {
	tdpHome, ok := os.LookupEnv("TDP_HOME")
	if !ok {
		t.Errorf("TDP_HOME environment variable is not set")
		return
	}
	expected, _ := util.ExpandHome(tdpHome + "/cache/huggingface/hub")
	actual, _ := HuggingfaceHubCache()
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

// func TestXdgCacheHome(t *testing.T) {
// 	expected, _ := util.ExpandHome("~/.cache")
// 	actual, _ := XdgCacheHome()
// 	if expected != actual {
// 		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
// 	}
// }
