package main

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/urfave/cli"
)

func TestParseTestCmdArgs(t *testing.T) {
	if err := os.Chdir("testdata/abc003"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir("../.."); err != nil {
			panic(err)
		}
	}()

	type result struct {
		problem string
		cmd     string
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
			args:   []string{"goj", "test"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				problem: "abc003_3",
				cmd:     "./abc003_3",
			},
		},
		{
			args:   []string{"goj", "test"},
			config: &Config{DefaultLanguage: "python", Languages: languages},
			want: result{
				problem: "abc003_2",
				cmd:     "python abc003_2.py",
			},
		},
		{
			args:   []string{"goj", "test", "-l", "python"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				problem: "abc003_2",
				cmd:     "python abc003_2.py",
			},
		},
		{
			args:   []string{"goj", "test", "1"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				problem: "abc003_1",
				cmd:     "./abc003_1",
			},
		},
		{
			args:   []string{"goj", "test", "abc003_3", "-c", "hogehoge"},
			config: &Config{DefaultLanguage: "c++", Languages: nil},
			want: result{
				problem: "abc003_3",
				cmd:     "hogehoge",
			},
		},
	}

	for _, test := range testsOK {
		app := cli.NewApp()
		app.Commands = []cli.Command{
			{
				Name: "test",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "command, c",
					},
					cli.StringFlag{
						Name: "language, l",
					},
				},
				Action: func(c *cli.Context) error {
					_, problem, cmd, err := ParseTestCmdArgs(c, test.config)
					if err != nil {
						return err
					}
					got := result{problem, cmd}
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
			args:   []string{"goj", "test"},
			config: &Config{DefaultLanguage: "c++", Languages: nil},
		},
		{
			args:   []string{"goj", "test"},
			config: &Config{DefaultLanguage: "go", Languages: languages},
		},
		{
			args:   []string{"goj", "test", "-l", "go"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
		{
			args:   []string{"goj", "test", "-c", "hogehoge"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
		{
			args:   []string{"goj", "test", "abc003_5", "hogehoge"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
	}

	for _, test := range testsNG {
		app := cli.NewApp()
		app.Commands = []cli.Command{
			{
				Name: "test",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "command, c",
					},
					cli.StringFlag{
						Name: "language, l",
					},
				},
				Action: func(c *cli.Context) error {
					_, _, _, err := ParseTestCmdArgs(c, test.config)
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
