package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
)

type Contest struct {
	Name        string
	URL         string
	ProblemURLs []string
}

type TestCase struct {
	Input  string
	Output string
}

type ProblemInfo struct {
	// 問題名の横にある問題ID
	// A - Payment だったら"A"
	ID string `toml:"id"`
	// 問題の名前
	// コンテスト名_ID
	// https://atcoder.jp/contests/abc173/tasks/abc173_aだったら"abc173_a"
	// https://atcoder.jp/contests/arc077/tasks/arc084_aだったら"arc077_c"
	// https://atcoder.jp/contests/abc001/tasks/abc001_1だったら"abc001_a"
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

type Problem struct {
	*ProblemInfo
	TestCases []*TestCase
}

type Problems struct {
	Problems []*ProblemInfo `toml:"problem"`
}

func LoadProblems() (*Problems, error) {
	_, err := os.Stat(".goj.toml")
	if err != nil {
		return nil, nil
	}
	var problems Problems
	_, err = toml.DecodeFile(".goj.toml", &problems)
	if err != nil {
		return nil, err
	}
	return &problems, nil
}

// ProblemInfoをtomlに追加で保存
func (p *ProblemInfo) AddTOML() error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	problems := &Problems{[]*ProblemInfo{p}}
	if err := enc.Encode(problems); err != nil {
		return err
	}

	file, err := os.OpenFile(".goj.toml", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(buf.String() + "\n")
	if err != nil {
		return err
	}

	return nil
}

func LoginAtCoder(client *http.Client, username, password string) error {
	submitURL := ATCODER_BASE_URL + "/login"
	res, err := client.Get(submitURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	csrf, err := ParseAtCoderCSRFToken(res.Body)
	if err != nil {
		return err
	}

	post, err := client.PostForm(submitURL, url.Values{
		"username":   {username},
		"password":   {password},
		"csrf_token": {csrf},
	})
	if err != nil {
		return err
	}
	defer post.Body.Close()

	doc, err := goquery.NewDocumentFromReader(post.Body)
	if err != nil {
		return err
	}

	fail := doc.Find("div.alert-danger")
	if len(fail.Nodes) != 0 {
		return errors.New("login failed")
	}

	success := doc.Find("div.alert-success")
	if len(success.Nodes) != 0 {
		return nil
	}

	return errors.New("unknown login error")
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
			[]byte(testcase.Input), 0666); err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(dir, fmt.Sprintf("sample-%d.out", i+1)),
			[]byte(testcase.Output), 0666); err != nil {
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

func ParseAtCoderCSRFToken(r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", nil
	}

	csrf, ok := doc.Find(`input[name="csrf_token"]`).Attr("value")
	if !ok {
		return "", errors.New("cannot find csrf_token")
	}
	return csrf, nil
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

	problemInfo := &ProblemInfo{
		ID:   id,
		Name: name,
		URL:  url,
	}
	if err := problemInfo.AddTOML(); err != nil {
		panic(err)
	}

	return &Problem{
		ProblemInfo: problemInfo,
		TestCases:   testcases,
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

func SubmitAtCoder(client *http.Client, contest, problem string, src_path string, lang string) error {
	submitURL := fmt.Sprintf("https://atcoder.jp/contests/%s/submit", contest)
	res, err := client.Get(submitURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 一つのio.Readerを二回読み込むには...？
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	languageID, err := ParseAtCoderLanguageID(bytes.NewReader(body), problem, lang)
	if err != nil {
		return err
	}

	code, err := ioutil.ReadFile(src_path)
	if err != nil {
		return err
	}

	csrf, err := ParseAtCoderCSRFToken(bytes.NewReader(body))
	if err != nil {
		return err
	}

	post, err := client.PostForm(submitURL, url.Values{
		"data.TaskScreenName": {problem},
		"data.LanguageId":     {languageID},
		"sourceCode":          {string(code)},
		"csrf_token":          {csrf},
	})
	if err != nil {
		return err
	}
	defer post.Body.Close()

	return nil
}

func ParseAtCoderLanguageID(r io.Reader, problem string, lang string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}

	var id string

	doc.Find(fmt.Sprintf("div[id=select-lang-%s] select option", problem)).Each(func(i int, s *goquery.Selection) {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.Text())), lang) {
			if v, ok := s.Attr("value"); id == "" && ok {
				id = v
			}
		}
	})

	if id == "" {
		return "", errors.New("cannot find language id")
	}
	return id, nil
}
