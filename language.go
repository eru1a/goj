package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Language struct {
	Name     string
	Ext      string
	BuildCmd string
	RunCmd   string
	Template string
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

var Languages = map[string]*Language{
	"cpp": {
		Name:     "cpp",
		Ext:      ".cpp",
		BuildCmd: "g++ -g -o [P] [P].cpp",
		RunCmd:   "./[P]",
		Template: `#include <bits/stdc++.h>

using namespace std;
using ll = long long;

int main() {
  cin.tie(nullptr);
  ios::sync_with_stdio(false);

  return 0;
}
`,
	},
	"python": {
		Name:   "python",
		Ext:    ".py",
		RunCmd: "python [P].py",
	},
}
