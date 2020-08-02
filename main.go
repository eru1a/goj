package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gookit/color"
	"github.com/urfave/cli"
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

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(filepath.Join(configDir, "goj", "config.toml"))
	if err != nil {
		if err := os.MkdirAll(filepath.Join(configDir, "goj"), 0755); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(filepath.Join(configDir, "goj", "config.toml"), []byte(defaultConfigToml), 0644); err != nil {
			panic(err)
		}
	}
	var config Config
	_, err = toml.DecodeFile(filepath.Join(configDir, "goj", "config.toml"), &config)
	if err != nil {
		panic(err)
	}

	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "download",
			Aliases: []string{"dl", "d"},
			Usage:   "download testcases",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "language, l",
					Value: "c++",
				},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					return errors.New("goj download <contest> or <contest/problem>")
				}
				lang, err := findLang(config.Languages, config.DefaultLanguage, c.String("l"))
				if err != nil {
					return err
				}

				split := strings.Split(c.Args().First(), "/")
				client := new(http.Client)
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
			Usage:   "test testcases",
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
				if len(c.Args()) != 1 {
					return errors.New("goj test <problem> -c <command> -l <language>")
				}
				lang, err := findLang(config.Languages, config.DefaultLanguage, c.String("l"))
				if err != nil {
					return err
				}

				problem := c.Args().First()
				cmd := c.String("c")
				if cmd == "<none>" {
					if err := lang.Build(problem); err != nil {
						return err
					}
					cmd = lang.GetRunCmd(problem)
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
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
