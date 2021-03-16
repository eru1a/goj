package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func NewLoginCmd(atcoder *AtCoder, jar *cookiejar.Jar, config *Config) *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "goj login",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "logout",
			},
			&cli.BoolFlag{
				Name: "check",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				return errors.New("goj login")
			}

			if c.Bool("logout") {
				cacheDir, _ := os.UserCacheDir()
				os.Remove(filepath.Join(cacheDir, "goj", "cookiejar"))
				return nil
			}

			if c.Bool("check") {
				// TODO: ログインしているアカウント名
				if err := atcoder.CheckLogin(); err != nil {
					if errors.Is(err, ErrNeedLogin) {
						LogWarning("you are not logged in")
						return nil
					}
					return err
				}
				LogSuccess("you are logged in")
				return nil
			}

			var username string
			fmt.Print("username: ")
			_, err := fmt.Scanln(&username)
			if err != nil {
				return err
			}
			fmt.Print("password: ")
			bytes, err := terminal.ReadPassword(syscall.Stdin)
			if err != nil {
				return err
			}
			fmt.Println()
			password := string(bytes)
			if err := atcoder.Login(username, password); err != nil {
				return err
			}
			if err := jar.Save(); err != nil {
				return err
			}
			return nil
		},
	}
}
