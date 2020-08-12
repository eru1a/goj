package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestLoadProblems(t *testing.T) {
	if err := os.Chdir("testdata"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir(".."); err != nil {
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
						ID:      "A",
						Name:    "abc001_1",
						Contest: "abc001",
						URL:     "https://atcoder.jp/contests/abc001/tasks/abc001_1",
					},
					{
						ID:      "B",
						Name:    "abc173_b",
						Contest: "abc173",
						URL:     "https://atcoder.jp/contests/abc173/tasks/abc173_b",
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
