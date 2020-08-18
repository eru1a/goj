package main

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestCmpOutput(t *testing.T) {
	tests := []struct {
		actual         string
		expected       string
		floatTolerance float64
		result         bool
	}{
		{
			actual:         "6.283185307179586\n",
			expected:       "6.28318530717958623200\n",
			floatTolerance: 0.01,
			result:         true,
		},
		{
			actual:         "6.28\n",
			expected:       "6.28318530717958623200\n",
			floatTolerance: 0.01,
			result:         true,
		},
		{
			actual:         "6\n",
			expected:       "6.28318530717958623200\n",
			floatTolerance: 0.01,
			result:         false,
		},
		{
			actual:         "a b c\n1 2 3\n",
			expected:       "a b c\n1 2 3\n",
			floatTolerance: 0,
			result:         true,
		},
		{
			actual:         "a b c\n1 2 3\n",
			expected:       "a b c\n1 3\n",
			floatTolerance: 0,
			result:         false,
		},
		{
			actual:         "a b c\n1 2 3 4\n",
			expected:       "a b c\n1 2 3\n",
			floatTolerance: 0,
			result:         false,
		},
	}

	for _, test := range tests {
		result := CmpOutput(test.actual, test.expected, test.floatTolerance)
		if result != test.result {
			t.Errorf("CmpOutput(%s, %s, %f): want %v, got %v",
				test.actual, test.expected, test.floatTolerance, test.result, result)
		}
	}
}

func TestJudge(t *testing.T) {
	os.Stderr = nil
	log.SetOutput(ioutil.Discard)

	if err := os.Chdir("testdata/judge"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir("../.."); err != nil {
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
		result, err := Judge(test.problem, test.cmd, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, test.result) {
			t.Errorf("[%s] %s", test.problem, pretty.Compare(test.result, result))
		}
	}
}
