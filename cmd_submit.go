package main

import (
	"fmt"
	"math"

	"github.com/urfave/cli"
)

func ParseSubmitCmdArgs(c *cli.Context, config *Config) (lang *Language, problem *ProblemInfo, err error) {
	langName := c.String("l")
	if langName == "" {
		langName = config.DefaultLanguage
	}
	lang, err = FindLang(config.Languages, langName)
	if err != nil {
		return nil, nil, err
	}
	// 最後に編集されたファイルから提出する問題を決める
	problemName, err := FindProblemName(c.Args().First(), lang.Ext)
	if err != nil {
		return nil, nil, err
	}
	problem, err = FindProblem(problemName)
	if err != nil {
		return nil, nil, err
	}

	return lang, problem, nil
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
			lang, problem, err := ParseSubmitCmdArgs(c, config)
			if err != nil {
				return err
			}

			if !c.Bool("skip") {
				if err := lang.Build(problem.Name); err != nil {
					return err
				}
				floatTolerance := problem.FloatTolerance
				if c.Uint("f") != 0 {
					floatTolerance = math.Pow10(-int(c.Uint("f")))
				}
				result, err := Judge(problem.Name, lang.GetRunCmd(problem.Name),
					problem.TimeLimitSec*1000, problem.MemoryLimitMB, floatTolerance)
				if err != nil {
					return err
				}
				if !result.IsAC {
					LogFailure("interrupted the submission because test failed")
					return nil
				}
			}

			src := problem.Name + lang.Ext

			// 提出するか確認する
			fmt.Printf("submit? %s/%s %s %s [y/n]: ", problem.Contest, problem.Name, src, lang.Name)
			var yes string
			fmt.Scan(&yes)
			if yes != "y" {
				LogInfo("submission interrupted")
				return nil
			}

			if err := atcoder.Submit(problem.Contest, problem.Name, src, lang.Name); err != nil {
				return fmt.Errorf("%v: submit failed %s/%s, %s, %s", err, problem.Contest, problem.Name, src, lang.Name)
			}
			if err := atcoder.WatchLastSubmissionStatus(problem.Contest); err != nil {
				return err
			}
			return nil
		},
	}
}
