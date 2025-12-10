package tui

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var reColor *regexp.Regexp = regexp.MustCompile(`(<[^>]+>)`)

var colors = map[string][]int{
	"black":    {59, 66, 82},
	"red":      {191, 97, 106},
	"green":    {163, 190, 140},
	"yellow":   {235, 203, 139},
	"blue":     {129, 161, 193},
	"magenta":  {180, 142, 173},
	"cyan":     {136, 192, 208},
	"white":    {229, 233, 240},
	"bblack":   {76, 86, 106},
	"bred":     {208, 111, 121},
	"bgreen":   {180, 215, 165},
	"byellow":  {242, 216, 167},
	"bblue":    {143, 173, 217},
	"bmagenta": {199, 168, 199},
	"bcyan":    {154, 213, 227},
	"bwhite":   {255, 255, 255},
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
					return fmt.Sprintf("\033[38;2;%d;%d;%dm", value[0], value[1], value[2])
				} else if lowerKey == "/" || lowerKey == "reset" {
					return "\033[0m"
				} else if lowerKey == "bold" {
					return "\033[1m"
				} else if lowerKey == "italic" {
					return "\033[3m"
				} else {
					return s
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
