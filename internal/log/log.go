package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	output.FormatLevel = func(i interface{}) string {
		var level = i.(string)
		switch level {
		case "info":
			return "\033[34m" + strings.ToUpper(fmt.Sprintf("%-6s|", level)) + "\033[0m" // Blue
		case "warn":
			return "\033[33m" + strings.ToUpper(fmt.Sprintf("%-6s|", level)) + "\033[0m" // Yellow
		case "error":
			return "\033[31m" + strings.ToUpper(fmt.Sprintf("%-6s|", level)) + "\033[0m" // Red
		case "fatal":
			return "\033[41;37m" + strings.ToUpper(fmt.Sprintf("%-6s|", level)) + "\033[0m" // White on Red
		case "debug":
			return "\033[32m" + strings.ToUpper(fmt.Sprintf("%-6s|", level)) + "\033[0m" // Blue
		default:
			return strings.ToUpper(fmt.Sprintf("%-6s|", level))
		}
	}
	output.FormatTimestamp = func(i interface{}) string {
		return ""
	}

	logger = zerolog.New(output).With().Logger()
}

func Info() *zerolog.Event {
	return logger.Info()
}

func Error() *zerolog.Event {
	return logger.Error()
}

func Fatal() *zerolog.Event {
	return logger.Fatal()
}

func Debug() *zerolog.Event {
	return logger.Debug()
}

func Warn() *zerolog.Event {
	return logger.Warn()
}
