package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestParseAtCoderContest(t *testing.T) {
	tests := []struct {
		file string
		urls []string
	}{
		{
			file: "testdata/abc001_tasks",
			urls: []string{
				"https://atcoder.jp/contests/abc001/tasks/abc001_1",
				"https://atcoder.jp/contests/abc001/tasks/abc001_2",
				"https://atcoder.jp/contests/abc001/tasks/abc001_3",
				"https://atcoder.jp/contests/abc001/tasks/abc001_4",
			},
		},
		{
			file: "testdata/abc173_tasks",
			urls: []string{
				"https://atcoder.jp/contests/abc173/tasks/abc173_a",
				"https://atcoder.jp/contests/abc173/tasks/abc173_b",
				"https://atcoder.jp/contests/abc173/tasks/abc173_c",
				"https://atcoder.jp/contests/abc173/tasks/abc173_d",
				"https://atcoder.jp/contests/abc173/tasks/abc173_e",
				"https://atcoder.jp/contests/abc173/tasks/abc173_f",
			},
		},
	}

	for _, test := range tests {
		f, err := os.Open(test.file)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		urls, err := ParseContest(f)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(urls, test.urls) {
			t.Errorf("[]\nwant\t%v\ngot\t%v", test.urls, urls)
		}
	}
}

func TestParseAtCoderProblem(t *testing.T) {
	tests := []struct {
		file      string
		id        string
		testcases []*TestCase
	}{
		{
			file: "testdata/abc001_1",
			id:   "A",
			testcases: []*TestCase{
				{
					Input:  "15\n10\n",
					Output: "5\n",
				},
				{
					Input:  "0\n0\n",
					Output: "0\n",
				},
				{
					Input:  "5\n20\n",
					Output: "-15\n",
				},
			},
		},
		{
			file: "testdata/abc173_b",
			id:   "B",
			testcases: []*TestCase{
				{
					Input:  "6\nAC\nTLE\nAC\nAC\nWA\nTLE\n",
					Output: "AC x 3\nWA x 1\nTLE x 2\nRE x 0\n",
				},
				{
					Input:  "10\nAC\nAC\nAC\nAC\nAC\nAC\nAC\nAC\nAC\nAC\n",
					Output: "AC x 10\nWA x 0\nTLE x 0\nRE x 0\n",
				},
			},
		},
	}

	for _, test := range tests {
		f, err := os.Open(test.file)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		id, testcases, err := ParseProblem(f)
		if err != nil {
			t.Error(err)
		}
		if id != test.id {
			t.Errorf("[%s] want %v, got %v\n", test.file, test.id, id)
		}
		if !reflect.DeepEqual(testcases, test.testcases) {
			t.Errorf("[%s] %s", test.file, pretty.Compare(test.testcases, testcases))
		}
	}
}

func TestParseCSRFToken(t *testing.T) {
	tests := []struct {
		file string
		csrf string
	}{
		{
			file: "testdata/login",
			csrf: "JmExC2cpP04lxScfq2TNW/1o0XDVKQfkJjYW3PEtnHM=",
		},
	}

	for _, test := range tests {
		f, err := os.Open(test.file)
		if err != nil {
			t.Fatal(err)
		}
		csrf, err := ParseCSRFToken(f)
		if err != nil {
			t.Fatal(err)
		}
		if csrf != test.csrf {
			t.Errorf("[%s] want %s, got %s", test.file, test.csrf, csrf)
		}
	}
}

func TestParseLanguageID(t *testing.T) {
	tests := []struct {
		lang       string
		languageID string
	}{
		{
			lang:       "c++",
			languageID: "4003",
		},
		{
			lang:       "python",
			languageID: "4006",
		},
		{
			lang:       "go",
			languageID: "4026",
		},
	}

	for _, test := range tests {
		f, err := os.Open("testdata/submit")
		if err != nil {
			t.Fatal(err)
		}
		languageID, err := ParseLanguageID(f, "abc174_a", test.lang)
		if err != nil {
			t.Fatal(err)
		}
		if languageID != test.languageID {
			t.Errorf("[%s] want %s, got %s", test.lang, test.languageID, languageID)
		}
	}
}

func TestParseSubmissionsStatus(t *testing.T) {
	tests := []struct {
		file   string
		status []*SubmissionStatus
	}{
		{
			file: "testdata/submissions",
			status: []*SubmissionStatus{
				{
					ID:         "15903812",
					Date:       "2020-08-15 18:30:10+0900",
					Problem:    "A - 積雪深差",
					User:       "eru1a ",
					Language:   "C++ (GCC 9.2.1)",
					Score:      "0",
					CodeLength: "138 Byte",
					Result:     "TLE",
					RunTime:    "2205 ms",
					Memory:     "3292 KB",
				},
				{
					ID:         "15903794",
					Date:       "2020-08-15 18:29:23+0900",
					Problem:    "A - 積雪深差",
					User:       "eru1a ",
					Language:   "C++ (GCC 9.2.1)",
					Score:      "0",
					CodeLength: "115 Byte",
					Result:     "WA",
					RunTime:    "6 ms",
					Memory:     "3632 KB",
				},
				{
					ID:         "15903787",
					Date:       "2020-08-15 18:29:10+0900",
					Problem:    "A - 積雪深差",
					User:       "eru1a ",
					Language:   "C++ (GCC 9.2.1)",
					Score:      "100",
					CodeLength: "119 Byte",
					Result:     "AC",
					RunTime:    "6 ms",
					Memory:     "3584 KB",
				},
				{
					ID:         "15903782",
					Date:       "2020-08-15 18:28:45+0900",
					Problem:    "A - 積雪深差",
					User:       "eru1a ",
					Language:   "C++ (GCC 9.2.1)",
					Score:      "0",
					CodeLength: "97 Byte",
					Result:     "CE",
					RunTime:    "Detail",
					Memory:     "",
				},
				{
					ID:         "13696846",
					Date:       "2020-05-30 08:33:40+0900",
					Problem:    "B - 視程の通報",
					User:       "eru1a ",
					Language:   "Rust (1.15.1)",
					Score:      "100",
					CodeLength: "1511 Byte",
					Result:     "AC",
					RunTime:    "2 ms",
					Memory:     "4352 KB",
				},
				{
					ID:         "13696841",
					Date:       "2020-05-30 08:33:30+0900",
					Problem:    "A - 積雪深差",
					User:       "eru1a ",
					Language:   "Rust (1.15.1)",
					Score:      "100",
					CodeLength: "1197 Byte",
					Result:     "AC",
					RunTime:    "2 ms",
					Memory:     "4352 KB",
				},
			},
		},
	}

	for _, test := range tests {
		f, err := os.Open(test.file)
		if err != nil {
			t.Fatal(err)
		}
		status, err := ParseSubmissionsStatus(f)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(status, test.status) {
			t.Errorf(pretty.Compare(test.status, status))
		}
	}
}
