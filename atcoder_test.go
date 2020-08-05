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
			file: "testdata/abc001_tasks.html",
			urls: []string{
				"https://atcoder.jp/contests/abc001/tasks/abc001_1",
				"https://atcoder.jp/contests/abc001/tasks/abc001_2",
				"https://atcoder.jp/contests/abc001/tasks/abc001_3",
				"https://atcoder.jp/contests/abc001/tasks/abc001_4",
			},
		},
		{
			file: "testdata/abc173_tasks.html",
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
			panic(err)
		}
		defer f.Close()

		urls, err := ParseAtCoderContest(f)
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
			file: "testdata/abc001_1.html",
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
			file: "testdata/abc173_b.html",
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
			panic(err)
		}
		defer f.Close()
		id, testcases, err := ParseAtCoderProblem(f)
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
