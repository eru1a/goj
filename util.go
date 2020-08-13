package main

import (
	"io/ioutil"
	"os"
)

func makeTemplateFile(problem string, lang *Language) error {
	if lang == nil {
		return nil
	}
	file := problem + lang.Ext
	// ファイルがある場合は何もしない
	if _, err := os.Stat(file); err != nil {
		if err := ioutil.WriteFile(file, []byte(lang.Template), 0666); err != nil {
			return err
		}
	}
	return nil
}
