package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

func ParseDownloadCmdArgs(c *cli.Context, config *Config) (lang *Language, contest string, problem string, err error) {
	if len(c.Args()) > 1 {
		return nil, "", "", errors.New("goj download <contest> or <contest/problem>")
	}
	langName := c.String("l")
	if langName == "" {
		langName = config.DefaultLanguage
	}
	lang, err = FindLang(config.Languages, langName)
	if err != nil {
		return nil, "", "", err
	}

	first := c.Args().First()
	split := strings.Split(first, "/")
	switch {
	case first == "":
		// コマンド引数としてコンテストが与えられなかった場合はカレントディレクトリの名前をコンテストと見なす
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", "", err
		}
		return lang, filepath.Base(cwd), "", nil
	case len(split) == 1:
		// <contest>
		return lang, first, "", nil
	case len(split) == 2:
		// <contest/problem>
		return lang, split[0], split[1], nil
	default:
		return nil, "", "", errors.New("goj download <contest> or <contest/problem>")
	}
}

func NewDownloadCmd(atcoder *AtCoder, config *Config) cli.Command {
	return cli.Command{
		Name:    "download",
		Aliases: []string{"dl", "d"},
		Usage:   "goj download <contest> or <contest/problem>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "language, l",
			},
		},
		Action: func(c *cli.Context) error {
			lang, contest, problem, err := ParseDownloadCmdArgs(c, config)
			if err != nil {
				return err
			}
			switch {
			case problem != "":
				if err := atcoder.DownloadProblem(contest, problem, lang); err != nil {
					return err
				}
			default:
				if err := atcoder.DownloadContest(contest, lang); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
