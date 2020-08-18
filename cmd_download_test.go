package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/urfave/cli"
)

func TestParseDownloadCmdArgs(t *testing.T) {
	os.Stderr = nil
	log.SetOutput(ioutil.Discard)

	if err := os.Chdir("testdata/parse_args/download/abc002"); err != nil {
		panic(err)
	}
	defer func() {
		if err := os.Chdir("../../../.."); err != nil {
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
	}

	for _, test := range tests {
		app := cli.NewApp()
		app.Commands = []cli.Command{
			{
				Name:    "download",
				Aliases: []string{"dl", "d"},
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "language, l",
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
