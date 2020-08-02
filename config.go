package main

type Config struct {
	DefaultLanguage string      `toml:"default_language"`
	Languages       []*Language `toml:"language"`
}
