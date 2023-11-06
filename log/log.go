package log

import (
	"github.com/rs/zerolog"
	"strings"
)

func SetGlobalLevel(level string) {
	level = strings.ToLower(level)
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		break
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		break
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		break
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		break
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
		break
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
		break
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
