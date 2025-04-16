package md

import (
	"os"
	"testing"

	util "github.com/qiuzhanghua/common/util"
)

func TestMdHome(t *testing.T) {
	tdpHome, ok := os.LookupEnv("TDP_HOME")
	if !ok {
		t.Errorf("TDP_HOME environment variable is not set")
		return
	}

	expected, _ := util.ExpandHome(tdpHome + "/cache/modelscope/hub")
	actual, _ := MdHome()
	if expected != actual {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}
