package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gookit/color"
)

type JudgeResult struct {
	AC           int
	WA           int
	RE           int
	TLE          int
	MLE          int
	IsAC         bool
	Result       string
	CntTestCases int
}

func NewJudgeResult(ac, wa, re, tle, mle int) *JudgeResult {
	result := "AC"
	switch {
	case tle > 0:
		result = "TLE"
	case mle > 0:
		result = "MLE"
	case re > 0:
		result = "RE"
	case wa > 0:
		result = "WA"
	}
	return &JudgeResult{
		AC:           ac,
		WA:           wa,
		RE:           re,
		TLE:          tle,
		MLE:          mle,
		IsAC:         wa == 0 && re == 0 && tle == 0 && mle == 0,
		Result:       result,
		CntTestCases: ac + wa + re + tle + mle,
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

func Judge(problem string, command string, timeLimitMS int, memoryLimitMB int, floatTolerance float64) (*JudgeResult, error) {
	// timeLimitは本当はsecで渡したかったけどテスト時にミリ秒で渡せないと待ち時間が長くなるので...
	LogInfo("test %s (%s)", problem, command)
	LogInfo("Time Limit: %.1f sec", float64(timeLimitMS)/1000)
	LogInfo("Memory Limit: %d MB", memoryLimitMB)
	LogInfo("Float Tolerance: %.9f", floatTolerance)
	dir := fmt.Sprintf("test_%s", problem)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var ac, wa, re, tle, mle int
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) != ".in" {
			continue
		}
		LogEmit("")

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
		ctx, cancel := context.WithTimeout(context.Background(),
			time.Millisecond*time.Duration(timeLimitMS+100))
		defer cancel()

		cmd := exec.CommandContext(ctx, c[0], c[1:]...)
		cmd.Stderr = os.Stderr

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}

		_, err = io.WriteString(stdin, string(in))
		if err != nil {
			return nil, err
		}

		start := time.Now()
		stdout, err := cmd.Output()

		// 経過時間
		elapsedMS := float64(time.Since(start)) / float64(time.Millisecond)
		// 使用メモリってこうやって測るの？
		memory := float64(cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss) / 1000.0

		LogStatus("time: %.5f sec", elapsedMS)
		LogStatus("memory: %.5f MB", memory)

		switch {
		case err != nil && ctx.Err() == context.DeadlineExceeded:
			tle++
			LogFailure(color.Red.Sprint("TLE"))
		case err != nil:
			re++
			LogFailure(color.Red.Sprint("RE"))
			LogEmit("  %s\n", err)
		case memory > float64(memoryLimitMB):
			mle++
			LogFailure(color.Red.Sprint("MLE"))
		case elapsedMS > float64(timeLimitMS):
			// ctxのタイムアウトは100ミリ秒余分に取ってるので起こりうる
			tle++
			LogFailure(color.Red.Sprint("TLE"))
		case CmpOutput(string(stdout), string(out), floatTolerance):
			ac++
			LogSuccess(color.Green.Sprint("AC"))
		default:
			wa++
			LogFailure(color.Red.Sprint("WA"))
			LogEmit("expected:")
			LogEmit(color.Bold.Sprint(string(out)))
			LogEmit("got:")
			LogEmit(color.Bold.Sprint(string(stdout)))
		}
	}

	LogEmit("")

	result := NewJudgeResult(ac, wa, re, tle, mle)
	if result.IsAC {
		LogSuccess("%s (AC:%d WA:%d RE:%d TLE:%d MLE:%d)", color.Green.Sprint(result.Result), ac, wa, re, tle, mle)
	} else {
		LogFailure("%s (AC:%d WA:%d RE:%d TLE:%d MLE:%d)", color.Red.Sprint(result.Result), ac, wa, re, tle, mle)
	}
	return result, nil
}
