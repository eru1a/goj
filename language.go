package main

import (
	"fmt"
	"os"
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

func (l *Language) Build(problem string) error {
	if l.BuildCmd == "" {
		return nil
	}
	buildCmd := strings.ReplaceAll(l.BuildCmd, "[P]", problem)
	LogInfo(buildCmd)
	c := strings.Split(buildCmd, " ")
	cmd := exec.Command(c[0], c[1:]...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (l *Language) GetRunCmd(problem string) string {
	return strings.ReplaceAll(l.RunCmd, "[P]", problem)
}

func FindLang(languages []*Language, name string) (*Language, error) {
	for _, lang := range languages {
		if lang.Name == name {
			return lang, nil
		}
	}
	return nil, fmt.Errorf("cannot find %s in languages", name)
}
