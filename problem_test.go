package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

func createTempFiles(dir string) {
	// この順番にファイルを作る
	files := []string{
		"abc001_1.cpp",
		"abc173_b.cpp",
		"abc173_b.py",
		"abc163_a.go",
		"abc163_a.cpp",
		"abc173_a.py",
		"abc173_a.cpp",
	}

	gojtoml := `[[problem]]
  id = "A"
  name = "abc001_1"
  contest = "abc001"
  url = "https://atcoder.jp/contests/abc001/tasks/abc001_1"
  time_limit = 2
  memory_limit = 64
  float_tolerance = 0.0

[[problem]]
  id = "B"
  name = "abc173_b"
  contest = "abc173"
  url = "https://atcoder.jp/contests/abc173/tasks/abc173_b"
  time_limit = 2
  memory_limit = 1024
  float_tolerance = 0.0

[[problem]]
  id = "A"
  name = "abc163_a"
  contest = "abc163"
  url = "https://atcoder.jp/contests/abc163/tasks/abc163_a"
  time_limit = 2
  memory_limit = 1024
  float_tolerance = 0.01

[[problem]]
  id = "A"
  name = "abc173_a"
  contest = "abc173"
  url = "https://atcoder.jp/contests/abc173/tasks/abc173_a"
  time_limit = 2
  memory_limit = 1024
  float_tolerance = 0.0
`

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	for i, file := range files {
		if _, err := os.Create(filepath.Join(dir, file)); err != nil {
			panic(err)
		}
		if err := os.Chtimes(filepath.Join(dir, file),
			time.Now(),
			time.Now().Add(time.Hour*time.Duration(i))); err != nil {
			panic(err)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "goj.toml"), []byte(gojtoml), 0666); err != nil {
		panic(err)
	}
}

func TestLoadProblems(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "goj", "problem")
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	createTempFiles(tmpDir)
	if err := os.Chdir(tmpDir); err != nil {
		panic(err)
	}
	defer func() {
		if os.Chdir(curDir); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
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
	tmpDir := filepath.Join(os.TempDir(), "goj", "problem")
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	createTempFiles(tmpDir)
	if err := os.Chdir(tmpDir); err != nil {
		panic(err)
	}
	defer func() {
		if os.Chdir(curDir); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			panic(err)
		}
	}()

	testsOK := []struct {
		fileName string
		ext      string
		want     string
	}{
		{
			fileName: "abc001_1",
			ext:      ".cpp",
			want:     "abc001_1",
		},
		{
			fileName: "1",
			ext:      ".cpp",
			want:     "abc001_1",
		},
		{
			fileName: "abc163_a",
			ext:      ".cpp",
			want:     "abc163_a",
		},
		{
			fileName: "a",
			ext:      ".cpp",
			want:     "abc173_a",
		},
		{
			fileName: "b",
			ext:      ".cpp",
			want:     "abc173_b",
		},
		{
			fileName: "b",
			ext:      ".py",
			want:     "abc173_b",
		},
		{
			fileName: "",
			ext:      ".go",
			want:     "abc163_a",
		},
		{
			fileName: "",
			ext:      ".py",
			want:     "abc173_a",
		},
	}

	for _, test := range testsOK {
		got, err := FindProblemName(test.fileName, test.ext)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("FindProblemName(%s, %s): want %s, got %s", test.fileName, test.ext, test.want, got)
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
			t.Fatalf("FindProblemName(%s, %s) should be error", test.fileName, test.ext)
		}
	}
}

func TestFindProblem(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "goj", "problem")
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	createTempFiles(tmpDir)
	if err := os.Chdir(tmpDir); err != nil {
		panic(err)
	}
	defer func() {
		if os.Chdir(curDir); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
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
