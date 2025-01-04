package hf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func XdgCacheHome() (string, error) {
	return get_dir_with_env("XDG_CACHE_HOME", "~/.cache")
}

func HfHome() (string, error) {
	return get_dir_with_env("HF_HOME", "~/.cache/huggingface")
}

func HuggingfaceHubCache() (string, error) {
	cache := ""
	result, ok := os.LookupEnv("HUGGINGFACE_HUB_CACHE")
	if ok {
		result, err := get_dir_with_env("HUGGINGFACE_HUB_CACHE", "~/.cache/huggingface/hub")
		if err != nil {
			return "", err
		}
		cache = result
	} else {
		hfHome, err := HfHome()
		if err != nil {
			return "", err
		} else {
			cache = filepath.Join(hfHome, "hub")
			cache = strings.ReplaceAll(cache, "/", string(os.PathSeparator))
		}
	}
	log.Debugf("HuggingfaceHubCache: %s", cache)
	fileInfo, err := os.Stat(cache)
	if err == nil {
		if fileInfo.IsDir() {
			return cache, nil
		} else {
			return "", fmt.Errorf("%s is not a directory", result)
		}
	} else {
		return "", err
	}
}

func HfDatasetsCache() (string, error) {
	result, err := get_dir_with_env("HF_DATASETS_CACHE", "~/.cache/huggingface/datasets")
	if err != nil {
		result, err = HfHome()
		if err != nil {
			return "", err
		} else {
			result = result + "/datasets"
			fileInfo, err := os.Stat(result)
			if err == nil {
				if fileInfo.IsDir() {
					return result, nil
				} else {
					return "", fmt.Errorf("%s is not a directory", result)
				}
			} else {
				return "", err
			}
		}
	}
	return result, nil
}

func HfModelPath(model string) (string, error) {
	cache, err := HuggingfaceHubCache()
	if err != nil {
		return "", err
	}
	modelPath := "models--" + strings.ReplaceAll(model, "/", "--")
	modelDir := filepath.Join(cache, modelPath)
	oid, err := readOidOf(modelDir)
	if err != nil {
		return "", err
	}
	result := filepath.Join(modelDir, "snapshots", oid)
	fileInfo, err := os.Stat(result)
	if err == nil {
		if fileInfo.IsDir() {
			return result, nil
		} else {
			return "", fmt.Errorf("%s is not a directory", result)
		}
	} else {
		return "", err
	}

}

func readOidOf(modelOrDs string) (string, error) {
	filePath := filepath.Join(modelOrDs, "refs", "main")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	oid := strings.TrimSpace(string(data))
	return oid, nil
}

func HfDatasetsPath(model string) (string, error) {
	cache, err := HuggingfaceHubCache()
	if err != nil {
		return "", err
	}
	dsPath := "datasets--" + strings.ReplaceAll(model, "/", "--")
	dsDir := filepath.Join(cache, dsPath)
	oid, err := readOidOf(dsDir)
	if err != nil {
		return "", err
	}
	result := filepath.Join(dsDir, "snapshots", oid)
	fileInfo, err := os.Stat(result)
	if err == nil {
		if fileInfo.IsDir() {
			return result, nil
		} else {
			return "", fmt.Errorf("%s is not a directory", result)
		}
	} else {
		return "", err
	}

}
