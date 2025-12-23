package tui

/*
	Based on prompter library by: https://github.com/Songmu/prompter
*/

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// Prompter is object for prompting
type Prompter struct {
	Message string
	// choices of answer
	Choices    []string
	IgnoreCase bool
	Default    string
	// specify answer pattern by regexp. When both Choices and Regexp are specified, Regexp takes a priority.
	Regexp *regexp.Regexp
	// for passwords and so on.
	NoEcho     bool
	UseDefault bool
	reg        *regexp.Regexp
}

// Prompt displays a prompt and returns answer
func (p *Prompter) Prompt() string {
	fmt.Print(Colorize(p.msg()))
	if p.UseDefault || skip() {
		return p.Default
	}
	input := ""
	if p.NoEcho {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err == nil {
			input = string(b)
		}
		fmt.Print("\n")
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		ok := scanner.Scan()
		if ok {
			input = strings.TrimRight(scanner.Text(), "\r\n")
		}
	}
	if input == "" {
		input = p.Default
	}
	if !p.inputIsValid(input) {
		fmt.Println(Colorize(p.errorMsg()))
		return p.Prompt()
	}
	return input
}

func skip() bool {
	if os.Getenv("GO_PROMPTER_USE_DEFAULT") != "" {
		return true
	}
	if isPipe() {
		return os.Getenv("GO_PROMPTER_USE_PIPE") == ""
	}
	return false
}

func isPipe() bool {
	return !(isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())) ||
		!(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))

}

func (p *Prompter) msg() string {
	msg := strings.Builder{}
	msg.WriteString("<bold><green>" + p.Message)
	if len(p.Choices) > 0 {
		msg.WriteString(fmt.Sprintf("<reset> <blue>(%s)", strings.Join(p.Choices, "/")))
	}
	if p.Default != "" {
		msg.WriteString(fmt.Sprintf(" <purple>[%s]", p.Default))
	}

	msg.WriteString("<green>:</> ")

	return msg.String()
}

func (p *Prompter) errorMsg() string {
	if p.Regexp != nil {
		return fmt.Sprintf("<bold><red>!! No coincide con /%s/</>", p.Regexp)
	}
	if len(p.Choices) > 0 {
		if len(p.Choices) == 1 {
			return fmt.Sprintf("<bold><red>!! Sólo puedes seleccionar `%s`</>", p.Choices[0])
		}
		choices := make([]string, len(p.Choices)-1)
		for i, v := range p.Choices[:len(p.Choices)-1] {
			choices[i] = "`" + v + "`"
		}
		return fmt.Sprintf(
			"<bold><red>!! Sólo puedes ingresar <bold>%s</><red> o <bold>`%s`</>",
			strings.Join(choices, ", "),
			p.Choices[len(p.Choices)-1],
		)
	}
	return ""
}

func (p *Prompter) inputIsValid(input string) bool {
	return p.regexp().MatchString(input)
}

var allReg = regexp.MustCompile(`.*`)

func (p *Prompter) regexp() *regexp.Regexp {
	if p.Regexp != nil {
		return p.Regexp
	}
	if p.reg != nil {
		return p.reg
	}
	if len(p.Choices) == 0 {
		p.reg = allReg
		return p.reg
	}

	choices := make([]string, len(p.Choices))
	for i, v := range p.Choices {
		choices[i] = regexp.QuoteMeta(v)
	}
	ignoreReg := ""
	if p.IgnoreCase {
		ignoreReg = "(?i)"
	}
	p.reg = regexp.MustCompile(fmt.Sprintf(`%s\A(?:%s)\z`, ignoreReg, strings.Join(choices, "|")))
	return p.reg
}
