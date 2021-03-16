package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gookit/color"
	cookiejar "github.com/juju/persistent-cookiejar"
)

const ATCODER_BASE_URL = "https://atcoder.jp"

var ErrNeedLogin = errors.New("you need to login")

type Contest struct {
	Name        string
	URL         string
	ProblemURLs []string
}

type TestCase struct {
	Input  string
	Output string
}

type AtCoder struct {
	Client *http.Client
}

func NewAtCoder(jar *cookiejar.Jar) *AtCoder {
	client := &http.Client{Jar: jar}
	return &AtCoder{Client: client}
}

func (a *AtCoder) DownloadContest(contest string, lang *Language) error {
	c, err := a.FetchContest(contest)
	if err != nil {
		return err
	}
	for _, url := range c.ProblemURLs {
		p, err := a.FetchProblemFromURL(url)
		if err != nil {
			return err
		}
		if err := p.AddTOML(); err != nil {
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

func (a *AtCoder) DownloadProblem(contest, problem string, lang *Language) error {
	// TODO: 上とほとんど同じ
	p, err := a.FetchProblem(contest, problem)
	if err != nil {
		return err
	}
	if err := p.AddTOML(); err != nil {
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

func (a *AtCoder) FetchContest(contest string) (*Contest, error) {
	url := fmt.Sprintf("%s/contests/%s/tasks", ATCODER_BASE_URL, contest)

	res, err := a.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	problemURLs, err := ParseContest(res.Body)
	if err != nil {
		return nil, err
	}
	return &Contest{
		Name:        contest,
		URL:         url,
		ProblemURLs: problemURLs,
	}, nil
}

func (a *AtCoder) CheckLogin() error {
	// リダイレクトが起きるかどうかでログインしているか確認する

	url := fmt.Sprintf("%s/contests/abc001/submit", ATCODER_BASE_URL)

	a.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 1 {
			return ErrNeedLogin
		}
		return nil
	}

	res, err := a.Client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

// url -> (contest, problem, error)
func ContestAndProblemFromURL(url string) (string, string, error) {
	// TODO: 正規表現で厳密に
	sp := strings.Split(url, "/")
	switch len(sp) {
	case 5, 6:
		// コンテストのトップページか問題一覧ページ
		return sp[4], "", nil
	case 7:
		// 問題ページ
		return sp[4], sp[6], nil
	default:
		return "", "", fmt.Errorf("invalid atcoder url: %s", url)
	}
}

func (a *AtCoder) FetchProblemFromURL(url string) (*Problem, error) {
	contest, problem, err := ContestAndProblemFromURL(url)
	if err != nil {
		return nil, err
	}
	return a.FetchProblem(contest, problem)
}

func (a *AtCoder) FetchProblem(contest, problem string) (*Problem, error) {
	url := fmt.Sprintf("%s/contests/%s/tasks/%s", ATCODER_BASE_URL, contest, problem)

	res, err := a.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	LogSuccess("fetched %s", url)

	p, err := ParseProblem(res.Body, contest, problem, url)
	if err != nil {
		LogFailure("failed to parse problem's testcase: url %s", url)
		return nil, err
	}
	return p, nil
}

func (a *AtCoder) Login(username, password string) error {
	submitURL := ATCODER_BASE_URL + "/login"
	res, err := a.Client.Get(submitURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	csrf, err := ParseCSRFToken(res.Body)
	if err != nil {
		return err
	}

	post, err := a.Client.PostForm(submitURL, url.Values{
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
		return errors.New("div.alert-danger found")
	}

	success := doc.Find("div.alert-success")
	if len(success.Nodes) != 0 {
		LogSuccess("login success")
		return nil
	}

	return errors.New("couldn't find div.alert-danger or div.alert-success")
}

func (a *AtCoder) Submit(contest, problem string, srcPath string, lang string) error {
	submitURL := fmt.Sprintf("%s/contests/%s/submit", ATCODER_BASE_URL, contest)

	a.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 1 {
			return ErrNeedLogin
		}
		return nil
	}

	res, err := a.Client.Get(submitURL)
	if err != nil {
		if errors.Is(err, ErrNeedLogin) {
			return ErrNeedLogin
		}
		return err
	}
	defer res.Body.Close()

	// 一つのio.Readerを二回読み込むには...？
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	languageID, err := ParseLanguageID(bytes.NewReader(body), problem, lang)
	if err != nil {
		return err
	}

	code, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return err
	}

	csrf, err := ParseCSRFToken(bytes.NewReader(body))
	if err != nil {
		return err
	}

	// 提出時のリダイレクトは必要
	a.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return nil
	}

	post, err := a.Client.PostForm(submitURL, url.Values{
		"data.TaskScreenName": {problem},
		"data.LanguageId":     {languageID},
		"sourceCode":          {string(code)},
		"csrf_token":          {csrf},
	})
	if err != nil {
		return err
	}
	defer post.Body.Close()

	LogSuccess("submit %s/%s %s(%s)", contest, problem, srcPath, lang)

	return nil
}

func (a *AtCoder) SubmissionsStatus(contest string) ([]*SubmissionStatus, error) {
	submissionsURL := fmt.Sprintf("%s/contests/%s/submissions/me", ATCODER_BASE_URL, contest)

	a.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 1 {
			return ErrNeedLogin
		}
		return nil
	}

	res, err := a.Client.Get(submissionsURL)
	if err != nil {
		if errors.Is(err, ErrNeedLogin) {
			return nil, ErrNeedLogin
		}
		return nil, err
	}
	defer res.Body.Close()
	status, err := ParseSubmissionsStatus(res.Body)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// 自分の最も最近の提出をジャッジが終わるまで観測する
// 提出IDから辿れる専用のページを見たほうがいいだろうか？
func (a *AtCoder) WatchLastSubmissionStatus(contest string) error {
	isFinish := func(result string) bool {
		for _, s := range []string{"AC", "WA", "TLE", "MLE", "RE", "CE", "QLE", "IE"} {
			if result == s {
				return true
			}
		}
		return false
	}

	isWA := func(result string) bool {
		for _, s := range []string{"WA", "TLE", "MLE", "RE", "CE", "QLE", "IE"} {
			if strings.Contains(result, s) {
				return true
			}
		}
		return false
	}

	for {
		status, err := a.SubmissionsStatus(contest)
		if err != nil {
			return err
		}

		result := status[0].Result
		if result == "AC" {
			result = color.Green.Sprint(result)
		}
		if isWA(status[0].Result) {
			result = color.Yellow.Sprint(result)
		}
		if isFinish(status[0].Result) {
			fmt.Printf("\r\033[KResult: %s\n", result)
			break
		}
		fmt.Printf("\r\033[KJudging... %s", result)

		// time.Tickerのほうがいいのかな
		time.Sleep(time.Second * 2)
	}

	return nil
}
