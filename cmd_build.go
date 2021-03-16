package main

import (
	"github.com/urfave/cli/v2"
)

func NewBuildCmd(config *Config) *cli.Command {
	return &cli.Command{
		Name:    "build",
		Aliases: []string{"b"},
		Usage:   "goj build [problem]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "language",
				Aliases: []string{"l"},
			},
		},
		Action: func(c *cli.Context) error {
			lang, problem, _, err := ParseTestCmdArgs(c, config)
			if err != nil {
				return err
			}

			if lang != nil {
				if err := lang.Build(problem.Name); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
