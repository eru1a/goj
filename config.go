package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DefaultLanguage string      `toml:"default_language"`
	Languages       []*Language `toml:"language"`
}

func LoadConfig() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(filepath.Join(configDir, "goj", "config.toml"))
	if err != nil {
		if err := os.MkdirAll(filepath.Join(configDir, "goj"), 0755); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(filepath.Join(configDir, "goj", "config.toml"), []byte(DefaultConfigToml), 0666); err != nil {
			return nil, err
		}
	}
	var config Config
	_, err = toml.DecodeFile(filepath.Join(configDir, "goj", "config.toml"), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

const DefaultConfigToml = `default_language = "c++"

[[language]]
name = "c++"
ext = ".cpp"
# [P] : Problem Name
build_cmd = "g++ -fdiagnostics-color=always -g -o [P] [P].cpp"
run_cmd = "./[P]"
template = """#include <bits/stdc++.h>

using namespace std;
using ll = long long;

int main() {
  cin.tie(nullptr);
  ios::sync_with_stdio(false);

  return 0;
}
"""

[[language]]
name = "python"
ext = ".py"
run_cmd = "python [P].py"
`
