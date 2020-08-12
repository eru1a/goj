package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/gookit/color"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

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

// 拡張子がextでファイルの名前がsuffixで終わる最も最近編集されたファイルを返す。
func getProblem(suffix string, ext string) (string, error) {
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
		case suffix == "":
			files2 = append(files2, file)
		case fileNameWithoutExt == suffix:
			files2 = append(files2, file)
		case strings.HasSuffix(fileNameWithoutExt, suffix):
			files2 = append(files2, file)
		}
	}
	if len(files2) == 0 {
		return "", fmt.Errorf("cannot find a file. suffix: %s, ext: %s", suffix, ext)
	}
	sort.SliceStable(files2, func(i, j int) bool {
		return files2[i].ModTime().Unix() > files2[j].ModTime().Unix()
	})
	return strings.TrimSuffix(files2[0].Name(), ext), nil
}

func judge(problem string, cmd string) bool {
	ac, wa, re, err := Judge(problem, cmd)
	if err != nil {
		panic(err)
	}
	result := color.Green.Sprint("AC")
	if re > 0 {
		result = color.Red.Sprint("RE")
	} else if wa > 0 {
		result = color.Red.Sprint("WA")
	}
	fmt.Printf("%s (AC:%d WA:%d RE:%d)\n", result, ac, wa, re)
	if re == 0 && wa == 0 {
		return true
	}
	return false
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
	atcoder := NewAtCoder(jar)

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
					if err := atcoder.DownloadContest(split[0], lang); err != nil {
						return err
					}
				case 2:
					if err := atcoder.DownloadProblem(split[0], split[1], lang); err != nil {
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
					Name: "command, c",
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

				// -commandが渡されている時は拡張子は考慮しない
				ext := lang.Ext
				if c.String("c") != "" {
					ext = ""
				}
				problem, err := getProblem(c.Args().First(), ext)
				if err != nil {
					return err
				}
				cmd := c.String("c")
				if cmd == "" {
					cmd = lang.GetRunCmd(problem)
					if err := lang.Build(problem); err != nil {
						return err
					}
				}
				judge(problem, cmd)
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
				if err := atcoder.Login(username, password); err != nil {
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
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "skip tests",
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
					if len(split) != 2 {
						return errors.New("goj submit <contest>/<problem> <source_file>")
					}
					contest = split[0]
					problem = split[1]
					src = c.Args().Get(1)
				case 0:
					// 最後に編集されたファイルから提出する問題を決める
					problem, err = getProblem("", lang.Ext)
					if err != nil {
						return err
					}
					for _, pp := range problems.Problems {
						if pp.Name == problem {
							contest = pp.Contest
							src = problem + lang.Ext
							break
						}
					}
					if contest == "" {
						return fmt.Errorf("cannot find problem: %s", problem)
					}
				default:
					return errors.New("goj submit <contest>/<problem> <source_file>")
				}

				if !c.Bool("f") {
					if err := lang.Build(problem); err != nil {
						return err
					}
					if ac := judge(problem, lang.GetRunCmd(problem)); !ac {
						fmt.Println("interrupted the submission because test failed")
						return nil
					}
				}

				if err := atcoder.Submit(contest, problem, src, lang.Name); err != nil {
					return fmt.Errorf("%v: submit failed (%s, %s, %s, %s)", err, contest, problem, src, lang.Name)
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
