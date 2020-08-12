package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
)

func CmpOutput(actual, expected string) bool {
	a := strings.Split(actual, "\n")
	e := strings.Split(expected, "\n")

	if len(a) != len(e) {
		return false
	}

	for i := range a {
		if strings.TrimSuffix(a[i], " ") != e[i] {
			return false
		}
	}

	return true
}

func Judge(problem string, command string) (ac, wa, re int, err error) {
	fmt.Printf("test %s (%s)\n", problem, command)
	dir := fmt.Sprintf("test_%s", problem)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return ac, wa, re, err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if filepath.Ext(f.Name()) != ".in" {
			continue
		}

		testName := strings.TrimSuffix(f.Name(), ".in")
		fmt.Print(fmt.Sprintf("%s ... ", testName))

		in, err := ioutil.ReadFile(filepath.Join(dir, testName+".in"))
		if err != nil {
			return ac, wa, re, err
		}

		out, err := ioutil.ReadFile(filepath.Join(dir, testName+".out"))
		if err != nil {
			return ac, wa, re, err
		}

		c := strings.Split(command, " ")
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stderr = os.Stderr

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return ac, wa, re, err
		}

		_, err = io.WriteString(stdin, string(in))
		if err != nil {
			return ac, wa, re, err
		}

		stdout, err := cmd.Output()
		if err != nil {
			re++
			color.Red.Println("RE")
			fmt.Printf("  %s\n", err)
			continue
		}

		if CmpOutput(string(stdout), string(out)) {
			ac++
			color.Green.Println("AC")
		} else {
			wa++
			color.Red.Println("WA")
			color.Bold.Println("\nexpected:")
			fmt.Println(string(out))
			color.Bold.Println("got:")
			fmt.Println(string(stdout))
		}
	}
	return ac, wa, re, nil
}
