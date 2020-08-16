package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

func NewStatusCmd(atcoder *AtCoder) cli.Command {
	return cli.Command{
		Name:  "status",
		Usage: "goj status",
		Action: func(c *cli.Context) error {
			if len(c.Args()) > 1 {
				return errors.New("goj status <contest>")
			}
			contest := c.Args().First()
			if contest == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				contest = filepath.Base(cwd)
			}
			status, err := atcoder.SubmissionsStatus(contest)
			if err != nil {
				return err
			}
			for _, s := range status {
				fmt.Println(s.DrawString())
			}
			return nil
		},
	}
}
