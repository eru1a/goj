package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseContest(r io.Reader) ([]string, error) {
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

func parseProblemID(doc *goquery.Document) (string, error) {
	title := doc.Find("title").Text()
	if !strings.Contains(title, "-") {
		return "", errors.New("cannot parse problem's title")
	}
	return strings.TrimSpace(strings.Split(title, "-")[0]), nil
}

func parseProblemTestCases(doc *goquery.Document) ([]*TestCase, error) {
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
		var input, output []string

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
				return nil, err
			}
			return testcases, nil
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
		var input, output []string

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
				return nil, err
			}
			return testcases, nil
		}
	}

	// もっと別のパターンもある？
	return nil, errors.New("cannot find sample testcase")
}

// 実行時間制限とメモリ制限を取得する
// 見つからなかった場合ログにエラーを出力しデフォルト値として(2, 1024)を返す
// 実用上はデフォルトの値で困らないのでエラーを返したりはしない
func parseProblemTimeLimitAndMemoryLimit(doc *goquery.Document) (int, int) {
	// Time Limit: 2 sec / Memory Limit: 1024 MB
	re := regexp.MustCompile("Time Limit: ([0-9]+) sec / Memory Limit: ([0-9]+) MB")
	text := strings.TrimSpace(doc.Find("#task-statement").Prev().Text())
	match := re.FindSubmatch([]byte(text))
	if len(match) != 3 {
		LogFailure("cannot find time limit and memory limit")
		return 2, 1024
	}
	timeLimit, err := strconv.Atoi(string(match[1]))
	if err != nil {
		LogFailure("%cannot find time limit: v", err)
		return 2, 1024
	}
	memoryLimit, err := strconv.Atoi(string(match[2]))
	if err != nil {
		LogFailure("cannot find memory limit: %v", err)
		return 2, 1024
	}
	return timeLimit, memoryLimit
}

// 小数の許容誤差を返す。見つからなければ0.0。
func parseProblemFloatTolerance(doc *goquery.Document) float64 {
	// 10^{−2}とか10^{-5}とか
	// 何で負の符号が2種類あるの...
	re := regexp.MustCompile(`10\^\{[−|-]([0-9]+)\}`)

	floatTolerance := 0.0
	doc.Find("section > h3").Each(func(_ int, s *goquery.Selection) {
		if s.Text() != "出力" {
			return
		}
		s.NextAll().Each(func(_ int, s *goquery.Selection) {
			match := re.FindSubmatch([]byte(s.Text()))
			if len(match) == 2 {
				i, err := strconv.Atoi(string(match[1]))
				// 一応エラーチェックしとく
				if err != nil {
					LogFailure("%v", err)
					return
				}
				floatTolerance = math.Pow10(-i)
			}
		})
	})
	return floatTolerance
}

// IDとテストケースを返す
func ParseProblem(r io.Reader, contest, problem, url string) (*Problem, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	id, err := parseProblemID(doc)
	if err != nil {
		return nil, err
	}

	testcases, err := parseProblemTestCases(doc)
	if err != nil {
		return nil, err
	}

	timeLimit, memoryLimit := parseProblemTimeLimitAndMemoryLimit(doc)
	floatTolerance := parseProblemFloatTolerance(doc)

	return &Problem{
		ProblemInfo: &ProblemInfo{
			Contest:        contest,
			Name:           problem,
			URL:            url,
			ID:             id,
			TimeLimitSec:   timeLimit,
			MemoryLimitMB:  memoryLimit,
			FloatTolerance: floatTolerance,
		},
		TestCases: testcases,
	}, nil
}

func ParseCSRFToken(r io.Reader) (string, error) {
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

func ParseLanguageID(r io.Reader, problem string, lang string) (string, error) {
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

func ParseSubmissionsStatus(r io.Reader) ([]*SubmissionStatus, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var submissions []*SubmissionStatus

	doc.Find("table > tbody > tr").Each(func(_ int, s *goquery.Selection) {
		submissions = append(submissions, &SubmissionStatus{
			ID:         s.Find("td:nth-child(5)").AttrOr("data-id", ""),
			Date:       s.Find("td:nth-child(1)").Text(),
			Problem:    s.Find("td:nth-child(2)").Text(),
			User:       s.Find("td:nth-child(3)").Text(),
			Language:   s.Find("td:nth-child(4)").Text(),
			Score:      s.Find("td:nth-child(5)").Text(),
			CodeLength: s.Find("td:nth-child(6)").Text(),
			Result:     s.Find("td:nth-child(7)").Text(),
			RunTime:    strings.TrimSpace(s.Find("td:nth-child(8)").Text()),
			Memory:     s.Find("td:nth-child(9)").Text(),
		})
	})

	return submissions, nil
}
