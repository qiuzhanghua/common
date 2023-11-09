package util

import (
	"github.com/labstack/gommon/log"
	"strings"
)

func LevelOf(level string) log.Lvl {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return log.DEBUG // 1
	case "info":
		return log.INFO // 2
	case "warn":
		return log.WARN // 3
	case "error":
		return log.ERROR // 4
	case "off":
		return log.OFF // 5
	case "panic":
		return 6
	case "fatal":
		return 7
	default:
		return log.INFO
	}
}
