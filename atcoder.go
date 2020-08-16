package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	cookiejar "github.com/juju/persistent-cookiejar"
)

const ATCODER_BASE_URL = "https://atcoder.jp"

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
	p, err := a.FetchProblem(contest, problem)
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

func (a *AtCoder) FetchProblemFromURL(url string) (*Problem, error) {
	contest := strings.Split(url, "/")[4]
	problem := strings.Split(url, "/")[6]
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

	id, testcases, err := ParseProblem(res.Body)
	if err != nil {
		LogFailure("failed to parse problem's testcase: url %s", url)
		return nil, err
	}

	problemInfo := &ProblemInfo{
		ID:      id,
		Name:    problem,
		Contest: contest,
		URL:     url,
	}
	if err := problemInfo.AddTOML(); err != nil {
		return nil, err
	}

	return &Problem{
		ProblemInfo: problemInfo,
		TestCases:   testcases,
	}, nil
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
	res, err := a.Client.Get(submitURL)
	if err != nil {
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
	res, err := a.Client.Get(submissionsURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	status, err := ParseSubmissionsStatus(res.Body)
	if err != nil {
		return nil, err
	}
	return status, nil
}
