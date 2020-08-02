package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "download",
			Aliases: []string{"dl", "d"},
			Usage:   "download testcases",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "language, l",
					Value: "cpp",
				},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					return errors.New("goj download <contest> or <contest/problem>")
				}
				lang := Languages[c.String("l")]
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
					Value: "cpp",
				},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					return errors.New("goj test <problem> -c <command> -l <language>")
				}
				problem := c.Args().First()
				cmd := c.String("c")
				lang := Languages[c.String("l")]
				if cmd == "<none>" {
					if err := lang.Build(problem); err != nil {
						panic(err)
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
