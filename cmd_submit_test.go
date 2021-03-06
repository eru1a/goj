package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/urfave/cli/v2"
)

func TestParseSubmitCmdArgs(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	if err := os.Chdir("testdata/problem"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir("../.."); err != nil {
			panic(err)
		}
	}()

	type result struct {
		contest string
		problem string
	}

	languages := []*Language{
		{
			Name:     "c++",
			Ext:      ".cpp",
			BuildCmd: "g++ -o [P] [P].cpp",
			RunCmd:   "./[P]",
		},
		{
			Name:   "python",
			Ext:    ".py",
			RunCmd: "python [P].py",
		},
	}

	testsOK := []struct {
		args   []string
		config *Config
		want   result
	}{
		{
			args:   []string{"goj", "submit"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				contest: "abc173",
				problem: "abc173_a",
			},
		},
		{
			args:   []string{"goj", "submit"},
			config: &Config{DefaultLanguage: "python", Languages: languages},
			want: result{
				contest: "abc173",
				problem: "abc173_a",
			},
		},
		{
			args:   []string{"goj", "submit", "-l", "python"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				contest: "abc173",
				problem: "abc173_a",
			},
		},
		{
			args:   []string{"goj", "submit", "abc163_a"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				contest: "abc163",
				problem: "abc163_a",
			},
		},
		{
			args:   []string{"goj", "submit", "b"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				contest: "abc173",
				problem: "abc173_b",
			},
		},
		{
			args:   []string{"goj", "submit", "1"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				contest: "abc001",
				problem: "abc001_1",
			},
		},
	}

	for _, test := range testsOK {
		app := cli.NewApp()
		app.Commands = []*cli.Command{
			{
				Name: "submit",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "skip tests",
					},
				},
				Action: func(c *cli.Context) error {
					_, problem, err := ParseSubmitCmdArgs(c, test.config)
					if err != nil {
						return err
					}
					got := result{problem.Contest, problem.Name}
					if !reflect.DeepEqual(got, test.want) {
						return errors.New(pretty.Compare(test.want, got))
					}
					return nil
				},
			},
		}
		if err := app.Run(test.args); err != nil {
			t.Error(err)
		}
	}

	testsNG := []struct {
		args   []string
		config *Config
	}{
		{
			args:   []string{"goj", "submit", "abc173_c"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
		{
			args:   []string{"goj", "submit", "5"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
		{
			args:   []string{"goj", "submit", "1"},
			config: &Config{DefaultLanguage: "python", Languages: languages},
		},
		{
			args:   []string{"goj", "submit", "1", "-l", "python"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
		{
			args:   []string{"goj", "submit"},
			config: &Config{DefaultLanguage: "go", Languages: languages},
		},
		{
			args:   []string{"goj", "submit", "-l", "go"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
	}

	for _, test := range testsNG {
		app := cli.NewApp()
		app.Commands = []*cli.Command{
			{
				Name: "submit",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "skip tests",
					},
				},
				Action: func(c *cli.Context) error {
					_, _, err := ParseSubmitCmdArgs(c, test.config)
					if err == nil {
						return errors.New("should error: " + strings.Join(test.args, " "))
					}
					return nil
				},
			},
		}
		if err := app.Run(test.args); err != nil {
			t.Error(err)
		}
	}
}
