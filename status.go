package main

import (
	"fmt"
	"strings"

	"github.com/gookit/color"
	"github.com/mattn/go-runewidth"
)

type SubmissionStatus struct {
	ID         string
	Date       string
	Problem    string
	User       string
	Language   string
	Score      string
	CodeLength string
	Result     string
	RunTime    string
	Memory     string
}

func (s *SubmissionStatus) DrawString() string {
	// 文字列を長さlimitにカットする(余ったら右空白埋め)。
	fix := func(s string, limit int) string {
		length := 0
		var ret string
		for _, c := range s {
			width := runewidth.RuneWidth(c)
			if length+width > limit {
				break
			}
			ret += string(c)
			length += width
		}
		return ret + strings.Repeat(" ", limit-length)
	}
	resultWithSpace := fmt.Sprintf("%3s", s.Result)
	result := color.Green.Sprint(resultWithSpace)
	if s.Result != "AC" {
		result = color.Yellow.Sprint(resultWithSpace)
	}
	return fmt.Sprintf("%s | %s | %s | %s | %s | %s | %s",
		fix(s.Problem, 16),
		result,
		fmt.Sprintf("%4s", s.Score),
		fmt.Sprintf("%.10s", s.Language),
		fmt.Sprintf("%8s", s.RunTime),
		fmt.Sprintf("%10s", s.Memory),
		s.Date[:19])
}
