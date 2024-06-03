package util

import (
	"os"
	"regexp"
	"strings"
)

// ReplaceString replaces ${VAR} or %VAR% with the value of the environment variable VAR.
func ReplaceString(s string, rep ...string) string {
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
				s = strings.ReplaceAll(s, e, val)
			}
		} else {
			env, ok := os.LookupEnv(e2)
			if !ok {
				continue
			}
			// not needed
			//if runtime.GOOS == "windows" {
			//	env = strings.ReplaceAll(env, "\\", "\\\\")
			//}
			s = strings.ReplaceAll(s, e, env)
		}
	}
	return s
}
