package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
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

	log.Logger = zerolog.New(output).With().Logger()

}
