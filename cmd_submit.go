package main

import (
	"fmt"
	"math"

	"github.com/urfave/cli"
)

func ParseSubmitCmdArgs(c *cli.Context, config *Config) (lang *Language, contest, problem string, err error) {
	problems, err := LoadProblems()
	if err != nil {
		return nil, "", "", err
	}
	lang, err = findLang(config.Languages, config.DefaultLanguage, c.String("l"))
	if err != nil {
		return nil, "", "", err
	}
	// 最後に編集されたファイルから提出する問題を決める
	problem, err = getProblem(c.Args().First(), lang.Ext)
	if err != nil {
		return nil, "", "", err
	}
	for _, p := range problems.Problems {
		if p.Name == problem {
			contest = p.Contest
			break
		}
	}
	if contest == "" {
		return nil, "", "", fmt.Errorf("cannot find problem: %s", problem)
	}

	return lang, contest, problem, nil
}

func NewSubmitCmd(atcoder *AtCoder, config *Config) cli.Command {
	return cli.Command{
		Name:  "submit",
		Usage: "goj submit <problem>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "language, l",
			},
			cli.BoolFlag{
				Name:  "skip",
				Usage: "skip tests",
			},
			cli.UintFlag{
				Name:  "f",
				Usage: "float tolerance (10^(-f))",
			},
		},
		Action: func(c *cli.Context) error {
			lang, contest, problem, err := ParseSubmitCmdArgs(c, config)
			if err != nil {
				return err
			}

			if !c.Bool("f") {
				if err := lang.Build(problem); err != nil {
					return err
				}
				floatTolerance := 0.0
				if c.Uint("f") != 0 {
					floatTolerance = math.Pow10(-int(c.Uint("f")))
				}
				result, err := Judge(problem, lang.GetRunCmd(problem), 2*1000, 1024, floatTolerance)
				if err != nil {
					return err
				}
				if !result.IsAC {
					LogFailure("interrupted the submission because test failed")
					return nil
				}
			}

			src := problem + lang.Ext
			if err := atcoder.Submit(contest, problem, src, lang.Name); err != nil {
				return fmt.Errorf("%v: submit failed (%s, %s, %s, %s)", err, contest, problem, src, lang.Name)
			}
			if err := atcoder.WatchLastSubmissionStatus(contest); err != nil {
				return err
			}
			return nil
		},
	}
}
