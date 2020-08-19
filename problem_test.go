package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestLoadProblems(t *testing.T) {
	if err := os.Chdir("testdata/problem"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir("../.."); err != nil {
			panic(err)
		}
	}()

	tests := []struct {
		problems *Problems
	}{
		{
			problems: &Problems{
				Problems: []*ProblemInfo{
					{
						ID:             "A",
						Name:           "abc001_1",
						Contest:        "abc001",
						URL:            "https://atcoder.jp/contests/abc001/tasks/abc001_1",
						TimeLimitSec:   2,
						MemoryLimitMB:  64,
						FloatTolerance: 0.0,
					},
					{
						ID:             "B",
						Name:           "abc173_b",
						Contest:        "abc173",
						URL:            "https://atcoder.jp/contests/abc173/tasks/abc173_b",
						TimeLimitSec:   2,
						MemoryLimitMB:  1024,
						FloatTolerance: 0.0,
					},
					{
						ID:             "A",
						Name:           "abc163_a",
						Contest:        "abc163",
						URL:            "https://atcoder.jp/contests/abc163/tasks/abc163_a",
						TimeLimitSec:   2,
						MemoryLimitMB:  1024,
						FloatTolerance: 0.01,
					},
					{
						ID:             "A",
						Name:           "abc173_a",
						Contest:        "abc173",
						URL:            "https://atcoder.jp/contests/abc173/tasks/abc173_a",
						TimeLimitSec:   2,
						MemoryLimitMB:  1024,
						FloatTolerance: 0.0,
					},
				},
			},
		},
	}

	for _, test := range tests {
		problems, err := LoadProblems()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(problems, test.problems) {
			t.Errorf("LoadProblems: %s", pretty.Compare(test.problems, problems))
		}
	}
}

func TestFindProblemName(t *testing.T) {
	if err := os.Chdir("testdata/problem"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir("../.."); err != nil {
			panic(err)
		}
	}()

	testsOK := []struct {
		fileName    string
		ext         string
		problemName string
	}{
		{
			fileName:    "abc001_1",
			ext:         ".cpp",
			problemName: "abc001_1",
		},
		{
			fileName:    "1",
			ext:         ".cpp",
			problemName: "abc001_1",
		},
		{
			fileName:    "abc163_a",
			ext:         ".cpp",
			problemName: "abc163_a",
		},
		{
			fileName:    "a",
			ext:         ".cpp",
			problemName: "abc173_a",
		},
		{
			fileName:    "b",
			ext:         ".cpp",
			problemName: "abc173_b",
		},
		{
			fileName:    "b",
			ext:         ".py",
			problemName: "abc173_b",
		},
		{
			fileName:    "",
			ext:         ".go",
			problemName: "abc163_a",
		},
		{
			fileName:    "",
			ext:         ".py",
			problemName: "abc173_a",
		},
	}

	for _, test := range testsOK {
		problemName, err := FindProblemName(test.fileName, test.ext)
		if err != nil {
			t.Fatal(err)
		}
		if problemName != test.problemName {
			t.Errorf("GetProblemName(%s, %s): want %s, got %s", test.fileName, test.ext, test.problemName, problemName)
		}
	}

	testsNG := []struct {
		fileName string
		ext      string
	}{
		{
			fileName: "abc001_2",
			ext:      ".cpp",
		},
		{
			fileName: "abc001_1",
			ext:      ".py",
		},
		{
			fileName: "",
			ext:      ".java",
		},
		{
			fileName: "c",
			ext:      ".cpp",
		},
		{
			fileName: "abc001",
			ext:      ".cpp",
		},
	}

	for _, test := range testsNG {
		_, err := FindProblemName(test.fileName, test.ext)
		if err == nil {
			t.Fatalf("GetProblemName(%s, %s) should be error", test.fileName, test.ext)
		}
	}
}

func TestFindProblem(t *testing.T) {
	if err := os.Chdir("testdata/problem"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir("../.."); err != nil {
			panic(err)
		}
	}()

	testsOK := []struct {
		problemName string
		problem     *ProblemInfo
	}{
		{
			problemName: "abc001_1",
			problem: &ProblemInfo{
				ID:             "A",
				Name:           "abc001_1",
				Contest:        "abc001",
				URL:            "https://atcoder.jp/contests/abc001/tasks/abc001_1",
				TimeLimitSec:   2,
				MemoryLimitMB:  64,
				FloatTolerance: 0.0,
			},
		},
		{
			problemName: "abc163_a",
			problem: &ProblemInfo{
				ID:             "A",
				Name:           "abc163_a",
				Contest:        "abc163",
				URL:            "https://atcoder.jp/contests/abc163/tasks/abc163_a",
				TimeLimitSec:   2,
				MemoryLimitMB:  1024,
				FloatTolerance: 0.01,
			},
		},
	}

	for _, test := range testsOK {
		problem, err := FindProblem(test.problemName)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(problem, test.problem) {
			t.Errorf("[%s]\n%s", test.problemName, pretty.Compare(test.problem, problem))
		}
	}

	testsNG := []struct {
		problemName string
	}{
		{"A"},
		{"abc001"},
		{"1"},
		{"abc173_c"},
	}

	for _, test := range testsNG {
		_, err := FindProblem(test.problemName)
		if err == nil {
			t.Errorf("GetProblem(%s): should be error", test.problemName)
		}
	}
}
