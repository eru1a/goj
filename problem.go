package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ProblemInfo struct {
	// 問題名の横にある問題ID (e.g. "A", "B")
	// AtCoder Beginners Selectionだと"ABC086A"みたいなIDになる。
	ID string `toml:"id"`
	// ファイル名等に使う問題の名前。(e.g. "abc174_a")
	// URLの最後の要素。
	// 昔のコンテストだと、
	// `abc001/abc001_1`とか`abc077/arc084_a`みたいに`コンテスト_ID`の形になってないことがある。
	// 最近のでも`m-solutions2020/m_solutions2020_a`みたいなのがあるか。
	Name string `toml:"name"`
	// コンテスト名。
	// e.g. "abc174"
	Contest string `toml:"contest"`
	URL     string `toml:"url"`
	// 実行制限時間(sec)
	TimeLimitSec int `toml:"time_limit"`
	// メモリ制限(mb)
	MemoryLimitMB int `toml:"memory_limit"`
	// 小数の許容誤差
	FloatTolerance float64 `toml:"float_tolerance"`
}

type Problem struct {
	*ProblemInfo
	TestCases []*TestCase
}

type Problems struct {
	Problems []*ProblemInfo `toml:"problem"`
}

func LoadProblems() (*Problems, error) {
	_, err := os.Stat("goj.toml")
	if err != nil {
		LogInfo("goj.toml not found")
		return nil, nil
	}
	var problems Problems
	_, err = toml.DecodeFile("goj.toml", &problems)
	if err != nil {
		LogFailure("failed to load goj.toml")
		return nil, err
	}
	LogSuccess("loaded goj.toml")
	return &problems, nil
}

// ProblemInfoをtomlに追加で保存
// TODO: 同じ問題を重複して記録しないようにする
func (p *ProblemInfo) AddTOML() error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	problems := &Problems{[]*ProblemInfo{p}}
	if err := enc.Encode(problems); err != nil {
		LogFailure("failed to encode problems to toml")
		return err
	}

	file, err := os.OpenFile("goj.toml", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		LogFailure("failed to (read/create) goj.toml")
		return err
	}
	defer file.Close()

	_, err = file.WriteString(buf.String() + "\n")
	if err != nil {
		LogFailure("failed to add %s to goj.toml", p.Name)
		return err
	}

	LogSuccess("added %s to goj.toml", p.Name)
	return nil
}

func (p *Problem) Save() error {
	dir := "test_" + p.Name
	if err := os.MkdirAll(dir, 0755); err != nil {
		LogFailure("failed to mkdir %s", dir)
		return err
	}
	for i, testcase := range p.TestCases {
		// 既にファイルがあった場合は上書きになる
		if err := ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("sample-%d.in", i+1)),
			[]byte(testcase.Input), 0666); err != nil {
			LogFailure("failed to save sample-%d.in", i+1)
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("sample-%d.out", i+1)),
			[]byte(testcase.Output), 0666); err != nil {
			LogFailure("failed to save sample-%d.out", i+1)
			return err
		}
	}
	LogSuccess("saved %s's testcases", p.Name)
	return nil
}
