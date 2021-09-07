package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	URL   string `yaml:"url"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Data  RuleData  `json:"data"`
	Match RuleMatch `json:"match"`
}

type RuleData struct {
	Internal    bool   `json:"internal"`
	Destination string `json:"destination"`
	Source      string `json:"source"`
}

type RuleMatch struct {
	Reciever string `json:"reciever,omitempty"`
	IBAN     string `json:"iban,omitempty"`
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
