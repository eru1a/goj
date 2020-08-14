package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestJudge(t *testing.T) {
	os.Stderr = nil
	log.SetOutput(ioutil.Discard)

	if err := os.Chdir("testdata"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir(".."); err != nil {
			panic(err)
		}
	}()

	tests := []struct {
		problem string
		cmd     string
		ac      int
		wa      int
		re      int
	}{
		{
			problem: "abc001_1",
			cmd:     "go run abc001_1_ac.go",
			ac:      3,
			wa:      0,
			re:      0,
		},
		{
			problem: "abc001_1",
			cmd:     "go run abc001_1_wa.go",
			ac:      1,
			wa:      2,
			re:      0,
		},
		{
			problem: "abc001_1",
			cmd:     "go run abc001_1_re.go",
			ac:      0,
			wa:      0,
			re:      3,
		},
	}

	for _, test := range tests {
		ac, wa, re, err := Judge(test.problem, test.cmd)
		if err != nil {
			t.Fatal(err)
		}
		if ac != test.ac || wa != test.wa || re != test.re {
			t.Errorf("[%s] want(ac=%d, wa=%d, re=%d), got(ac=%d, wa=%d, re=%d)", test.problem, test.ac, test.wa, test.re, ac, wa, re)
		}
	}
}
