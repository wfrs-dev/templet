package tui

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var reColor *regexp.Regexp = regexp.MustCompile(`(<[^>]+>)`)

var colors = map[string]int{
	// Foreground
	"black":  30,
	"red":    31,
	"green":  32,
	"yellow": 33,
	"blue":   34,
	"purple": 35,
	"cyan":   36,
	"white":  37,
	// Background
	"bgBlack":  40,
	"bgRed":    41,
	"bgGreen":  42,
	"bgYellow": 43,
	"bgBlue":   44,
	"bgPurple": 45,
	"bgCyan":   46,
	"bgWhite":  47,
	// Bright Foreground
	"brightBlack":  90,
	"brightRed":    91,
	"brightGreen":  92,
	"brightYellow": 93,
	"brightBlue":   94,
	"brightPurple": 95,
	"brightCyan":   96,
	"brightWhite":  97,
	// Bright Background
	"bgBrightBlack":  100,
	"bgBrightRed":    101,
	"bgBrightGreen":  102,
	"bgBrightYellow": 103,
	"bgBrightBlue":   104,
	"bgBrightPurple": 105,
	"bgBrightCyan":   106,
	"bgBrightWhite":  107,
	// Properties
	"bold":          1,
	"italic":        3,
	"underline":     4,
	"strikethrough": 9,

	"reset": 0,
	"/":     0,
}

func Colorize(format string, a ...any) string {
	c := IsColor()
	txt := format
	out := reColor.ReplaceAllStringFunc(
		txt,
		func(s string) string {
			if c {
				lowerKey := strings.ToLower(strings.Trim(s, " <>"))
				if value, ok := colors[lowerKey]; ok {
					return fmt.Sprintf("\033[%dm", value)
				}
			}

			return ""
		},
	)

	return fmt.Sprintf(out, a...)
}

func IsColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")

	return !ok
}
