package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/urfave/cli"
)

func findLang(languages []*Language, langName, argLang string) (*Language, error) {
	if argLang != "" {
		langName = argLang
	}
	for _, l := range languages {
		if l.Name == langName {
			return l, nil
		}
	}
	return nil, fmt.Errorf("cannot find %s in languages", langName)
}

// 拡張子がextでファイルの名前がsuffixで終わる最も最近編集されたファイルを返す。
func getProblem(suffix string, ext string) (string, error) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return "", err
	}
	var files2 []os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ext) {
			continue
		}
		fileNameWithoutExt := strings.TrimSuffix(file.Name(), ext)
		switch {
		case suffix == "":
			files2 = append(files2, file)
		case fileNameWithoutExt == suffix:
			files2 = append(files2, file)
		case strings.HasSuffix(fileNameWithoutExt, suffix):
			files2 = append(files2, file)
		}
	}
	if len(files2) == 0 {
		return "", fmt.Errorf("cannot find a file. suffix: %s, ext: %s", suffix, ext)
	}
	sort.SliceStable(files2, func(i, j int) bool {
		return files2[i].ModTime().Unix() > files2[j].ModTime().Unix()
	})
	return strings.TrimSuffix(files2[0].Name(), ext), nil
}

func main() {
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
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
