package md

import (
	"fmt"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/qiuzhanghua/common/util"
)

func get_dir_with_env(env string, default_dir string) (string, error) {
	result, ok := os.LookupEnv(env)
	if ok {
		fileInfo, err := os.Stat(result)
		if err == nil {
			if fileInfo.IsDir() {
				return result, nil
			} else {
				log.Errorf("%s is not a directory", fileInfo.Name())
				return "", fmt.Errorf("%s is not a directory", fileInfo.Name())
			}
		} else {
			return "", err
		}
	}

	result, err := util.ExpandHome(default_dir)
	if err != nil {
		return "", err
	}
	fileInfo, err := os.Stat(result)
	if err == nil {
		if fileInfo.IsDir() {
			return result, nil
		} else {
			return "", fmt.Errorf("%s is not a directory", default_dir)

		}
	} else {
		return "", err
	}
}

func MdHome() (string, error) {
	home, err := get_dir_with_env("MODELSCOPE_CACHE", "~/.cache/modelscope/hub")
	if err != nil {
		return "", err
	}

	info, err := os.Stat(home)
	if err != nil {
		return "", fmt.Errorf("cannot access MODELSCOPE_CACHE directory: %s", home)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", home)
	}

	return home, nil
}
