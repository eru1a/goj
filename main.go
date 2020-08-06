package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/gookit/color"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

const defaultConfigToml = `default_language = "c++"

[[language]]
name = "c++"
ext = ".cpp"
# [P] : Problem Name
build_cmd = "g++ -g -o [P] [P].cpp"
run_cmd = "./[P]"
template = """#include <bits/stdc++.h>

using namespace std;
using ll = long long;

int main() {
  cin.tie(nullptr);
  ios::sync_with_stdio(false);

  return 0;
}
"""

[[language]]
name = "python"
ext = ".py"
run_cmd = "python [P].py"
`

func findLang(languages []*Language, defaultLang, argLang string) (*Language, error) {
	langName := defaultLang
	if argLang != "" {
		langName = argLang
	}
	for _, l := range languages {
		if l.Name == langName {
			return l, nil
		}
	}
	return nil, fmt.Errorf("cannot find %s in languages", langName)
}

// keywordから問題を推測する。
// - keywordが問題の名前('abc173_a'等)ならファイル名が一致している。
// - keywordが問題のID('a'等)ならファイル名が'_ID'で終わっている。
// - keywordが空文字列なら条件なし。
// 上記の条件を満たすファイルの内、拡張子がextで最も最近編集されたファイルの名前を返す。
func getProblem(keyword string, ext string) (string, error) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return "", err
	}
	var files2 []os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ext) {
			continue
		}
		fileNameWithoutExt := strings.TrimSuffix(file.Name(), ext)
		switch {
		case keyword == "":
			files2 = append(files2, file)
		case fileNameWithoutExt == keyword:
			files2 = append(files2, file)
		case strings.HasSuffix(fileNameWithoutExt, "_"+keyword):
			files2 = append(files2, file)
		}
	}
	if len(files2) == 0 {
		return "", fmt.Errorf("cannot find a file that meets the requirements. keyword: %s, ext: %s", keyword, ext)
	}
	sort.SliceStable(files2, func(i, j int) bool {
		return files2[i].ModTime().Unix() > files2[j].ModTime().Unix()
	})
	return strings.TrimSuffix(files2[0].Name(), ext), nil
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	gojCacheDir := filepath.Join(cacheDir, "goj")
	if err := os.MkdirAll(gojCacheDir, 0755); err != nil {
		panic(err)
	}
	cookieJarFile := filepath.Join(gojCacheDir, "cookiejar")
	jar, err := cookiejar.New(&cookiejar.Options{Filename: cookieJarFile})
	if err != nil {
		panic(err)
	}
	client := &http.Client{Jar: jar}

	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "download",
			Aliases: []string{"dl", "d"},
			Usage:   "goj download <contest> or <contest/problem>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "language, l",
					Value: "c++",
				},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args()) > 1 {
					return errors.New("goj download <contest> or <contest/problem>")
				}
				lang, err := findLang(config.Languages, config.DefaultLanguage, c.String("l"))
				if err != nil {
					return err
				}

				contest := c.Args().First()
				if contest == "" {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					contest = filepath.Base(cwd)
				}

				split := strings.Split(contest, "/")
				switch len(split) {
				case 1:
					if err := DownloadAtCoderContest(client, split[0], lang); err != nil {
						return err
					}
				case 2:
					if err := DownloadAtCoderProblem(client, split[0], split[1], lang); err != nil {
						return err
					}
				default:
					return errors.New("goj d <contest> or <contest/problem>")
				}
				return nil
			},
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "goj test <problem>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "command, c",
					Value: "<none>",
				},
				cli.StringFlag{
					Name:  "language, l",
					Value: "c++",
				},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args()) > 1 {
					return errors.New("goj test <problem>")
				}
				lang, err := findLang(config.Languages, config.DefaultLanguage, c.String("l"))
				if err != nil {
					return err
				}

				problem, err := getProblem(c.Args().First(), lang.Ext)
				if err != nil {
					return err
				}
				cmd := c.String("c")
				if cmd == "<none>" {
					cmd = lang.GetRunCmd(problem)
					if err := lang.Build(problem); err != nil {
						return err
					}
				}
				ac, wa, re := Judge(problem, cmd)
				result := color.Green.Sprint("AC")
				if re > 0 {
					result = color.Red.Sprint("RE")
				} else if wa > 0 {
					result = color.Red.Sprint("WA")
				}
				fmt.Printf("%s (AC:%d WA:%d RE:%d)\n", result, ac, wa, re)
				return nil
			},
		},
		{
			Name:  "login",
			Usage: "goj login",
			Action: func(c *cli.Context) error {
				if len(c.Args()) > 0 {
					return errors.New("goj login")
				}

				var username string
				fmt.Print("username: ")
				_, err := fmt.Scanln(&username)
				if err != nil {
					return err
				}
				fmt.Print("password: ")
				bytes, err := terminal.ReadPassword(syscall.Stdin)
				if err != nil {
					return err
				}
				password := string(bytes)
				if err := LoginAtCoder(client, username, password); err != nil {
					return err
				}
				if err := jar.Save(); err != nil {
					return err
				}

				fmt.Println("login success")
				return nil
			},
		},
		{
			Name:    "submit",
			Aliases: []string{"s"},
			Usage:   "goj submit <contest>/<problem> <source_file>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "language, l",
					Value: "c++",
				},
			},
			Action: func(c *cli.Context) error {
				problems, err := LoadProblems()
				if err != nil {
					return err
				}
				lang, err := findLang(config.Languages, config.DefaultLanguage, c.String("l"))
				if err != nil {
					return err
				}
				var contest, problem, src string
				switch len(c.Args()) {
				case 2:
					split := strings.Split(c.Args().Get(0), "/")
					contest = split[0]
					problem = split[1]
					src = c.Args().Get(1)
				case 0:
					// 最後に編集されたファイルから提出する問題を推測する
					p, err := getProblem("", lang.Ext)
					if err != nil {
						return err
					}
					for _, pp := range problems.Problems {
						if pp.Name == p {
							contest = strings.TrimSuffix(pp.Name, "_"+strings.ToLower(pp.ID))
							problem = path.Base(pp.URL)
							src = p + lang.Ext
							break
						}
					}
					if contest == "" {
						return fmt.Errorf("cannot find problem: %s", p)
					}
				default:
					return errors.New("goj submit <contest>/<problem> <source_file>")
				}
				if err := SubmitAtCoder(client, contest, problem, src, lang.Name); err != nil {
					return err
				}
				fmt.Println("submit success:", contest, problem, src, lang.Name)
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
