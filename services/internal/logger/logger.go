package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

func Err(err error) slog.Attr {
	return slog.String("error", err.Error())
}

func BuildLogger(debug bool) {
	var RESET = "\033[0m"
	var RED = "\033[31m"
	var ORANGE = "\033[33m"
	var GREEN = "\033[32m"
	var PURPLE = "\033[35m"

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	config := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				a.Value = slog.StringValue(source.Function + ":" + strconv.Itoa(source.Line))
				a.Key = "src"
			} else if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				if level == slog.LevelDebug {
					fmt.Print(PURPLE + "DEBUG" + RESET + " ")
				} else if level == slog.LevelInfo {
					fmt.Print(GREEN + "INFO" + RESET + " ")
				} else if level == slog.LevelWarn {
					fmt.Print(ORANGE + "WARN" + RESET + " ")
				} else if level == slog.LevelError {
					fmt.Print(RED + "ERROR" + RESET + " ")
				}
				return slog.Attr{}
			}
			return a
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, config))
	slog.SetDefault(logger)
}
