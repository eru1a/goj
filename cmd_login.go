package main

import (
	"errors"
	"fmt"
	"syscall"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

func NewLoginCmd(atcoder *AtCoder, jar *cookiejar.Jar, config *Config) cli.Command {
	return cli.Command{
		Name:  "login",
		Usage: "goj login",
		Action: func(c *cli.Context) error {
			if len(c.Args()) > 0 {
				return errors.New("goj login")
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
			password := string(bytes)
			if err := atcoder.Login(username, password); err != nil {
				return err
			}
			if err := jar.Save(); err != nil {
				return err
			}

			fmt.Println("login success")
			return nil
		},
	}
}
