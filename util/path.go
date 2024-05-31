package util

import (
	"errors"
	"github.com/labstack/gommon/log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func EnsurePathSeparator(p string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(p, "/", "\\")
	} else {
		return strings.ReplaceAll(p, "\\", "/")
	}
}

// ExpandHome expend dir start with ~
func ExpandHome(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if path[0] != '~' {
		return path, nil
	}

	path = strings.ReplaceAll(path, "/", string(os.PathSeparator))
	path = strings.ReplaceAll(path, "\\", string(os.PathSeparator))

	if len(path) > 1 && path[1] != os.PathSeparator {
		return "", errors.New("cannot expand home dir")
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("Error getting user home dir: %v", err)
		return "", err
	}
	return filepath.Join(dir, path[1:]), nil
}

func AbsPath(p string) (string, error) {
	absPath, err := filepath.EvalSymlinks(p)
	if err != nil {
		log.Errorf("Error evaluating symlinks: %v", err)
		return ".", err
	}
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		log.Errorf("Error getting absolute path: %v", err)
		return ".", err
	}
	return absPath, nil
}

// AppHome 获取当前执行文件的绝对路径
func AppHome() (string, error) {
	execPath, _ := os.Executable()
	absPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		log.Errorf("Error evaluating symlinks: %v", err)
		return ".", err
	}
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		log.Errorf("Error getting absolute path: %v", err)
		return ".", err
	}
	absPath, err = filepath.Abs(path.Dir(absPath))
	if err != nil {
		log.Errorf("Error getting absolute path: %v", err)
		return ".", err
	}
	return absPath, nil
}

// ExecPath 获取当前执行文件全名，含路径
func ExecPath() (string, error) {
	execPath, _ := os.Executable()
	absPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		log.Errorf("Error evaluating symlinks: %v", err)
		return ".", err
	}
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		log.Errorf("Error getting absolute path: %v", err)
		return execPath, err
	}
	return absPath, nil
}
