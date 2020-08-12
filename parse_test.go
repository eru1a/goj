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
		f, err := os.Open("testdata/submission")
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
