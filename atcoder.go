package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Contest struct {
	Name        string
	URL         string
	ProblemURLs []string
}

type Problem struct {
	// 問題名の横にある問題ID
	// A - Payment だったら"A"
	ID string
	// 問題の名前
	// コンテスト名_ID
	// https://atcoder.jp/contests/abc173/tasks/abc173_aだったら"abc173_a"
	// https://atcoder.jp/contests/arc077/tasks/arc084_aだったら"arc077_c"
	// https://atcoder.jp/contests/abc001/tasks/abc001_1だったら"abc001_a"
	Name      string
	URL       string
	TestCases []*TestCase
}

type TestCase struct {
	Input  string
	Output string
}

func (c *Contest) Save(client *http.Client) error {
	for _, url := range c.ProblemURLs {
		p, err := FetchAtCoderProblemFromURL(client, url)
		if err != nil {
			return err
		}
		if err := p.Save(); err != nil {
			return err
		}
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
			[]byte(testcase.Input), 0644); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("sample-%d.out", i+1)),
			[]byte(testcase.Output), 0644); err != nil {
			return err
		}
	}
	fmt.Printf("save %s's testcases\n", p.Name)
	return nil
}

func makeTemplateFile(problemName string, lang *Language) error {
	file := problemName + lang.Ext
	_, err := os.Stat(file)
	// ファイルが存在しない
	if err != nil {
		if err := ioutil.WriteFile(file, []byte(lang.Template), 0666); err != nil {
			return err
		}
	}
	return nil
}

func DownloadAtCoderContest(client *http.Client, contest string, lang *Language) error {
	c, err := FetchAtCoderContest(client, contest)
	if err != nil {
		return err
	}
	for _, url := range c.ProblemURLs {
		p, err := FetchAtCoderProblemFromURL(client, url)
		if err != nil {
			return err
		}
		if err := p.Save(); err != nil {
			return err
		}
		if err := makeTemplateFile(p.Name, lang); err != nil {
			return err
		}
	}
	return nil
}

func DownloadAtCoderProblem(client *http.Client, contest, problem string, lang *Language) error {
	p, err := FetchAtCoderProblem(client, contest, problem)
	if err != nil {
		return err
	}
	if err := p.Save(); err != nil {
		return err
	}
	if err := makeTemplateFile(p.Name, lang); err != nil {
		return err
	}
	return nil
}

const ATCODER_BASE_URL = "https://atcoder.jp"

func FetchAtCoderContest(client *http.Client, contest string) (*Contest, error) {
	url := fmt.Sprintf("%s/contests/%s/tasks", ATCODER_BASE_URL, contest)
	fmt.Printf("fetch %s\n", url)

	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	problemURLs, err := ParseAtCoderContest(res.Body)
	if err != nil {
		return nil, err
	}
	return &Contest{
		Name:        contest,
		URL:         url,
		ProblemURLs: problemURLs,
	}, nil
}

func ParseAtCoderContest(r io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var urls []string
	doc.Find("table > tbody > tr > td:first-child > a").Each(func(_ int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		urls = append(urls, ATCODER_BASE_URL+url)
	})
	if len(urls) == 0 {
		return nil, errors.New("cannot parse problem urls")
	}
	return urls, nil
}

func FetchAtCoderProblemFromURL(client *http.Client, url string) (*Problem, error) {
	contest := strings.Split(url, "/")[4]
	problem := strings.Split(url, "/")[6]
	return FetchAtCoderProblem(client, contest, problem)
}

func FetchAtCoderProblem(client *http.Client, contest, problem string) (*Problem, error) {
	url := fmt.Sprintf("%s/contests/%s/tasks/%s", ATCODER_BASE_URL, contest, problem)
	fmt.Printf("fetch %s\n", url)

	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	id, testcases, err := ParseAtCoderProblem(res.Body)
	if err != nil {
		return nil, err
	}
	name := contest + "_" + strings.ToLower(id)
	return &Problem{
		ID:        id,
		Name:      name,
		URL:       url,
		TestCases: testcases,
	}, nil
}

// IDとテストケースを返す
func ParseAtCoderProblem(r io.Reader) (string, []*TestCase, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", nil, err
	}

	var id string

	title := doc.Find("title").Text()
	if !strings.Contains(title, "-") {
		return "", nil, errors.New("cannot parse problem's title")
	}
	id = strings.TrimSpace(strings.Split(title, "-")[0])

	newTestCases := func(input, output []string) ([]*TestCase, error) {
		if len(input) != len(output) {
			return nil, errors.New("The lengths of input and output are different.")
		}
		var testcases []*TestCase
		for i := range input {
			testcases = append(testcases, &TestCase{input[i], output[i]})
		}
		return testcases, nil
	}

	var input, output []string

	// 最近のパターン
	// sectionの中のh3の中のpre
	//
	// https://atcoder.jp/contests/m-solutions2020/tasks/m_solutions2020_a
	// <div class="part">
	// <section>
	// <h3>入力例 1 <span class="btn btn-default btn-sm btn-copy" tabindex="0" data-toggle="tooltip" data-trigger="manual" title="" data-target="pre-sample0" data-original-title="Copied!">Copy</span></h3><div class="div-btn-copy"><span class="btn-copy btn-pre" tabindex="0" data-toggle="tooltip" data-trigger="manual" title="" data-target="pre-sample0" data-original-title="Copied!">Copy</span></div><pre id="pre-sample0">725
	// </pre>
	//
	// </section>
	// </div>
	{
		h3sel := doc.Find(".part > section > h3")
		h3sel.Each(func(_ int, s *goquery.Selection) {
			switch {
			case strings.HasPrefix(s.Text(), "入力例"):
				input = append(input, s.Parent().Find("pre").Text())
			case strings.HasPrefix(s.Text(), "出力例"):
				output = append(output, s.Parent().Find("pre").Text())
			}
		})
		if len(input) != 0 {
			testcases, err := newTestCases(input, output)
			if err != nil {
				return "", nil, err
			}
			return id, testcases, nil
		}
	}

	// 古いパターン
	// h3の下のselectionの中のpre
	//
	// https://atcoder.jp/contests/arc001/tasks/arc001_1
	// <h3>入力例 1</h3>
	// <section>
	// <div class="div-btn-copy"><span class="btn-copy btn-pre" tabindex="0" data-toggle="tooltip" data-trigger="manual" title="" data-target="for_copy0" data-original-title="Copied!">Copy</span></div><pre class="prettyprint linenums source-code prettyprinted" style=""><ol class="linenums"><li class="L0"><span class="lit">9</span></li><li class="L1"><span class="lit">131142143</span></li></ol></pre><pre id="for_copy0" class="source-code-for-copy">9
	// 131142143
	// </pre>
	// </section>
	{
		h3sel := doc.Find("h3")
		h3sel.Each(func(_ int, s *goquery.Selection) {
			switch {
			case strings.HasPrefix(s.Text(), "入力例"):
				input = append(input, s.Next().Find("pre").Text())
			case strings.HasPrefix(s.Text(), "出力例"):
				output = append(output, s.Next().Find("pre").Text())
			}
		})
		if len(input) != 0 {
			testcases, err := newTestCases(input, output)
			if err != nil {
				return "", nil, err
			}
			return id, testcases, nil
		}
	}

	// もっと別のパターンもある？

	return "", nil, errors.New("cannot find sample testcase")
}
