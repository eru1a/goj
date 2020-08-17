package main

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
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
		result  *JudgeResult
	}{
		{
			problem: "abc001_1",
			cmd:     "go run abc001_1_ac.go",
			result: &JudgeResult{
				AC:           3,
				WA:           0,
				RE:           0,
				IsAC:         true,
				Result:       "AC",
				CntTestCases: 3,
			},
		},
		{
			problem: "abc001_1",
			cmd:     "go run abc001_1_wa.go",
			result: &JudgeResult{
				AC:           1,
				WA:           2,
				RE:           0,
				IsAC:         false,
				Result:       "WA",
				CntTestCases: 3,
			},
		},
		{
			problem: "abc001_1",
			cmd:     "go run abc001_1_re.go",
			result: &JudgeResult{
				AC:           0,
				WA:           0,
				RE:           3,
				IsAC:         false,
				Result:       "RE",
				CntTestCases: 3,
			},
		},
	}

	for _, test := range tests {
		result, err := Judge(test.problem, test.cmd)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("[%s] %s", test.problem, pretty.Compare(test.result, result))
		}
	}
}