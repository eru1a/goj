package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/urfave/cli"
)

func findLang(languages []*Language, defaultLang, argLang string) (*Language, error) {
	langName := defaultLang
	if argLang != "" {
		langName = argLang
	}
	for _, l := range languages {
		if l.Name == langName {
			return l, nil
		}
	}
	return nil, fmt.Errorf("cannot find %s in languages", defaultLang)
}

func main() {
	log.SetFlags(0)

	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	gojCacheDir := filepath.Join(cacheDir, "goj")
	if err := os.MkdirAll(gojCacheDir, 0755); err != nil {
		panic(err)
	}
	cookieJarFile := filepath.Join(gojCacheDir, "cookiejar")
	jar, err := cookiejar.New(&cookiejar.Options{Filename: cookieJarFile})
	if err != nil {
		panic(err)
	}
	atcoder := NewAtCoder(jar)

	app := cli.NewApp()

	app.Commands = []cli.Command{
		NewDownloadCmd(atcoder, config),
		NewTestCmd(config),
		NewLoginCmd(atcoder, jar, config),
		NewSubmitCmd(atcoder, config),
		NewStatusCmd(atcoder),
	}

	if err := app.Run(os.Args); err != nil {
		LogFailure(err.Error())
		os.Exit(1)
	}
}
