package main

import (
	"errors"
	"math"

	"github.com/urfave/cli/v2"
)

func ParseTestCmdArgs(c *cli.Context, config *Config) (lang *Language, problem *ProblemInfo, cmd string, err error) {
	if c.Args().Len() > 1 {
		return nil, nil, "", errors.New("goj test [problem]")
	}

	var problemName string

	switch {
	case c.String("c") != "" && c.Args().First() == "":
		// -commandが与えられている場合は<problem>も必要
		return nil, nil, "", errors.New("goj test [problem]")

	case c.String("c") != "":
		problemName = c.Args().First()
		cmd = c.String("c")

	default:
		langName := c.String("l")
		if langName == "" {
			langName = config.DefaultLanguage
		}
		lang, err = FindLang(config.Languages, langName)
		if err != nil {
			return nil, nil, "", errors.New("couldn't find lang, and command was not given")
		}

		problemName, err = FindProblemName(c.Args().First(), lang.Ext)
		if err != nil {
			return nil, nil, "", err
		}
		cmd = lang.GetRunCmd(problemName)
	}

	problem, err = FindProblem(problemName)
	if err != nil {
		return nil, nil, "", err
	}

	return lang, problem, cmd, nil
}

func NewTestCmd(config *Config) *cli.Command {
	return &cli.Command{
		Name:    "test",
		Aliases: []string{"t"},
		Usage:   "goj test [problem]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "command",
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:    "language",
				Aliases: []string{"l"},
			},
			&cli.UintFlag{
				Name:  "f",
				Usage: "float tolerance",
			},
		},
		Action: func(c *cli.Context) error {
			lang, problem, cmd, err := ParseTestCmdArgs(c, config)
			if err != nil {
				return err
			}

			if lang != nil {
				if err := lang.Build(problem.Name); err != nil {
					return err
				}
			}

			floatTolerance := problem.FloatTolerance
			if c.Int("f") != 0 {
				floatTolerance = math.Pow10(-int(c.Uint("f")))
			}
			if _, err := Judge(problem.Name, cmd, problem.TimeLimitSec*1000, problem.MemoryLimitMB, floatTolerance); err != nil {
				return err
			}
			return nil
		},
	}
}
