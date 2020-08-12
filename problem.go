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
		return nil, nil
	}
	var problems Problems
	_, err = toml.DecodeFile("goj.toml", &problems)
	if err != nil {
		return nil, err
	}
	return &problems, nil
}

// ProblemInfoをtomlに追加で保存
func (p *ProblemInfo) AddTOML() error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	problems := &Problems{[]*ProblemInfo{p}}
	if err := enc.Encode(problems); err != nil {
		return err
	}

	file, err := os.OpenFile("goj.toml", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(buf.String() + "\n")
	if err != nil {
		return err
	}

	return nil
}

func (p *Problem) Save() error {
	dir := "test_" + p.Name
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for i, testcase := range p.TestCases {
		// 既にファイルがあった場合は上書きになる
		if err := ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("sample-%d.in", i+1)),
			[]byte(testcase.Input), 0666); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("sample-%d.out", i+1)),
			[]byte(testcase.Output), 0666); err != nil {
			return err
		}
	}
	fmt.Printf("save %s's testcases\n", p.Name)
	return nil
}
