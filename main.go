package main

import (
	"log"
	"os"
	"path/filepath"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/urfave/cli/v2"
)

func main() {
	log.SetFlags(0)

	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	gojCacheDir := filepath.Join(userCacheDir, "goj")
	if err := os.MkdirAll(gojCacheDir, 0755); err != nil {
		panic(err)
	}
	cookieJarFile := filepath.Join(gojCacheDir, "cookiejar")
	jar, err := cookiejar.New(&cookiejar.Options{Filename: cookieJarFile})
	if err != nil {
		panic(err)
	}
	atcoder := NewAtCoder(jar)

	// app := cli.NewApp()
	app := &cli.App{
		Name:  "goj",
		Usage: "AtCoder support tool.",
	}

	app.Commands = []*cli.Command{
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
