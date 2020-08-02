package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Language struct {
	Name     string `toml:"name"`
	Ext      string `toml:"ext"`
	BuildCmd string `toml:"build_cmd"`
	RunCmd   string `toml:"run_cmd"`
	Template string `toml:"template"`
}

func runCmd(problem string, cmd string) error {
	c := strings.ReplaceAll(cmd, "[P]", problem)
	fmt.Println(c)
	if err := exec.Command(c).Run(); err != nil {
		return err
	}
	return nil
}

func (l *Language) Build(problem string) error {
	if l.BuildCmd == "" {
		return nil
	}
	buildCmd := strings.ReplaceAll(l.BuildCmd, "[P]", problem)
	fmt.Println(buildCmd)
	cmd := strings.Split(buildCmd, " ")
	if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
		return err
	}
	return nil
}

func (l *Language) GetRunCmd(problem string) string {
	return strings.ReplaceAll(l.RunCmd, "[P]", problem)
}
