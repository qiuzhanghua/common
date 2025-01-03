package hf

import (
	"testing"

	util "github.com/qiuzhanghua/common/util"
)

// TDP should be activated before running this test
func TestHfHome(t *testing.T) {
	expected, _ := util.ExpandHome("~/tdp/cache/huggingface")
	actual, _ := HfHome()
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

// TDP should be activated before running this test, and the model should be downloaded
func TestHfModelPath(t *testing.T) {
	expected, _ := util.ExpandHome("~/tdp/cache/huggingface/hub/models--intfloat--e5-mistral-7b-instruct/snapshots/07163b72af1488142a360786df853f237b1a3ca1")
	actual, _ := HfModelPath("intfloat/e5-mistral-7b-instruct")
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

// TDP should be activated before running this test
func TestHuggingfaceHubCache(t *testing.T) {
	expected, _ := util.ExpandHome("~/tdp/cache/huggingface/hub")
	actual, _ := HuggingfaceHubCache()
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

func TestXdgCacheHome(t *testing.T) {
	expected, _ := util.ExpandHome("~/.cache")
	actual, _ := XdgCacheHome()
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}
