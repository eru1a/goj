package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/urfave/cli/v2"
)

func TestParseTestCmdArgs(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	tmpDir := filepath.Join(os.TempDir(), "goj", "testing")
	curDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	createTempFiles(tmpDir)
	if err := os.Chdir(tmpDir); err != nil {
		panic(err)
	}
	defer func() {
		if os.Chdir(curDir); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
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
				problem: "abc173_a",
				cmd:     "./abc173_a",
			},
		},
		{
			args:   []string{"goj", "test"},
			config: &Config{DefaultLanguage: "python", Languages: languages},
			want: result{
				problem: "abc173_a",
				cmd:     "python abc173_a.py",
			},
		},
		{
			args:   []string{"goj", "test", "-l", "python"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				problem: "abc173_a",
				cmd:     "python abc173_a.py",
			},
		},
		{
			args:   []string{"goj", "test", "1"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
			want: result{
				problem: "abc001_1",
				cmd:     "./abc001_1",
			},
		},
		{
			args:   []string{"goj", "test", "-c", "hogehoge", "abc173_a"},
			config: &Config{DefaultLanguage: "c++", Languages: nil},
			want: result{
				problem: "abc173_a",
				cmd:     "hogehoge",
			},
		},
	}

	for _, test := range testsOK {
		app := cli.NewApp()
		app.Commands = []*cli.Command{
			{
				Name: "test",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "command",
						Aliases: []string{"c"},
					},
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
					},
				},
				Action: func(c *cli.Context) error {
					_, problem, cmd, err := ParseTestCmdArgs(c, test.config)
					if err != nil {
						return err
					}
					got := result{problem.Name, cmd}
					if !reflect.DeepEqual(got, test.want) {
						return errors.New(pretty.Compare(test.want, got))
					}
					return nil
				},
			},
		}
		if err := app.Run(test.args); err != nil {
			t.Error(test.args, err)
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
			args:   []string{"goj", "test", "abc173_c", "hogehoge"},
			config: &Config{DefaultLanguage: "c++", Languages: languages},
		},
	}

	for _, test := range testsNG {
		app := cli.NewApp()
		app.Commands = []*cli.Command{
			{
				Name: "test",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "command",
						Aliases: []string{"c"},
					},
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
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
