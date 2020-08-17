package main

import (
	"errors"
	"math"

	"github.com/urfave/cli"
)

func ParseTestCmdArgs(c *cli.Context, config *Config) (lang *Language, problem string, cmd string, err error) {
	if len(c.Args()) > 1 {
		return nil, "", "", errors.New("goj test <problem>")
	}

	switch {
	case c.String("c") != "" && c.Args().First() == "":
		// -commandが与えられている場合は<problem>も必要
		return nil, "", "", errors.New("goj test <problem>")

	case c.String("c") == "":
		lang, err = findLang(config.Languages, config.DefaultLanguage, c.String("l"))
		if err != nil {
			return nil, "", "", errors.New("couldn't find lang, and command was not given")
		}

		problem, err = getProblem(c.Args().First(), lang.Ext)
		if err != nil {
			return nil, "", "", err
		}
		cmd = lang.GetRunCmd(problem)

	default:
		problem = c.Args().First()
		cmd = c.String("c")
	}

	return lang, problem, cmd, nil
}

func NewTestCmd(config *Config) cli.Command {
	return cli.Command{
		Name:    "test",
		Aliases: []string{"t"},
		Usage:   "goj test <problem>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "command, c",
			},
			cli.StringFlag{
				Name: "language, l",
			},
			cli.IntFlag{
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
				if err := lang.Build(problem); err != nil {
					return err
				}
			}

			floatTolerance := 0.0
			if c.Int("f") != 0 {
				floatTolerance = math.Pow10(-c.Int("f"))
			}
			if _, err := Judge(problem, cmd, floatTolerance); err != nil {
				return err
			}
			return nil
		},
	}
}
