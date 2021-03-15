package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/urfave/cli/v2"
)

func TestParseDownloadCmdArgs(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	tmpDir := filepath.Join(os.TempDir(), "goj", "abc002")
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
		contest string
		problem string
	}

	config := &Config{
		DefaultLanguage: "c++",
		Languages:       []*Language{{Name: "c++"}},
	}

	tests := []struct {
		args   []string
		config *Config
		want   result
	}{
		{
			args:   []string{"goj", "download"},
			config: config,
			want: result{
				contest: "abc002",
				problem: "",
			},
		},
		{
			args:   []string{"goj", "download", "abc173"},
			config: config,
			want: result{
				contest: "abc173",
				problem: "",
			},
		},
		{
			args:   []string{"goj", "download", "abc173/abc173_c"},
			config: config,
			want: result{
				contest: "abc173",
				problem: "abc173_c",
			},
		},
		{
			args:   []string{"goj", "download", "https://atcoder.jp/contests/abc173/tasks/abc173_c"},
			config: config,
			want: result{
				contest: "abc173",
				problem: "abc173_c",
			},
		},
		{
			args:   []string{"goj", "download", "https://atcoder.jp/contests/abc173/"},
			config: config,
			want: result{
				contest: "abc173",
				problem: "",
			},
		},
		{
			args:   []string{"goj", "download", "https://atcoder.jp/contests/abc173/tasks"},
			config: config,
			want: result{
				contest: "abc173",
				problem: "",
			},
		},
	}

	for _, test := range tests {
		app := cli.NewApp()
		app.Commands = []*cli.Command{
			{
				Name:    "download",
				Aliases: []string{"dl", "d"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
					},
				},
				Action: func(c *cli.Context) error {
					_, contest, problem, err := ParseDownloadCmdArgs(c, test.config)
					if err != nil {
						return err
					}
					got := result{contest, problem}
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
}
