package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

type JudgeResult struct {
	AC           int
	WA           int
	RE           int
	IsAC         bool
	Result       string
	CntTestCases int
}

func NewJudgeResult(ac, wa, re int) *JudgeResult {
	result := "AC"
	if re > 0 {
		result = "RE"
	} else if wa > 0 {
		result = "WA"
	}
	return &JudgeResult{
		AC:           ac,
		WA:           wa,
		RE:           re,
		IsAC:         wa == 0 && re == 0,
		Result:       result,
		CntTestCases: ac + wa + re,
	}
}

func CmpOutput(actual, expected string, floatTolerance float64) bool {
	a := strings.Split(actual, "\n")
	e := strings.Split(expected, "\n")

	if len(a) != len(e) {
		return false
	}

	for i := range a {
		aa := strings.Split(a[i], " ")
		ee := strings.Split(e[i], " ")
		if len(aa) != len(ee) {
			return false
		}
		for j := range aa {
			af, err1 := strconv.ParseFloat(aa[j], 64)
			ef, err2 := strconv.ParseFloat(ee[j], 64)
			switch {
			case err1 == nil && err2 == nil:
				// float同士の比較なので許容誤差を考慮する
				if math.Abs(af-ef) > floatTolerance {
					return false
				}
			case err1 != nil && err2 == nil, err1 == nil && err1 != nil:
				// 片方しかfloatに変換出来てないのでおかしい
				return false
			default:
				// どっちもfloatに変換出来ない
				if aa[j] != ee[j] {
					return false
				}
			}
		}
	}

	return true
}

func Judge(problem string, command string, floatTolerance float64) (*JudgeResult, error) {
	LogInfo("test %s (%s)", problem, command)
	LogInfo("Float Tolerance: %f", floatTolerance)
	dir := fmt.Sprintf("test_%s", problem)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var ac, wa, re int
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) != ".in" {
			continue
		}

		testName := strings.TrimSuffix(f.Name(), ".in")
		LogInfo(testName)

		in, err := ioutil.ReadFile(filepath.Join(dir, testName+".in"))
		if err != nil {
			LogFailure("cannot find %s.in", filepath.Join(dir, testName))
			return nil, err
		}

		out, err := ioutil.ReadFile(filepath.Join(dir, testName+".out"))
		if err != nil {
			LogFailure("cannot find %s.out", filepath.Join(dir, testName))
			return nil, err
		}

		c := strings.Split(command, " ")
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stderr = os.Stderr

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		_, err = io.WriteString(stdin, string(in))
		if err != nil {
			return nil, err
		}

		stdout, err := cmd.Output()
		if err != nil {
			re++
			LogFailure(color.Red.Sprint("RE"))
			LogEmit("  %s\n", err)
			continue
		}

		if CmpOutput(string(stdout), string(out), floatTolerance) {
			ac++
			LogSuccess(color.Green.Sprint("AC"))
		} else {
			wa++
			LogFailure(color.Red.Sprint("WA"))
			LogEmit("expected:")
			LogEmit(color.Bold.Sprint(string(out)))
			LogEmit("got:")
			LogEmit(color.Bold.Sprint(string(stdout)))
		}
	}

	result := NewJudgeResult(ac, wa, re)
	if result.IsAC {
		LogSuccess("%s (AC:%d WA:%d RE:%d)", color.Green.Sprint(result.Result), ac, wa, re)
	} else {
		LogFailure("%s (AC:%d WA:%d RE:%d)", color.Red.Sprint(result.Result), ac, wa, re)
	}
	return result, nil
}
