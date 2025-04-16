package mc

import (
	"testing"

	util "github.com/qiuzhanghua/common/util"
)

func TestMcHome(t *testing.T) {
	expected, _ := util.ExpandHome("~/tdp/cache/modelscope")
	actual, _ := McHome()
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}
