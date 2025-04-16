package mc

import (
	"fmt"
	"os"
	"path/filepath"

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

func McHome() (string, error) {
	if cache, exists := os.LookupEnv("MODELSCOPE_CACHE"); exists {
		cachePath, err := util.ExpandHome(cache)
		if err != nil {
			return "", fmt.Errorf("invalid MODELSCOPE_CACHE path: %s", cache)
		}

		// Get parent directory
		parent := filepath.Dir(cachePath)
		info, err := os.Stat(parent)
		if err != nil {
			return "", fmt.Errorf("cannot access parent directory of MODELSCOPE_CACHE: %s", parent)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("%s is not a directory", parent)
		}

		return parent, nil
	} else {
		cache, err := get_dir_with_env("MC_HOME", "~/.cache/modelscope")
		if err != nil {
			return "", err
		}

		info, err := os.Stat(cache)
		if err != nil {
			return "", fmt.Errorf("cannot access MC_HOME directory: %s", cache)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("%s is not a directory", cache)
		}

		return cache, nil
	}
}

// modelscopeHubCache retrieves the MODELSCOPE_CACHE directory.
// It ensures that the path exists and is a directory.
func ModelscopeHubCache() (string, error) {
	cache, err := get_dir_with_env("MODELSCOPE_CACHE", "~/.cache/modelscope/hub")
	if err != nil {
		return "", err
	}

	info, err := os.Stat(cache)
	if err != nil {
		return "", fmt.Errorf("cannot access MODELSCOPE_CACHE directory: %s", cache)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", cache)
	}

	return cache, nil
}
