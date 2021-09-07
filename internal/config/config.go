package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Data  RuleData  `yaml:"data"`
	Match RuleMatch `yaml:"match"`
}

type RuleData struct {
	Internal    bool   `yaml:"internal"`
	Destination string `yaml:"destination"`
	Source      string `yaml:"source"`
	Category    string `yaml:"category"`
	Description string `yaml:"description"`
}

type RuleMatch struct {
	Reciever string `yaml:"reciever,omitempty"`
	IBAN     string `yaml:"iban,omitempty"`
}

func GetConfig(path string) Config {
	var config Config
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		panic(err)
	}

	return config
}
