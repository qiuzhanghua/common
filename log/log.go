package log

import (
	"github.com/labstack/gommon/log"
	"strings"
)

func SetGlobalLevel(level string) {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		log.SetLevel(log.DEBUG) // 1
		break
	case "info":
		log.SetLevel(log.INFO) // 2
		break
	case "warn":
		log.SetLevel(log.WARN) // 3
		break
	case "error":
		log.SetLevel(log.ERROR) // 4
		break
	case "off":
		log.SetLevel(log.OFF) // 5
		break
	case "panic":
		log.SetLevel(6)
		break
	case "fatal":
		log.SetLevel(7)
		break
	default:
		log.SetLevel(log.INFO)
	}
}
