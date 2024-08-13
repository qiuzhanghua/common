package util

import (
	"os"
	"regexp"
	"strings"
)

type ReplaceMode int

const (
	Keep            = iota // 不变，无论是/, \还是\\
	Slash                  //将\替换为/
	BackSlash              // 将/替换为\
	DoubleBackSlash        // 将\替换为\\
)

func ConvertString(s string, mode ReplaceMode) string {
	switch mode {
	case Slash:
		return strings.ReplaceAll(s, "\\", "/")
	case BackSlash:
		return strings.ReplaceAll(s, "/", "\\")
	case DoubleBackSlash:
		return strings.ReplaceAll(s, "\\", "\\\\")
	default:
		return s
	}
}

func ReplaceString(s string, rep ...string) string {
	return ReplaceStringWithMode(s, Keep, rep...)
}

// ReplaceStringWithMode ReplaceString replaces ${VAR} or %VAR% with the value of the environment variable VAR.
func ReplaceStringWithMode(s string, mode ReplaceMode, rep ...string) string {
	var regex = regexp.MustCompile("[$][{].+?[}]|[%].+?[%]")
	m := make(map[string]string, len(rep))
	for _, r := range rep {
		arr := strings.Split(r, "=")
		if len(arr) == 2 {
			m[strings.TrimSpace(arr[0])] = strings.TrimSpace(arr[1])
		}
	}
	envs := regex.FindAllString(s, -1)
	if len(envs) == 0 && len(m) == 0 {
		return s
	}
	for _, e := range envs {
		if len(e) < 3 {
			continue
		}
		var e2 string
		if strings.HasPrefix(e, "$") {
			e2 = e[2 : len(e)-1]
		} else {
			e2 = e[1 : len(e)-1]
		}
		if len(m) > 0 {
			val, ok := m[e2]
			if ok {
				s = strings.ReplaceAll(s, e, ConvertString(val, mode))
			}
		} else {
			env, ok := os.LookupEnv(e2)
			if !ok {
				continue
			}
			s = strings.ReplaceAll(s, e, ConvertString(env, mode))
		}
	}
	return s
}
